package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
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

// game.JudgmentRepository adapter methods for PostgresStore

func (s *PostgresStore) Stunt(ctx context.Context, stuntID string) (game.StuntSnapshot, bool, error) {
	var snap game.StuntSnapshot
	var seasonID sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, player_id, season_id, status, source, destination, food
FROM jumps
WHERE id = $1`, stuntID).Scan(
		&snap.ID,
		&snap.GroupID,
		&snap.PlayerID,
		&seasonID,
		&snap.Status,
		&snap.Source,
		&snap.Destination,
		&snap.Food,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return game.StuntSnapshot{}, false, nil
	}
	if err != nil {
		return game.StuntSnapshot{}, false, err
	}
	if !visiblePerformedStatus(snap.Status) {
		return game.StuntSnapshot{}, false, nil
	}
	if seasonID.Valid {
		snap.SeasonID = &seasonID.String
	}
	return snap, true, nil
}

func (s *PostgresStore) Season(ctx context.Context, seasonID string) (game.SeasonSnapshot, error) {
	var snap game.SeasonSnapshot
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons WHERE id = $1`, seasonID).Scan(
		&snap.ID, &snap.GroupID, &snap.CommissionerPlayerID, &snap.Status, &snap.SubmissionDeadline, &snap.JudgingDeadline)
	if errors.Is(err, sql.ErrNoRows) {
		return game.SeasonSnapshot{}, nil
	}
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return snap, nil
}

func (s *PostgresStore) GroupMembership(ctx context.Context, playerID, groupID string) (game.MembershipSnapshot, bool, error) {
	var role string
	err := s.db.QueryRowContext(ctx, `
SELECT role FROM group_memberships
WHERE player_id = $1 AND group_id = $2`, playerID, groupID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return game.MembershipSnapshot{}, false, nil
	}
	if err != nil {
		return game.MembershipSnapshot{}, false, err
	}
	return game.MembershipSnapshot{Role: role}, true, nil
}

// game.GroupRepository adapter methods for PostgresStore

func (s *PostgresStore) InsertGroup(ctx context.Context, groupID, name string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO groups (id, name) VALUES ($1, $2)`, groupID, name)
	return err
}

func (s *PostgresStore) InsertMembership(ctx context.Context, groupID, playerID, role string) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO group_memberships (group_id, player_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (group_id, player_id) DO UPDATE SET role = EXCLUDED.role`, groupID, playerID, role)
	return err
}

func (s *PostgresStore) Group(ctx context.Context, groupID string) (game.GroupSnapshot, bool, error) {
	var snap game.GroupSnapshot
	err := s.db.QueryRowContext(ctx, `SELECT id, name FROM groups WHERE id = $1`, groupID).Scan(&snap.ID, &snap.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return game.GroupSnapshot{}, false, nil
	}
	if err != nil {
		return game.GroupSnapshot{}, false, err
	}
	return snap, true, nil
}

func (s *PostgresStore) Membership(ctx context.Context, playerID, groupID string) (game.GroupMembershipSnapshot, bool, error) {
	var snap game.GroupMembershipSnapshot
	err := s.db.QueryRowContext(ctx, `
SELECT group_id, player_id, role FROM group_memberships
WHERE player_id = $1 AND group_id = $2`, playerID, groupID).Scan(&snap.GroupID, &snap.PlayerID, &snap.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return game.GroupMembershipSnapshot{}, false, nil
	}
	if err != nil {
		return game.GroupMembershipSnapshot{}, false, err
	}
	return snap, true, nil
}

