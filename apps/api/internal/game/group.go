package game

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInviteNotFound      = errors.New("Invite not found")
	ErrInviteExpired       = errors.New("Invite expired")
	ErrInviteUsed          = errors.New("Invite already used")
	ErrPlayerAlreadyMember = errors.New("Player already has a Group Membership")
)

type GroupSnapshot struct {
	ID   string
	Name string
}

type GroupMembershipSnapshot struct {
	GroupID  string
	PlayerID string
	Role     string
}

type MembershipWithGroupSnapshot struct {
	Group      GroupSnapshot
	Membership GroupMembershipSnapshot
}

type InviteSnapshot struct {
	ID        string
	GroupID   string
	Token     string
	CreatedBy string
	ExpiresAt int64
	UsedBy    *string
}

type GroupRepository interface {
	InsertGroup(ctx context.Context, groupID, name string) error
	InsertMembership(ctx context.Context, groupID, playerID, role string) error
	Group(ctx context.Context, groupID string) (GroupSnapshot, bool, error)
	Membership(ctx context.Context, playerID, groupID string) (GroupMembershipSnapshot, bool, error)
	MembershipsForPlayer(ctx context.Context, playerID string) ([]MembershipWithGroupSnapshot, error)
	InsertInvite(ctx context.Context, groupID, createdByPlayerID string, expiresAt int64) (InviteSnapshot, error)
	InviteByToken(ctx context.Context, token string) (InviteSnapshot, bool, error)
	MarkInviteUsed(ctx context.Context, token, playerID string) error
}

type CreateGroupInput struct {
	GroupID         string
	GroupName       string
	CreatorPlayerID string
}

type CreateGroupResult struct {
	Group      GroupSnapshot
	Membership GroupMembershipSnapshot
	Err        error
}

type CreateInviteInput struct {
	GroupID  string
	PlayerID string
	Now      time.Time
}

type CreateInviteResult struct {
	Invite  InviteSnapshot
	Allowed bool
	Err     error
}

type AcceptInviteInput struct {
	Token    string
	PlayerID string
	Now      time.Time
}

type InviteAcceptStatus string

const (
	InviteAccepted InviteAcceptStatus = "accepted"
	InviteInvalid  InviteAcceptStatus = "invalid"
	InviteUsed     InviteAcceptStatus = "used"
	InviteExpired  InviteAcceptStatus = "expired"
	InviteMember   InviteAcceptStatus = "member"
)

type AcceptInviteResult struct {
	Group      GroupSnapshot
	Membership GroupMembershipSnapshot
	Status     InviteAcceptStatus
	Err        error
}

type GroupHomeInput struct {
	PlayerID string
	GroupID  string
}

type GroupHomeResult struct {
	Group      GroupSnapshot
	Membership GroupMembershipSnapshot
	Allowed    bool
	Err        error
}

func CreateGroup(ctx context.Context, repo GroupRepository, input CreateGroupInput) CreateGroupResult {
	if err := repo.InsertGroup(ctx, input.GroupID, input.GroupName); err != nil {
		return CreateGroupResult{Err: err}
	}
	if err := repo.InsertMembership(ctx, input.GroupID, input.CreatorPlayerID, "Group Admin"); err != nil {
		return CreateGroupResult{Err: err}
	}
	return CreateGroupResult{
		Group:      GroupSnapshot{ID: input.GroupID, Name: input.GroupName},
		Membership: GroupMembershipSnapshot{GroupID: input.GroupID, PlayerID: input.CreatorPlayerID, Role: "Group Admin"},
	}
}

func CreateInvite(ctx context.Context, repo GroupRepository, input CreateInviteInput) CreateInviteResult {
	_, ok, err := repo.Membership(ctx, input.PlayerID, input.GroupID)
	if err != nil {
		return CreateInviteResult{Err: err}
	}
	if !ok {
		return CreateInviteResult{Allowed: false}
	}

	expiresAt := input.Now.Add(7 * 24 * time.Hour).Unix()
	invite, err := repo.InsertInvite(ctx, input.GroupID, input.PlayerID, expiresAt)
	if err != nil {
		return CreateInviteResult{Err: err}
	}
	return CreateInviteResult{Invite: invite, Allowed: true}
}

func AcceptInvite(ctx context.Context, repo GroupRepository, input AcceptInviteInput) AcceptInviteResult {
	invite, ok, err := repo.InviteByToken(ctx, input.Token)
	if err != nil {
		return AcceptInviteResult{Err: err}
	}
	if !ok {
		return AcceptInviteResult{Status: InviteInvalid}
	}
	if invite.UsedBy != nil {
		return AcceptInviteResult{Status: InviteUsed}
	}
	if input.Now.Unix() > invite.ExpiresAt {
		return AcceptInviteResult{Status: InviteExpired}
	}

	_, ok, err = repo.Membership(ctx, input.PlayerID, invite.GroupID)
	if err != nil {
		return AcceptInviteResult{Err: err}
	}
	if ok {
		return AcceptInviteResult{Status: InviteMember}
	}

	if err := repo.MarkInviteUsed(ctx, input.Token, input.PlayerID); err != nil {
		return AcceptInviteResult{Err: err}
	}
	if err := repo.InsertMembership(ctx, invite.GroupID, input.PlayerID, "Player"); err != nil {
		return AcceptInviteResult{Err: err}
	}

	group, ok, err := repo.Group(ctx, invite.GroupID)
	if err != nil {
		return AcceptInviteResult{Err: err}
	}
	if !ok {
		return AcceptInviteResult{Status: InviteInvalid}
	}

	return AcceptInviteResult{
		Group:      group,
		Membership: GroupMembershipSnapshot{GroupID: invite.GroupID, PlayerID: input.PlayerID, Role: "Player"},
		Status:     InviteAccepted,
	}
}

func GroupHome(ctx context.Context, repo GroupRepository, input GroupHomeInput) GroupHomeResult {
	group, ok, err := repo.Group(ctx, input.GroupID)
	if err != nil {
		return GroupHomeResult{Err: err}
	}
	if !ok {
		return GroupHomeResult{Allowed: false}
	}

	membership, ok, err := repo.Membership(ctx, input.PlayerID, input.GroupID)
	if err != nil {
		return GroupHomeResult{Err: err}
	}
	if !ok {
		return GroupHomeResult{Allowed: false}
	}

	return GroupHomeResult{
		Group:      group,
		Membership: membership,
		Allowed:    true,
	}
}

func ListGroups(ctx context.Context, repo GroupRepository, playerID string) ([]MembershipWithGroupSnapshot, error) {
	return repo.MembershipsForPlayer(ctx, playerID)
}
