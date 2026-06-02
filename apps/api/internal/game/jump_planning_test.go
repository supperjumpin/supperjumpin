package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockJumpRepo struct {
	insertIdeaFn          func(ctx context.Context, groupID, playerID, source, destination, food string) (JumpSnapshot, error)
	ideaFn                func(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	groupMembershipFn     func(ctx context.Context, playerID, groupID string) (MembershipSnapshot, bool, error)
	activeSeasonForGroupFn func(ctx context.Context, groupID string) (SeasonSnapshot, error)
	updateJumpToPlannedFn func(ctx context.Context, jumpID, playerID string, seasonID *string) (JumpSnapshot, error)
	insertPerformedJumpFn func(ctx context.Context, params InsertPerformedJumpParams) (JumpSnapshot, EvidenceSnapshot, error)
}

func (m *mockJumpRepo) InsertIdea(_ context.Context, groupID, playerID, source, destination, food string) (JumpSnapshot, error) {
	return m.insertIdeaFn(nil, groupID, playerID, source, destination, food)
}

func (m *mockJumpRepo) Idea(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
	return m.ideaFn(nil, jumpID)
}

func (m *mockJumpRepo) GroupMembership(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
	return m.groupMembershipFn(nil, playerID, groupID)
}

func (m *mockJumpRepo) ActiveSeasonForGroup(_ context.Context, groupID string) (SeasonSnapshot, error) {
	return m.activeSeasonForGroupFn(nil, groupID)
}

func (m *mockJumpRepo) UpdateJumpToPlanned(_ context.Context, jumpID, playerID string, seasonID *string) (JumpSnapshot, error) {
	return m.updateJumpToPlannedFn(nil, jumpID, playerID, seasonID)
}

func (m *mockJumpRepo) InsertPerformedJump(_ context.Context, params InsertPerformedJumpParams) (JumpSnapshot, EvidenceSnapshot, error) {
	return m.insertPerformedJumpFn(nil, params)
}

func testCreateIdea_NonMemberIsRejected(t *testing.T) {
	repo := &mockJumpRepo{
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{}, false, nil
		},
	}

	result := CreateIdea(context.Background(), repo, CreateIdeaInput{
		GroupID:     "group_1",
		PlayerID:    "stranger_1",
		Source:      "Taco Bell",
		Destination: "Olive Garden",
		Food:        "Crunchwrap",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected non-member to be rejected")
	}
}

func testCreateIdea_MemberCanCreateIdea(t *testing.T) {
	var insertedGroupID, insertedPlayerID, insertedSource, insertedDestination, insertedFood string

	repo := &mockJumpRepo{
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		insertIdeaFn: func(_ context.Context, groupID, playerID, source, destination, food string) (JumpSnapshot, error) {
			insertedGroupID = groupID
			insertedPlayerID = playerID
			insertedSource = source
			insertedDestination = destination
			insertedFood = food
			return JumpSnapshot{
				ID:          "jump_1",
				GroupID:     groupID,
				PlayerID:    playerID,
				Source:      source,
				Destination: destination,
				Food:        food,
				Status:      "Idea",
			}, nil
		},
	}

	result := CreateIdea(context.Background(), repo, CreateIdeaInput{
		GroupID:     "group_1",
		PlayerID:    "player_alice",
		Source:      "Taco Bell",
		Destination: "Olive Garden",
		Food:        "Crunchwrap",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected member to be allowed")
	}
	if result.Jump.Status != "Idea" {
		t.Fatalf("expected Idea status, got %q", result.Jump.Status)
	}
	if insertedGroupID != "group_1" || insertedPlayerID != "player_alice" || insertedSource != "Taco Bell" || insertedDestination != "Olive Garden" || insertedFood != "Crunchwrap" {
		t.Fatalf("unexpected insert args: group=%s player=%s source=%s dest=%s food=%s", insertedGroupID, insertedPlayerID, insertedSource, insertedDestination, insertedFood)
	}
}

func testCreateIdea_PersistenceErrorPropagates(t *testing.T) {
	repo := &mockJumpRepo{
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		insertIdeaFn: func(_ context.Context, groupID, playerID, source, destination, food string) (JumpSnapshot, error) {
			return JumpSnapshot{}, errors.New("db error")
		},
	}

	result := CreateIdea(context.Background(), repo, CreateIdeaInput{
		GroupID:  "group_1",
		PlayerID: "player_alice",
	})

	if result.Err == nil || result.Err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", result.Err)
	}
}

func testCreatePlannedJump_NonOwnerIsRejected(t *testing.T) {
	repo := &mockJumpRepo{
		ideaFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", PlayerID: "owner_1", Status: "Idea"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
	}

	result := CreatePlannedJump(context.Background(), repo, CreatePlannedJumpInput{
		IdeaID:   "jump_1",
		PlayerID: "stranger_1",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected non-owner to be rejected")
	}
}

func testCreatePlannedJump_NonMemberIsRejected(t *testing.T) {
	repo := &mockJumpRepo{
		ideaFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", PlayerID: "owner_1", Status: "Idea"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{}, false, nil
		},
	}

	result := CreatePlannedJump(context.Background(), repo, CreatePlannedJumpInput{
		IdeaID:   "jump_1",
		PlayerID: "owner_1",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected non-member to be rejected")
	}
}

func testCreatePlannedJump_OwnerCanPlanJump(t *testing.T) {
	var updatedJumpID, updatedPlayerID string
	var updatedSeasonID *string

	repo := &mockJumpRepo{
		ideaFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", PlayerID: "owner_1", Status: "Idea"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		activeSeasonForGroupFn: func(_ context.Context, groupID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{}, nil // no active season
		},
		updateJumpToPlannedFn: func(_ context.Context, jumpID, playerID string, seasonID *string) (JumpSnapshot, error) {
			updatedJumpID = jumpID
			updatedPlayerID = playerID
			updatedSeasonID = seasonID
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", PlayerID: playerID, Status: "Planned Jump"}, nil
		},
	}

	result := CreatePlannedJump(context.Background(), repo, CreatePlannedJumpInput{
		IdeaID:   "jump_1",
		PlayerID: "owner_1",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected owner to be allowed")
	}
	if result.Jump.Status != "Planned Jump" {
		t.Fatalf("expected Planned Jump status, got %q", result.Jump.Status)
	}
	if updatedJumpID != "jump_1" || updatedPlayerID != "owner_1" {
		t.Fatalf("unexpected update args: jump=%s player=%s", updatedJumpID, updatedPlayerID)
	}
	if updatedSeasonID != nil {
		t.Fatal("expected nil seasonID when no active season")
	}
}

func testCreatePlannedJump_JumpNotFoundReturnsError(t *testing.T) {
	repo := &mockJumpRepo{
		ideaFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{}, false, nil
		},
	}

	result := CreatePlannedJump(context.Background(), repo, CreatePlannedJumpInput{
		IdeaID:   "jump_1",
		PlayerID: "owner_1",
	})

	if !errors.Is(result.Err, ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound, got %v", result.Err)
	}
}

func testCreatePlannedJump_WrongStatusReturnsError(t *testing.T) {
	repo := &mockJumpRepo{
		ideaFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", PlayerID: "owner_1", Status: "Planned Jump"}, true, nil
		},
	}

	result := CreatePlannedJump(context.Background(), repo, CreatePlannedJumpInput{
		IdeaID:   "jump_1",
		PlayerID: "owner_1",
	})

	if !errors.Is(result.Err, ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound for wrong status, got %v", result.Err)
	}
}

