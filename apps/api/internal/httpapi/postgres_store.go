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
	return groupHome(group, membership, nil, []PerformedStuntView{}, []StandingEntry{}), nil
}

func (s *PostgresStore) GroupHome(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error) {
	if err := s.ensureSeasonStatusesForGroup(ctx, groupID); err != nil {
		return GroupHomeResponse{}, false, err
	}
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
	season, err := s.currentSeasonForGroup(ctx, group.ID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	recentStunts, err := recentPerformedStuntsForGroupQuery(ctx, s.db, group.ID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	standings, err := s.standingsForGroup(ctx, group.ID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	return groupHome(group, membership, season, recentStunts, standings), true, nil
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
	if err := tx.Commit(); err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	if err := s.ensureSeasonStatusesForGroup(ctx, group.ID); err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	season, err := s.currentSeasonForGroup(ctx, group.ID)
	if err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	recentStunts, err := recentPerformedStuntsForGroupQuery(ctx, s.db, invite.GroupID)
	if err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	standings, err := s.standingsForGroup(ctx, group.ID)
	if err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	return groupHome(group, membership, season, recentStunts, standings), InviteAccepted, nil
}

func (s *PostgresStore) StartSeason(ctx context.Context, player Player, groupID string, submissionDeadline time.Time, judgingDeadline time.Time) (GroupHomeResponse, bool, error) {
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
		SubmissionDeadline:   submissionDeadline.UTC(),
		JudgingDeadline:      judgingDeadline.UTC(),
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO seasons (id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline)
VALUES ($1, $2, $3, $4, $5, $6)`, season.ID, season.GroupID, season.CommissionerPlayerID, season.Status, season.SubmissionDeadline, season.JudgingDeadline); err != nil {
		if isSeasonOpenConflict(err) {
			return GroupHomeResponse{}, true, ErrSeasonAlreadyOpen
		}
		return GroupHomeResponse{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return GroupHomeResponse{}, false, err
	}
	recentStunts, err := recentPerformedStuntsForGroupQuery(ctx, s.db, groupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	return groupHome(group, membership, &season, recentStunts, []StandingEntry{}), true, nil
}

func (s *PostgresStore) CloseSeasonSubmissions(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error) {
	if err := s.ensureSeasonStatusesForSeason(ctx, seasonID); err != nil {
		return GroupHomeResponse{}, false, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	defer tx.Rollback()

	var group Group
	var membership GroupMembership
	var season Season
	if err := tx.QueryRowContext(ctx, `
SELECT seasons.id, seasons.group_id, seasons.commissioner_player_id, seasons.status, seasons.submission_deadline, seasons.judging_deadline,
       groups.name, group_memberships.player_id, group_memberships.role
FROM seasons
JOIN groups ON groups.id = seasons.group_id
JOIN group_memberships ON group_memberships.group_id = seasons.group_id
WHERE seasons.id = $1 AND group_memberships.player_id = $2`, seasonID, player.ID).Scan(
		&season.ID,
		&season.GroupID,
		&season.CommissionerPlayerID,
		&season.Status,
		&season.SubmissionDeadline,
		&season.JudgingDeadline,
		&group.Name,
		&membership.PlayerID,
		&membership.Role,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			var exists bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM seasons WHERE id = $1)`, seasonID).Scan(&exists); err != nil {
				return GroupHomeResponse{}, false, err
			}
			if !exists {
				return GroupHomeResponse{}, false, ErrSeasonNotFound
			}
			return GroupHomeResponse{}, false, nil
		}
		return GroupHomeResponse{}, false, err
	}
	group.ID = season.GroupID
	membership.GroupID = season.GroupID
	if player.ID != season.CommissionerPlayerID && membership.Role != "Group Admin" {
		return GroupHomeResponse{}, false, nil
	}
	if season.Status == "Active" {
		fromStatus := season.Status
		if _, err := tx.ExecContext(ctx, `UPDATE seasons SET status = 'Judging Grace Period' WHERE id = $1`, season.ID); err != nil {
			return GroupHomeResponse{}, false, err
		}
		season.Status = "Judging Grace Period"
		if err := insertSeasonHistoryEntry(ctx, tx, season.ID, "Submissions Closed", player.ID, membership.Role, player.ID != season.CommissionerPlayerID, fromStatus, season.Status); err != nil {
			return GroupHomeResponse{}, false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return GroupHomeResponse{}, false, err
	}

	recentStunts, err := recentPerformedStuntsForGroupQuery(ctx, s.db, season.GroupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	standings, err := s.standingsForGroup(ctx, season.GroupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	currentSeason, err := s.currentSeasonForGroup(ctx, season.GroupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	return groupHome(group, membership, currentSeason, recentStunts, standings), true, nil
}

func (s *PostgresStore) FinalizeSeason(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error) {
	if err := s.ensureSeasonStatusesForSeason(ctx, seasonID); err != nil {
		return GroupHomeResponse{}, false, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	defer tx.Rollback()

	var group Group
	var membership GroupMembership
	var season Season
	if err := tx.QueryRowContext(ctx, `
SELECT seasons.id, seasons.group_id, seasons.commissioner_player_id, seasons.status, seasons.submission_deadline, seasons.judging_deadline,
       groups.name, group_memberships.player_id, group_memberships.role
FROM seasons
JOIN groups ON groups.id = seasons.group_id
JOIN group_memberships ON group_memberships.group_id = seasons.group_id
WHERE seasons.id = $1 AND group_memberships.player_id = $2`, seasonID, player.ID).Scan(
		&season.ID,
		&season.GroupID,
		&season.CommissionerPlayerID,
		&season.Status,
		&season.SubmissionDeadline,
		&season.JudgingDeadline,
		&group.Name,
		&membership.PlayerID,
		&membership.Role,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			var exists bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM seasons WHERE id = $1)`, seasonID).Scan(&exists); err != nil {
				return GroupHomeResponse{}, false, err
			}
			if !exists {
				return GroupHomeResponse{}, false, ErrSeasonNotFound
			}
			return GroupHomeResponse{}, false, nil
		}
		return GroupHomeResponse{}, false, err
	}
	group.ID = season.GroupID
	membership.GroupID = season.GroupID
	if player.ID != season.CommissionerPlayerID && membership.Role != "Group Admin" {
		return GroupHomeResponse{}, false, nil
	}
	if season.Status != "Finalized" {
		fromStatus := season.Status
		if _, err := tx.ExecContext(ctx, `UPDATE seasons SET status = 'Finalized' WHERE id = $1`, season.ID); err != nil {
			return GroupHomeResponse{}, false, err
		}
		if err := finalizeSeasonStuntsInTx(ctx, tx, season.ID); err != nil {
			return GroupHomeResponse{}, false, err
		}
		if err := insertSeasonHistoryEntry(ctx, tx, season.ID, "Season Finalized", player.ID, membership.Role, player.ID != season.CommissionerPlayerID, fromStatus, "Finalized"); err != nil {
			return GroupHomeResponse{}, false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return GroupHomeResponse{}, false, err
	}

	recentStunts, err := recentPerformedStuntsForGroupQuery(ctx, s.db, season.GroupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	standings, err := s.standingsForGroup(ctx, season.GroupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	currentSeason, err := s.currentSeasonForGroup(ctx, season.GroupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	return groupHome(group, membership, currentSeason, recentStunts, standings), true, nil
}

func (s *PostgresStore) SeasonHistory(ctx context.Context, player Player, seasonID string) (SeasonHistoryResponse, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SeasonHistoryResponse{}, false, err
	}
	defer tx.Rollback()

	var groupID string
	if err := tx.QueryRowContext(ctx, `SELECT group_id FROM seasons WHERE id = $1`, seasonID).Scan(&groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SeasonHistoryResponse{}, false, ErrSeasonNotFound
		}
		return SeasonHistoryResponse{}, false, err
	}
	if _, ok, err := groupMembershipInTx(ctx, tx, player, groupID); err != nil {
		return SeasonHistoryResponse{}, false, err
	} else if !ok {
		return SeasonHistoryResponse{}, false, nil
	}
	rows, err := tx.QueryContext(ctx, `
SELECT id, season_id, action, actor_player_id, actor_role, override, from_status, to_status, created_at
FROM season_history
WHERE season_id = $1
ORDER BY created_at ASC, id ASC`, seasonID)
	if err != nil {
		return SeasonHistoryResponse{}, false, err
	}
	defer rows.Close()

	entries := []SeasonHistoryEntry{}
	for rows.Next() {
		var entry SeasonHistoryEntry
		if err := rows.Scan(&entry.ID, &entry.SeasonID, &entry.Action, &entry.ActorPlayerID, &entry.ActorRole, &entry.Override, &entry.FromStatus, &entry.ToStatus, &entry.CreatedAt); err != nil {
			return SeasonHistoryResponse{}, false, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return SeasonHistoryResponse{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return SeasonHistoryResponse{}, false, err
	}
	return SeasonHistoryResponse{Entries: entries}, true, nil
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
	stunt := Stunt{
		GroupID:     groupID,
		PlayerID:    player.ID,
		Status:      "Idea",
		Source:      source,
		Destination: destination,
		Food:        food,
		OffSeason:   true,
	}
	for attempts := 0; attempts < 3; attempts++ {
		id, err := randomToken("stunt")
		if err != nil {
			return Stunt{}, false, err
		}
		stunt.ID = id
		result, err := tx.ExecContext(ctx, `
INSERT INTO stunts (id, group_id, player_id, status, source, destination, food)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT DO NOTHING`, stunt.ID, stunt.GroupID, stunt.PlayerID, stunt.Status, stunt.Source, stunt.Destination, stunt.Food)
		if err != nil {
			return Stunt{}, false, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return Stunt{}, false, err
		}
		if rows == 1 {
			break
		}
		if attempts == 2 {
			return Stunt{}, false, fmt.Errorf("create unique Idea after retries")
		}
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
	seasonID := (*string)(nil)
	if !offSeason {
		season, err := activeSeasonForGroupInTx(ctx, tx, stunt.GroupID)
		if err != nil {
			return Stunt{}, false, err
		}
		if season != nil {
			seasonID = &season.ID
		}
	}
	var updated Stunt
	var updatedSeasonID sql.NullString
	if err := tx.QueryRowContext(ctx, `
UPDATE stunts
SET status = 'Planned Stunt', season_id = $2
WHERE id = $1
  AND status = 'Idea'
  AND player_id = $3
  AND EXISTS (
    SELECT 1
    FROM group_memberships
    WHERE group_memberships.group_id = stunts.group_id
      AND group_memberships.player_id = $3
  )
RETURNING id, group_id, player_id, season_id, status, source, destination, food, final_score`, stunt.ID, seasonID, player.ID).Scan(
		&updated.ID,
		&updated.GroupID,
		&updated.PlayerID,
		&updatedSeasonID,
		&updated.Status,
		&updated.Source,
		&updated.Destination,
		&updated.Food,
		&updated.FinalScore,
	); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return Stunt{}, false, err
		}
		latest, latestErr := stuntInTx(ctx, tx, ideaID)
		if errors.Is(latestErr, sql.ErrNoRows) {
			return Stunt{}, false, ErrStuntNotFound
		}
		if latestErr != nil {
			return Stunt{}, false, latestErr
		}
		if _, ok, err := groupMembershipInTx(ctx, tx, player, latest.GroupID); err != nil {
			return Stunt{}, false, err
		} else if !ok || latest.PlayerID != player.ID {
			return Stunt{}, false, nil
		}
		return Stunt{}, false, ErrStuntNotFound
	}
	if updatedSeasonID.Valid {
		updated.SeasonID = &updatedSeasonID.String
		updated.OffSeason = false
	} else {
		updated.OffSeason = true
	}
	if err := tx.Commit(); err != nil {
		return Stunt{}, false, err
	}
	return updated, true, nil
}

func (s *PostgresStore) AuthorizeEvidenceUpload(ctx context.Context, player Player, stuntID string, contentType string) (EvidenceUploadAuthorization, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EvidenceUploadAuthorization{}, false, err
	}
	defer tx.Rollback()

	stunt, err := stuntInTx(ctx, tx, stuntID)
	if errors.Is(err, sql.ErrNoRows) {
		return EvidenceUploadAuthorization{}, false, ErrStuntNotFound
	}
	if err != nil {
		return EvidenceUploadAuthorization{}, false, err
	}
	if stunt.Status != "Planned Stunt" {
		return EvidenceUploadAuthorization{}, false, ErrStuntNotFound
	}
	if _, ok, err := groupMembershipInTx(ctx, tx, player, stunt.GroupID); err != nil {
		return EvidenceUploadAuthorization{}, false, err
	} else if !ok || stunt.PlayerID != player.ID {
		return EvidenceUploadAuthorization{}, false, nil
	}

	id, err := randomToken("evidence_upload")
	if err != nil {
		return EvidenceUploadAuthorization{}, false, err
	}
	mediaObjectKey, err := randomToken("evidence_object")
	if err != nil {
		return EvidenceUploadAuthorization{}, false, err
	}
	authorization := EvidenceUploadAuthorization{
		ID:             id,
		StuntID:        stuntID,
		UploadURL:      "https://storage.supperjumpin.test/uploads/" + mediaObjectKey,
		UploadMethod:   httpMethodPut,
		UploadHeaders:  map[string]string{"Content-Type": contentType},
		MediaObjectKey: mediaObjectKey,
		ExpiresAt:      time.Now().Add(15 * time.Minute).UTC(),
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO evidence_upload_authorizations (id, stunt_id, player_id, content_type, media_object_key, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)`,
		authorization.ID,
		authorization.StuntID,
		player.ID,
		contentType,
		authorization.MediaObjectKey,
		authorization.ExpiresAt,
	); err != nil {
		return EvidenceUploadAuthorization{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return EvidenceUploadAuthorization{}, false, err
	}
	return authorization, true, nil
}

func (s *PostgresStore) SubmitEvidence(ctx context.Context, player Player, stuntID string, uploadAuthorizationID string, caption string) (EvidenceSubmission, bool, error) {
	if err := s.ensureSeasonStatusesForStunt(ctx, stuntID); err != nil {
		return EvidenceSubmission{}, false, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EvidenceSubmission{}, false, err
	}
	defer tx.Rollback()

	stunt, err := stuntInTx(ctx, tx, stuntID)
	if errors.Is(err, sql.ErrNoRows) {
		return EvidenceSubmission{}, false, ErrStuntNotFound
	}
	if err != nil {
		return EvidenceSubmission{}, false, err
	}
	if stunt.Status != "Planned Stunt" {
		return EvidenceSubmission{}, false, ErrStuntNotFound
	}
	if _, ok, err := groupMembershipInTx(ctx, tx, player, stunt.GroupID); err != nil {
		return EvidenceSubmission{}, false, err
	} else if !ok || stunt.PlayerID != player.ID {
		return EvidenceSubmission{}, false, nil
	}
	if open, err := submissionWindowOpenInTx(ctx, tx, stunt); err != nil {
		return EvidenceSubmission{}, false, err
	} else if !open {
		return EvidenceSubmission{}, false, ErrSubmissionWindowClosed
	}

	var authorization EvidenceUploadAuthorization
	var contentType string
	if err := tx.QueryRowContext(ctx, `
SELECT id, stunt_id, content_type, media_object_key, expires_at
FROM evidence_upload_authorizations
WHERE id = $1 AND stunt_id = $2 AND player_id = $3`, uploadAuthorizationID, stuntID, player.ID).Scan(
		&authorization.ID,
		&authorization.StuntID,
		&contentType,
		&authorization.MediaObjectKey,
		&authorization.ExpiresAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EvidenceSubmission{}, false, ErrEvidenceUploadAuthorizationNotFound
		}
		return EvidenceSubmission{}, false, err
	}
	if time.Now().After(authorization.ExpiresAt) {
		return EvidenceSubmission{}, false, ErrEvidenceUploadAuthorizationNotFound
	}
	authorization.UploadURL = "https://storage.supperjumpin.test/uploads/" + authorization.MediaObjectKey
	authorization.UploadMethod = httpMethodPut
	authorization.UploadHeaders = map[string]string{"Content-Type": contentType}

	evidence := Evidence{
		ID:             stableID("evidence", stuntID+":"+uploadAuthorizationID),
		StuntID:        stuntID,
		Caption:        caption,
		MediaObjectKey: authorization.MediaObjectKey,
		CreatedAt:      time.Now().UTC(),
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO evidences (id, stunt_id, player_id, upload_authorization_id, caption, media_object_key, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		evidence.ID,
		evidence.StuntID,
		player.ID,
		uploadAuthorizationID,
		evidence.Caption,
		evidence.MediaObjectKey,
		evidence.CreatedAt,
	); err != nil {
		return EvidenceSubmission{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE stunts
SET status = 'Performed Stunt'
WHERE id = $1 AND status = 'Planned Stunt'`, stuntID); err != nil {
		return EvidenceSubmission{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM evidence_upload_authorizations WHERE id = $1`, uploadAuthorizationID); err != nil {
		return EvidenceSubmission{}, false, err
	}

	stunt.Status = "Performed Stunt"
	if err := tx.Commit(); err != nil {
		return EvidenceSubmission{}, false, err
	}
	return EvidenceSubmission{Stunt: stunt, Evidence: evidence}, true, nil
}

func (s *PostgresStore) SubmitJudgment(ctx context.Context, player Player, stuntID string, difficulty int, transgression int, creativity int, documentation int) (Judgment, bool, bool, error) {
	if !validJudgmentScores(difficulty, transgression, creativity, documentation) {
		return Judgment{}, false, false, ErrInvalidJudgmentScore
	}
	if err := s.ensureSeasonStatusesForStunt(ctx, stuntID); err != nil {
		return Judgment{}, false, false, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Judgment{}, false, false, err
	}
	defer tx.Rollback()

	stunt, err := stuntInTx(ctx, tx, stuntID)
	if errors.Is(err, sql.ErrNoRows) {
		return Judgment{}, false, false, ErrStuntNotFound
	}
	if err != nil {
		return Judgment{}, false, false, err
	}
	if !visiblePerformedStatus(stunt.Status) {
		return Judgment{}, false, false, ErrStuntNotFound
	}
	if stunt.Status != "Performed Stunt" {
		return Judgment{}, false, false, ErrJudgingWindowClosed
	}
	if _, ok, err := groupMembershipInTx(ctx, tx, player, stunt.GroupID); err != nil {
		return Judgment{}, false, false, err
	} else if !ok || stunt.PlayerID == player.ID {
		return Judgment{}, false, false, nil
	}
	if open, err := judgingWindowOpenInTx(ctx, tx, stunt); err != nil {
		return Judgment{}, false, false, err
	} else if !open {
		return Judgment{}, false, false, ErrJudgingWindowClosed
	}

	judgmentID := stableID("judgment", stuntID+":"+player.ID)
	var created bool
	if err := tx.QueryRowContext(ctx, `
WITH upsert AS (
  INSERT INTO judgments (id, stunt_id, player_id, difficulty, transgression, creativity, documentation)
  VALUES ($1, $2, $3, $4, $5, $6, $7)
  ON CONFLICT (stunt_id, player_id) DO UPDATE SET
    difficulty = EXCLUDED.difficulty,
    transgression = EXCLUDED.transgression,
    creativity = EXCLUDED.creativity,
    documentation = EXCLUDED.documentation,
    updated_at = now()
  RETURNING (xmax = 0) AS created
)
SELECT created FROM upsert`, judgmentID, stuntID, player.ID, difficulty, transgression, creativity, documentation).Scan(&created); err != nil {
		return Judgment{}, false, false, err
	}
	if err := tx.Commit(); err != nil {
		return Judgment{}, false, false, err
	}
	return Judgment{
		ID:            judgmentID,
		StuntID:       stuntID,
		PlayerID:      player.ID,
		Difficulty:    difficulty,
		Transgression: transgression,
		Creativity:    creativity,
		Documentation: documentation,
	}, true, created, nil
}

func (s *PostgresStore) CreateDispute(ctx context.Context, player Player, stuntID string, concern string, details string) (Dispute, bool, error) {
	if !validDisputeConcern(concern) {
		return Dispute{}, false, ErrInvalidDisputeConcern
	}
	if err := s.ensureSeasonStatusesForStunt(ctx, stuntID); err != nil {
		return Dispute{}, false, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Dispute{}, false, err
	}
	defer tx.Rollback()

	stunt, err := stuntInTx(ctx, tx, stuntID)
	if errors.Is(err, sql.ErrNoRows) {
		return Dispute{}, false, ErrStuntNotFound
	}
	if err != nil {
		return Dispute{}, false, err
	}
	if !disputableStuntStatus(stunt.Status) {
		return Dispute{}, false, ErrStuntNotFound
	}
	if _, ok, err := groupMembershipInTx(ctx, tx, player, stunt.GroupID); err != nil {
		return Dispute{}, false, err
	} else if !ok {
		return Dispute{}, false, nil
	}

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM disputes WHERE stunt_id = $1`, stuntID).Scan(&count); err != nil {
		return Dispute{}, false, err
	}
	dispute := Dispute{
		ID:               stableID("dispute", stuntID+":"+player.ID+":"+strconv.Itoa(count+1)),
		StuntID:          stuntID,
		RaisedByPlayerID: player.ID,
		Concern:          concern,
		Details:          details,
		Status:           "Open",
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO disputes (id, stunt_id, raised_by_player_id, concern, details, status)
VALUES ($1, $2, $3, $4, $5, $6)`, dispute.ID, dispute.StuntID, dispute.RaisedByPlayerID, dispute.Concern, dispute.Details, dispute.Status); err != nil {
		return Dispute{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Dispute{}, false, err
	}
	return dispute, true, nil
}

func (s *PostgresStore) ResolveDispute(ctx context.Context, player Player, disputeID string, resolution string, resolutionReason string) (DisputeResolution, bool, error) {
	if !validDisputeResolution(resolution) {
		return DisputeResolution{}, false, ErrInvalidDisputeResolution
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DisputeResolution{}, false, err
	}
	defer tx.Rollback()

	dispute, err := disputeInTx(ctx, tx, disputeID)
	if errors.Is(err, sql.ErrNoRows) {
		return DisputeResolution{}, false, ErrDisputeNotFound
	}
	if err != nil {
		return DisputeResolution{}, false, err
	}
	stunt, err := stuntInTx(ctx, tx, dispute.StuntID)
	if errors.Is(err, sql.ErrNoRows) {
		return DisputeResolution{}, false, ErrStuntNotFound
	}
	if err != nil {
		return DisputeResolution{}, false, err
	}
	membership, ok, err := groupMembershipInTx(ctx, tx, player, stunt.GroupID)
	if err != nil {
		return DisputeResolution{}, false, err
	}
	if !ok {
		return DisputeResolution{}, false, nil
	}

	if dispute.Status == "Open" {
		if stunt.SeasonID == nil || resolution == "Removed Stunt" {
			return DisputeResolution{}, false, nil
		}
		season, err := seasonInTx(ctx, tx, *stunt.SeasonID)
		if errors.Is(err, sql.ErrNoRows) {
			return DisputeResolution{}, false, nil
		}
		if err != nil {
			return DisputeResolution{}, false, err
		}
		if season.CommissionerPlayerID != player.ID {
			return DisputeResolution{}, false, nil
		}
		dispute.Status = "Resolved"
		dispute.Resolution = stringPointer(resolution)
		dispute.ResolutionReason = stringPointer(resolutionReason)
		dispute.ResolvedByPlayerID = stringPointer(player.ID)
		if _, err := tx.ExecContext(ctx, `
UPDATE disputes
SET status = 'Resolved', resolution = $2, resolution_reason = $3, resolved_by_player_id = $4
WHERE id = $1`, dispute.ID, resolution, resolutionReason, player.ID); err != nil {
			return DisputeResolution{}, false, err
		}
	} else {
		if membership.Role != "Group Admin" || resolution == "No Action" {
			return DisputeResolution{}, false, nil
		}
		dispute.Status = "Overridden"
		dispute.OverrideResolution = stringPointer(resolution)
		dispute.OverrideReason = stringPointer(resolutionReason)
		dispute.OverrideByPlayerID = stringPointer(player.ID)
		if _, err := tx.ExecContext(ctx, `
UPDATE disputes
SET status = 'Overridden', override_resolution = $2, override_reason = $3, override_by_player_id = $4
WHERE id = $1`, dispute.ID, resolution, resolutionReason, player.ID); err != nil {
			return DisputeResolution{}, false, err
		}
	}

	stunt = applyDisputeResolutionToStunt(stunt, effectiveDisputeResolution(dispute))
	if _, err := tx.ExecContext(ctx, `UPDATE stunts SET status = $2, final_score = $3 WHERE id = $1`, stunt.ID, stunt.Status, stunt.FinalScore); err != nil {
		return DisputeResolution{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return DisputeResolution{}, false, err
	}
	return DisputeResolution{Stunt: stunt, Dispute: dispute}, true, nil
}

func (s *PostgresStore) activeSeasonForGroup(ctx context.Context, groupID string) (*Season, error) {
	var season Season
	if err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1 AND status = 'Active'
ORDER BY created_at DESC
LIMIT 1`, groupID).Scan(
		&season.ID,
		&season.GroupID,
		&season.CommissionerPlayerID,
		&season.Status,
		&season.SubmissionDeadline,
		&season.JudgingDeadline,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &season, nil
}

func (s *PostgresStore) currentSeasonForGroup(ctx context.Context, groupID string) (*Season, error) {
	var season Season
	if err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1 AND status IN ('Active', 'Judging Grace Period')
ORDER BY created_at DESC
LIMIT 1`, groupID).Scan(
		&season.ID,
		&season.GroupID,
		&season.CommissionerPlayerID,
		&season.Status,
		&season.SubmissionDeadline,
		&season.JudgingDeadline,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &season, nil
}

func (s *PostgresStore) ensureSeasonStatusesForGroup(ctx context.Context, groupID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := ensureSeasonStatusesForGroupInTx(ctx, tx, groupID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) ensureSeasonStatusesForStunt(ctx context.Context, stuntID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var groupID string
	if err := tx.QueryRowContext(ctx, `SELECT group_id FROM stunts WHERE id = $1`, stuntID).Scan(&groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if err := ensureSeasonStatusesForGroupInTx(ctx, tx, groupID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) ensureSeasonStatusesForSeason(ctx context.Context, seasonID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var groupID string
	if err := tx.QueryRowContext(ctx, `SELECT group_id FROM seasons WHERE id = $1`, seasonID).Scan(&groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if err := ensureSeasonStatusesForGroupInTx(ctx, tx, groupID); err != nil {
		return err
	}
	return tx.Commit()
}

func ensureSeasonStatusesForGroupInTx(ctx context.Context, tx *sql.Tx, groupID string) error {
	rows, err := tx.QueryContext(ctx, `
SELECT id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1`, groupID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type seasonStatus struct {
		id                 string
		status             string
		submissionDeadline time.Time
		judgingDeadline    time.Time
	}
	seasons := []seasonStatus{}
	for rows.Next() {
		var season seasonStatus
		if err := rows.Scan(&season.id, &season.status, &season.submissionDeadline, &season.judgingDeadline); err != nil {
			return err
		}
		seasons = append(seasons, season)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	now := time.Now()
	for _, season := range seasons {
		newStatus := season.status
		if newStatus == "Active" && now.After(season.submissionDeadline) {
			newStatus = "Judging Grace Period"
		}
		if newStatus == "Judging Grace Period" && now.After(season.judgingDeadline) {
			newStatus = "Finalized"
		}
		if newStatus == season.status {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE seasons SET status = $2 WHERE id = $1`, season.id, newStatus); err != nil {
			return err
		}
		if newStatus == "Finalized" {
			if err := finalizeSeasonStuntsInTx(ctx, tx, season.id); err != nil {
				return err
			}
		}
	}
	return nil
}

func finalizeSeasonStuntsInTx(ctx context.Context, tx *sql.Tx, seasonID string) error {
	if _, err := tx.ExecContext(ctx, `
UPDATE stunts
SET status = 'Judged Stunt',
    final_score = scores.final_score
FROM (
    SELECT stunt_id, ((sum(difficulty + transgression + creativity + documentation)) / count(*))::int AS final_score
    FROM judgments
    GROUP BY stunt_id
) AS scores
WHERE stunts.id = scores.stunt_id
  AND stunts.season_id = $1
  AND stunts.status = 'Performed Stunt'`, seasonID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
UPDATE stunts
SET status = 'Unjudged Stunt', final_score = NULL
WHERE season_id = $1
  AND status = 'Performed Stunt'
  AND NOT EXISTS (SELECT 1 FROM judgments WHERE judgments.stunt_id = stunts.id)`, seasonID)
	return err
}

func insertSeasonHistoryEntry(ctx context.Context, tx *sql.Tx, seasonID string, action string, actorPlayerID string, actorRole string, override bool, fromStatus string, toStatus string) error {
	id, err := randomToken("season_history")
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO season_history (id, season_id, action, actor_player_id, actor_role, override, from_status, to_status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, id, seasonID, action, actorPlayerID, actorRole, override, fromStatus, toStatus)
	return err
}

func (s *PostgresStore) standingsForGroup(ctx context.Context, groupID string) ([]StandingEntry, error) {
	seasonID, err := latestSeasonIDForGroup(ctx, s.db, groupID)
	if err != nil || seasonID == "" {
		return []StandingEntry{}, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT players.id, players.display_name, sum(stunts.final_score)::int AS season_score, count(*)::int AS judged_stunts
FROM stunts
JOIN players ON players.id = stunts.player_id
WHERE stunts.group_id = $1
  AND stunts.season_id = $2
  AND stunts.status = 'Judged Stunt'
  AND stunts.final_score IS NOT NULL
GROUP BY players.id, players.display_name
ORDER BY season_score DESC, players.display_name ASC`, groupID, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	standings := []StandingEntry{}
	for rows.Next() {
		var entry StandingEntry
		if err := rows.Scan(&entry.Player.ID, &entry.Player.DisplayName, &entry.SeasonScore, &entry.JudgedStunts); err != nil {
			return nil, err
		}
		standings = append(standings, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return standings, nil
}

func latestSeasonIDForGroup(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, groupID string) (string, error) {
	var seasonID string
	if err := queryer.QueryRowContext(ctx, `
SELECT id
FROM seasons
WHERE group_id = $1
ORDER BY created_at DESC
LIMIT 1`, groupID).Scan(&seasonID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return seasonID, nil
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
SELECT id, group_id, player_id, season_id, status, source, destination, food, final_score
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
		&stunt.FinalScore,
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

func judgingWindowOpenInTx(ctx context.Context, tx *sql.Tx, stunt Stunt) (bool, error) {
	if stunt.SeasonID == nil {
		return true, nil
	}
	var status string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM seasons WHERE id = $1`, *stunt.SeasonID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return isOpenSeasonStatus(status), nil
}

func submissionWindowOpenInTx(ctx context.Context, tx *sql.Tx, stunt Stunt) (bool, error) {
	if stunt.SeasonID == nil {
		return true, nil
	}
	var status string
	var submissionDeadline time.Time
	if err := tx.QueryRowContext(ctx, `SELECT status, submission_deadline FROM seasons WHERE id = $1`, *stunt.SeasonID).Scan(&status, &submissionDeadline); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return status == "Active" && time.Now().Before(submissionDeadline), nil
}

type stuntViewQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func recentPerformedStuntsForGroupQuery(ctx context.Context, queryer stuntViewQueryer, groupID string) ([]PerformedStuntView, error) {
	rows, err := queryer.QueryContext(ctx, `
SELECT stunts.id, stunts.group_id, stunts.player_id, stunts.season_id, stunts.status, stunts.source, stunts.destination, stunts.food, stunts.final_score,
       evidences.id, evidences.caption, evidences.media_object_key, evidences.created_at,
       players.id, players.display_name
FROM stunts
JOIN evidences ON evidences.stunt_id = stunts.id
JOIN players ON players.id = stunts.player_id
WHERE stunts.group_id = $1 AND stunts.status IN ('Performed Stunt', 'Judged Stunt', 'Unjudged Stunt', 'Disqualified Stunt')
ORDER BY evidences.created_at DESC, stunts.id DESC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	performed := []PerformedStuntView{}
	for rows.Next() {
		var stunt Stunt
		var seasonID sql.NullString
		var evidence Evidence
		var performer Player
		if err := rows.Scan(
			&stunt.ID,
			&stunt.GroupID,
			&stunt.PlayerID,
			&seasonID,
			&stunt.Status,
			&stunt.Source,
			&stunt.Destination,
			&stunt.Food,
			&stunt.FinalScore,
			&evidence.ID,
			&evidence.Caption,
			&evidence.MediaObjectKey,
			&evidence.CreatedAt,
			&performer.ID,
			&performer.DisplayName,
		); err != nil {
			return nil, err
		}
		if seasonID.Valid {
			stunt.SeasonID = &seasonID.String
			stunt.OffSeason = false
		} else {
			stunt.OffSeason = true
		}
		disputes, err := disputesForStuntQuery(ctx, queryer, stunt.ID)
		if err != nil {
			return nil, err
		}
		performed = append(performed, PerformedStuntView{Stunt: stunt, Performer: performer, Evidence: evidence, Disputes: disputes})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return performed, nil
}

func disputesForStuntQuery(ctx context.Context, queryer stuntViewQueryer, stuntID string) ([]Dispute, error) {
	rows, err := queryer.QueryContext(ctx, `
SELECT id, stunt_id, raised_by_player_id, concern, details, status,
       resolution, resolution_reason, resolved_by_player_id,
       override_resolution, override_reason, override_by_player_id
FROM disputes
WHERE stunt_id = $1
ORDER BY created_at ASC, id ASC`, stuntID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	disputes := []Dispute{}
	for rows.Next() {
		var dispute Dispute
		var resolution sql.NullString
		var resolutionReason sql.NullString
		var resolvedByPlayerID sql.NullString
		var overrideResolution sql.NullString
		var overrideReason sql.NullString
		var overrideByPlayerID sql.NullString
		if err := rows.Scan(
			&dispute.ID,
			&dispute.StuntID,
			&dispute.RaisedByPlayerID,
			&dispute.Concern,
			&dispute.Details,
			&dispute.Status,
			&resolution,
			&resolutionReason,
			&resolvedByPlayerID,
			&overrideResolution,
			&overrideReason,
			&overrideByPlayerID,
		); err != nil {
			return nil, err
		}
		if resolution.Valid {
			dispute.Resolution = &resolution.String
		}
		if resolutionReason.Valid {
			dispute.ResolutionReason = &resolutionReason.String
		}
		if resolvedByPlayerID.Valid {
			dispute.ResolvedByPlayerID = &resolvedByPlayerID.String
		}
		if overrideResolution.Valid {
			dispute.OverrideResolution = &overrideResolution.String
		}
		if overrideReason.Valid {
			dispute.OverrideReason = &overrideReason.String
		}
		if overrideByPlayerID.Valid {
			dispute.OverrideByPlayerID = &overrideByPlayerID.String
		}
		disputes = append(disputes, dispute)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return disputes, nil
}

func disputeInTx(ctx context.Context, tx *sql.Tx, disputeID string) (Dispute, error) {
	var dispute Dispute
	var resolution sql.NullString
	var resolutionReason sql.NullString
	var resolvedByPlayerID sql.NullString
	var overrideResolution sql.NullString
	var overrideReason sql.NullString
	var overrideByPlayerID sql.NullString
	if err := tx.QueryRowContext(ctx, `
SELECT id, stunt_id, raised_by_player_id, concern, details, status,
       resolution, resolution_reason, resolved_by_player_id,
       override_resolution, override_reason, override_by_player_id
FROM disputes
WHERE id = $1`, disputeID).Scan(
		&dispute.ID,
		&dispute.StuntID,
		&dispute.RaisedByPlayerID,
		&dispute.Concern,
		&dispute.Details,
		&dispute.Status,
		&resolution,
		&resolutionReason,
		&resolvedByPlayerID,
		&overrideResolution,
		&overrideReason,
		&overrideByPlayerID,
	); err != nil {
		return Dispute{}, err
	}
	if resolution.Valid {
		dispute.Resolution = &resolution.String
	}
	if resolutionReason.Valid {
		dispute.ResolutionReason = &resolutionReason.String
	}
	if resolvedByPlayerID.Valid {
		dispute.ResolvedByPlayerID = &resolvedByPlayerID.String
	}
	if overrideResolution.Valid {
		dispute.OverrideResolution = &overrideResolution.String
	}
	if overrideReason.Valid {
		dispute.OverrideReason = &overrideReason.String
	}
	if overrideByPlayerID.Valid {
		dispute.OverrideByPlayerID = &overrideByPlayerID.String
	}
	return dispute, nil
}

func seasonInTx(ctx context.Context, tx *sql.Tx, seasonID string) (Season, error) {
	var season Season
	if err := tx.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons
WHERE id = $1`, seasonID).Scan(
		&season.ID,
		&season.GroupID,
		&season.CommissionerPlayerID,
		&season.Status,
		&season.SubmissionDeadline,
		&season.JudgingDeadline,
	); err != nil {
		return Season{}, err
	}
	return season, nil
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
