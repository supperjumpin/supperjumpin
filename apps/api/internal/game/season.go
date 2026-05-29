package game

import (
	"context"
	"errors"
	"sort"
	"time"
)

var (
	ErrSeasonAlreadyOpen = errors.New("Group already has an active or closing Season")
	ErrSeasonNotFound    = errors.New("Season not found")
)

// SeasonHistoryEntry is a read-only view of a season history event.
type SeasonHistoryEntry struct {
	ID            string
	SeasonID      string
	Action        string
	ActorPlayerID string
	ActorRole     string
	Override      bool
	FromStatus    string
	ToStatus      string
}

// PlayerSnapshot is a read-only view of a Player.
type PlayerSnapshot struct {
	ID          string
	DisplayName string
}

// StandingEntry is a computed standing for a player in a season.
type StandingEntry struct {
	PlayerID     string
	DisplayName  string
	SeasonScore  int
	JudgedStunts int
}

// SeasonRepository defines persistence operations for the season lifecycle flow.
type SeasonRepository interface {
	// GroupMembership returns the membership for a player in a group.
	// ok is false when there is no membership.
	GroupMembership(ctx context.Context, playerID, groupID string) (MembershipSnapshot, bool, error)

	// OpenSeasonForGroup returns the open season for a group, if any.
	// SeasonSnapshot.ID is empty when no open season exists.
	OpenSeasonForGroup(ctx context.Context, groupID string) (SeasonSnapshot, error)

	// InsertSeason persists a new season.
	InsertSeason(ctx context.Context, groupID, commissionerPlayerID string, submissionDeadline, judgingDeadline time.Time) (SeasonSnapshot, error)

	// Season returns the season for the given ID.
	Season(ctx context.Context, seasonID string) (SeasonSnapshot, error)

	// UpdateSeasonStatus atomically updates a season's status and records history.
	UpdateSeasonStatus(ctx context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error

	// SeasonHistoryEntries returns the history entries for a season.
	SeasonHistoryEntries(ctx context.Context, seasonID string) ([]SeasonHistoryEntry, error)

	// StuntsForSeason returns all stunts linked to the given season.
	StuntsForSeason(ctx context.Context, seasonID string) ([]StuntSnapshot, error)

	// JudgmentsForStunt returns all judgments for a given stunt.
	JudgmentsForStunt(ctx context.Context, stuntID string) ([]Judgment, error)

	// UpdateStuntFinalization updates a stunt's status and final score.
	UpdateStuntFinalization(ctx context.Context, stuntID string, status string, finalScore *int) error

	// LatestSeasonForGroup returns the latest season (by creation time) for a group.
	LatestSeasonForGroup(ctx context.Context, groupID string) (SeasonSnapshot, error)

	// GroupPlayers returns all players in a group.
	GroupPlayers(ctx context.Context, groupID string) ([]PlayerSnapshot, error)
}

// StartSeasonInput bundles the parameters for starting a Season.
type StartSeasonInput struct {
	GroupID            string
	PlayerID           string
	SubmissionDeadline time.Time
	JudgingDeadline    time.Time
}

// StartSeasonResult is the outcome of starting a Season.
type StartSeasonResult struct {
	Season  SeasonSnapshot
	Allowed bool
	Err     error
}

// CloseSeasonSubmissionsInput bundles the parameters for closing Season submissions.
type CloseSeasonSubmissionsInput struct {
	SeasonID string
	PlayerID string
}

// CloseSeasonSubmissionsResult is the outcome of closing Season submissions.
type CloseSeasonSubmissionsResult struct {
	Season  SeasonSnapshot
	Allowed bool
	Err     error
}

// FinalizeSeasonInput bundles the parameters for finalizing a Season.
type FinalizeSeasonInput struct {
	SeasonID string
	PlayerID string
}

// FinalizeSeasonResult is the outcome of finalizing a Season.
type FinalizeSeasonResult struct {
	Season  SeasonSnapshot
	Allowed bool
	Err     error
}

// SeasonHistoryInput bundles the parameters for viewing Season history.
type SeasonHistoryInput struct {
	SeasonID string
	PlayerID string
}

// SeasonHistoryResult is the outcome of viewing Season history.
type SeasonHistoryResult struct {
	Entries []SeasonHistoryEntry
	Allowed bool
	Err     error
}

// StartSeason evaluates season start rules and persists the result.
//
// Returns Allowed=false when the player is not a group member.
// Returns an error when the group already has an open season or persistence fails.
func StartSeason(ctx context.Context, repo SeasonRepository, input StartSeasonInput) StartSeasonResult {
	// 1. Player must be a group member
	_, ok, err := repo.GroupMembership(ctx, input.PlayerID, input.GroupID)
	if err != nil {
		return StartSeasonResult{Err: err}
	}
	if !ok {
		return StartSeasonResult{Allowed: false}
	}

	// 2. Check no open season exists
	openSeason, err := repo.OpenSeasonForGroup(ctx, input.GroupID)
	if err != nil {
		return StartSeasonResult{Err: err}
	}
	if openSeason.ID != "" {
		return StartSeasonResult{Err: ErrSeasonAlreadyOpen}
	}

	// 3. Create the season
	season, err := repo.InsertSeason(ctx, input.GroupID, input.PlayerID, input.SubmissionDeadline, input.JudgingDeadline)
	if err != nil {
		return StartSeasonResult{Err: err}
	}

	return StartSeasonResult{Season: season, Allowed: true}
}

// CloseSeasonSubmissions evaluates season close-submission rules and persists the result.
//
// Returns ErrSeasonNotFound when the season does not exist.
// Returns Allowed=false when the player is not authorized.
func CloseSeasonSubmissions(ctx context.Context, repo SeasonRepository, input CloseSeasonSubmissionsInput) CloseSeasonSubmissionsResult {
	// 1. Look up the season
	season, err := repo.Season(ctx, input.SeasonID)
	if err != nil {
		return CloseSeasonSubmissionsResult{Err: err}
	}
	if season.ID == "" {
		return CloseSeasonSubmissionsResult{Err: ErrSeasonNotFound}
	}

	// 2. Player must be a group member
	membership, ok, err := repo.GroupMembership(ctx, input.PlayerID, season.GroupID)
	if err != nil {
		return CloseSeasonSubmissionsResult{Err: err}
	}
	if !ok {
		return CloseSeasonSubmissionsResult{Allowed: false}
	}

	// 3. Authority: commissioner or Group Admin
	if input.PlayerID != season.CommissionerPlayerID && membership.Role != "Group Admin" {
		return CloseSeasonSubmissionsResult{Allowed: false}
	}

	// 4. Transition from Active to Judging Grace Period
	if season.Status == "Active" {
		override := input.PlayerID != season.CommissionerPlayerID
		if err := repo.UpdateSeasonStatus(ctx, season.ID, "Submissions Closed", input.PlayerID, membership.Role, override, season.Status, "Judging Grace Period"); err != nil {
			return CloseSeasonSubmissionsResult{Err: err}
		}
		season.Status = "Judging Grace Period"
	}

	return CloseSeasonSubmissionsResult{Season: season, Allowed: true}
}

// FinalizeSeason evaluates season finalization rules and persists the result.
//
// Returns ErrSeasonNotFound when the season does not exist.
// Returns Allowed=false when the player is not authorized.
func FinalizeSeason(ctx context.Context, repo SeasonRepository, input FinalizeSeasonInput) FinalizeSeasonResult {
	// 1. Look up the season
	season, err := repo.Season(ctx, input.SeasonID)
	if err != nil {
		return FinalizeSeasonResult{Err: err}
	}
	if season.ID == "" {
		return FinalizeSeasonResult{Err: ErrSeasonNotFound}
	}

	// 2. Player must be a group member
	membership, ok, err := repo.GroupMembership(ctx, input.PlayerID, season.GroupID)
	if err != nil {
		return FinalizeSeasonResult{Err: err}
	}
	if !ok {
		return FinalizeSeasonResult{Allowed: false}
	}

	// 3. Authority: commissioner or Group Admin
	if input.PlayerID != season.CommissionerPlayerID && membership.Role != "Group Admin" {
		return FinalizeSeasonResult{Allowed: false}
	}

	// 4. Finalize from any non-Finalized status
	if season.Status != "Finalized" {
		override := input.PlayerID != season.CommissionerPlayerID
		if err := repo.UpdateSeasonStatus(ctx, season.ID, "Season Finalized", input.PlayerID, membership.Role, override, season.Status, "Finalized"); err != nil {
			return FinalizeSeasonResult{Err: err}
		}
		if err := finalizeStuntsForSeason(ctx, repo, season.ID); err != nil {
			return FinalizeSeasonResult{Err: err}
		}
		season.Status = "Finalized"
	}

	return FinalizeSeasonResult{Season: season, Allowed: true}
}

// AutoFinalizeSeason evaluates deadline-based finalization for a season.
// It applies stunt finalizations when the season is already Finalized.
// This is safe to call repeatedly and returns nil when the season is not Finalized.
func AutoFinalizeSeason(ctx context.Context, repo SeasonRepository, seasonID string) error {
	season, err := repo.Season(ctx, seasonID)
	if err != nil {
		return err
	}
	if season.Status != "Finalized" {
		return nil
	}
	return finalizeStuntsForSeason(ctx, repo, seasonID)
}

// finalizeStuntsForSeason computes and persists final scores for all performed stunts
// in the season, transitioning them to Judged Stunt or Unjudged Stunt.
func finalizeStuntsForSeason(ctx context.Context, repo SeasonRepository, seasonID string) error {
	stunts, err := repo.StuntsForSeason(ctx, seasonID)
	if err != nil {
		return err
	}
	for _, stunt := range stunts {
		if stunt.Status != "Performed Stunt" {
			continue
		}
		judgments, err := repo.JudgmentsForStunt(ctx, stunt.ID)
		if err != nil {
			return err
		}
		if len(judgments) > 0 {
			score := finalScore(judgments)
			if err := repo.UpdateStuntFinalization(ctx, stunt.ID, "Judged Stunt", &score); err != nil {
				return err
			}
		} else {
			if err := repo.UpdateStuntFinalization(ctx, stunt.ID, "Unjudged Stunt", nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func finalScore(judgments []Judgment) int {
	total := 0
	for _, j := range judgments {
		total += j.Difficulty + j.Transgression + j.Creativity + j.Documentation
	}
	return total / len(judgments)
}

// Standings computes standings for the latest season of a group.
func Standings(ctx context.Context, repo SeasonRepository, groupID string) ([]StandingEntry, error) {
	season, err := repo.LatestSeasonForGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if season.ID == "" {
		return []StandingEntry{}, nil
	}

	stunts, err := repo.StuntsForSeason(ctx, season.ID)
	if err != nil {
		return nil, err
	}

	players, err := repo.GroupPlayers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	displayName := make(map[string]string, len(players))
	for _, p := range players {
		displayName[p.ID] = p.DisplayName
	}

	byPlayer := map[string]*StandingEntry{}
	for _, stunt := range stunts {
		if stunt.Status != "Judged Stunt" || stunt.FinalScore == nil {
			continue
		}
		entry := byPlayer[stunt.PlayerID]
		if entry == nil {
			entry = &StandingEntry{
				PlayerID:    stunt.PlayerID,
				DisplayName: displayName[stunt.PlayerID],
			}
			byPlayer[stunt.PlayerID] = entry
		}
		entry.SeasonScore += *stunt.FinalScore
		entry.JudgedStunts++
	}

	standings := make([]StandingEntry, 0, len(byPlayer))
	for _, entry := range byPlayer {
		standings = append(standings, *entry)
	}
	sort.Slice(standings, func(i, j int) bool {
		if standings[i].SeasonScore == standings[j].SeasonScore {
			return standings[i].DisplayName < standings[j].DisplayName
		}
		return standings[i].SeasonScore > standings[j].SeasonScore
	})

	return standings, nil
}

// SeasonHistory evaluates season history view rules and returns the history.
//
// Returns ErrSeasonNotFound when the season does not exist.
// Returns Allowed=false when the player is not a group member.
func SeasonHistory(ctx context.Context, repo SeasonRepository, input SeasonHistoryInput) SeasonHistoryResult {
	// 1. Look up the season
	season, err := repo.Season(ctx, input.SeasonID)
	if err != nil {
		return SeasonHistoryResult{Err: err}
	}
	if season.ID == "" {
		return SeasonHistoryResult{Err: ErrSeasonNotFound}
	}

	// 2. Player must be a group member
	_, ok, err := repo.GroupMembership(ctx, input.PlayerID, season.GroupID)
	if err != nil {
		return SeasonHistoryResult{Err: err}
	}
	if !ok {
		return SeasonHistoryResult{Allowed: false}
	}

	// 3. Fetch history entries
	entries, err := repo.SeasonHistoryEntries(ctx, input.SeasonID)
	if err != nil {
		return SeasonHistoryResult{Err: err}
	}

	return SeasonHistoryResult{Entries: entries, Allowed: true}
}
