package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockOpenRepo struct {
	jumpsForOpenMonthFn       func(ctx context.Context, year, month int) ([]JumpSnapshot, error)
	judgmentsForJumpFn        func(ctx context.Context, jumpID string) ([]Judgment, error)
	updateJumpOpenFinalScoreFn func(ctx context.Context, jumpID string, score *int) error
	playersForOpenMonthFn     func(ctx context.Context, year, month int) ([]PlayerSnapshot, error)
	upsertOpenStandingFn      func(ctx context.Context, year, month int, playerID string, score, judgedJumps int) error
}

func (m *mockOpenRepo) JumpsForOpenMonth(ctx context.Context, year, month int) ([]JumpSnapshot, error) {
	return m.jumpsForOpenMonthFn(ctx, year, month)
}

func (m *mockOpenRepo) JudgmentsForJump(ctx context.Context, jumpID string) ([]Judgment, error) {
	return m.judgmentsForJumpFn(ctx, jumpID)
}

func (m *mockOpenRepo) UpdateJumpOpenFinalScore(ctx context.Context, jumpID string, score *int) error {
	return m.updateJumpOpenFinalScoreFn(ctx, jumpID, score)
}

func (m *mockOpenRepo) PlayersForOpenMonth(ctx context.Context, year, month int) ([]PlayerSnapshot, error) {
	return m.playersForOpenMonthFn(ctx, year, month)
}

func (m *mockOpenRepo) UpsertOpenStanding(ctx context.Context, year, month int, playerID string, score, judgedJumps int) error {
	return m.upsertOpenStandingFn(ctx, year, month, playerID, score, judgedJumps)
}

func TestComputeOpenScores_AveragesMultipleQualifyingJudgments(t *testing.T) {
	var updatedScore *int

	repo := &mockOpenRepo{
		jumpsForOpenMonthFn: func(_ context.Context, year, month int) ([]JumpSnapshot, error) {
			return []JumpSnapshot{
				{ID: "jump_1", PlayerID: "player_1", Status: "Judged Jump"},
			}, nil
		},
		judgmentsForJumpFn: func(_ context.Context, jumpID string) ([]Judgment, error) {
			return []Judgment{
				{JumpID: jumpID, Provenance: "public", Commitment: 4, Transgression: 4, Creativity: 4, Presentation: 4}, // 16
				{JumpID: jumpID, Provenance: "open", Commitment: 2, Transgression: 2, Creativity: 2, Presentation: 2},   // 8
				{JumpID: jumpID, Provenance: "season", Commitment: 1, Transgression: 1, Creativity: 1, Presentation: 1}, // 4, excluded
			}, nil
		},
		updateJumpOpenFinalScoreFn: func(_ context.Context, jumpID string, score *int) error {
			updatedScore = score
			return nil
		},
		playersForOpenMonthFn: func(_ context.Context, year, month int) ([]PlayerSnapshot, error) {
			return []PlayerSnapshot{{ID: "player_1", DisplayName: "Alice"}}, nil
		},
		upsertOpenStandingFn: func(_ context.Context, year, month int, playerID string, score, judgedJumps int) error {
			return nil
		},
	}

	now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	result := ComputeOpenScores(context.Background(), repo, ComputeOpenScoresInput{Year: 2026, Month: 6}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if updatedScore == nil {
		t.Fatal("expected score to be set")
	}
	// (16 + 8) / 2 = 12
	if *updatedScore != 12 {
		t.Fatalf("expected score 12, got %d", *updatedScore)
	}
}

func TestComputeOpenScores_SetsNilScoreForSeasonOnlyJudgments(t *testing.T) {
	var updatedJumpID string
	var updatedScore *int

	repo := &mockOpenRepo{
		jumpsForOpenMonthFn: func(_ context.Context, year, month int) ([]JumpSnapshot, error) {
			return []JumpSnapshot{
				{ID: "jump_1", PlayerID: "player_1", Status: "Judged Jump"},
			}, nil
		},
		judgmentsForJumpFn: func(_ context.Context, jumpID string) ([]Judgment, error) {
			return []Judgment{
				{JumpID: jumpID, Provenance: "season", Commitment: 4, Transgression: 4, Creativity: 4, Presentation: 4},
			}, nil
		},
		updateJumpOpenFinalScoreFn: func(_ context.Context, jumpID string, score *int) error {
			updatedJumpID = jumpID
			updatedScore = score
			return nil
		},
		playersForOpenMonthFn: func(_ context.Context, year, month int) ([]PlayerSnapshot, error) {
			return nil, nil
		},
		upsertOpenStandingFn: func(_ context.Context, year, month int, playerID string, score, judgedJumps int) error {
			return nil
		},
	}

	now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	result := ComputeOpenScores(context.Background(), repo, ComputeOpenScoresInput{Year: 2026, Month: 6}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed")
	}
	if updatedJumpID != "jump_1" {
		t.Fatalf("expected jump_1, got %q", updatedJumpID)
	}
	if updatedScore != nil {
		t.Fatalf("expected nil score for season-only judgments, got %d", *updatedScore)
	}
}

func TestComputeOpenScores_RejectsCurrentMonth(t *testing.T) {
	repo := &mockOpenRepo{}

	now := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC) // June 15, computing June's open
	result := ComputeOpenScores(context.Background(), repo, ComputeOpenScoresInput{Year: 2026, Month: 6}, now)

	if !errors.Is(result.Err, ErrOpenMonthNotClosed) {
		t.Fatalf("expected ErrOpenMonthNotClosed, got %v", result.Err)
	}
}