func (s *PostgresStore) MembershipsForPlayer(ctx context.Context, playerID string) ([]game.MembershipWithGroupSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT groups.id, groups.name, group_memberships.group_id, group_memberships.player_id, group_memberships.role
FROM group_memberships
JOIN groups ON groups.id = group_memberships.group_id
WHERE group_memberships.player_id = $1`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []game.MembershipWithGroupSnapshot
	for rows.Next() {
		var m game.MembershipWithGroupSnapshot
		if err := rows.Scan(&m.Group.ID, &m.Group.Name, &m.Membership.GroupID, &m.Membership.PlayerID, &m.Membership.Role); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) InsertInvite(ctx context.Context, groupID, createdByPlayerID string, expiresAt int64) (game.InviteSnapshot, error) {
	id, err := randomToken("invite")
	if err != nil {
		return game.InviteSnapshot{}, err
	}
	token, err := randomToken("invite_token")
	if err != nil {
		return game.InviteSnapshot{}, err
	}
	expiresAtTime := time.Unix(expiresAt, 0).UTC()
	_, err = s.db.ExecContext(ctx, `
INSERT INTO invites (id, group_id, token, created_by_player_id, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO NOTHING`, id, groupID, token, createdByPlayerID, expiresAtTime)
	if err != nil {
		return game.InviteSnapshot{}, err
	}
	return game.InviteSnapshot{
		ID:        id,
		GroupID:   groupID,
		Token:     token,
		CreatedBy: createdByPlayerID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *PostgresStore) InviteByToken(ctx context.Context, token string) (game.InviteSnapshot, bool, error) {
	var snap game.InviteSnapshot
	var usedBy sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, token, created_by_player_id, used_by_player_id, EXTRACT(EPOCH FROM expires_at)::bigint
FROM invites
WHERE token = $1`, token).Scan(&snap.ID, &snap.GroupID, &snap.Token, &snap.CreatedBy, &usedBy, &snap.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return game.InviteSnapshot{}, false, nil
	}
	if err != nil {
		return game.InviteSnapshot{}, false, err
	}
	if usedBy.Valid {
		snap.UsedBy = &usedBy.String
	}
	return snap, true, nil
}

func (s *PostgresStore) MarkInviteUsed(ctx context.Context, token, playerID string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE invites
SET used_by_player_id = $1
WHERE token = $2 AND used_by_player_id IS NULL`, playerID, token)
	return err
}

func (s *PostgresStore) UpsertJudgment(ctx context.Context, stuntID, playerID string, difficulty, transgression, creativity, presentation int) (game.Judgment, bool, error) {
	judgmentID := stableID("judgment", stuntID+":"+playerID)
	var created bool
	err := s.db.QueryRowContext(ctx, `
WITH upsert AS (
  INSERT INTO judgments (id, jump_id, player_id, difficulty, transgression, creativity, presentation)
  VALUES ($1, $2, $3, $4, $5, $6, $7)
  ON CONFLICT (jump_id, player_id) DO UPDATE SET
    difficulty = EXCLUDED.difficulty,
    transgression = EXCLUDED.transgression,
    creativity = EXCLUDED.creativity,
    presentation = EXCLUDED.presentation,
    updated_at = now()
  RETURNING (xmax = 0) AS created
)
SELECT created FROM upsert`, judgmentID, stuntID, playerID, difficulty, transgression, creativity, presentation).Scan(&created)
	if err != nil {
		return game.Judgment{}, false, err
	}
	return game.Judgment{
		ID:            judgmentID,
		JumpID:        stuntID,
		PlayerID:      playerID,
		Difficulty:    difficulty,
		Transgression: transgression,
		Creativity:    creativity,
		Presentation:  presentation,
	}, created, nil
}

// game.StuntPlanningRepository adapter methods for PostgresStore

func (s *PostgresStore) InsertIdea(ctx context.Context, groupID, playerID, source, destination, food string) (game.StuntSnapshot, error) {
	for attempts := 0; attempts < 3; attempts++ {
		id, err := randomToken("stunt")
		if err != nil {
			return game.StuntSnapshot{}, err
		}
		result, err := s.db.ExecContext(ctx, `
INSERT INTO jumps (id, group_id, player_id, status, source, destination, food)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT DO NOTHING`, id, groupID, playerID, "Idea", source, destination, food)
		if err != nil {
			return game.StuntSnapshot{}, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return game.StuntSnapshot{}, err
		}
		if rows == 1 {
			return game.StuntSnapshot{
				ID:          id,
				GroupID:     groupID,
				PlayerID:    playerID,
				Status:      "Idea",
				Source:      source,
				Destination: destination,
				Food:        food,
			}, nil
		}
		if attempts == 2 {
			return game.StuntSnapshot{}, fmt.Errorf("create unique Idea after retries")
		}
	}
	return game.StuntSnapshot{}, fmt.Errorf("create unique Idea: unreachable")
}

func (s *PostgresStore) Idea(ctx context.Context, stuntID string) (game.StuntSnapshot, bool, error) {
	var snap game.StuntSnapshot
	var seasonID sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, player_id, season_id, status, source, destination, food
FROM jumps
WHERE id = $1`, stuntID).Scan(
		&snap.ID,
		&snap.GroupID,
		&snap.PlayerID,
		&seasonID,
		&snap.Status,
		&snap.Source,
		&snap.Destination,
		&snap.Food,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return game.StuntSnapshot{}, false, nil
	}
	if err != nil {
		return game.StuntSnapshot{}, false, err
	}
	if seasonID.Valid {
		snap.SeasonID = &seasonID.String
	}
	return snap, true, nil
}

