package game

import (
	"context"
	"errors"
	"testing"
)

type mockDisputeRepo struct {
	jumpByIDFn                     func(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	seasonFn                       func(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	groupMembershipFn              func(ctx context.Context, playerID, groupID string) (MembershipSnapshot, bool, error)
	insertDisputeFn                func(ctx context.Context, jumpID, raisedByPlayerID, concern, details string) (DisputeSnapshot, error)
	disputeFn                      func(ctx context.Context, disputeID string) (DisputeSnapshot, error)
	updateDisputeResolutionFn      func(ctx context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error
	updateDisputeOverrideFn        func(ctx context.Context, disputeID, overrideResolution, overrideReason, overrideByPlayerID string) error
	updateJumpStatusAfterDisputeFn func(ctx context.Context, jumpID, status string) error
}

func (m *mockDisputeRepo) JumpByID(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
	return m.jumpByIDFn(nil, jumpID)
}

func (m *mockDisputeRepo) Season(_ context.Context, seasonID string) (SeasonSnapshot, error) {
	return m.seasonFn(nil, seasonID)
}

func (m *mockDisputeRepo) GroupMembership(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
	return m.groupMembershipFn(nil, playerID, groupID)
}

func (m *mockDisputeRepo) InsertDispute(_ context.Context, jumpID, raisedByPlayerID, concern, details string) (DisputeSnapshot, error) {
	return m.insertDisputeFn(nil, jumpID, raisedByPlayerID, concern, details)
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

func (m *mockDisputeRepo) UpdateJumpStatusAfterDispute(_ context.Context, jumpID, status string) error {
	return m.updateJumpStatusAfterDisputeFn(nil, jumpID, status)
}

func testCreateDispute_GroupMemberCanRaiseDisputeOnPerformedJump(t *testing.T) {
	var insertedJumpID, insertedPlayerID string
	repo := &mockDisputeRepo{
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Performed Jump"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		insertDisputeFn: func(_ context.Context, jumpID, raisedByPlayerID, concern, details string) (DisputeSnapshot, error) {
			insertedJumpID = jumpID
			insertedPlayerID = raisedByPlayerID
			return DisputeSnapshot{
				ID:               "dispute_abc123",
				JumpID:           jumpID,
				RaisedByPlayerID: raisedByPlayerID,
				Concern:          concern,
				Details:          details,
				Status:           "Open",
			}, nil
		},
	}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_bob",
		JumpID:   "jump_1",
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
	if insertedJumpID != "jump_1" || insertedPlayerID != "player_bob" {
		t.Fatalf("unexpected insert args: jump=%s, player=%s", insertedJumpID, insertedPlayerID)
	}
}

func testCreateDispute_NonMemberNotAllowed(t *testing.T) {
	repo := &mockDisputeRepo{
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Performed Jump"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{}, false, nil
		},
	}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_stranger",
		JumpID:   "jump_1",
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

func testCreateDispute_InvalidConcernReturnsError(t *testing.T) {
	repo := &mockDisputeRepo{}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_1",
		JumpID:   "jump_1",
		Concern:  "Invalid Concern",
		Details:  "Details.",
	})
	if !errors.Is(result.Err, ErrInvalidDisputeConcern) {
		t.Fatalf("expected ErrInvalidDisputeConcern, got %v", result.Err)
	}
}

func testCreateDispute_RemovedJumpNotDisputable(t *testing.T) {
	repo := &mockDisputeRepo{
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Removed Jump"}, true, nil
		},
	}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_1",
		JumpID:   "jump_1",
		Concern:  "House Rules",
		Details:  "Issues.",
	})
	if !errors.Is(result.Err, ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound for Removed Jump, got %v", result.Err)
	}
}

func testCreateDispute_IdeaNotDisputable(t *testing.T) {
	repo := &mockDisputeRepo{
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Idea"}, true, nil
		},
	}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_1",
		JumpID:   "jump_1",
		Concern:  "House Rules",
		Details:  "Issues.",
	})
	if !errors.Is(result.Err, ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound for Idea, got %v", result.Err)
	}
}

