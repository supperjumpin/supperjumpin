package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockSeasonRepo struct {
	groupMembershipFn        func(ctx context.Context, playerID, groupID string) (MembershipSnapshot, bool, error)
	openSeasonForGroupFn     func(ctx context.Context, groupID string) (SeasonSnapshot, error)
	insertSeasonFn           func(ctx context.Context, groupID, commissionerPlayerID string, submissionDeadline, judgingDeadline time.Time) (SeasonSnapshot, error)
	seasonFn                 func(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	updateSeasonStatusFn     func(ctx context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error
	seasonHistoryEntriesFn   func(ctx context.Context, seasonID string) ([]SeasonHistoryEntry, error)
	jumpsForSeasonFn         func(ctx context.Context, seasonID string) ([]JumpSnapshot, error)
	judgmentsForJumpFn       func(ctx context.Context, jumpID string) ([]Judgment, error)
	updateJumpFinalizationFn func(ctx context.Context, jumpID string, status string, finalScore *int) error
	latestSeasonForGroupFn   func(ctx context.Context, groupID string) (SeasonSnapshot, error)
	groupPlayersFn           func(ctx context.Context, groupID string) ([]PlayerSnapshot, error)
}

func (m *mockSeasonRepo) GroupMembership(ctx context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
	return m.groupMembershipFn(ctx, playerID, groupID)
}

func (m *mockSeasonRepo) OpenSeasonForGroup(ctx context.Context, groupID string) (SeasonSnapshot, error) {
	return m.openSeasonForGroupFn(ctx, groupID)
}

func (m *mockSeasonRepo) InsertSeason(ctx context.Context, groupID, commissionerPlayerID string, submissionDeadline, judgingDeadline time.Time) (SeasonSnapshot, error) {
	return m.insertSeasonFn(ctx, groupID, commissionerPlayerID, submissionDeadline, judgingDeadline)
}

func (m *mockSeasonRepo) Season(ctx context.Context, seasonID string) (SeasonSnapshot, error) {
	return m.seasonFn(ctx, seasonID)
}

func (m *mockSeasonRepo) UpdateSeasonStatus(ctx context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error {
	return m.updateSeasonStatusFn(ctx, seasonID, action, actorPlayerID, actorRole, override, fromStatus, toStatus)
}

func (m *mockSeasonRepo) JumpsForSeason(ctx context.Context, seasonID string) ([]JumpSnapshot, error) {
	return m.jumpsForSeasonFn(ctx, seasonID)
}

func (m *mockSeasonRepo) JudgmentsForJump(ctx context.Context, jumpID string) ([]Judgment, error) {
	return m.judgmentsForJumpFn(ctx, jumpID)
}

func (m *mockSeasonRepo) UpdateJumpFinalization(ctx context.Context, jumpID string, status string, finalScore *int) error {
	return m.updateJumpFinalizationFn(ctx, jumpID, status, finalScore)
}

func (m *mockSeasonRepo) LatestSeasonForGroup(ctx context.Context, groupID string) (SeasonSnapshot, error) {
	return m.latestSeasonForGroupFn(ctx, groupID)
}

func (m *mockSeasonRepo) GroupPlayers(ctx context.Context, groupID string) ([]PlayerSnapshot, error) {
	return m.groupPlayersFn(ctx, groupID)
}

func (m *mockSeasonRepo) SeasonHistoryEntries(ctx context.Context, seasonID string) ([]SeasonHistoryEntry, error) {
	return m.seasonHistoryEntriesFn(ctx, seasonID)
}

func testStartSeason_GroupMemberCanStartSeason(t *testing.T) {
	var insertedGroupID, insertedCommissionerPlayerID string

	repo := &mockSeasonRepo{
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		openSeasonForGroupFn: func(_ context.Context, groupID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{}, nil
		},
		insertSeasonFn: func(_ context.Context, groupID, commissionerPlayerID string, submissionDeadline, judgingDeadline time.Time) (SeasonSnapshot, error) {
			insertedGroupID = groupID
			insertedCommissionerPlayerID = commissionerPlayerID
			return SeasonSnapshot{
				ID:                   "season_1",
				GroupID:              groupID,
				CommissionerPlayerID: commissionerPlayerID,
				Status:               "Active",
				SubmissionDeadline:   submissionDeadline,
				JudgingDeadline:      judgingDeadline,
			}, nil
		},
	}

	submissionDeadline := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	judgingDeadline := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	result := StartSeason(context.Background(), repo, StartSeasonInput{
		GroupID:            "group_1",
		PlayerID:           "player_1",
		SubmissionDeadline: submissionDeadline,
		JudgingDeadline:    judgingDeadline,
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true")
	}
	if result.Season.ID != "season_1" {
		t.Fatalf("expected season_1, got %q", result.Season.ID)
	}
	if result.Season.CommissionerPlayerID != "player_1" {
		t.Fatalf("expected commissioner player_1, got %q", result.Season.CommissionerPlayerID)
	}
	if result.Season.Status != "Active" {
		t.Fatalf("expected Active, got %q", result.Season.Status)
	}
	if insertedGroupID != "group_1" {
		t.Fatalf("expected group_1, got %q", insertedGroupID)
	}
	if insertedCommissionerPlayerID != "player_1" {
		t.Fatalf("expected player_1, got %q", insertedCommissionerPlayerID)
	}
}

func testStartSeason_NonMemberCannotStartSeason(t *testing.T) {
	repo := &mockSeasonRepo{
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{}, false, nil
		},
	}

	result := StartSeason(context.Background(), repo, StartSeasonInput{
		GroupID:  "group_1",
		PlayerID: "player_2",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatalf("expected Allowed=false for non-member")
	}
}

func testStartSeason_ReturnsErrorWhenOpenSeasonExists(t *testing.T) {
	repo := &mockSeasonRepo{
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		openSeasonForGroupFn: func(_ context.Context, groupID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{ID: "existing_season", Status: "Active"}, nil
		},
	}

	result := StartSeason(context.Background(), repo, StartSeasonInput{
		GroupID:  "group_1",
		PlayerID: "player_1",
	})

	if !errors.Is(result.Err, ErrSeasonAlreadyOpen) {
		t.Fatalf("expected ErrSeasonAlreadyOpen, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatalf("expected Allowed=false when season already open")
	}
}

func testCloseSeasonSubmissions_CommissionerCanClose(t *testing.T) {
	var updatedSeasonID, updatedAction, updatedActorPlayerID string
	var updatedOverride bool
	var updatedFromStatus, updatedToStatus string

	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:                   "season_1",
				GroupID:              "group_1",
				CommissionerPlayerID: "player_1",
				Status:               "Active",
			}, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		updateSeasonStatusFn: func(_ context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error {
			updatedSeasonID = seasonID
			updatedAction = action
			updatedActorPlayerID = actorPlayerID
			updatedOverride = override
			updatedFromStatus = fromStatus
			updatedToStatus = toStatus
			return nil
		},
	}

	result := CloseSeasonSubmissions(context.Background(), repo, CloseSeasonSubmissionsInput{
		SeasonID: "season_1",
		PlayerID: "player_1",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true for commissioner")
	}
	if result.Season.Status != "Judging Grace Period" {
		t.Fatalf("expected Judging Grace Period, got %q", result.Season.Status)
	}
	if updatedSeasonID != "season_1" {
		t.Fatalf("expected season_1, got %q", updatedSeasonID)
	}
	if updatedAction != "Submissions Closed" {
		t.Fatalf("expected Submissions Closed, got %q", updatedAction)
	}
	if updatedActorPlayerID != "player_1" {
		t.Fatalf("expected player_1, got %q", updatedActorPlayerID)
	}
	if updatedOverride {
		t.Fatalf("expected override=false for commissioner")
	}
	if updatedFromStatus != "Active" {
		t.Fatalf("expected fromStatus Active, got %q", updatedFromStatus)
	}
	if updatedToStatus != "Judging Grace Period" {
		t.Fatalf("expected toStatus Judging Grace Period, got %q", updatedToStatus)
	}
}

func testCloseSeasonSubmissions_GroupAdminCanCloseWithOverride(t *testing.T) {
	var updatedOverride bool
	var updatedActorPlayerID string

	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:                   "season_1",
				GroupID:              "group_1",
				CommissionerPlayerID: "player_1",
				Status:               "Active",
			}, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Group Admin"}, true, nil
		},
		updateSeasonStatusFn: func(_ context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error {
			updatedOverride = override
			updatedActorPlayerID = actorPlayerID
			return nil
		},
	}

	result := CloseSeasonSubmissions(context.Background(), repo, CloseSeasonSubmissionsInput{
		SeasonID: "season_1",
		PlayerID: "admin_1",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true for Group Admin")
	}
	if !updatedOverride {
		t.Fatalf("expected override=true for Group Admin")
	}
	if updatedActorPlayerID != "admin_1" {
		t.Fatalf("expected admin_1, got %q", updatedActorPlayerID)
	}
}

func testCloseSeasonSubmissions_UnauthorisedPlayerCannotClose(t *testing.T) {
	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:                   "season_1",
				GroupID:              "group_1",
				CommissionerPlayerID: "player_1",
				Status:               "Active",
			}, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
	}

	result := CloseSeasonSubmissions(context.Background(), repo, CloseSeasonSubmissionsInput{
		SeasonID: "season_1",
		PlayerID: "player_2",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatalf("expected Allowed=false for unauthorised player")
	}
}

func testCloseSeasonSubmissions_ReturnsErrorForUnknownSeason(t *testing.T) {
	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{}, nil
		},
	}

	result := CloseSeasonSubmissions(context.Background(), repo, CloseSeasonSubmissionsInput{
		SeasonID: "unknown",
		PlayerID: "player_1",
	})

	if !errors.Is(result.Err, ErrSeasonNotFound) {
		t.Fatalf("expected ErrSeasonNotFound, got %v", result.Err)
	}
}

func testFinalizeSeason_CommissionerCanFinalize(t *testing.T) {
	var updatedSeasonID string
	var jumpsQueried bool

	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:                   "season_1",
				GroupID:              "group_1",
				CommissionerPlayerID: "player_1",
				Status:               "Judging Grace Period",
			}, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		updateSeasonStatusFn: func(_ context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error {
			updatedSeasonID = seasonID
			return nil
		},
		jumpsForSeasonFn: func(_ context.Context, seasonID string) ([]JumpSnapshot, error) {
			jumpsQueried = true
			return nil, nil
		},
		judgmentsForJumpFn: func(_ context.Context, jumpID string) ([]Judgment, error) {
			return nil, nil
		},
		updateJumpFinalizationFn: func(_ context.Context, jumpID string, status string, finalScore *int) error {
			return nil
		},
	}

	result := FinalizeSeason(context.Background(), repo, FinalizeSeasonInput{
		SeasonID: "season_1",
		PlayerID: "player_1",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true for commissioner")
	}
	if result.Season.Status != "Finalized" {
		t.Fatalf("expected Finalized, got %q", result.Season.Status)
	}
	if updatedSeasonID != "season_1" {
		t.Fatalf("expected season_1, got %q", updatedSeasonID)
	}
	if !jumpsQueried {
		t.Fatalf("expected jumps to be queried for finalization")
	}
}

func testFinalizeSeason_CommissionerCanFinalizeFromAnyStatus(t *testing.T) {
	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:                   "season_1",
				GroupID:              "group_1",
				CommissionerPlayerID: "player_1",
				Status:               "Active",
			}, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		updateSeasonStatusFn: func(_ context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error {
			return nil
		},
		jumpsForSeasonFn: func(_ context.Context, seasonID string) ([]JumpSnapshot, error) {
			return nil, nil
		},
		judgmentsForJumpFn: func(_ context.Context, jumpID string) ([]Judgment, error) {
			return nil, nil
		},
		updateJumpFinalizationFn: func(_ context.Context, jumpID string, status string, finalScore *int) error {
			return nil
		},
	}

	result := FinalizeSeason(context.Background(), repo, FinalizeSeasonInput{
		SeasonID: "season_1",
		PlayerID: "player_1",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true for commissioner")
	}
	if result.Season.Status != "Finalized" {
		t.Fatalf("expected Finalized, got %q", result.Season.Status)
	}
}

func testFinalizeSeason_GroupAdminCanFinalizeWithOverride(t *testing.T) {
	var updatedOverride bool

	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:                   "season_1",
				GroupID:              "group_1",
				CommissionerPlayerID: "player_1",
				Status:               "Judging Grace Period",
			}, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Group Admin"}, true, nil
		},
		updateSeasonStatusFn: func(_ context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error {
			updatedOverride = override
			return nil
		},
		jumpsForSeasonFn: func(_ context.Context, seasonID string) ([]JumpSnapshot, error) {
			return nil, nil
		},
		judgmentsForJumpFn: func(_ context.Context, jumpID string) ([]Judgment, error) {
			return nil, nil
		},
		updateJumpFinalizationFn: func(_ context.Context, jumpID string, status string, finalScore *int) error {
			return nil
		},
	}

	result := FinalizeSeason(context.Background(), repo, FinalizeSeasonInput{
		SeasonID: "season_1",
		PlayerID: "admin_1",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true for Group Admin")
	}
	if !updatedOverride {
		t.Fatalf("expected override=true for Group Admin")
	}
}

