package game

import (
	"context"
	"errors"
	"testing"
)

type mockDisputeRepo struct {
	stuntByIDFn                func(ctx context.Context, stuntID string) (StuntSnapshot, bool, error)
	seasonFn                   func(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	groupMembershipFn          func(ctx context.Context, playerID, groupID string) (MembershipSnapshot, bool, error)
	insertDisputeFn            func(ctx context.Context, stuntID, raisedByPlayerID, concern, details string) (DisputeSnapshot, error)
	disputeFn                  func(ctx context.Context, disputeID string) (DisputeSnapshot, error)
	updateDisputeResolutionFn  func(ctx context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error
	updateDisputeOverrideFn    func(ctx context.Context, disputeID, overrideResolution, overrideReason, overrideByPlayerID string) error
	updateStuntStatusAfterDisputeFn func(ctx context.Context, stuntID, status string) error
}

func (m *mockDisputeRepo) StuntByID(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
	return m.stuntByIDFn(nil, stuntID)
}

func (m *mockDisputeRepo) Season(_ context.Context, seasonID string) (SeasonSnapshot, error) {
	return m.seasonFn(nil, seasonID)
}

func (m *mockDisputeRepo) GroupMembership(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
	return m.groupMembershipFn(nil, playerID, groupID)
}

func (m *mockDisputeRepo) InsertDispute(_ context.Context, stuntID, raisedByPlayerID, concern, details string) (DisputeSnapshot, error) {
	return m.insertDisputeFn(nil, stuntID, raisedByPlayerID, concern, details)
}

func (m *mockDisputeRepo) Dispute(_ context.Context, disputeID string) (DisputeSnapshot, error) {
	return m.disputeFn(nil, disputeID)
}

func (m *mockDisputeRepo) UpdateDisputeResolution(_ context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error {
	return m.updateDisputeResolutionFn(nil, disputeID, resolution, resolutionReason, resolvedByPlayerID)
}

func (m *mockDisputeRepo) UpdateDisputeOverride(_ context.Context, disputeID, overrideResolution, overrideReason, overrideByPlayerID string) error {
	return m.updateDisputeOverrideFn(nil, disputeID, overrideResolution, overrideReason, overrideByPlayerID)
}

func (m *mockDisputeRepo) UpdateStuntStatusAfterDispute(_ context.Context, stuntID, status string) error {
	return m.updateStuntStatusAfterDisputeFn(nil, stuntID, status)
}

func TestCreateDispute_GroupMemberCanRaiseDisputeOnPerformedStunt(t *testing.T) {
	var insertedStuntID, insertedPlayerID string
	repo := &mockDisputeRepo{
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Performed Stunt"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		insertDisputeFn: func(_ context.Context, stuntID, raisedByPlayerID, concern, details string) (DisputeSnapshot, error) {
			insertedStuntID = stuntID
			insertedPlayerID = raisedByPlayerID
			return DisputeSnapshot{
				ID:               "dispute_abc123",
				StuntID:          stuntID,
				RaisedByPlayerID: raisedByPlayerID,
				Concern:          concern,
				Details:          details,
				Status:           "Open",
			}, nil
		},
	}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_bob",
		StuntID:  "stunt_1",
		Concern:  "House Rules",
		Details:  "This blocked the emergency exit.",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected Allowed=true")
	}
	if result.Dispute.ID == "" {
		t.Fatal("expected created dispute to have ID")
	}
	if result.Dispute.Status != "Open" {
		t.Fatalf("expected Open dispute, got %q", result.Dispute.Status)
	}
	if result.Dispute.Concern != "House Rules" {
		t.Fatalf("expected House Rules concern, got %q", result.Dispute.Concern)
	}
	if insertedStuntID != "stunt_1" || insertedPlayerID != "player_bob" {
		t.Fatalf("unexpected insert args: stunt=%s, player=%s", insertedStuntID, insertedPlayerID)
	}
}

func TestCreateDispute_NonMemberNotAllowed(t *testing.T) {
	repo := &mockDisputeRepo{
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Performed Stunt"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{}, false, nil
		},
	}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_stranger",
		StuntID:  "stunt_1",
		Concern:  "House Rules",
		Details:  "Issues.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for non-member")
	}
}

