package game

import (
	"context"
	"errors"
	"time"
)

var (
	ErrOpenMonthNotClosed = errors.New("Open month has not soft-closed yet")
)

// PlayerSnapshot is a read-only view of a Player needed for Open standings.
type PlayerSnapshot struct {
	ID          string
	DisplayName string
}

// StandingEntry accumulates scores for a player within a competition period.
type StandingEntry struct {
	PlayerID    string
	SeasonScore int
	JudgedJumps int
}

// OpenRepository defines persistence operations for the Open scoring flow.
type OpenRepository interface {
	// JumpsForOpenMonth returns all jumps created in the given calendar month
	// that are eligible for Open scoring.
	JumpsForOpenMonth(ctx context.Context, year, month int) ([]JumpSnapshot, error)

	// JudgmentsForJump returns all judgments for a given jump.
	JudgmentsForJump(ctx context.Context, jumpID string) ([]Judgment, error)

	// UpdateJumpOpenFinalScore persists the computed open_final_score for a jump.
	UpdateJumpOpenFinalScore(ctx context.Context, jumpID string, score *int) error

	// PlayersForOpenMonth returns players who performed jumps in the given month.
	PlayersForOpenMonth(ctx context.Context, year, month int) ([]PlayerSnapshot, error)

	// UpsertOpenStanding persists a computed standing for a player in an Open month.
	UpsertOpenStanding(ctx context.Context, year, month int, playerID string, score, judgedJumps int) error
}

// ComputeOpenScoresInput bundles parameters for computing Open Final Scores.
type ComputeOpenScoresInput struct {
	Year  int
	Month int
}

// ComputeOpenScoresResult is the outcome of computing Open scores.
type ComputeOpenScoresResult struct {
	Allowed bool
	Err     error
}

// ComputeOpenScores evaluates Open final scores for a completed calendar month.
//
// It returns ErrOpenMonthNotClosed when the requested month has not yet ended
// relative to the provided clock. Season-provenance Judgments are excluded.
func ComputeOpenScores(ctx context.Context, repo OpenRepository, input ComputeOpenScoresInput, now time.Time) ComputeOpenScoresResult {
	// 1. Soft-close gate: the requested month must be fully in the past.
	deadline := time.Date(input.Year, time.Month(input.Month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
	if now.Before(deadline) {
		return ComputeOpenScoresResult{Err: ErrOpenMonthNotClosed}
	}

	// 2. Fetch jumps for the month.
	jumps, err := repo.JumpsForOpenMonth(ctx, input.Year, input.Month)
	if err != nil {
		return ComputeOpenScoresResult{Err: err}
	}

	// 3. Compute open_final_score for each eligible jump.
	byPlayer := map[string]*StandingEntry{}
	for _, jump := range jumps {
		if jump.Status != "Performed Jump" && jump.Status != "Judged Jump" {
			continue
		}
		judgments, err := repo.JudgmentsForJump(ctx, jump.ID)
		if err != nil {
			return ComputeOpenScoresResult{Err: err}
		}
		qualifying := filterOpenJudgments(judgments)
		var score *int
		if len(qualifying) > 0 {
			s := finalScore(qualifying)
			score = &s
			entry := byPlayer[jump.PlayerID]
			if entry == nil {
				entry = &StandingEntry{PlayerID: jump.PlayerID}
				byPlayer[jump.PlayerID] = entry
			}
			entry.SeasonScore += s
			entry.JudgedJumps++
		}
		if err := repo.UpdateJumpOpenFinalScore(ctx, jump.ID, score); err != nil {
			return ComputeOpenScoresResult{Err: err}
		}
	}

	// 4. Fetch players for display names and upsert standings.
	players, err := repo.PlayersForOpenMonth(ctx, input.Year, input.Month)
	if err != nil {
		return ComputeOpenScoresResult{Err: err}
	}
	displayName := make(map[string]string, len(players))
	for _, p := range players {
		displayName[p.ID] = p.DisplayName
	}
	for playerID, entry := range byPlayer {
		if err := repo.UpsertOpenStanding(ctx, input.Year, input.Month, playerID, entry.SeasonScore, entry.JudgedJumps); err != nil {
			return ComputeOpenScoresResult{Err: err}
		}
	}

	return ComputeOpenScoresResult{Allowed: true}
}

func filterOpenJudgments(judgments []Judgment) []Judgment {
	var result []Judgment
	for _, j := range judgments {
		if j.Provenance != "season" {
			result = append(result, j)
		}
	}
	return result
}

func finalScore(judgments []Judgment) int {
	total := 0
	for _, j := range judgments {
		total += j.Commitment + j.Transgression + j.Creativity + j.Presentation
	}
	return total / len(judgments)
}
