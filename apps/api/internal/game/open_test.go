package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockOpenRepo struct {
	jumpsForOpenMonthFn        func(ctx context.Context, year, month int) ([]JumpSnapshot, error)
	judgmentsForJumpFn         func(ctx context.Context, jumpID string) ([]Judgment, error)
	updateJumpOpenFinalScoreFn func(ctx context.Context, jumpID string, score *int) error
	playersForOpenMonthFn      func(ctx context.Context, year, month int) ([]PlayerSnapshot, error)
	upsertOpenStandingFn       func(ctx context.Context, year, month int, playerID string, score, judgedJumps int) error
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

func TestComputeOpenScores(t *testing.T) {
	t.Run("jump scoring", func(t *testing.T) {
		t.Run("averages multiple qualifying judgments", func(t *testing.T) {
			var updatedScore *int

			repo := &mockOpenRepo{
				jumpsForOpenMonthFn: func(_ context.Context, year, month int) ([]JumpSnapshot, error) {
					return []JumpSnapshot{{ID: "jump_1", PlayerID: "player_1", Status: "Judged Jump"}}, nil
				},
				judgmentsForJumpFn: func(_ context.Context, jumpID string) ([]Judgment, error) {
					return []Judgment{
						{JumpID: jumpID, Provenance: "public", Commitment: 4, Transgression: 4, Creativity: 4, Presentation: 4},
						{JumpID: jumpID, Provenance: "open", Commitment: 2, Transgression: 2, Creativity: 2, Presentation: 2},
						{JumpID: jumpID, Provenance: "season", Commitment: 1, Transgression: 1, Creativity: 1, Presentation: 1},
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

			result := ComputeOpenScores(context.Background(), repo, ComputeOpenScoresInput{Year: 2026, Month: 6}, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))

			if result.Err != nil {
				t.Fatalf("expected no error, got %v", result.Err)
			}
			if updatedScore == nil {
				t.Fatal("expected score to be set")
			}
			if *updatedScore != 12 {
				t.Fatalf("expected score 12, got %d", *updatedScore)
			}
		})

		t.Run("sets nil score for season-only judgments", func(t *testing.T) {
			var updatedJumpID string
			var updatedScore *int

			repo := &mockOpenRepo{
				jumpsForOpenMonthFn: func(_ context.Context, year, month int) ([]JumpSnapshot, error) {
					return []JumpSnapshot{{ID: "jump_1", PlayerID: "player_1", Status: "Judged Jump"}}, nil
				},
				judgmentsForJumpFn: func(_ context.Context, jumpID string) ([]Judgment, error) {
					return []Judgment{{JumpID: jumpID, Provenance: "season", Commitment: 4, Transgression: 4, Creativity: 4, Presentation: 4}}, nil
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

			result := ComputeOpenScores(context.Background(), repo, ComputeOpenScoresInput{Year: 2026, Month: 6}, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))

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
		})

		t.Run("excludes season provenance", func(t *testing.T) {
			var updatedJumpID string
			var updatedScore *int

			repo := &mockOpenRepo{
				jumpsForOpenMonthFn: func(_ context.Context, year, month int) ([]JumpSnapshot, error) {
					return []JumpSnapshot{{ID: "jump_1", PlayerID: "player_1", Status: "Judged Jump"}}, nil
				},
				judgmentsForJumpFn: func(_ context.Context, jumpID string) ([]Judgment, error) {
					return []Judgment{
						{JumpID: jumpID, Provenance: "public", Commitment: 4, Transgression: 4, Creativity: 4, Presentation: 4},
						{JumpID: jumpID, Provenance: "season", Commitment: 1, Transgression: 1, Creativity: 1, Presentation: 1},
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

			result := ComputeOpenScores(context.Background(), repo, ComputeOpenScoresInput{Year: 2026, Month: 6}, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))

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
			if *updatedScore != 16 {
				t.Fatalf("expected score 16, got %d", *updatedScore)
			}
		})
	})

	t.Run("soft close gate", func(t *testing.T) {
		t.Run("rejects current month", func(t *testing.T) {
			repo := &mockOpenRepo{}

			result := ComputeOpenScores(context.Background(), repo, ComputeOpenScoresInput{Year: 2026, Month: 6}, time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC))

			if !errors.Is(result.Err, ErrOpenMonthNotClosed) {
				t.Fatalf("expected ErrOpenMonthNotClosed, got %v", result.Err)
			}
		})
	})

	t.Run("standings", func(t *testing.T) {
		t.Run("computes standings across players", func(t *testing.T) {
			var standings []struct {
				playerID    string
				score       int
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
						return []Judgment{{JumpID: jumpID, Provenance: "public", Commitment: 4, Transgression: 4, Creativity: 4, Presentation: 4}}, nil
					case "jump_2":
						return []Judgment{{JumpID: jumpID, Provenance: "public", Commitment: 2, Transgression: 2, Creativity: 2, Presentation: 2}}, nil
					case "jump_3":
						return []Judgment{{JumpID: jumpID, Provenance: "open", Commitment: 3, Transgression: 3, Creativity: 3, Presentation: 3}}, nil
					default:
						return nil, nil
					}
				},
				updateJumpOpenFinalScoreFn: func(_ context.Context, jumpID string, score *int) error {
					return nil
				},
				playersForOpenMonthFn: func(_ context.Context, year, month int) ([]PlayerSnapshot, error) {
					return []PlayerSnapshot{{ID: "player_1", DisplayName: "Alice"}, {ID: "player_2", DisplayName: "Bob"}}, nil
				},
				upsertOpenStandingFn: func(_ context.Context, year, month int, playerID string, score, judgedJumps int) error {
					standings = append(standings, struct {
						playerID    string
						score       int
						judgedJumps int
					}{playerID, score, judgedJumps})
					return nil
				},
			}

			result := ComputeOpenScores(context.Background(), repo, ComputeOpenScoresInput{Year: 2026, Month: 6}, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))

			if result.Err != nil {
				t.Fatalf("expected no error, got %v", result.Err)
			}
			if !result.Allowed {
				t.Fatal("expected allowed")
			}
			if len(standings) != 2 {
				t.Fatalf("expected 2 standings, got %d", len(standings))
			}

			byPlayer := map[string]struct{ score, judgedJumps int }{}
			for _, s := range standings {
				byPlayer[s.playerID] = struct{ score, judgedJumps int }{s.score, s.judgedJumps}
			}

			alice := byPlayer["player_1"]
			if alice.score != 24 || alice.judgedJumps != 2 {
				t.Fatalf("expected Alice score=24 jumps=2, got score=%d jumps=%d", alice.score, alice.judgedJumps)
			}

			bob := byPlayer["player_2"]
			if bob.score != 12 || bob.judgedJumps != 1 {
				t.Fatalf("expected Bob score=12 jumps=1, got score=%d jumps=%d", bob.score, bob.judgedJumps)
			}
		})
	})
}
