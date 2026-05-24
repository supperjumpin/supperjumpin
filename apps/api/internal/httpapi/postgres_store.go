package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MeResponse{}, err
	}
	defer tx.Rollback()

	profile, err := getProfileByAuthIdentity(ctx, tx, identity.Provider, identity.Subject)
	if err == nil {
		return profile, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return MeResponse{}, err
	}

	key := identity.Provider + ":" + identity.Subject
	account := Account{ID: stableID("account", key), Email: identity.Email}
	player := Player{ID: stableID("player", account.ID), DisplayName: displayName(identity.Email)}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO accounts (id, email)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email`, account.ID, account.Email); err != nil {
		return MeResponse{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO auth_identities (provider, subject, account_id)
VALUES ($1, $2, $3)
ON CONFLICT (provider, subject) DO NOTHING`, identity.Provider, identity.Subject, account.ID); err != nil {
		return MeResponse{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO players (id, account_id, display_name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO NOTHING`, player.ID, account.ID, player.DisplayName); err != nil {
		return MeResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return MeResponse{}, err
	}
	return MeResponse{Account: account, Player: player}, nil
}

func (s *PostgresStore) CreateGroup(ctx context.Context, player Player, name string) (GroupHomeResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return GroupHomeResponse{}, err
	}
	defer tx.Rollback()

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM groups`).Scan(&count); err != nil {
		return GroupHomeResponse{}, err
	}
	group := Group{ID: stableID("group", player.ID+":"+name+":"+strconv.Itoa(count+1)), Name: name}
	membership := GroupMembership{GroupID: group.ID, PlayerID: player.ID, Role: "Group Admin"}
	if _, err := tx.ExecContext(ctx, `INSERT INTO groups (id, name) VALUES ($1, $2)`, group.ID, group.Name); err != nil {
		return GroupHomeResponse{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO group_memberships (group_id, player_id, role)
VALUES ($1, $2, $3)`, membership.GroupID, membership.PlayerID, membership.Role); err != nil {
		return GroupHomeResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return GroupHomeResponse{}, err
	}
	return groupHome(group, membership, nil), nil
}

func (s *PostgresStore) GroupHome(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error) {
	var group Group
	var membership GroupMembership
	if err := s.db.QueryRowContext(ctx, `
SELECT groups.id, groups.name, group_memberships.player_id, group_memberships.role
FROM group_memberships
JOIN groups ON groups.id = group_memberships.group_id
WHERE group_memberships.group_id = $1 AND group_memberships.player_id = $2`, groupID, player.ID).Scan(
		&group.ID,
		&group.Name,
		&membership.PlayerID,
		&membership.Role,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GroupHomeResponse{}, false, nil
		}
		return GroupHomeResponse{}, false, err
	}
	membership.GroupID = group.ID
	season, err := s.activeSeasonForGroup(ctx, group.ID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	return groupHome(group, membership, season), true, nil
}

func (s *PostgresStore) ListGroups(ctx context.Context, player Player) (ListGroupsResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT groups.id, groups.name, group_memberships.player_id, group_memberships.role
FROM group_memberships
JOIN groups ON groups.id = group_memberships.group_id
WHERE group_memberships.player_id = $1
ORDER BY groups.name`, player.ID)
	if err != nil {
		return ListGroupsResponse{}, err
	}
	defer rows.Close()

	memberships := []GroupMembershipSummary{}
	for rows.Next() {
		var group Group
		var membership GroupMembership
		if err := rows.Scan(&group.ID, &group.Name, &membership.PlayerID, &membership.Role); err != nil {
			return ListGroupsResponse{}, err
		}
		membership.GroupID = group.ID
		memberships = append(memberships, GroupMembershipSummary{Group: group, Membership: membership})
	}
	if err := rows.Err(); err != nil {
		return ListGroupsResponse{}, err
	}
	sort.Slice(memberships, func(i, j int) bool {
		return memberships[i].Group.Name < memberships[j].Group.Name
	})
	return ListGroupsResponse{Memberships: memberships}, nil
}

func (s *PostgresStore) CreateInvite(ctx context.Context, player Player, groupID string) (Invite, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Invite{}, false, err
	}
	defer tx.Rollback()

	var existing string
	if err := tx.QueryRowContext(ctx, `
SELECT player_id
FROM group_memberships
WHERE group_id = $1 AND player_id = $2`, groupID, player.ID).Scan(&existing); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Invite{}, false, nil
		}
		return Invite{}, false, err
	}

	var invite Invite
	for attempts := 0; attempts < 3; attempts++ {
		id, err := randomToken("invite")
		if err != nil {
			return Invite{}, false, err
		}
		token, err := randomToken("invite_token")
		if err != nil {
			return Invite{}, false, err
		}
		invite = Invite{
			ID:        id,
			GroupID:   groupID,
			Token:     token,
			CreatedBy: player.ID,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour).UTC(),
		}
		result, err := tx.ExecContext(ctx, `
INSERT INTO invites (id, group_id, token, created_by_player_id, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO NOTHING`, invite.ID, invite.GroupID, invite.Token, invite.CreatedBy, invite.ExpiresAt)
		if err != nil {
			return Invite{}, false, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return Invite{}, false, err
		}
		if rows == 1 {
			break
		}
		if attempts == 2 {
			return Invite{}, false, fmt.Errorf("create unique Invite after retries")
		}
	}
	if err := tx.Commit(); err != nil {
		return Invite{}, false, err
	}
	return invite, true, nil
}