func testCreatePlannedJump_LinksToActiveSeason(t *testing.T) {
	seasonID := "season_1"
	var updatedSeasonID *string

	repo := &mockJumpRepo{
		ideaFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", PlayerID: "owner_1", Status: "Idea"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		activeSeasonForGroupFn: func(_ context.Context, groupID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{ID: seasonID, Status: "Active"}, nil
		},
		updateJumpToPlannedFn: func(_ context.Context, jumpID, playerID string, seasonID *string) (JumpSnapshot, error) {
			updatedSeasonID = seasonID
			return JumpSnapshot{ID: jumpID, Status: "Planned Jump", SeasonID: seasonID}, nil
		},
	}

	result := CreatePlannedJump(context.Background(), repo, CreatePlannedJumpInput{
		IdeaID:   "jump_1",
		PlayerID: "owner_1",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed")
	}
	if updatedSeasonID == nil || *updatedSeasonID != "season_1" {
		t.Fatalf("expected season_1 linked, got %v", updatedSeasonID)
	}
}

func testCreatePlannedJump_OffSeasonSkipsSeasonLinking(t *testing.T) {
	var updatedSeasonID *string

	repo := &mockJumpRepo{
		ideaFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", PlayerID: "owner_1", Status: "Idea"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		updateJumpToPlannedFn: func(_ context.Context, jumpID, playerID string, seasonID *string) (JumpSnapshot, error) {
			updatedSeasonID = seasonID
			return JumpSnapshot{ID: jumpID, Status: "Planned Jump"}, nil
		},
	}

	result := CreatePlannedJump(context.Background(), repo, CreatePlannedJumpInput{
		IdeaID:    "jump_1",
		PlayerID:  "owner_1",
		OffSeason: true,
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed")
	}
	if updatedSeasonID != nil {
		t.Fatal("expected nil seasonID when off-season")
	}
}

