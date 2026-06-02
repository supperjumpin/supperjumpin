package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockGroupRepo struct {
	insertGroupFn          func(ctx context.Context, groupID, name string) error
	insertMembershipFn     func(ctx context.Context, groupID, playerID, role string) error
	groupFn                func(ctx context.Context, groupID string) (GroupSnapshot, bool, error)
	membershipFn           func(ctx context.Context, playerID, groupID string) (GroupMembershipSnapshot, bool, error)
	membershipsForPlayerFn func(ctx context.Context, playerID string) ([]MembershipWithGroupSnapshot, error)
	insertInviteFn         func(ctx context.Context, groupID, createdByPlayerID string, expiresAt int64) (InviteSnapshot, error)
	inviteByTokenFn        func(ctx context.Context, token string) (InviteSnapshot, bool, error)
	markInviteUsedFn       func(ctx context.Context, token, playerID string) error
}

func (m *mockGroupRepo) InsertGroup(_ context.Context, groupID, name string) error {
	return m.insertGroupFn(nil, groupID, name)
}

func (m *mockGroupRepo) InsertMembership(_ context.Context, groupID, playerID, role string) error {
	return m.insertMembershipFn(nil, groupID, playerID, role)
}

func (m *mockGroupRepo) Group(_ context.Context, groupID string) (GroupSnapshot, bool, error) {
	return m.groupFn(nil, groupID)
}

func (m *mockGroupRepo) Membership(_ context.Context, playerID, groupID string) (GroupMembershipSnapshot, bool, error) {
	return m.membershipFn(nil, playerID, groupID)
}

func (m *mockGroupRepo) MembershipsForPlayer(_ context.Context, playerID string) ([]MembershipWithGroupSnapshot, error) {
	return m.membershipsForPlayerFn(nil, playerID)
}

func (m *mockGroupRepo) InsertInvite(_ context.Context, groupID, createdByPlayerID string, expiresAt int64) (InviteSnapshot, error) {
	return m.insertInviteFn(nil, groupID, createdByPlayerID, expiresAt)
}

func (m *mockGroupRepo) InviteByToken(_ context.Context, token string) (InviteSnapshot, bool, error) {
	return m.inviteByTokenFn(nil, token)
}

func (m *mockGroupRepo) MarkInviteUsed(_ context.Context, token, playerID string) error {
	return m.markInviteUsedFn(nil, token, playerID)
}

func testCreateInvite_NonMemberIsRejected(t *testing.T) {
	repo := &mockGroupRepo{
		membershipFn: func(_ context.Context, playerID, groupID string) (GroupMembershipSnapshot, bool, error) {
			return GroupMembershipSnapshot{}, false, nil
		},
	}

	result := CreateInvite(context.Background(), repo, CreateInviteInput{
		GroupID:  "group_1",
		PlayerID: "bob",
		Now:      time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	})

	if result.Allowed {
		t.Fatal("expected non-member to be rejected")
	}
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
}

func testCreateInvite_MemberCanCreateInvite(t *testing.T) {
	var insertedGroupID, insertedCreatedBy string
	var insertedExpiresAt int64

	repo := &mockGroupRepo{
		membershipFn: func(_ context.Context, playerID, groupID string) (GroupMembershipSnapshot, bool, error) {
			return GroupMembershipSnapshot{Role: "Group Admin"}, true, nil
		},
		insertInviteFn: func(_ context.Context, groupID, createdByPlayerID string, expiresAt int64) (InviteSnapshot, error) {
			insertedGroupID = groupID
			insertedCreatedBy = createdByPlayerID
			insertedExpiresAt = expiresAt
			return InviteSnapshot{
				ID:        "invite_1",
				GroupID:   groupID,
				Token:     "invite_token_abc123",
				CreatedBy: createdByPlayerID,
				ExpiresAt: expiresAt,
			}, nil
		},
	}

	result := CreateInvite(context.Background(), repo, CreateInviteInput{
		GroupID:  "group_1",
		PlayerID: "alice",
		Now:      time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	})

	if !result.Allowed {
		t.Fatal("expected member to be allowed to create invite")
	}
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Invite.GroupID != "group_1" || result.Invite.CreatedBy != "alice" {
		t.Fatalf("expected invite for group_1 by alice, got %#v", result.Invite)
	}
	if insertedGroupID != "group_1" || insertedCreatedBy != "alice" {
		t.Fatalf("expected InsertInvite called with group_1/alice, got %q/%q", insertedGroupID, insertedCreatedBy)
	}
	expectedExpiry := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC).Unix()
	if insertedExpiresAt != expectedExpiry {
		t.Fatalf("expected invite expiry 7 days later (%d), got %d", expectedExpiry, insertedExpiresAt)
	}
}