func TestCreateDispute_InvalidConcernReturnsError(t *testing.T) {
	repo := &mockDisputeRepo{}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_1",
		StuntID:  "stunt_1",
		Concern:  "Invalid Concern",
		Details:  "Details.",
	})
	if !errors.Is(result.Err, ErrInvalidDisputeConcern) {
		t.Fatalf("expected ErrInvalidDisputeConcern, got %v", result.Err)
	}
}

func TestCreateDispute_RemovedStuntNotDisputable(t *testing.T) {
	repo := &mockDisputeRepo{
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Removed Stunt"}, true, nil
		},
	}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_1",
		StuntID:  "stunt_1",
		Concern:  "House Rules",
		Details:  "Issues.",
	})
	if !errors.Is(result.Err, ErrStuntNotFound) {
		t.Fatalf("expected ErrStuntNotFound for Removed Stunt, got %v", result.Err)
	}
}

func TestCreateDispute_IdeaNotDisputable(t *testing.T) {
	repo := &mockDisputeRepo{
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Idea"}, true, nil
		},
	}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_1",
		StuntID:  "stunt_1",
		Concern:  "House Rules",
		Details:  "Issues.",
	})
	if !errors.Is(result.Err, ErrStuntNotFound) {
		t.Fatalf("expected ErrStuntNotFound for Idea, got %v", result.Err)
	}
}

func TestCreateDispute_AllValidConcernsAccepted(t *testing.T) {
	for _, concern := range []string{"House Rules", "Credibility", "Source", "Destination", "Food", "duplicate", "other"} {
		t.Run(concern, func(t *testing.T) {
			repo := &mockDisputeRepo{
				stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
					return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Performed Stunt"}, true, nil
				},
				groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
					return MembershipSnapshot{Role: "Player"}, true, nil
				},
				insertDisputeFn: func(_ context.Context, stuntID, raisedByPlayerID, concern, details string) (DisputeSnapshot, error) {
					return DisputeSnapshot{ID: "dispute_1", Status: "Open", Concern: concern}, nil
				},
			}

			result := CreateDispute(context.Background(), repo, CreateDisputeInput{
				PlayerID: "player_1",
				StuntID:  "stunt_1",
				Concern:  concern,
				Details:  "Test.",
			})
			if result.Err != nil {
				t.Fatalf("unexpected error for concern %q: %v", concern, result.Err)
			}
			if !result.Allowed {
				t.Fatalf("expected Allowed=true for concern %q", concern)
			}
		})
	}
}

func TestCreateDispute_StuntNotFound(t *testing.T) {
	repo := &mockDisputeRepo{
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{}, false, nil
		},
	}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_1",
		StuntID:  "stunt_1",
		Concern:  "House Rules",
		Details:  "Issues.",
	})
	if !errors.Is(result.Err, ErrStuntNotFound) {
		t.Fatalf("expected ErrStuntNotFound, got %v", result.Err)
	}
}

func TestResolveDispute_CommissionerCanDisqualifySeasonLinkedStunt(t *testing.T) {
	seasonID := "season_1"
	var resolvedResolution string
	var updatedStatus string

	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{
				ID: disputeID, StuntID: "stunt_1", Status: "Open",
			}, nil
		},
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{
				ID: stuntID, GroupID: "group_1", Status: "Judged Stunt",
				SeasonID: &seasonID, FinalScore: intPtr(35),
			}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		seasonFn: func(_ context.Context, sid string) (SeasonSnapshot, error) {
			return SeasonSnapshot{ID: sid, CommissionerPlayerID: "player_commissioner"}, nil
		},
		updateDisputeResolutionFn: func(_ context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error {
			resolvedResolution = resolution
			return nil
		},
		updateStuntStatusAfterDisputeFn: func(_ context.Context, stuntID, status string) error {
			updatedStatus = status
			return nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_commissioner",
		DisputeID:        "dispute_1",
		Resolution:       "Disqualified Stunt",
		ResolutionReason: "Evidence does not support the claim.",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected Allowed=true")
	}
	if result.Dispute.Status != "Resolved" {
		t.Fatalf("expected Resolved status, got %q", result.Dispute.Status)
	}
	if resolvedResolution != "Disqualified Stunt" {
		t.Fatalf("expected Disqualified Stunt resolution, got %q", resolvedResolution)
	}
	if result.Stunt.Status != "Disqualified Stunt" {
		t.Fatalf("expected stunt status Disqualified Stunt, got %q", result.Stunt.Status)
	}
	if result.Stunt.FinalScore != nil {
		t.Fatal("expected FinalScore to be nil after disqualification")
	}
	if updatedStatus != "Disqualified Stunt" {
		t.Fatalf("expected updated stunt status Disqualified Stunt, got %q", updatedStatus)
	}
}