func testCreateDispute_AllValidConcernsAccepted(t *testing.T) {
	for _, concern := range []string{"House Rules", "Credibility", "Source", "Destination", "Food", "duplicate", "other"} {
		t.Run(concern, func(t *testing.T) {
			repo := &mockDisputeRepo{
				jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
					return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Performed Jump"}, true, nil
				},
				groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
					return MembershipSnapshot{Role: "Player"}, true, nil
				},
				insertDisputeFn: func(_ context.Context, jumpID, raisedByPlayerID, concern, details string) (DisputeSnapshot, error) {
					return DisputeSnapshot{ID: "dispute_1", Status: "Open", Concern: concern}, nil
				},
			}

			result := CreateDispute(context.Background(), repo, CreateDisputeInput{
				PlayerID: "player_1",
				JumpID:   "jump_1",
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

func testCreateDispute_JumpNotFound(t *testing.T) {
	repo := &mockDisputeRepo{
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{}, false, nil
		},
	}

	result := CreateDispute(context.Background(), repo, CreateDisputeInput{
		PlayerID: "player_1",
		JumpID:   "jump_1",
		Concern:  "House Rules",
		Details:  "Issues.",
	})
	if !errors.Is(result.Err, ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound, got %v", result.Err)
	}
}

func testResolveDispute_CommissionerCanDisqualifySeasonLinkedJump(t *testing.T) {
	seasonID := "season_1"
	var resolvedResolution string
	var updatedStatus string

	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{
				ID: disputeID, JumpID: "jump_1", Status: "Open",
			}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID: jumpID, GroupID: "group_1", Status: "Judged Jump",
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
		updateJumpStatusAfterDisputeFn: func(_ context.Context, jumpID, status string) error {
			updatedStatus = status
			return nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_commissioner",
		DisputeID:        "dispute_1",
		Resolution:       "Disqualified Jump",
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
	if resolvedResolution != "Disqualified Jump" {
		t.Fatalf("expected Disqualified Jump resolution, got %q", resolvedResolution)
	}
	if result.Jump.Status != "Disqualified Jump" {
		t.Fatalf("expected jump status Disqualified Jump, got %q", result.Jump.Status)
	}
	if result.Jump.FinalScore != nil {
		t.Fatal("expected FinalScore to be nil after disqualification")
	}
	if updatedStatus != "Disqualified Jump" {
		t.Fatalf("expected updated jump status Disqualified Jump, got %q", updatedStatus)
	}
}

func testResolveDispute_CommissionerCannotRemoveSeasonLinkedJump(t *testing.T) {
	seasonID := "season_1"
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, JumpID: "jump_1", Status: "Open"}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Judged Jump", SeasonID: &seasonID}, true, nil
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
		Resolution:       "Removed Jump",
		ResolutionReason: "Privacy issue.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for commissioner trying Removed Jump on season-linked jump")
	}
}

func testResolveDispute_GroupAdminCanRemoveOffSeasonJump(t *testing.T) {
	var updatedStatus string
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, JumpID: "jump_1", Status: "Open"}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Performed Jump"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Group Admin"}, true, nil
		},
		updateDisputeResolutionFn: func(_ context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error {
			return nil
		},
		updateJumpStatusAfterDisputeFn: func(_ context.Context, jumpID, status string) error {
			updatedStatus = status
			return nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_admin",
		DisputeID:        "dispute_1",
		Resolution:       "Removed Jump",
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
	if result.Jump.Status != "Removed Jump" {
		t.Fatalf("expected Removed Jump, got %q", result.Jump.Status)
	}
	if updatedStatus != "Removed Jump" {
		t.Fatalf("expected updated status Removed Jump, got %q", updatedStatus)
	}
}

func testResolveDispute_GroupAdminCannotDisqualifyOffSeasonJump(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, JumpID: "jump_1", Status: "Open"}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Performed Jump"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Group Admin"}, true, nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_admin",
		DisputeID:        "dispute_1",
		Resolution:       "Disqualified Jump",
		ResolutionReason: "House rules.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for Group Admin trying Disqualified Jump on off-season")
	}
}

func testResolveDispute_NonMemberNotAllowed(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, JumpID: "jump_1", Status: "Open"}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Performed Jump"}, true, nil
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