func testAcceptInvite_InvalidTokenReturnsInvalid(t *testing.T) {
	repo := &mockGroupRepo{
		inviteByTokenFn: func(_ context.Context, token string) (InviteSnapshot, bool, error) {
			return InviteSnapshot{}, false, nil
		},
	}

	result := AcceptInvite(context.Background(), repo, AcceptInviteInput{
		Token:    "bad-token",
		PlayerID: "bob",
		Now:      time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	})

	if result.Status != InviteInvalid {
		t.Fatalf("expected InviteInvalid, got %q", result.Status)
	}
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
}

func testAcceptInvite_UsedTokenReturnsUsed(t *testing.T) {
	usedBy := "alice"
	repo := &mockGroupRepo{
		inviteByTokenFn: func(_ context.Context, token string) (InviteSnapshot, bool, error) {
			return InviteSnapshot{
				ID:        "invite_1",
				GroupID:   "group_1",
				Token:     token,
				CreatedBy: "alice",
				ExpiresAt: time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC).Unix(),
				UsedBy:    &usedBy,
			}, true, nil
		},
	}

	result := AcceptInvite(context.Background(), repo, AcceptInviteInput{
		Token:    "used-token",
		PlayerID: "bob",
		Now:      time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	})

	if result.Status != InviteUsed {
		t.Fatalf("expected InviteUsed, got %q", result.Status)
	}
}

func testAcceptInvite_ExpiredTokenReturnsExpired(t *testing.T) {
	repo := &mockGroupRepo{
		inviteByTokenFn: func(_ context.Context, token string) (InviteSnapshot, bool, error) {
			return InviteSnapshot{
				ID:        "invite_1",
				GroupID:   "group_1",
				Token:     token,
				CreatedBy: "alice",
				ExpiresAt: time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC).Unix(),
			}, true, nil
		},
	}

	result := AcceptInvite(context.Background(), repo, AcceptInviteInput{
		Token:    "expired-token",
		PlayerID: "bob",
		Now:      time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	})

	if result.Status != InviteExpired {
		t.Fatalf("expected InviteExpired, got %q", result.Status)
	}
}

func testAcceptInvite_ExistingMemberReturnsMember(t *testing.T) {
	repo := &mockGroupRepo{
		inviteByTokenFn: func(_ context.Context, token string) (InviteSnapshot, bool, error) {
			return InviteSnapshot{
				ID:        "invite_1",
				GroupID:   "group_1",
				Token:     token,
				CreatedBy: "alice",
				ExpiresAt: time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC).Unix(),
			}, true, nil
		},
		membershipFn: func(_ context.Context, playerID, groupID string) (GroupMembershipSnapshot, bool, error) {
			return GroupMembershipSnapshot{Role: "Group Admin"}, true, nil
		},
	}

	result := AcceptInvite(context.Background(), repo, AcceptInviteInput{
		Token:    "valid-token",
		PlayerID: "bob",
		Now:      time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	})

	if result.Status != InviteMember {
		t.Fatalf("expected InviteMember, got %q", result.Status)
	}
}

func testAcceptInvite_ValidTokenCreatesMembershipAndReturnsAccepted(t *testing.T) {
	var markedToken, markedPlayerID string
	var insertedGroupID, insertedPlayerID, insertedRole string

	repo := &mockGroupRepo{
		inviteByTokenFn: func(_ context.Context, token string) (InviteSnapshot, bool, error) {
			return InviteSnapshot{
				ID:        "invite_1",
				GroupID:   "group_1",
				Token:     token,
				CreatedBy: "alice",
				ExpiresAt: time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC).Unix(),
			}, true, nil
		},
		membershipFn: func(_ context.Context, playerID, groupID string) (GroupMembershipSnapshot, bool, error) {
			return GroupMembershipSnapshot{}, false, nil
		},
		groupFn: func(_ context.Context, groupID string) (GroupSnapshot, bool, error) {
			return GroupSnapshot{ID: groupID, Name: "Breakfast Crew"}, true, nil
		},
		markInviteUsedFn: func(_ context.Context, token, playerID string) error {
			markedToken = token
			markedPlayerID = playerID
			return nil
		},
		insertMembershipFn: func(_ context.Context, groupID, playerID, role string) error {
			insertedGroupID = groupID
			insertedPlayerID = playerID
			insertedRole = role
			return nil
		},
	}

	result := AcceptInvite(context.Background(), repo, AcceptInviteInput{
		Token:    "valid-token",
		PlayerID: "bob",
		Now:      time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	})

	if result.Status != InviteAccepted {
		t.Fatalf("expected InviteAccepted, got %q", result.Status)
	}
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Group.Name != "Breakfast Crew" {
		t.Fatalf("expected group name Breakfast Crew, got %q", result.Group.Name)
	}
	if result.Membership.Role != "Player" || result.Membership.PlayerID != "bob" {
		t.Fatalf("expected Player membership for bob, got %#v", result.Membership)
	}
	if markedToken != "valid-token" || markedPlayerID != "bob" {
		t.Fatalf("expected MarkInviteUsed called with valid-token/bob, got %q/%q", markedToken, markedPlayerID)
	}
	if insertedGroupID != "group_1" || insertedPlayerID != "bob" || insertedRole != "Player" {
		t.Fatalf("expected InsertMembership called with group_1/bob/Player, got %q/%q/%q", insertedGroupID, insertedPlayerID, insertedRole)
	}
}