func TestResolveDispute_CommissionerCannotRemoveSeasonLinkedStunt(t *testing.T) {
	seasonID := "season_1"
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, StuntID: "stunt_1", Status: "Open"}, nil
		},
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Judged Stunt", SeasonID: &seasonID}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		seasonFn: func(_ context.Context, sid string) (SeasonSnapshot, error) {
			return SeasonSnapshot{ID: sid, CommissionerPlayerID: "player_commissioner"}, nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_commissioner",
		DisputeID:        "dispute_1",
		Resolution:       "Removed Stunt",
		ResolutionReason: "Privacy issue.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for commissioner trying Removed Stunt on season-linked stunt")
	}
}

func TestResolveDispute_GroupAdminCanRemoveOffSeasonStunt(t *testing.T) {
	var updatedStatus string
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, StuntID: "stunt_1", Status: "Open"}, nil
		},
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Performed Stunt"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Group Admin"}, true, nil
		},
		updateDisputeResolutionFn: func(_ context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error {
			return nil
		},
		updateStuntStatusAfterDisputeFn: func(_ context.Context, stuntID, status string) error {
			updatedStatus = status
			return nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_admin",
		DisputeID:        "dispute_1",
		Resolution:       "Removed Stunt",
		ResolutionReason: "Serious privacy issue.",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected Allowed=true")
	}
	if result.Dispute.Status != "Resolved" {
		t.Fatalf("expected Resolved status, got %q", result.Dispute.Status)
	}
	if result.Stunt.Status != "Removed Stunt" {
		t.Fatalf("expected Removed Stunt, got %q", result.Stunt.Status)
	}
	if updatedStatus != "Removed Stunt" {
		t.Fatalf("expected updated status Removed Stunt, got %q", updatedStatus)
	}
}

func TestResolveDispute_GroupAdminCannotDisqualifyOffSeasonStunt(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, StuntID: "stunt_1", Status: "Open"}, nil
		},
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Performed Stunt"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Group Admin"}, true, nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_admin",
		DisputeID:        "dispute_1",
		Resolution:       "Disqualified Stunt",
		ResolutionReason: "House rules.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for Group Admin trying Disqualified Stunt on off-season")
	}
}

func TestResolveDispute_NonMemberNotAllowed(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, StuntID: "stunt_1", Status: "Open"}, nil
		},
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Performed Stunt"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{}, false, nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_stranger",
		DisputeID:        "dispute_1",
		Resolution:       "No Action",
		ResolutionReason: "Not applicable.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for non-member")
	}
}

func TestResolveDispute_GroupAdminCanOverrideResolvedDispute(t *testing.T) {
	var overriddenResolution string
	seasonID := "season_1"
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{
				ID: disputeID, StuntID: "stunt_1", Status: "Resolved",
				Resolution: stringPtr("Disqualified Stunt"),
			}, nil
		},
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{
				ID: stuntID, GroupID: "group_1", Status: "Disqualified Stunt",
				SeasonID: &seasonID, FinalScore: nil,
			}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Group Admin"}, true, nil
		},
		updateDisputeOverrideFn: func(_ context.Context, disputeID, overrideResolution, overrideReason, overrideByPlayerID string) error {
			overriddenResolution = overrideResolution
			return nil
		},
		updateStuntStatusAfterDisputeFn: func(_ context.Context, stuntID, status string) error {
			return nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_admin",
		DisputeID:        "dispute_1",
		Resolution:       "Removed Stunt",
		ResolutionReason: "Serious privacy violation.",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected Allowed=true")
	}
	if result.Dispute.Status != "Overridden" {
		t.Fatalf("expected Overridden status, got %q", result.Dispute.Status)
	}
	if overriddenResolution != "Removed Stunt" {
		t.Fatalf("expected override Removed Stunt, got %q", overriddenResolution)
	}
	if result.Stunt.Status != "Removed Stunt" {
		t.Fatalf("expected Removed Stunt, got %q", result.Stunt.Status)
	}
}