func testResolveDispute_GroupAdminCanOverrideResolvedDispute(t *testing.T) {
	var overriddenResolution string
	seasonID := "season_1"
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{
				ID: disputeID, JumpID: "jump_1", Status: "Resolved",
				Resolution: stringPtr("Disqualified Jump"),
			}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{
				ID: jumpID, GroupID: "group_1", Status: "Disqualified Jump",
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
		updateJumpStatusAfterDisputeFn: func(_ context.Context, jumpID, status string) error {
			return nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_admin",
		DisputeID:        "dispute_1",
		Resolution:       "Removed Jump",
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
	if overriddenResolution != "Removed Jump" {
		t.Fatalf("expected override Removed Jump, got %q", overriddenResolution)
	}
	if result.Jump.Status != "Removed Jump" {
		t.Fatalf("expected Removed Jump, got %q", result.Jump.Status)
	}
}

func testResolveDispute_GroupAdminCannotOverrideWithNoAction(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{
				ID: disputeID, JumpID: "jump_1", Status: "Resolved",
				Resolution: stringPtr("Disqualified Jump"),
			}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Disqualified Jump"}, true, nil
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

func testResolveDispute_InvalidResolutionReturnsError(t *testing.T) {
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

func testResolveDispute_DisputeNotFound(t *testing.T) {
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

func testResolveDispute_CommissionerCanTakeNoActionOnSeasonLinkedJump(t *testing.T) {
	seasonID := "season_1"
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, JumpID: "jump_1", Status: "Open"}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Performed Jump", SeasonID: &seasonID}, true, nil
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
	if result.Jump.Status != "Performed Jump" {
		t.Fatalf("expected jump to remain unchanged, got %q", result.Jump.Status)
	}
}

func testResolveDispute_NonCommissionerPlayerNotAllowed(t *testing.T) {
	seasonID := "season_1"
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, JumpID: "jump_1", Status: "Open"}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Judged Jump", SeasonID: &seasonID}, true, nil
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

func testResolveDispute_PlainPlayerNotAllowedToOverride(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{
				ID: disputeID, JumpID: "jump_1", Status: "Resolved",
				Resolution: stringPtr("Disqualified Jump"),
			}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Disqualified Jump"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_bob",
		DisputeID:        "dispute_1",
		Resolution:       "Removed Jump",
		ResolutionReason: "Should not be allowed.",
	})
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for non-admin override")
	}
}

func testResolveDispute_JumpNotFoundDuringResolve(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, JumpID: "jump_1", Status: "Open"}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{}, false, nil
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_1",
		DisputeID:        "dispute_1",
		Resolution:       "No Action",
		ResolutionReason: "Jump not found.",
	})
	if !errors.Is(result.Err, ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound, got %v", result.Err)
	}
}

func testResolveDispute_PersistenceErrorPropagates(t *testing.T) {
	repo := &mockDisputeRepo{
		disputeFn: func(_ context.Context, disputeID string) (DisputeSnapshot, error) {
			return DisputeSnapshot{ID: disputeID, JumpID: "jump_1", Status: "Open"}, nil
		},
		jumpByIDFn: func(_ context.Context, jumpID string) (JumpSnapshot, bool, error) {
			return JumpSnapshot{ID: jumpID, GroupID: "group_1", Status: "Performed Jump"}, true, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Group Admin"}, true, nil
		},
		updateDisputeResolutionFn: func(_ context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error {
			return errors.New("db error")
		},
	}

	result := ResolveDispute(context.Background(), repo, ResolveDisputeInput{
		PlayerID:         "player_admin",
		DisputeID:        "dispute_1",
		Resolution:       "Removed Jump",
		ResolutionReason: "Error test.",
	})
	if result.Err == nil || result.Err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", result.Err)
	}
}

func TestDispute(t *testing.T) {
	t.Run("create dispute", func(t *testing.T) {
		t.Run("group member can raise dispute on performed jump", testCreateDispute_GroupMemberCanRaiseDisputeOnPerformedJump)
		t.Run("non-member not allowed", testCreateDispute_NonMemberNotAllowed)
		t.Run("invalid concern returns error", testCreateDispute_InvalidConcernReturnsError)
		t.Run("removed jump not disputable", testCreateDispute_RemovedJumpNotDisputable)
		t.Run("idea not disputable", testCreateDispute_IdeaNotDisputable)
		t.Run("all valid concerns accepted", testCreateDispute_AllValidConcernsAccepted)
		t.Run("jump not found", testCreateDispute_JumpNotFound)
	})

	t.Run("resolve dispute", func(t *testing.T) {
		t.Run("commissioner can disqualify season-linked jump", testResolveDispute_CommissionerCanDisqualifySeasonLinkedJump)
		t.Run("commissioner cannot remove season-linked jump", testResolveDispute_CommissionerCannotRemoveSeasonLinkedJump)
		t.Run("group admin can remove off-season jump", testResolveDispute_GroupAdminCanRemoveOffSeasonJump)
		t.Run("group admin cannot disqualify off-season jump", testResolveDispute_GroupAdminCannotDisqualifyOffSeasonJump)
		t.Run("non-member not allowed", testResolveDispute_NonMemberNotAllowed)
		t.Run("group admin can override resolved dispute", testResolveDispute_GroupAdminCanOverrideResolvedDispute)
		t.Run("group admin cannot override with no action", testResolveDispute_GroupAdminCannotOverrideWithNoAction)
		t.Run("invalid resolution returns error", testResolveDispute_InvalidResolutionReturnsError)
		t.Run("dispute not found", testResolveDispute_DisputeNotFound)
		t.Run("commissioner can take no action on season-linked jump", testResolveDispute_CommissionerCanTakeNoActionOnSeasonLinkedJump)
		t.Run("non-commissioner player not allowed", testResolveDispute_NonCommissionerPlayerNotAllowed)
		t.Run("plain player not allowed to override", testResolveDispute_PlainPlayerNotAllowedToOverride)
		t.Run("jump not found during resolve", testResolveDispute_JumpNotFoundDuringResolve)
		t.Run("persistence error propagates", testResolveDispute_PersistenceErrorPropagates)
	})
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