func (s *PostgresStore) AcceptInvite(ctx context.Context, player Player, token string) (GroupHomeResponse, InviteAcceptStatus, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	defer tx.Rollback()

	var invite Invite
	var usedBy sql.NullString
	if err := tx.QueryRowContext(ctx, `
SELECT id, group_id, token, created_by_player_id, used_by_player_id, expires_at
FROM invites
WHERE token = $1`, token).Scan(&invite.ID, &invite.GroupID, &invite.Token, &invite.CreatedBy, &usedBy, &invite.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GroupHomeResponse{}, InviteInvalid, nil
		}
		return GroupHomeResponse{}, InviteInvalid, err
	}
	if usedBy.Valid {
		return GroupHomeResponse{}, InviteUsed, nil
	}
	if time.Now().After(invite.ExpiresAt) {
		return GroupHomeResponse{}, InviteExpired, nil
	}
	var existingRole string
	if err := tx.QueryRowContext(ctx, `
SELECT role
FROM group_memberships
WHERE group_id = $1 AND player_id = $2`, invite.GroupID, player.ID).Scan(&existingRole); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return GroupHomeResponse{}, InviteInvalid, err
	} else if err == nil {
		return GroupHomeResponse{}, InviteMember, nil
	}

	var group Group
	if err := tx.QueryRowContext(ctx, `SELECT id, name FROM groups WHERE id = $1`, invite.GroupID).Scan(&group.ID, &group.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GroupHomeResponse{}, InviteInvalid, nil
		}
		return GroupHomeResponse{}, InviteInvalid, err
	}
	if err := tx.QueryRowContext(ctx, `
UPDATE invites
SET used_by_player_id = $1
WHERE token = $2 AND used_by_player_id IS NULL AND expires_at > now()
RETURNING id`, player.ID, token).Scan(&invite.ID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return GroupHomeResponse{}, InviteInvalid, err
		}
		if status, err := s.inviteStatus(ctx, tx, token); err != nil || status != InviteInvalid {
			return GroupHomeResponse{}, status, err
		}
		return GroupHomeResponse{}, InviteInvalid, nil
	}
	membership := GroupMembership{GroupID: invite.GroupID, PlayerID: player.ID, Role: "Player"}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO group_memberships (group_id, player_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (group_id, player_id) DO NOTHING`, membership.GroupID, membership.PlayerID, membership.Role); err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	if err := tx.QueryRowContext(ctx, `
SELECT role
FROM group_memberships
WHERE group_id = $1 AND player_id = $2`, invite.GroupID, player.ID).Scan(&membership.Role); err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	season, err := s.activeSeasonForGroup(ctx, group.ID)
	if err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	if err := tx.Commit(); err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	return groupHome(group, membership, season), InviteAccepted, nil
}

func (s *PostgresStore) StartSeason(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	defer tx.Rollback()

	var group Group
	var membership GroupMembership
	if err := tx.QueryRowContext(ctx, `
SELECT groups.id, groups.name, group_memberships.player_id, group_memberships.role
FROM group_memberships
JOIN groups ON groups.id = group_memberships.group_id
WHERE group_memberships.group_id = $1 AND group_memberships.player_id = $2`, groupID, player.ID).Scan(
		&group.ID,
		&group.Name,
		&membership.PlayerID,
		&membership.Role,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GroupHomeResponse{}, false, nil
		}
		return GroupHomeResponse{}, false, err
	}
	membership.GroupID = group.ID
	var openCount int
	if err := tx.QueryRowContext(ctx, `
SELECT count(*)
FROM seasons
WHERE group_id = $1 AND status IN ('Active', 'Judging Grace Period')`, groupID).Scan(&openCount); err != nil {
		return GroupHomeResponse{}, false, err
	}
	if openCount > 0 {
		return GroupHomeResponse{}, true, ErrSeasonAlreadyOpen
	}

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM seasons`).Scan(&count); err != nil {
		return GroupHomeResponse{}, false, err
	}
	season := Season{
		ID:                   stableID("season", groupID+":"+strconv.Itoa(count+1)),
		GroupID:              groupID,
		CommissionerPlayerID: player.ID,
		Status:               "Active",
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO seasons (id, group_id, commissioner_player_id, status)
VALUES ($1, $2, $3, $4)`, season.ID, season.GroupID, season.CommissionerPlayerID, season.Status); err != nil {
		if isSeasonOpenConflict(err) {
			return GroupHomeResponse{}, true, ErrSeasonAlreadyOpen
		}
		return GroupHomeResponse{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return GroupHomeResponse{}, false, err
	}
	return groupHome(group, membership, &season), true, nil
}

func (s *PostgresStore) CreateIdea(ctx context.Context, player Player, groupID string, source string, destination string, food string) (Stunt, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Stunt{}, false, err
	}
	defer tx.Rollback()

	if _, ok, err := groupMembershipInTx(ctx, tx, player, groupID); err != nil {
		return Stunt{}, false, err
	} else if !ok {
		return Stunt{}, false, nil
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM stunts`).Scan(&count); err != nil {
		return Stunt{}, false, err
	}
	stunt := Stunt{
		ID:          stableID("stunt", groupID+":"+player.ID+":"+strconv.Itoa(count+1)),
		GroupID:     groupID,
		PlayerID:    player.ID,
		Status:      "Idea",
		Source:      source,
		Destination: destination,
		Food:        food,
		OffSeason:   true,
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO stunts (id, group_id, player_id, status, source, destination, food)
VALUES ($1, $2, $3, $4, $5, $6, $7)`, stunt.ID, stunt.GroupID, stunt.PlayerID, stunt.Status, stunt.Source, stunt.Destination, stunt.Food); err != nil {
		return Stunt{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Stunt{}, false, err
	}
	return stunt, true, nil
}

func (s *PostgresStore) CreatePlannedStunt(ctx context.Context, player Player, ideaID string, offSeason bool) (Stunt, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Stunt{}, false, err
	}
	defer tx.Rollback()

	stunt, err := stuntInTx(ctx, tx, ideaID)
	if errors.Is(err, sql.ErrNoRows) {
		return Stunt{}, false, ErrStuntNotFound
	}
	if err != nil {
		return Stunt{}, false, err
	}
	if stunt.Status != "Idea" {
		return Stunt{}, false, ErrStuntNotFound
	}
	if _, ok, err := groupMembershipInTx(ctx, tx, player, stunt.GroupID); err != nil {
		return Stunt{}, false, err
	} else if !ok || stunt.PlayerID != player.ID {
		return Stunt{}, false, nil
	}
	stunt.Status = "Planned Stunt"
	stunt.SeasonID = nil
	stunt.OffSeason = true
	if !offSeason {
		season, err := activeSeasonForGroupInTx(ctx, tx, stunt.GroupID)
		if err != nil {
			return Stunt{}, false, err
		}
		if season != nil {
			stunt.SeasonID = &season.ID
			stunt.OffSeason = false
		}
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE stunts
SET status = $1, season_id = $2
WHERE id = $3`, stunt.Status, stunt.SeasonID, stunt.ID); err != nil {
		return Stunt{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Stunt{}, false, err
	}
	return stunt, true, nil
}