func TestComputeOpenScores_ComputesStandingsAcrossPlayers(t *testing.T) {
	var standings []struct {
		playerID     string
		score        int
		judgedJumps int
	}

	repo := &mockOpenRepo{
		jumpsForOpenMonthFn: func(_ context.Context, year, month int) ([]JumpSnapshot, error) {
			return []JumpSnapshot{
				{ID: "jump_1", PlayerID: "player_1", Status: "Judged Jump"},
				{ID: "jump_2", PlayerID: "player_1", Status: "Judged Jump"},
				{ID: "jump_3", PlayerID: "player_2", Status: "Judged Jump"},
			}, nil
		},
		judgmentsForJumpFn: func(_ context.Context, jumpID string) ([]Judgment, error) {
			switch jumpID {
			case "jump_1":
				return []Judgment{{JumpID: jumpID, Provenance: "public", Commitment: 4, Transgression: 4, Creativity: 4, Presentation: 4}}, nil // 16
			case "jump_2":
				return []Judgment{{JumpID: jumpID, Provenance: "public", Commitment: 2, Transgression: 2, Creativity: 2, Presentation: 2}}, nil // 8
			case "jump_3":
				return []Judgment{{JumpID: jumpID, Provenance: "open", Commitment: 3, Transgression: 3, Creativity: 3, Presentation: 3}}, nil // 12
			default:
				return nil, nil
			}
		},
		updateJumpOpenFinalScoreFn: func(_ context.Context, jumpID string, score *int) error {
			return nil
		},
		playersForOpenMonthFn: func(_ context.Context, year, month int) ([]PlayerSnapshot, error) {
			return []PlayerSnapshot{
				{ID: "player_1", DisplayName: "Alice"},
				{ID: "player_2", DisplayName: "Bob"},
			}, nil
		},
		upsertOpenStandingFn: func(_ context.Context, year, month int, playerID string, score, judgedJumps int) error {
			standings = append(standings, struct {
				playerID     string
				score        int
				judgedJumps int
			}{playerID, score, judgedJumps})
			return nil
		},
	}

	now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	result := ComputeOpenScores(context.Background(), repo, ComputeOpenScoresInput{Year: 2026, Month: 6}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed")
	}
	if len(standings) != 2 {
		t.Fatalf("expected 2 standings, got %d", len(standings))
	}

	byPlayer := map[string]struct{ score int; judgedJumps int }{}
	for _, s := range standings {
		byPlayer[s.playerID] = struct{ score int; judgedJumps int }{s.score, s.judgedJumps}
	}

	alice := byPlayer["player_1"]
	if alice.score != 24 || alice.judgedJumps != 2 {
		t.Fatalf("expected Alice score=24 jumps=2, got score=%d jumps=%d", alice.score, alice.judgedJumps)
	}

	bob := byPlayer["player_2"]
	if bob.score != 12 || bob.judgedJumps != 1 {
		t.Fatalf("expected Bob score=12 jumps=1, got score=%d jumps=%d", bob.score, bob.judgedJumps)
	}
}

func TestComputeOpenScores_ExcludesSeasonProvenance(t *testing.T) {
	var updatedJumpID string
	var updatedScore *int

	repo := &mockOpenRepo{
		jumpsForOpenMonthFn: func(_ context.Context, year, month int) ([]JumpSnapshot, error) {
			return []JumpSnapshot{
				{ID: "jump_1", PlayerID: "player_1", Status: "Judged Jump"},
			}, nil
		},
		judgmentsForJumpFn: func(_ context.Context, jumpID string) ([]Judgment, error) {
			return []Judgment{
				{JumpID: jumpID, Provenance: "public", Commitment: 4, Transgression: 4, Creativity: 4, Presentation: 4}, // total = 16
				{JumpID: jumpID, Provenance: "season", Commitment: 1, Transgression: 1, Creativity: 1, Presentation: 1}, // total = 4, excluded
			}, nil
		},
		updateJumpOpenFinalScoreFn: func(_ context.Context, jumpID string, score *int) error {
			updatedJumpID = jumpID
			updatedScore = score
			return nil
		},
		playersForOpenMonthFn: func(_ context.Context, year, month int) ([]PlayerSnapshot, error) {
			return []PlayerSnapshot{{ID: "player_1", DisplayName: "Player One"}}, nil
		},
		upsertOpenStandingFn: func(_ context.Context, year, month int, playerID string, score, judgedJumps int) error {
			return nil
		},
	}

	now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC) // July, computing June's open
	result := ComputeOpenScores(context.Background(), repo, ComputeOpenScoresInput{Year: 2026, Month: 6}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed")
	}
	if updatedJumpID != "jump_1" {
		t.Fatalf("expected jump_1, got %q", updatedJumpID)
	}
	if updatedScore == nil {
		t.Fatal("expected score to be set")
	}
	// Only the public judgment (16) should count, so score = 16
	if *updatedScore != 16 {
		t.Fatalf("expected score 16, got %d", *updatedScore)
	}
}