func testGroupHome_NonMemberIsRejected(t *testing.T) {
	repo := &mockGroupRepo{
		groupFn: func(_ context.Context, groupID string) (GroupSnapshot, bool, error) {
			return GroupSnapshot{ID: groupID, Name: "Breakfast Crew"}, true, nil
		},
		membershipFn: func(_ context.Context, playerID, groupID string) (GroupMembershipSnapshot, bool, error) {
			return GroupMembershipSnapshot{}, false, nil
		},
	}

	result := GroupHome(context.Background(), repo, GroupHomeInput{
		PlayerID: "bob",
		GroupID:  "group_1",
	})

	if result.Allowed {
		t.Fatal("expected non-member to be rejected")
	}
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
}

func testGroupHome_NotFoundIsRejected(t *testing.T) {
	repo := &mockGroupRepo{
		groupFn: func(_ context.Context, groupID string) (GroupSnapshot, bool, error) {
			return GroupSnapshot{}, false, nil
		},
	}

	result := GroupHome(context.Background(), repo, GroupHomeInput{
		PlayerID: "alice",
		GroupID:  "nonexistent",
	})

	if result.Allowed {
		t.Fatal("expected nonexistent group to be rejected")
	}
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
}

func testGroupHome_MemberCanViewHome(t *testing.T) {
	repo := &mockGroupRepo{
		groupFn: func(_ context.Context, groupID string) (GroupSnapshot, bool, error) {
			return GroupSnapshot{ID: groupID, Name: "Breakfast Crew"}, true, nil
		},
		membershipFn: func(_ context.Context, playerID, groupID string) (GroupMembershipSnapshot, bool, error) {
			return GroupMembershipSnapshot{GroupID: groupID, PlayerID: playerID, Role: "Player"}, true, nil
		},
	}

	result := GroupHome(context.Background(), repo, GroupHomeInput{
		PlayerID: "alice",
		GroupID:  "group_1",
	})

	if !result.Allowed {
		t.Fatal("expected member to be allowed")
	}
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Group.Name != "Breakfast Crew" {
		t.Fatalf("expected group name Breakfast Crew, got %q", result.Group.Name)
	}
	if result.Membership.Role != "Player" {
		t.Fatalf("expected Player role, got %q", result.Membership.Role)
	}
}

func testListGroups_ReturnsMembershipsForPlayer(t *testing.T) {
	repo := &mockGroupRepo{
		membershipsForPlayerFn: func(_ context.Context, playerID string) ([]MembershipWithGroupSnapshot, error) {
			return []MembershipWithGroupSnapshot{
				{Group: GroupSnapshot{ID: "group_1", Name: "Breakfast Crew"}, Membership: GroupMembershipSnapshot{GroupID: "group_1", PlayerID: playerID, Role: "Group Admin"}},
				{Group: GroupSnapshot{ID: "group_2", Name: "Dinner Weirdos"}, Membership: GroupMembershipSnapshot{GroupID: "group_2", PlayerID: playerID, Role: "Player"}},
			}, nil
		},
	}

	memberships, err := ListGroups(context.Background(), repo, "player_alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(memberships) != 2 {
		t.Fatalf("expected 2 memberships, got %d", len(memberships))
	}
	if memberships[0].Group.Name != "Breakfast Crew" || memberships[0].Membership.Role != "Group Admin" {
		t.Fatalf("expected first membership Breakfast Crew/Group Admin, got %#v", memberships[0])
	}
	if memberships[1].Group.Name != "Dinner Weirdos" || memberships[1].Membership.Role != "Player" {
		t.Fatalf("expected second membership Dinner Weirdos/Player, got %#v", memberships[1])
	}
}