func (s *PostgresStore) activeSeasonForGroup(ctx context.Context, groupID string) (*Season, error) {
	var season Season
	if err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status
FROM seasons
WHERE group_id = $1 AND status = 'Active'
ORDER BY created_at DESC
LIMIT 1`, groupID).Scan(
		&season.ID,
		&season.GroupID,
		&season.CommissionerPlayerID,
		&season.Status,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &season, nil
}

func groupMembershipInTx(ctx context.Context, tx *sql.Tx, player Player, groupID string) (GroupMembership, bool, error) {
	var membership GroupMembership
	if err := tx.QueryRowContext(ctx, `
SELECT group_id, player_id, role
FROM group_memberships
WHERE group_id = $1 AND player_id = $2`, groupID, player.ID).Scan(&membership.GroupID, &membership.PlayerID, &membership.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GroupMembership{}, false, nil
		}
		return GroupMembership{}, false, err
	}
	return membership, true, nil
}

func stuntInTx(ctx context.Context, tx *sql.Tx, stuntID string) (Stunt, error) {
	var stunt Stunt
	var seasonID sql.NullString
	if err := tx.QueryRowContext(ctx, `
SELECT id, group_id, player_id, season_id, status, source, destination, food
FROM stunts
WHERE id = $1`, stuntID).Scan(
		&stunt.ID,
		&stunt.GroupID,
		&stunt.PlayerID,
		&seasonID,
		&stunt.Status,
		&stunt.Source,
		&stunt.Destination,
		&stunt.Food,
	); err != nil {
		return Stunt{}, err
	}
	if seasonID.Valid {
		stunt.SeasonID = &seasonID.String
		stunt.OffSeason = false
	} else {
		stunt.OffSeason = true
	}
	return stunt, nil
}

func activeSeasonForGroupInTx(ctx context.Context, tx *sql.Tx, groupID string) (*Season, error) {
	var season Season
	if err := tx.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status
FROM seasons
WHERE group_id = $1 AND status = 'Active'
ORDER BY created_at DESC
LIMIT 1`, groupID).Scan(
		&season.ID,
		&season.GroupID,
		&season.CommissionerPlayerID,
		&season.Status,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &season, nil
}

