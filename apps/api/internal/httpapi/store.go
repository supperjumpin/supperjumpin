package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var ErrSeasonAlreadyOpen = errors.New("Group already has an active or closing Season")

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

type Season struct {
	ID                   string `json:"id"`
	GroupID              string `json:"groupId"`
	CommissionerPlayerID string `json:"commissionerPlayerId"`
	Status               string `json:"status"`
}

type GroupHomeResponse struct {
	Group        Group           `json:"group"`
	Membership   GroupMembership `json:"membership"`
	ActiveSeason *Season         `json:"activeSeason"`
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

type Store interface {
	BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error)
	CreateGroup(ctx context.Context, player Player, name string) (GroupHomeResponse, error)
	GroupHome(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error)
	ListGroups(ctx context.Context, player Player) (ListGroupsResponse, error)
	StartSeason(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error)
}

type MemoryStore struct {
	mu           sync.Mutex
	accounts     map[string]MeResponse
	groups       map[string]Group
	memberships  map[string]map[string]GroupMembership
	seasons      map[string]Season
	groupNumber  int
	seasonNumber int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		accounts:    map[string]MeResponse{},
		groups:      map[string]Group{},
		memberships: map[string]map[string]GroupMembership{},
		seasons:     map[string]Season{},
	}
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
	return groupHome(group, membership, nil), nil
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
	season := s.activeSeasonForGroup(groupID)
	return groupHome(group, membership, season), true, nil
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

func (s *MemoryStore) StartSeason(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error) {
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
	if s.openSeasonForGroup(groupID) != nil {
		return GroupHomeResponse{}, true, ErrSeasonAlreadyOpen
	}

	s.seasonNumber++
	season := Season{
		ID:                   stableID("season", groupID+":"+strconv.Itoa(s.seasonNumber)),
		GroupID:              groupID,
		CommissionerPlayerID: player.ID,
		Status:               "Active",
	}
	s.seasons[season.ID] = season
	return groupHome(group, membership, &season), true, nil
}

func (s *MemoryStore) openSeasonForGroup(groupID string) *Season {
	for _, season := range s.seasons {
		if season.GroupID == groupID && isOpenSeasonStatus(season.Status) {
			return &season
		}
	}
	return nil
}

func (s *MemoryStore) activeSeasonForGroup(groupID string) *Season {
	season := s.openSeasonForGroup(groupID)
	if season != nil && season.Status == "Active" {
		return season
	}
	return nil
}

func isOpenSeasonStatus(status string) bool {
	return status == "Active" || status == "Judging Grace Period"
}

func groupHome(group Group, membership GroupMembership, activeSeason *Season) GroupHomeResponse {
	return GroupHomeResponse{
		Group:        group,
		Membership:   membership,
		ActiveSeason: activeSeason,
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