func (s *PostgresStore) ActiveSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	season, err := s.activeSeasonForGroup(ctx, groupID)
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	if season == nil {
		return game.SeasonSnapshot{}, nil
	}
	return game.SeasonSnapshot{ID: season.ID, Status: "Active"}, nil
}

func (s *PostgresStore) UpdateStuntToPlanned(ctx context.Context, stuntID, playerID string, seasonID *string) (game.StuntSnapshot, error) {
	var snap game.StuntSnapshot
	var resultSeasonID sql.NullString
	err := s.db.QueryRowContext(ctx, `
UPDATE jumps
SET status = 'Planned Jump', season_id = $2
WHERE id = $1
  AND status = 'Idea'
RETURNING id, group_id, player_id, season_id, status, source, destination, food`, stuntID, seasonID).Scan(
		&snap.ID,
		&snap.GroupID,
		&snap.PlayerID,
		&resultSeasonID,
		&snap.Status,
		&snap.Source,
		&snap.Destination,
		&snap.Food,
	)
	if err != nil {
		return game.StuntSnapshot{}, err
	}
	if resultSeasonID.Valid {
		snap.SeasonID = &resultSeasonID.String
	}
	return snap, nil
}

// game.EvidenceRepository adapter methods for PostgresStore

func (s *PostgresStore) PlannedStunt(ctx context.Context, stuntID string) (game.StuntSnapshot, bool, error) {
	var snap game.StuntSnapshot
	var seasonID sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, player_id, season_id, status, source, destination, food
FROM jumps
WHERE id = $1`, stuntID).Scan(
		&snap.ID,
		&snap.GroupID,
		&snap.PlayerID,
		&seasonID,
		&snap.Status,
		&snap.Source,
		&snap.Destination,
		&snap.Food,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return game.StuntSnapshot{}, false, nil
	}
	if err != nil {
		return game.StuntSnapshot{}, false, err
	}
	if snap.Status != "Planned Jump" {
		return game.StuntSnapshot{}, false, nil
	}
	if seasonID.Valid {
		snap.SeasonID = &seasonID.String
	}
	return snap, true, nil
}

func (s *PostgresStore) CreateAuthorization(ctx context.Context, stuntID, playerID, contentType string) (game.AuthorizationSnapshot, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	defer tx.Rollback()

	now := time.Now()
	id, err := randomToken("evidence_upload")
	if err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	mediaObjectKey, err := randomToken("evidence_object")
	if err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	expiresAt := now.Add(15 * time.Minute).UTC()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO evidence_upload_authorizations (id, jump_id, player_id, content_type, media_object_key, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)`,
		id, stuntID, playerID, contentType, mediaObjectKey, expiresAt,
	); err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.AuthorizationSnapshot{}, err
	}
	return game.AuthorizationSnapshot{
		ID:             id,
		JumpID:         stuntID,
		MediaObjectKey: mediaObjectKey,
		ExpiresAt:      expiresAt,
	}, nil
}