func (s *PostgresStore) inviteStatus(ctx context.Context, tx *sql.Tx, token string) (InviteAcceptStatus, error) {
	var usedBy sql.NullString
	var expiresAt time.Time
	if err := tx.QueryRowContext(ctx, `
SELECT used_by_player_id, expires_at
FROM invites
WHERE token = $1`, token).Scan(&usedBy, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InviteInvalid, nil
		}
		return InviteInvalid, err
	}
	if usedBy.Valid {
		return InviteUsed, nil
	}
	if time.Now().After(expiresAt) {
		return InviteExpired, nil
	}
	return InviteInvalid, nil
}

func isSeasonOpenConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}
	return pgErr.ConstraintName == "seasons_one_open_per_group_idx" || pgErr.ConstraintName == "seasons_pkey"
}

func getProfileByAuthIdentity(ctx context.Context, tx *sql.Tx, provider string, subject string) (MeResponse, error) {
	var profile MeResponse
	if err := tx.QueryRowContext(ctx, `
SELECT accounts.id, accounts.email, players.id, players.display_name
FROM accounts
JOIN auth_identities ON auth_identities.account_id = accounts.id
JOIN players ON players.account_id = accounts.id
WHERE auth_identities.provider = $1 AND auth_identities.subject = $2`, provider, subject).Scan(
		&profile.Account.ID,
		&profile.Account.Email,
		&profile.Player.ID,
		&profile.Player.DisplayName,
	); err != nil {
		return MeResponse{}, err
	}
	if profile.Account.ID == "" || profile.Player.ID == "" {
		return MeResponse{}, fmt.Errorf("incomplete profile for auth identity")
	}
	return profile, nil
}