func testFinalizeSeason_UnauthorisedPlayerCannotFinalize(t *testing.T) {
	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:                   "season_1",
				GroupID:              "group_1",
				CommissionerPlayerID: "player_1",
				Status:               "Active",
			}, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
	}

	result := FinalizeSeason(context.Background(), repo, FinalizeSeasonInput{
		SeasonID: "season_1",
		PlayerID: "player_2",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatalf("expected Allowed=false for unauthorised player")
	}
}

func testFinalizeSeason_ReturnsErrorForUnknownSeason(t *testing.T) {
	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{}, nil
		},
	}

	result := FinalizeSeason(context.Background(), repo, FinalizeSeasonInput{
		SeasonID: "unknown",
		PlayerID: "player_1",
	})

	if !errors.Is(result.Err, ErrSeasonNotFound) {
		t.Fatalf("expected ErrSeasonNotFound, got %v", result.Err)
	}
}

func testSeasonHistory_GroupMemberCanViewHistory(t *testing.T) {
	expectedEntries := []SeasonHistoryEntry{
		{ID: "entry_1", Action: "Submissions Closed", FromStatus: "Active", ToStatus: "Judging Grace Period"},
		{ID: "entry_2", Action: "Season Finalized", FromStatus: "Judging Grace Period", ToStatus: "Finalized"},
	}

	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:      "season_1",
				GroupID: "group_1",
			}, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{Role: "Player"}, true, nil
		},
		seasonHistoryEntriesFn: func(_ context.Context, seasonID string) ([]SeasonHistoryEntry, error) {
			return expectedEntries, nil
		},
	}

	result := SeasonHistory(context.Background(), repo, SeasonHistoryInput{
		SeasonID: "season_1",
		PlayerID: "player_1",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true for group member")
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
	if result.Entries[0].Action != "Submissions Closed" {
		t.Fatalf("expected Submissions Closed, got %q", result.Entries[0].Action)
	}
}

func testSeasonHistory_NonMemberCannotViewHistory(t *testing.T) {
	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{
				ID:      "season_1",
				GroupID: "group_1",
			}, nil
		},
		groupMembershipFn: func(_ context.Context, playerID, groupID string) (MembershipSnapshot, bool, error) {
			return MembershipSnapshot{}, false, nil
		},
	}

	result := SeasonHistory(context.Background(), repo, SeasonHistoryInput{
		SeasonID: "season_1",
		PlayerID: "player_2",
	})

	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Allowed {
		t.Fatalf("expected Allowed=false for non-member")
	}
}