func TestResolveDispute_GroupAdminCannotOverrideWithNoAction(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{
				ID: disputeID, StuntID: "stunt_1", Status: "Resolved",
				Resolution: stringPtr("Disqualified Stunt"),
			}, nil
		},
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Disqualified Stunt"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Group Admin"}, true, nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_admin",
		DisputeID:        "dispute_1",
		Resolution:       "No Action",
		ResolutionReason: "Never mind.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for override with No Action")
	}
}

func TestResolveDispute_InvalidResolutionReturnsError(t *testing.T) {
	repo := &mockDisputeRepo{}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_1",
		DisputeID:        "dispute_1",
		Resolution:       "Invalid",
		ResolutionReason: "Bad.",
	})
	if !errors.Is(result.Err, ErrInvalidDisputeResolution) {
		t.Fatalf("expected ErrInvalidDisputeResolution, got %v", result.Err)
	}
}

func TestResolveDispute_DisputeNotFound(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{}, nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_1",
		DisputeID:        "dispute_1",
		Resolution:       "No Action",
		ResolutionReason: "Not found.",
	})
	if !errors.Is(result.Err, ErrDisputeNotFound) {
		t.Fatalf("expected ErrDisputeNotFound, got %v", result.Err)
	}
}

func TestResolveDispute_CommissionerCanTakeNoActionOnSeasonLinkedStunt(t *testing.T) {
	seasonID := "season_1"
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, StuntID: "stunt_1", Status: "Open"}, nil
		},
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Performed Stunt", SeasonID: &seasonID}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		seasonFn: func(_ context.Context, sid string) (SeasonSnapshot, error) {
			return SeasonSnapshot{ID: sid, CommissionerPlayerID: "player_commissioner"}, nil
		},
		updateDisputeResolutionFn: func(_ context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error {
			return nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_commissioner",
		DisputeID:        "dispute_1",
		Resolution:       "No Action",
		ResolutionReason: "Evidence checks out.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatal("expected Allowed=true")
	}
	if result.Dispute.Status != "Resolved" {
		t.Fatalf("expected Resolved, got %q", result.Dispute.Status)
	}
	if result.Stunt.Status != "Performed Stunt" {
		t.Fatalf("expected stunt to remain unchanged, got %q", result.Stunt.Status)
	}
}

func TestResolveDispute_NonCommissionerPlayerNotAllowed(t *testing.T) {
	seasonID := "season_1"
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, StuntID: "stunt_1", Status: "Open"}, nil
		},
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Judged Stunt", SeasonID: &seasonID}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		seasonFn: func(_ context.Context, sid string) (SeasonSnapshot, error) {
			return SeasonSnapshot{ID: sid, CommissionerPlayerID: "player_commissioner"}, nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_bob",
		DisputeID:        "dispute_1",
		Resolution:       "No Action",
		ResolutionReason: "Looks fine.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for non-commissioner")
	}
}

func TestResolveDispute_PlainPlayerNotAllowedToOverride(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{
				ID: disputeID, StuntID: "stunt_1", Status: "Resolved",
				Resolution: stringPtr("Disqualified Stunt"),
			}, nil
		},
		stuntByIDFn: func(_ context.Context, stuntID string) (StuntSnapshot, bool, error) {
			return StuntSnapshot{ID: stuntID, GroupID: "group_1", Status: "Disqualified Stunt"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_bob",
		DisputeID:        "dispute_1",
		Resolution:       "Removed Stunt",
		ResolutionReason: "Should not be allowed.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for non-admin override")
	}
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
