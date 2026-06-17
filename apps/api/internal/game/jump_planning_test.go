package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockJumpRepo struct {
	insertPerformedJumpFn func(ctx context.Context, params InsertPerformedJumpParams) (JumpSnapshot, EvidenceSnapshot, error)
}

func (m *mockJumpRepo) InsertPerformedJump(_ context.Context, params InsertPerformedJumpParams) (JumpSnapshot, EvidenceSnapshot, error) {
	return m.insertPerformedJumpFn(nil, params)
}

func testCreatePerformedJump_CreatesPerformedJumpDirectly(t *testing.T) {
	var insertedParams InsertPerformedJumpParams

	repo := &mockJumpRepo{
		insertPerformedJumpFn: func(_ context.Context, params InsertPerformedJumpParams) (JumpSnapshot, EvidenceSnapshot, error) {
			insertedParams = params
			return JumpSnapshot{
				ID:          "jump_1",
				PlayerID:    params.PlayerID,
				Status:      "Performed Jump",
				Source:      params.Source,
				Destination: params.Destination,
				Food:        params.Food,
			}, EvidenceSnapshot{
				ID:             "evidence_1",
				JumpID:         "jump_1",
				PlayerID:       params.PlayerID,
				MediaObjectKey: params.MediaObjectKey,
				Caption:        params.Caption,
			}, nil
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := CreatePerformedJump(context.Background(), repo, CreatePerformedJumpInput{
		PlayerID:       "player_alice",
		Source:         "Taco Bell",
		Destination:    "Olive Garden",
		Food:           "Crunchwrap",
		Caption:        "Direct performed jump",
		MediaObjectKey: "uploads/photo_1",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Jump.Status != "Performed Jump" {
		t.Fatalf("expected Performed Jump status, got %q", result.Jump.Status)
	}
	if result.Evidence.Caption != "Direct performed jump" {
		t.Fatalf("expected evidence caption, got %q", result.Evidence.Caption)
	}
	if insertedParams.GracePeriodExpiresAt.IsZero() {
		t.Fatal("expected grace period expiry to be set")
	}
	expectedExpiry := now.Add(10 * time.Minute).UTC()
	if !insertedParams.GracePeriodExpiresAt.Equal(expectedExpiry) {
		t.Fatalf("expected grace period expiry %v, got %v", expectedExpiry, insertedParams.GracePeriodExpiresAt)
	}
}

func testCreatePerformedJump_PersistenceErrorPropagates(t *testing.T) {
	repo := &mockJumpRepo{
		insertPerformedJumpFn: func(_ context.Context, params InsertPerformedJumpParams) (JumpSnapshot, EvidenceSnapshot, error) {
			return JumpSnapshot{}, EvidenceSnapshot{}, errors.New("db error")
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := CreatePerformedJump(context.Background(), repo, CreatePerformedJumpInput{
		PlayerID:       "player_alice",
		Source:         "A",
		Destination:    "B",
		Food:           "C",
		Caption:        "Test",
		MediaObjectKey: "uploads/photo_1",
	}, now)

	if result.Err == nil || result.Err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", result.Err)
	}
}

func TestJumpPlanning(t *testing.T) {
	t.Run("create performed jump", func(t *testing.T) {
		t.Run("creates performed jump directly", testCreatePerformedJump_CreatesPerformedJumpDirectly)
		t.Run("persistence error propagates", testCreatePerformedJump_PersistenceErrorPropagates)
	})
}
