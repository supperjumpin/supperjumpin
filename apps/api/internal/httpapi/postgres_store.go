package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/db"
	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type PostgresStore struct {
	db      *sql.DB
	queries *db.Queries
}

// isUniqueViolation reports whether err (or one of its wrapped errors) is a
// Postgres unique-constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	d, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := d.PingContext(ctx); err != nil {
		d.Close()
		return nil, err
	}
	return &PostgresStore{
		db:      d,
		queries: db.New(d),
	}, nil
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
	qtx := s.queries.WithTx(tx)

	profile, err := qtx.GetProfileByAuthIdentity(ctx, db.GetProfileByAuthIdentityParams{
		Provider: identity.Provider,
		Subject:  identity.Subject,
	})
	if err == nil {
		return MeResponse{
			Account: Account{ID: profile.ID, Email: profile.Email},
			Player:  Player{ID: profile.ID_2, DisplayName: profile.DisplayName},
		}, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return MeResponse{}, err
	}
	key := identity.Provider + ":" + identity.Subject
	account := Account{ID: stableID("account", key), Email: identity.Email}
	player := Player{ID: stableID("player", account.ID), DisplayName: displayName(identity.Email)}
	if err := qtx.UpsertAccount(ctx, db.UpsertAccountParams{
		ID: account.ID, Email: account.Email,
	}); err != nil {
		return MeResponse{}, err
	}
	if err := qtx.InsertAuthIdentity(ctx, db.InsertAuthIdentityParams{
		Provider: identity.Provider, Subject: identity.Subject, AccountID: account.ID,
	}); err != nil {
		return MeResponse{}, err
	}
	if err := qtx.InsertPlayer(ctx, db.InsertPlayerParams{
		ID: player.ID, AccountID: account.ID, DisplayName: player.DisplayName,
	}); err != nil {
		return MeResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return MeResponse{}, err
	}
	return MeResponse{Account: account, Player: player}, nil
}
func (s *PostgresStore) UpdateDisplayName(ctx context.Context, playerID string, displayName string) (Player, error) {
	if err := s.queries.UpdatePlayerDisplayName(ctx, db.UpdatePlayerDisplayNameParams{
		ID:          playerID,
		DisplayName: displayName,
	}); err != nil {
		return Player{}, err
	}
	return Player{ID: playerID, DisplayName: displayName}, nil
}
func (s *PostgresStore) Now() time.Time {
	return time.Now()
}
func (s *PostgresStore) GroupHomeForGroup(ctx context.Context, groupID string, player Player) (GroupHomeResponse, bool, error) {
	newlyFinalized, err := s.ensureSeasonStatusesForGroup(ctx, groupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	for _, id := range newlyFinalized {
		if err := game.AutoFinalizeSeason(ctx, s, id); err != nil {
			return GroupHomeResponse{}, false, err
		}
	}
	return s.groupHomeForGroup(ctx, groupID, player)
}
func (s *PostgresStore) GroupHomeForSeason(ctx context.Context, seasonID string, player Player) (GroupHomeResponse, bool, error) {
	season, err := seasonInDB(ctx, s.db, seasonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GroupHomeResponse{}, false, nil
		}
		return GroupHomeResponse{}, false, err
	}
	return s.GroupHomeForGroup(ctx, season.GroupID, player)
}