func testSeasonHistory_ReturnsErrorForUnknownSeason(t *testing.T) {
	repo := &mockSeasonRepo{
		seasonFn: func(_ context.Context, seasonID string) (SeasonSnapshot, error) {
			return SeasonSnapshot{}, nil
		},
	}

	result := SeasonHistory(context.Background(), repo, SeasonHistoryInput{
		SeasonID: "unknown",
		PlayerID: "player_1",
	})

	if !errors.Is(result.Err, ErrSeasonNotFound) {
		t.Fatalf("expected ErrSeasonNotFound, got %v", result.Err)
	}
}

func TestSeason(t *testing.T) {
	t.Run("start season", func(t *testing.T) {
		t.Run("group member can start season", testStartSeason_GroupMemberCanStartSeason)
		t.Run("non-member cannot start season", testStartSeason_NonMemberCannotStartSeason)
		t.Run("returns error when open season exists", testStartSeason_ReturnsErrorWhenOpenSeasonExists)
	})

	t.Run("close season submissions", func(t *testing.T) {
		t.Run("commissioner can close", testCloseSeasonSubmissions_CommissionerCanClose)
		t.Run("group admin can close with override", testCloseSeasonSubmissions_GroupAdminCanCloseWithOverride)
		t.Run("unauthorised player cannot close", testCloseSeasonSubmissions_UnauthorisedPlayerCannotClose)
		t.Run("returns error for unknown season", testCloseSeasonSubmissions_ReturnsErrorForUnknownSeason)
	})

	t.Run("finalize season", func(t *testing.T) {
		t.Run("commissioner can finalize", testFinalizeSeason_CommissionerCanFinalize)
		t.Run("commissioner can finalize from any status", testFinalizeSeason_CommissionerCanFinalizeFromAnyStatus)
		t.Run("group admin can finalize with override", testFinalizeSeason_GroupAdminCanFinalizeWithOverride)
		t.Run("unauthorised player cannot finalize", testFinalizeSeason_UnauthorisedPlayerCannotFinalize)
		t.Run("returns error for unknown season", testFinalizeSeason_ReturnsErrorForUnknownSeason)
	})

	t.Run("season history", func(t *testing.T) {
		t.Run("group member can view history", testSeasonHistory_GroupMemberCanViewHistory)
		t.Run("non-member cannot view history", testSeasonHistory_NonMemberCannotViewHistory)
		t.Run("returns error for unknown season", testSeasonHistory_ReturnsErrorForUnknownSeason)
	})
}