func (s *PostgresStore) ClaimAndAdvance(ctx context.Context, authorizationID, stuntID, playerID, caption string) (game.EvidenceCreateResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.EvidenceCreateResult{}, err
	}
	defer tx.Rollback()

	var authMediaObjectKey string
	var authExpiresAt time.Time
	var foundPlayerID string
	err = tx.QueryRowContext(ctx, `
SELECT id, player_id, media_object_key, expires_at
FROM evidence_upload_authorizations
WHERE id = $1 AND jump_id = $2 FOR UPDATE`, authorizationID, stuntID).Scan(
		&authorizationID, &foundPlayerID, &authMediaObjectKey, &authExpiresAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return game.EvidenceCreateResult{}, game.ErrEvidenceUploadAuthorizationNotFound
	}
	if err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if foundPlayerID != playerID || time.Now().After(authExpiresAt) {
		return game.EvidenceCreateResult{}, game.ErrEvidenceUploadAuthorizationNotFound
	}

	evidenceID := stableID("evidence", stuntID+":"+authorizationID)
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO evidences (id, jump_id, player_id, upload_authorization_id, caption, media_object_key, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		evidenceID, stuntID, playerID, authorizationID, caption, authMediaObjectKey, now,
	); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE jumps
SET status = 'Performed Jump'
WHERE id = $1 AND status = 'Planned Jump'`, stuntID,
	); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM evidence_upload_authorizations WHERE id = $1`, authorizationID); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.EvidenceCreateResult{}, err
	}
	return game.EvidenceCreateResult{
		EvidenceID:     evidenceID,
		MediaObjectKey: authMediaObjectKey,
	}, nil
}

// game.DisputeRepository adapter methods for PostgresStore

func (s *PostgresStore) StuntByID(ctx context.Context, stuntID string) (game.StuntSnapshot, bool, error) {
	var snap game.StuntSnapshot
	var seasonID sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, player_id, season_id, status, source, destination, food, final_score
FROM jumps
WHERE id = $1`, stuntID).Scan(
		&snap.ID,
		&snap.GroupID,
		&snap.PlayerID,
		&seasonID,
		&snap.Status,
		&snap.Source,
		&snap.Destination,
		&snap.Food,
		&snap.FinalScore,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return game.StuntSnapshot{}, false, nil
	}
	if err != nil {
		return game.StuntSnapshot{}, false, err
	}
	if seasonID.Valid {
		snap.SeasonID = &seasonID.String
	}
	return snap, true, nil
}

func (s *PostgresStore) InsertDispute(ctx context.Context, stuntID, raisedByPlayerID, concern, details string) (game.DisputeSnapshot, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.DisputeSnapshot{}, err
	}
	defer tx.Rollback()

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM disputes WHERE jump_id = $1`, stuntID).Scan(&count); err != nil {
		return game.DisputeSnapshot{}, err
	}
	dispute := Dispute{
		ID:               stableID("dispute", stuntID+":"+raisedByPlayerID+":"+strconv.Itoa(count+1)),
		JumpID:           stuntID,
		RaisedByPlayerID: raisedByPlayerID,
		Concern:          concern,
		Details:          details,
		Status:           "Open",
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO disputes (id, jump_id, raised_by_player_id, concern, details, status)
VALUES ($1, $2, $3, $4, $5, $6)`, dispute.ID, dispute.JumpID, dispute.RaisedByPlayerID, dispute.Concern, dispute.Details, dispute.Status); err != nil {
		return game.DisputeSnapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.DisputeSnapshot{}, err
	}
	return disputeToSnapshot(dispute), nil
}

func (s *PostgresStore) Dispute(ctx context.Context, disputeID string) (game.DisputeSnapshot, error) {
	dispute, err := disputeInDB(ctx, s.db, disputeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.DisputeSnapshot{}, nil
		}
		return game.DisputeSnapshot{}, err
	}
	return disputeToSnapshot(dispute), nil
}

func (s *PostgresStore) UpdateDisputeResolution(ctx context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE disputes
SET status = 'Resolved', resolution = $2, resolution_reason = $3, resolved_by_player_id = $4
WHERE id = $1`, disputeID, resolution, resolutionReason, resolvedByPlayerID)
	return err
}

