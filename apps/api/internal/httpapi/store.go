package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Account struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type Player struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type MeResponse struct {
	Account Account `json:"account"`
	Player  Player  `json:"player"`
}

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GroupMembership struct {
	GroupID  string `json:"groupId"`
	PlayerID string `json:"playerId"`
	Role     string `json:"role"`
}

type GroupHomeResponse struct {
	Group        Group           `json:"group"`
	Membership   GroupMembership `json:"membership"`
	ActiveSeason any             `json:"activeSeason"`
	RecentStunts []any           `json:"recentStunts"`
	Standings    []any           `json:"standings"`
}

type GroupMembershipSummary struct {
	Group      Group           `json:"group"`
	Membership GroupMembership `json:"membership"`
}

type ListGroupsResponse struct {
	Memberships []GroupMembershipSummary `json:"memberships"`
}

type Invite struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"groupId"`
	Token     string    `json:"token"`
	CreatedBy string    `json:"createdBy"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type InviteAcceptStatus string

const (
	InviteAccepted InviteAcceptStatus = "accepted"
	InviteInvalid  InviteAcceptStatus = "invalid"
	InviteUsed     InviteAcceptStatus = "used"
	InviteExpired  InviteAcceptStatus = "expired"
	InviteMember   InviteAcceptStatus = "member"
)

type Store interface {
	BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error)
	CreateGroup(ctx context.Context, player Player, name string) (GroupHomeResponse, error)
	GroupHome(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error)
	ListGroups(ctx context.Context, player Player) (ListGroupsResponse, error)
	CreateInvite(ctx context.Context, player Player, groupID string) (Invite, bool, error)
	AcceptInvite(ctx context.Context, player Player, token string) (GroupHomeResponse, InviteAcceptStatus, error)
}

type MemoryStore struct {
	mu           sync.Mutex
	accounts     map[string]MeResponse
	groups       map[string]Group
	memberships  map[string]map[string]GroupMembership
	invites      map[string]memoryInvite
	now          func() time.Time
	groupNumber  int
	inviteNumber int
}

type memoryInvite struct {
	Invite
	UsedBy string
}

func NewMemoryStore() *MemoryStore {
	return NewMemoryStoreWithClock(time.Now)
}

func NewMemoryStoreWithClock(now func() time.Time) *MemoryStore {
	return &MemoryStore{
		accounts:    map[string]MeResponse{},
		groups:      map[string]Group{},
		memberships: map[string]map[string]GroupMembership{},
		invites:     map[string]memoryInvite{},
		now:         now,
	}
}

func (s *MemoryStore) SetClock(now func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.now = now
}

func (s *MemoryStore) BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := identity.Provider + ":" + identity.Subject
	if profile, ok := s.accounts[key]; ok {
		return profile, nil
	}

	accountID := stableID("account", key)
	profile := MeResponse{
		Account: Account{ID: accountID, Email: identity.Email},
		Player:  Player{ID: stableID("player", accountID), DisplayName: displayName(identity.Email)},
	}
	s.accounts[key] = profile
	return profile, nil
}

func (s *MemoryStore) CreateGroup(ctx context.Context, player Player, name string) (GroupHomeResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.groupNumber++
	group := Group{ID: stableID("group", player.ID+":"+name+":"+strconv.Itoa(s.groupNumber)), Name: name}
	membership := GroupMembership{GroupID: group.ID, PlayerID: player.ID, Role: "Group Admin"}
	s.groups[group.ID] = group
	if s.memberships[group.ID] == nil {
		s.memberships[group.ID] = map[string]GroupMembership{}
	}
	s.memberships[group.ID][player.ID] = membership
	return groupHome(group, membership), nil
}

func (s *MemoryStore) GroupHome(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, ok := s.groups[groupID]
	if !ok {
		return GroupHomeResponse{}, false, nil
	}
	membership, ok := s.memberships[groupID][player.ID]
	if !ok {
		return GroupHomeResponse{}, false, nil
	}
	return groupHome(group, membership), true, nil
}

func (s *MemoryStore) ListGroups(ctx context.Context, player Player) (ListGroupsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	memberships := []GroupMembershipSummary{}
	for groupID, groupMemberships := range s.memberships {
		membership, ok := groupMemberships[player.ID]
		if !ok {
			continue
		}
		memberships = append(memberships, GroupMembershipSummary{Group: s.groups[groupID], Membership: membership})
	}
	sort.Slice(memberships, func(i, j int) bool {
		return memberships[i].Group.Name < memberships[j].Group.Name
	})
	return ListGroupsResponse{Memberships: memberships}, nil
}

func (s *MemoryStore) CreateInvite(ctx context.Context, player Player, groupID string) (Invite, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.memberships[groupID][player.ID]; !ok {
		return Invite{}, false, nil
	}

	s.inviteNumber++
	token, err := randomToken("invite_token")
	if err != nil {
		return Invite{}, false, err
	}
	invite := Invite{
		ID:        stableID("invite", groupID+":"+strconv.Itoa(s.inviteNumber)),
		GroupID:   groupID,
		Token:     token,
		CreatedBy: player.ID,
		ExpiresAt: s.now().Add(7 * 24 * time.Hour).UTC(),
	}
	s.invites[invite.Token] = memoryInvite{Invite: invite}
	return invite, true, nil
}

func (s *MemoryStore) AcceptInvite(ctx context.Context, player Player, token string) (GroupHomeResponse, InviteAcceptStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	invite, ok := s.invites[token]
	if !ok {
		return GroupHomeResponse{}, InviteInvalid, nil
	}
	if invite.UsedBy != "" {
		return GroupHomeResponse{}, InviteUsed, nil
	}
	if s.now().After(invite.ExpiresAt) {
		return GroupHomeResponse{}, InviteExpired, nil
	}
	group, ok := s.groups[invite.GroupID]
	if !ok {
		return GroupHomeResponse{}, InviteInvalid, nil
	}
	membership := GroupMembership{GroupID: invite.GroupID, PlayerID: player.ID, Role: "Player"}
	if existing, ok := s.memberships[invite.GroupID][player.ID]; ok {
		_ = existing
		return GroupHomeResponse{}, InviteMember, nil
	}
	s.memberships[invite.GroupID][player.ID] = membership
	invite.UsedBy = player.ID
	s.invites[token] = invite
	return groupHome(group, membership), InviteAccepted, nil
}

func randomToken(kind string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate %s: %w", kind, err)
	}
	return kind + "_" + hex.EncodeToString(bytes), nil
}

func groupHome(group Group, membership GroupMembership) GroupHomeResponse {
	return GroupHomeResponse{
		Group:        group,
		Membership:   membership,
		ActiveSeason: nil,
		RecentStunts: []any{},
		Standings:    []any{},
	}
}

func stableID(kind string, value string) string {
	sum := sha256.Sum256([]byte(kind + ":" + value))
	return kind + "_" + hex.EncodeToString(sum[:])[:12]
}

func displayName(email string) string {
	name, _, ok := strings.Cut(email, "@")
	if !ok || name == "" {
		return "player"
	}
	return name
}