func testCreatePlannedJump_PersistenceErrorPropagates(t *testing.T) {
	repo := &mockJumpRepo{
		ideaFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", PlayerID: "owner_1", Status: "Idea"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		activeSeasonForGroupFn: func(_ context.Context, groupID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{}, nil
		},
		updateJumpToPlannedFn: func(_ context.Context, jumpID, playerID string, seasonID *string) (JumpSnapshot, error) {
			return JumpSnapshot{}, errors.New("db error")
		},
	}

	result := CreatePlannedJump(context.Background(), repo, CreatePlannedJumpInput{
		IdeaID:   "jump_1",
		PlayerID: "owner_1",
	})

	if result.Err == nil || result.Err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", result.Err)
	}
}

func testCreatePerformedJump_CreatesPerformedJumpDirectly(t *testing.T) {
	var insertedParams InsertPerformedJumpParams

	repo := &mockJumpRepo{
		insertPerformedJumpFn: func(_ context.Context, params InsertPerformedJumpParams) (JumpSnapshot, EvidenceSnapshot, error) {
			insertedParams = params
			return JumpSnapshot{
				ID:       "jump_1",
				PlayerID: params.PlayerID,
				Status:   "Performed Jump",
				Source:   params.Source,
				Destination: params.Destination,
				Food:     params.Food,
			}, EvidenceSnapshot{
				ID:             "evidence_1",
				JumpID:        "jump_1",
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
		GroupID:        "",
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
	if insertedParams.GroupID != "" {
		t.Fatalf("expected empty groupID, got %q", insertedParams.GroupID)
	}
	if insertedParams.SeasonID != nil {
		t.Fatalf("expected nil seasonID for ungrouped jump, got %v", insertedParams.SeasonID)
	}
	expectedExpiry := now.Add(10 * time.Minute).UTC()
	if !insertedParams.GracePeriodExpiresAt.Equal(expectedExpiry) {
		t.Fatalf("expected grace period expiry %v, got %v", expectedExpiry, insertedParams.GracePeriodExpiresAt)
	}
}

func testCreatePerformedJump_LinksToActiveSeasonWhenGroupProvided(t *testing.T) {
	var insertedParams InsertPerformedJumpParams

	repo := &mockJumpRepo{
		activeSeasonForGroupFn: func(_ context.Context, groupID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{ID: "season_1", Status: "Active"}, nil
		},
		insertPerformedJumpFn: func(_ context.Context, params InsertPerformedJumpParams) (JumpSnapshot, EvidenceSnapshot, error) {
			insertedParams = params
			return JumpSnapshot{ID: "jump_1", Status: "Performed Jump", SeasonID: params.SeasonID}, EvidenceSnapshot{ID: "evidence_1"}, nil
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := CreatePerformedJump(context.Background(), repo, CreatePerformedJumpInput{
		PlayerID:       "player_alice",
		Source:         "Taco Bell",
		Destination:    "Olive Garden",
		Food:           "Crunchwrap",
		Caption:        "Group jump",
		MediaObjectKey: "uploads/photo_1",
		GroupID:        "group_1",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if insertedParams.SeasonID == nil || *insertedParams.SeasonID != "season_1" {
		t.Fatalf("expected season_1 linked, got %v", insertedParams.SeasonID)
	}
}

func testCreatePerformedJump_NoActiveSeasonWhenGroupProvided(t *testing.T) {
	var insertedParams InsertPerformedJumpParams

	repo := &mockJumpRepo{
		activeSeasonForGroupFn: func(_ context.Context, groupID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{}, nil // no active season
		},
		insertPerformedJumpFn: func(_ context.Context, params InsertPerformedJumpParams) (JumpSnapshot, EvidenceSnapshot, error) {
			insertedParams = params
			return JumpSnapshot{ID: "jump_1", Status: "Performed Jump"}, EvidenceSnapshot{ID: "evidence_1"}, nil
		},
	}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	result := CreatePerformedJump(context.Background(), repo, CreatePerformedJumpInput{
		PlayerID:       "player_alice",
		GroupID:        "group_1",
		Source:         "A",
		Destination:    "B",
		Food:           "C",
		Caption:        "Test",
		MediaObjectKey: "uploads/photo_1",
	}, now)

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if insertedParams.SeasonID != nil {
		t.Fatalf("expected nil seasonID when no active season, got %v", insertedParams.SeasonID)
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
	t.Run("create idea", func(t *testing.T) {
		t.Run("non-member is rejected", testCreateIdea_NonMemberIsRejected)
		t.Run("member can create idea", testCreateIdea_MemberCanCreateIdea)
		t.Run("persistence error propagates", testCreateIdea_PersistenceErrorPropagates)
	})

	t.Run("create planned jump", func(t *testing.T) {
		t.Run("non-owner is rejected", testCreatePlannedJump_NonOwnerIsRejected)
		t.Run("non-member is rejected", testCreatePlannedJump_NonMemberIsRejected)
		t.Run("owner can plan jump", testCreatePlannedJump_OwnerCanPlanJump)
		t.Run("jump not found returns error", testCreatePlannedJump_JumpNotFoundReturnsError)
		t.Run("wrong status returns error", testCreatePlannedJump_WrongStatusReturnsError)
		t.Run("links to active season", testCreatePlannedJump_LinksToActiveSeason)
		t.Run("off-season skips season linking", testCreatePlannedJump_OffSeasonSkipsSeasonLinking)
		t.Run("persistence error propagates", testCreatePlannedJump_PersistenceErrorPropagates)
	})

	t.Run("create performed jump", func(t *testing.T) {
		t.Run("creates performed jump directly", testCreatePerformedJump_CreatesPerformedJumpDirectly)
		t.Run("links to active season when group provided", testCreatePerformedJump_LinksToActiveSeasonWhenGroupProvided)
		t.Run("no active season when group provided", testCreatePerformedJump_NoActiveSeasonWhenGroupProvided)
		t.Run("persistence error propagates", testCreatePerformedJump_PersistenceErrorPropagates)
	})
}