func (s *PostgresStore) UpdateDisputeOverride(ctx context.Context, disputeID, overrideResolution, overrideReason, overrideByPlayerID string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE disputes
SET status = 'Overridden', override_resolution = $2, override_reason = $3, override_by_player_id = $4
WHERE id = $1`, disputeID, overrideResolution, overrideReason, overrideByPlayerID)
	return err
}

func (s *PostgresStore) UpdateStuntStatusAfterDispute(ctx context.Context, stuntID, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE jumps SET status = $2, final_score = NULL WHERE id = $1`, stuntID, status)
	return err
}

// game.SeasonRepository adapter methods for PostgresStore

func (s *PostgresStore) OpenSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	var snap game.SeasonSnapshot
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1 AND status IN ('Active', 'Judging Grace Period')
LIMIT 1`, groupID).Scan(
		&snap.ID, &snap.GroupID, &snap.CommissionerPlayerID, &snap.Status, &snap.SubmissionDeadline, &snap.JudgingDeadline)
	if errors.Is(err, sql.ErrNoRows) {
		return game.SeasonSnapshot{}, nil
	}
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return snap, nil
}

func (s *PostgresStore) InsertSeason(ctx context.Context, groupID, commissionerPlayerID string, submissionDeadline, judgingDeadline time.Time) (game.SeasonSnapshot, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM seasons`).Scan(&count); err != nil {
		return game.SeasonSnapshot{}, err
	}
	season := Season{
		ID:                   stableID("season", groupID+":"+strconv.Itoa(count+1)),
		GroupID:              groupID,
		CommissionerPlayerID: commissionerPlayerID,
		Status:               "Active",
		SubmissionDeadline:   submissionDeadline.UTC(),
		JudgingDeadline:      judgingDeadline.UTC(),
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO seasons (id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline)
VALUES ($1, $2, $3, $4, $5, $6)`, season.ID, season.GroupID, season.CommissionerPlayerID, season.Status, season.SubmissionDeadline, season.JudgingDeadline)
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return game.SeasonSnapshot{
		ID:                   season.ID,
		GroupID:              season.GroupID,
		CommissionerPlayerID: season.CommissionerPlayerID,
		Status:               season.Status,
		SubmissionDeadline:   season.SubmissionDeadline,
		JudgingDeadline:      season.JudgingDeadline,
	}, nil
}

func (s *PostgresStore) UpdateSeasonStatus(ctx context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE seasons SET status = $2 WHERE id = $1`, seasonID, toStatus); err != nil {
		return err
	}
	if err := insertSeasonHistoryEntry(ctx, tx, seasonID, action, actorPlayerID, actorRole, override, fromStatus, toStatus); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) StuntsForSeason(ctx context.Context, seasonID string) ([]game.StuntSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, group_id, player_id, season_id, status, source, destination, food, final_score
FROM jumps
WHERE season_id = $1`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []game.StuntSnapshot
	for rows.Next() {
		var snap game.StuntSnapshot
		var sid sql.NullString
		if err := rows.Scan(&snap.ID, &snap.GroupID, &snap.PlayerID, &sid, &snap.Status, &snap.Source, &snap.Destination, &snap.Food, &snap.FinalScore); err != nil {
			return nil, err
		}
		if sid.Valid {
			snap.SeasonID = &sid.String
		}
		result = append(result, snap)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) JudgmentsForStunt(ctx context.Context, stuntID string) ([]game.Judgment, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, jump_id, player_id, difficulty, transgression, creativity, presentation
FROM judgments
WHERE jump_id = $1`, stuntID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []game.Judgment
	for rows.Next() {
		var j game.Judgment
		if err := rows.Scan(&j.ID, &j.JumpID, &j.PlayerID, &j.Difficulty, &j.Transgression, &j.Creativity, &j.Presentation); err != nil {
			return nil, err
		}
		result = append(result, j)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) UpdateStuntFinalization(ctx context.Context, stuntID string, status string, finalScore *int) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE jumps
SET status = $2, final_score = $3
WHERE id = $1`, stuntID, status, finalScore)
	return err
}

func (s *PostgresStore) LatestSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	var snap game.SeasonSnapshot
	err := s.db.QueryRowContext(ctx, `
SELECT id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline
FROM seasons
WHERE group_id = $1
ORDER BY created_at DESC
LIMIT 1`, groupID).Scan(
		&snap.ID, &snap.GroupID, &snap.CommissionerPlayerID, &snap.Status, &snap.SubmissionDeadline, &snap.JudgingDeadline)
	if errors.Is(err, sql.ErrNoRows) {
		return game.SeasonSnapshot{}, nil
	}
	if err != nil {
		return game.SeasonSnapshot{}, err
	}
	return snap, nil
}

func (s *PostgresStore) GroupPlayers(ctx context.Context, groupID string) ([]game.PlayerSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT players.id, players.display_name
FROM group_memberships
JOIN players ON players.id = group_memberships.player_id
WHERE group_memberships.group_id = $1`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []game.PlayerSnapshot
	for rows.Next() {
		var p game.PlayerSnapshot
		if err := rows.Scan(&p.ID, &p.DisplayName); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) SeasonHistoryEntries(ctx context.Context, seasonID string) ([]game.SeasonHistoryEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, season_id, action, actor_player_id, actor_role, override, from_status, to_status
FROM season_history
WHERE season_id = $1
ORDER BY created_at ASC, id ASC`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []game.SeasonHistoryEntry{}
	for rows.Next() {
		var entry game.SeasonHistoryEntry
		if err := rows.Scan(&entry.ID, &entry.SeasonID, &entry.Action, &entry.ActorPlayerID, &entry.ActorRole, &entry.Override, &entry.FromStatus, &entry.ToStatus); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
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
	recentStunts, err := recentPerformedStuntsForGroupQuery(ctx, s.db, groupID)
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
	return groupHome(group, membership, currentSeason, recentStunts, standingsFromGame(standings)), true, nil
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

func (s *PostgresStore) ensureSeasonStatusesForStunt(ctx context.Context, stuntID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var groupID string
	if err := tx.QueryRowContext(ctx, `SELECT group_id FROM jumps WHERE id = $1`, stuntID).Scan(&groupID); err != nil {
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

// stuntViewQueryer is satisfied by *sql.DB and *sql.Tx for querying stunt views.
type stuntViewQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func recentPerformedStuntsForGroupQuery(ctx context.Context, queryer stuntViewQueryer, groupID string) ([]PerformedJumpView, error) {
	rows, err := queryer.QueryContext(ctx, `
SELECT jumps.id, jumps.group_id, jumps.player_id, jumps.season_id, jumps.status, jumps.source, jumps.destination, jumps.food, jumps.final_score,
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
		var stunt Jump
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
		performed = append(performed, PerformedJumpView{Jump: jumpForResponse(stunt), Performer: performer, Evidence: evidence, Disputes: disputes})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return performed, nil
}

func disputesForStuntQuery(ctx context.Context, queryer stuntViewQueryer, stuntID string) ([]Dispute, error) {
	rows, err := queryer.QueryContext(ctx, `
SELECT id, jump_id, raised_by_player_id, concern, details, status,
       resolution, resolution_reason, resolved_by_player_id,
       override_resolution, override_reason, override_by_player_id
FROM disputes
WHERE jump_id = $1
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
