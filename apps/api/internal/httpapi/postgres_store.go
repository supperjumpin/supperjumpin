package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"

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
		return GroupHomeResponse{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return GroupHomeResponse{}, false, err
	}
	return groupHome(group, membership, &season), true, nil
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