func testCreateGroup_InsertGroupErrorPropagates(t *testing.T) {
	repo := &mockGroupRepo{
		insertGroupFn: func(_ context.Context, groupID, name string) error {
			return errors.New("db error")
		},
	}

	result := CreateGroup(context.Background(), repo, CreateGroupInput{
		GroupID:         "group_1",
		GroupName:       "Breakfast Crew",
		CreatorPlayerID: "player_alice",
	})

	if result.Err == nil || result.Err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", result.Err)
	}
}

func testCreateGroup_InsertMembershipErrorPropagates(t *testing.T) {
	repo := &mockGroupRepo{
		insertGroupFn: func(_ context.Context, groupID, name string) error {
			return nil
		},
		insertMembershipFn: func(_ context.Context, groupID, playerID, role string) error {
			return errors.New("db error")
		},
	}

	result := CreateGroup(context.Background(), repo, CreateGroupInput{
		GroupID:         "group_1",
		GroupName:       "Breakfast Crew",
		CreatorPlayerID: "player_alice",
	})

	if result.Err == nil || result.Err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", result.Err)
	}
}

func testCreateGroup_CreatorBecomesGroupAdmin(t *testing.T) {
	var insertedGroupID, insertedGroupName string
	var insertedMembershipGroupID, insertedMembershipPlayerID, insertedMembershipRole string

	repo := &mockGroupRepo{
		insertGroupFn: func(_ context.Context, groupID, name string) error {
			insertedGroupID = groupID
			insertedGroupName = name
			return nil
		},
		insertMembershipFn: func(_ context.Context, groupID, playerID, role string) error {
			insertedMembershipGroupID = groupID
			insertedMembershipPlayerID = playerID
			insertedMembershipRole = role
			return nil
		},
	}

	result := CreateGroup(context.Background(), repo, CreateGroupInput{
		GroupID:         "group_1",
		GroupName:       "Breakfast Crew",
		CreatorPlayerID: "player_alice",
	})

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Group.ID != "group_1" || result.Group.Name != "Breakfast Crew" {
		t.Fatalf("expected group id group_1 and name Breakfast Crew, got %#v", result.Group)
	}
	if result.Membership.PlayerID != "player_alice" || result.Membership.Role != "Group Admin" || result.Membership.GroupID != "group_1" {
		t.Fatalf("expected creator Group Admin membership, got %#v", result.Membership)
	}
	if insertedGroupID != "group_1" || insertedGroupName != "Breakfast Crew" {
		t.Fatalf("expected InsertGroup called with group_1/Breakfast Crew, got %q/%q", insertedGroupID, insertedGroupName)
	}
	if insertedMembershipGroupID != "group_1" || insertedMembershipPlayerID != "player_alice" || insertedMembershipRole != "Group Admin" {
		t.Fatalf("expected InsertMembership called with group_1/player_alice/Group Admin, got %q/%q/%q", insertedMembershipGroupID, insertedMembershipPlayerID, insertedMembershipRole)
	}
}

func TestGroup(t *testing.T) {
	t.Run("create invite", func(t *testing.T) {
		t.Run("non-member is rejected", testCreateInvite_NonMemberIsRejected)
		t.Run("member can create invite", testCreateInvite_MemberCanCreateInvite)
	})

	t.Run("accept invite", func(t *testing.T) {
		t.Run("invalid token returns invalid", testAcceptInvite_InvalidTokenReturnsInvalid)
		t.Run("used token returns used", testAcceptInvite_UsedTokenReturnsUsed)
		t.Run("expired token returns expired", testAcceptInvite_ExpiredTokenReturnsExpired)
		t.Run("existing member returns member", testAcceptInvite_ExistingMemberReturnsMember)
		t.Run("valid token creates membership and returns accepted", testAcceptInvite_ValidTokenCreatesMembershipAndReturnsAccepted)
	})

	t.Run("group home", func(t *testing.T) {
		t.Run("non-member is rejected", testGroupHome_NonMemberIsRejected)
		t.Run("not found is rejected", testGroupHome_NotFoundIsRejected)
		t.Run("member can view home", testGroupHome_MemberCanViewHome)
	})

	t.Run("list groups", func(t *testing.T) {
		t.Run("returns memberships for player", testListGroups_ReturnsMembershipsForPlayer)
	})

	t.Run("create group", func(t *testing.T) {
		t.Run("creator becomes group admin", testCreateGroup_CreatorBecomesGroupAdmin)
		t.Run("insert group error propagates", testCreateGroup_InsertGroupErrorPropagates)
		t.Run("insert membership error propagates", testCreateGroup_InsertMembershipErrorPropagates)
	})
}
