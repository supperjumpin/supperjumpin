package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrSeasonAlreadyOpen = errors.New("Group already has an active or closing Season")

var ErrStuntNotFound = errors.New("Stunt not found")

var ErrEvidenceUploadAuthorizationNotFound = errors.New("Evidence upload authorization not found")

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

type Stunt struct {
	ID          string  `json:"id"`
	GroupID     string  `json:"groupId"`
	PlayerID    string  `json:"playerId"`
	SeasonID    *string `json:"seasonId"`
	Status      string  `json:"status"`
	Source      string  `json:"source"`
	Destination string  `json:"destination"`
	Food        string  `json:"food"`
	OffSeason   bool    `json:"offSeason"`
}

type EvidenceUploadAuthorization struct {
	ID             string            `json:"id"`
	StuntID        string            `json:"stuntId"`
	UploadURL      string            `json:"uploadUrl"`
	UploadMethod   string            `json:"uploadMethod"`
	UploadHeaders  map[string]string `json:"uploadHeaders"`
	MediaObjectKey string            `json:"mediaObjectKey"`
	ExpiresAt      time.Time         `json:"expiresAt"`
}

type Evidence struct {
	ID             string    `json:"id"`
	StuntID        string    `json:"stuntId"`
	Caption        string    `json:"caption"`
	MediaObjectKey string    `json:"mediaObjectKey"`
	CreatedAt      time.Time `json:"createdAt"`
}

type EvidenceSubmission struct {
	Stunt    Stunt    `json:"stunt"`
	Evidence Evidence `json:"evidence"`
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
	StartSeason(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error)
	CreateIdea(ctx context.Context, player Player, groupID string, source string, destination string, food string) (Stunt, bool, error)
	CreatePlannedStunt(ctx context.Context, player Player, ideaID string, offSeason bool) (Stunt, bool, error)
	AuthorizeEvidenceUpload(ctx context.Context, player Player, stuntID string, contentType string) (EvidenceUploadAuthorization, bool, error)
	SubmitEvidence(ctx context.Context, player Player, stuntID string, uploadAuthorizationID string, caption string) (EvidenceSubmission, bool, error)
}

type MemoryStore struct {
	mu           sync.Mutex
	accounts     map[string]MeResponse
	groups       map[string]Group
	memberships  map[string]map[string]GroupMembership
	invites      map[string]memoryInvite
	seasons      map[string]Season
	stunts       map[string]Stunt
	uploads      map[string]EvidenceUploadAuthorization
	evidences    map[string]Evidence
	now          func() time.Time
	groupNumber  int
	inviteNumber int
	seasonNumber int
	stuntNumber  int
	uploadNumber int
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
		seasons:     map[string]Season{},
		stunts:      map[string]Stunt{},
		uploads:     map[string]EvidenceUploadAuthorization{},
		evidences:   map[string]Evidence{},
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
	if _, ok := s.memberships[invite.GroupID][player.ID]; ok {
		return GroupHomeResponse{}, InviteMember, nil
	}
	s.memberships[invite.GroupID][player.ID] = membership
	invite.UsedBy = player.ID
	s.invites[token] = invite
	return groupHome(group, membership, s.activeSeasonForGroup(invite.GroupID)), InviteAccepted, nil
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

func (s *MemoryStore) CreateIdea(ctx context.Context, player Player, groupID string, source string, destination string, food string) (Stunt, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.memberships[groupID][player.ID]; !ok {
		return Stunt{}, false, nil
	}
	s.stuntNumber++
	stunt := Stunt{
		ID:          stableID("stunt", groupID+":"+player.ID+":"+strconv.Itoa(s.stuntNumber)),
		GroupID:     groupID,
		PlayerID:    player.ID,
		Status:      "Idea",
		Source:      source,
		Destination: destination,
		Food:        food,
		OffSeason:   true,
	}
	s.stunts[stunt.ID] = stunt
	return stunt, true, nil
}

func (s *MemoryStore) CreatePlannedStunt(ctx context.Context, player Player, ideaID string, offSeason bool) (Stunt, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stunt, ok := s.stunts[ideaID]
	if !ok || stunt.Status != "Idea" {
		return Stunt{}, false, ErrStuntNotFound
	}
	if _, ok := s.memberships[stunt.GroupID][player.ID]; !ok || stunt.PlayerID != player.ID {
		return Stunt{}, false, nil
	}
	stunt.Status = "Planned Stunt"
	stunt.SeasonID = nil
	stunt.OffSeason = true
	if !offSeason {
		if season := s.activeSeasonForGroup(stunt.GroupID); season != nil {
			stunt.SeasonID = &season.ID
			stunt.OffSeason = false
		}
	}
	s.stunts[stunt.ID] = stunt
	return stunt, true, nil
}

func (s *MemoryStore) AuthorizeEvidenceUpload(ctx context.Context, player Player, stuntID string, contentType string) (EvidenceUploadAuthorization, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stunt, ok := s.stunts[stuntID]
	if !ok || stunt.Status != "Planned Stunt" {
		return EvidenceUploadAuthorization{}, false, ErrStuntNotFound
	}
	if stunt.PlayerID != player.ID {
		return EvidenceUploadAuthorization{}, false, nil
	}

	s.uploadNumber++
	authorization := EvidenceUploadAuthorization{
		ID:             stableID("evidence_upload", stuntID+":"+strconv.Itoa(s.uploadNumber)),
		StuntID:        stuntID,
		UploadURL:      "https://storage.supperjumpin.test/uploads/" + stuntID,
		UploadMethod:   httpMethodPut,
		UploadHeaders:  map[string]string{"Content-Type": contentType},
		MediaObjectKey: "uploads/" + stuntID + "/" + strconv.Itoa(s.uploadNumber),
		ExpiresAt:      s.now().Add(15 * time.Minute).UTC(),
	}
	s.uploads[authorization.ID] = authorization
	return authorization, true, nil
}

func (s *MemoryStore) SubmitEvidence(ctx context.Context, player Player, stuntID string, uploadAuthorizationID string, caption string) (EvidenceSubmission, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stunt, ok := s.stunts[stuntID]
	if !ok || stunt.Status != "Planned Stunt" {
		return EvidenceSubmission{}, false, ErrStuntNotFound
	}
	if stunt.PlayerID != player.ID {
		return EvidenceSubmission{}, false, nil
	}
	authorization, ok := s.uploads[uploadAuthorizationID]
	if !ok || authorization.StuntID != stuntID || s.now().After(authorization.ExpiresAt) {
		return EvidenceSubmission{}, false, ErrEvidenceUploadAuthorizationNotFound
	}

	delete(s.uploads, uploadAuthorizationID)
	stunt.Status = "Performed Stunt"
	s.stunts[stunt.ID] = stunt

	evidence := Evidence{
		ID:             stableID("evidence", stuntID+":"+uploadAuthorizationID),
		StuntID:        stuntID,
		Caption:        caption,
		MediaObjectKey: authorization.MediaObjectKey,
		CreatedAt:      s.now().UTC(),
	}
	s.evidences[stuntID] = evidence
	return EvidenceSubmission{Stunt: stunt, Evidence: evidence}, true, nil
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

func randomToken(kind string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate %s: %w", kind, err)
	}
	return kind + "_" + hex.EncodeToString(bytes), nil
}

const httpMethodPut = "PUT"

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