// --- PostgresStore private helpers ---
func (s *PostgresStore) groupHomeForGroup(ctx context.Context, groupID string, player Player) (GroupHomeResponse, bool, error) {
	group, membership, ok, err := s.groupMembershipLoad(ctx, groupID, player)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	if !ok {
		return GroupHomeResponse{}, false, nil
	}
	recentJumps, err := recentPerformedJumpsForGroupQuery(ctx, s.db, groupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	standings, err := game.Standings(ctx, s, groupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	currentSeason, err := s.currentSeasonForGroup(ctx, groupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}
	return groupHome(group, membership, currentSeason, recentJumps, standingsFromGame(standings)), true, nil
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
func (s *PostgresStore) ensureSeasonStatusesForGroup(ctx context.Context, groupID string) ([]string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	newlyFinalized, err := ensureSeasonStatusesForGroupInTx(ctx, tx, groupID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return newlyFinalized, nil
}
func (s *PostgresStore) ensureSeasonStatusesForJump(ctx context.Context, jumpID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var groupID string
	if err := tx.QueryRowContext(ctx, `SELECT group_id FROM jumps WHERE id = $1`, jumpID).Scan(&groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if _, err := ensureSeasonStatusesForGroupInTx(ctx, tx, groupID); err != nil {
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
	if _, err := ensureSeasonStatusesForGroupInTx(ctx, tx, groupID); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *PostgresStore) groupMembershipLoad(ctx context.Context, groupID string, player Player) (Group, GroupMembership, bool, error) {
	var group Group
	var membership GroupMembership
	err := s.db.QueryRowContext(ctx, `
SELECT groups.id, groups.name, group_memberships.player_id, group_memberships.role
FROM group_memberships
JOIN groups ON groups.id = group_memberships.group_id
WHERE group_memberships.group_id = $1 AND group_memberships.player_id = $2`, groupID, player.ID).Scan(
		&group.ID, &group.Name, &membership.PlayerID, &membership.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return Group{}, GroupMembership{}, false, nil
	}
	if err != nil {
		return Group{}, GroupMembership{}, false, err
	}
	membership.GroupID = group.ID
	return group, membership, true, nil
}

// --- Package-level helpers for PostgresStore ---
func ensureSeasonStatusesForGroupInTx(ctx context.Context, tx *sql.Tx, groupID string) ([]string, error) {
	// TODO: If Groups/Seasons become active again, move deadline-based status
	// progression into a pure game helper and let adapters only persist changes.
	rows, err := tx.QueryContext(ctx, `
SELECT id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1`, groupID)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		seasons = append(seasons, season)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var newlyFinalized []string
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
			return nil, err
		}
		if newStatus == "Finalized" {
			newlyFinalized = append(newlyFinalized, season.id)
		}
	}
	return newlyFinalized, nil
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

// jumpViewQueryer is satisfied by *sql.DB and *sql.Tx for querying jump views.
type jumpViewQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func recentPerformedJumpsForGroupQuery(ctx context.Context, queryer jumpViewQueryer, groupID string) ([]PerformedJumpView, error) {
	rows, err := queryer.QueryContext(ctx, `
SELECT jumps.id, jumps.group_id, jumps.player_id, jumps.season_id, jumps.status, jumps.source, jumps.destination, jumps.food, jumps.final_score, jumps.grace_period_expires_at,
       evidences.id, evidences.caption, evidences.media_object_key, evidences.created_at,
       players.id, players.display_name
FROM jumps
JOIN evidences ON evidences.jump_id = jumps.id
JOIN players ON players.id = jumps.player_id
WHERE jumps.group_id = $1 AND jumps.status IN ('Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump')
ORDER BY evidences.created_at DESC, jumps.id DESC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	performed := []PerformedJumpView{}
	for rows.Next() {
		var jump Jump
		var seasonID sql.NullString
		var gracePeriodExpiresAt sql.NullTime
		var evidence Evidence
		var performer Player
		if err := rows.Scan(
			&jump.ID,
			&jump.GroupID,
			&jump.PlayerID,
			&seasonID,
			&jump.Status,
			&jump.Source,
			&jump.Destination,
			&jump.Food,
			&jump.FinalScore,
			&gracePeriodExpiresAt,
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
			jump.SeasonID = &seasonID.String
			jump.OffSeason = false
		} else {
			jump.OffSeason = true
		}
		if gracePeriodExpiresAt.Valid {
			jump.GracePeriodExpiresAt = gracePeriodExpiresAt.Time
		}
		disputes, err := disputesForJumpQuery(ctx, queryer, jump.ID)
		if err != nil {
			return nil, err
		}
		performed = append(performed, PerformedJumpView{Jump: jump, Performer: performer, Evidence: evidence, Disputes: disputes})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return performed, nil
}
func disputesForJumpQuery(ctx context.Context, queryer jumpViewQueryer, jumpID string) ([]Dispute, error) {
	rows, err := queryer.QueryContext(ctx, `
SELECT id, jump_id, raised_by_player_id, concern, details, status,
       resolution, resolution_reason, resolved_by_player_id,
       override_resolution, override_reason, override_by_player_id
FROM disputes
WHERE jump_id = $1
ORDER BY created_at ASC, id ASC`, jumpID)
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
			&dispute.JumpID,
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
func disputeInDB(ctx context.Context, db *sql.DB, disputeID string) (Dispute, error) {
	var dispute Dispute
	var resolution sql.NullString
	var resolutionReason sql.NullString
	var resolvedByPlayerID sql.NullString
	var overrideResolution sql.NullString
	var overrideReason sql.NullString
	var overrideByPlayerID sql.NullString
	if err := db.QueryRowContext(ctx, `
SELECT id, jump_id, raised_by_player_id, concern, details, status,
       resolution, resolution_reason, resolved_by_player_id,
       override_resolution, override_reason, override_by_player_id
FROM disputes
WHERE id = $1`, disputeID).Scan(
		&dispute.ID,
		&dispute.JumpID,
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
func seasonInDB(ctx context.Context, db *sql.DB, seasonID string) (Season, error) {
	var season Season
	err := db.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons WHERE id = $1`, seasonID).Scan(
		&season.ID, &season.GroupID, &season.CommissionerPlayerID, &season.Status, &season.SubmissionDeadline, &season.JudgingDeadline)
	return season, err
}
