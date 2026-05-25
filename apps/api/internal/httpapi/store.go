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

var ErrSeasonNotFound = errors.New("Season not found")

var ErrStuntNotFound = errors.New("Stunt not found")

var ErrEvidenceUploadAuthorizationNotFound = errors.New("Evidence upload authorization not found")

var ErrJudgingWindowClosed = errors.New("Judging Window closed")

var ErrSubmissionWindowClosed = errors.New("Submission Window closed")

var ErrInvalidJudgmentScore = errors.New("Judgment scores must be between 0 and 10")

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
	ID                   string    `json:"id"`
	GroupID              string    `json:"groupId"`
	CommissionerPlayerID string    `json:"commissionerPlayerId"`
	Status               string    `json:"status"`
	SubmissionDeadline   time.Time `json:"submissionDeadline"`
	JudgingDeadline      time.Time `json:"judgingDeadline"`
}

type SeasonHistoryEntry struct {
	ID            string    `json:"id"`
	SeasonID      string    `json:"seasonId"`
	Action        string    `json:"action"`
	ActorPlayerID string    `json:"actorPlayerId"`
	ActorRole     string    `json:"actorRole"`
	Override      bool      `json:"override"`
	FromStatus    string    `json:"fromStatus"`
	ToStatus      string    `json:"toStatus"`
	CreatedAt     time.Time `json:"createdAt"`
}

type SeasonHistoryResponse struct {
	Entries []SeasonHistoryEntry `json:"entries"`
}

type GroupHomeResponse struct {
	Group        Group                `json:"group"`
	Membership   GroupMembership      `json:"membership"`
	ActiveSeason *Season              `json:"activeSeason"`
	RecentStunts []PerformedStuntView `json:"recentStunts"`
	Standings    []StandingEntry      `json:"standings"`
}

type StandingEntry struct {
	Player       Player `json:"player"`
	SeasonScore  int    `json:"seasonScore"`
	JudgedStunts int    `json:"judgedStunts"`
}

type PerformedStuntView struct {
	Stunt     Stunt    `json:"stunt"`
	Performer Player   `json:"performer"`
	Evidence  Evidence `json:"evidence"`
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
	FinalScore  *int    `json:"finalScore"`
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

type Judgment struct {
	ID            string `json:"id"`
	StuntID       string `json:"stuntId"`
	PlayerID      string `json:"playerId"`
	Difficulty    int    `json:"difficulty"`
	Transgression int    `json:"transgression"`
	Creativity    int    `json:"creativity"`
	Documentation int    `json:"documentation"`
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
	StartSeason(ctx context.Context, player Player, groupID string, submissionDeadline time.Time, judgingDeadline time.Time) (GroupHomeResponse, bool, error)
	CloseSeasonSubmissions(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error)
	FinalizeSeason(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error)
	SeasonHistory(ctx context.Context, player Player, seasonID string) (SeasonHistoryResponse, bool, error)
	CreateIdea(ctx context.Context, player Player, groupID string, source string, destination string, food string) (Stunt, bool, error)
	CreatePlannedStunt(ctx context.Context, player Player, ideaID string, offSeason bool) (Stunt, bool, error)
	AuthorizeEvidenceUpload(ctx context.Context, player Player, stuntID string, contentType string) (EvidenceUploadAuthorization, bool, error)
	SubmitEvidence(ctx context.Context, player Player, stuntID string, uploadAuthorizationID string, caption string) (EvidenceSubmission, bool, error)
	SubmitJudgment(ctx context.Context, player Player, stuntID string, difficulty int, transgression int, creativity int, documentation int) (Judgment, bool, bool, error)
}

type MemoryStore struct {
	mu                sync.Mutex
	accounts          map[string]MeResponse
	players           map[string]Player
	groups            map[string]Group
	memberships       map[string]map[string]GroupMembership
	invites           map[string]memoryInvite
	seasons           map[string]Season
	seasonEvents      map[string][]SeasonHistoryEntry
	seasonOrder       map[string]int
	stunts            map[string]Stunt
	uploads           map[string]EvidenceUploadAuthorization
	evidences         map[string]Evidence
	judgments         map[string]Judgment
	now               func() time.Time
	groupNumber       int
	inviteNumber      int
	seasonNumber      int
	stuntNumber       int
	uploadNumber      int
	seasonEventNumber int
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
		accounts:     map[string]MeResponse{},
		players:      map[string]Player{},
		groups:       map[string]Group{},
		memberships:  map[string]map[string]GroupMembership{},
		invites:      map[string]memoryInvite{},
		seasons:      map[string]Season{},
		seasonEvents: map[string][]SeasonHistoryEntry{},
		seasonOrder:  map[string]int{},
		stunts:       map[string]Stunt{},
		uploads:      map[string]EvidenceUploadAuthorization{},
		evidences:    map[string]Evidence{},
		judgments:    map[string]Judgment{},
		now:          now,
	}
}

func (s *MemoryStore) SetClock(now func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.now = now
}

func (s *MemoryStore) SetSeasonStatus(seasonID string, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	season := s.seasons[seasonID]
	season.Status = status
	s.seasons[seasonID] = season
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
	s.players[profile.Player.ID] = profile.Player
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
	return groupHome(group, membership, nil, s.recentPerformedStuntsForGroup(group.ID), s.standingsForGroup(group.ID)), nil
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
	s.ensureSeasonStatusesForGroup(groupID)
	season := s.currentSeasonForGroup(groupID)
	return groupHome(group, membership, season, s.recentPerformedStuntsForGroup(groupID), s.standingsForGroup(groupID)), true, nil
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
	return groupHome(group, membership, s.currentSeasonForGroup(invite.GroupID), s.recentPerformedStuntsForGroup(invite.GroupID), s.standingsForGroup(invite.GroupID)), InviteAccepted, nil
}

func (s *MemoryStore) StartSeason(ctx context.Context, player Player, groupID string, submissionDeadline time.Time, judgingDeadline time.Time) (GroupHomeResponse, bool, error) {
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
		SubmissionDeadline:   submissionDeadline.UTC(),
		JudgingDeadline:      judgingDeadline.UTC(),
	}
	s.seasons[season.ID] = season
	s.seasonOrder[season.ID] = s.seasonNumber
	return groupHome(group, membership, &season, s.recentPerformedStuntsForGroup(groupID), s.standingsForGroup(groupID)), true, nil
}

func (s *MemoryStore) CloseSeasonSubmissions(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	season, ok := s.seasons[seasonID]
	if !ok {
		return GroupHomeResponse{}, false, ErrSeasonNotFound
	}
	group, ok := s.groups[season.GroupID]
	if !ok {
		return GroupHomeResponse{}, false, ErrSeasonNotFound
	}
	membership, ok := s.memberships[season.GroupID][player.ID]
	if !ok {
		return GroupHomeResponse{}, false, nil
	}
	if player.ID != season.CommissionerPlayerID && membership.Role != "Group Admin" {
		return GroupHomeResponse{}, false, nil
	}
	if season.Status == "Active" {
		fromStatus := season.Status
		season.Status = "Judging Grace Period"
		s.seasons[season.ID] = season
		s.recordSeasonHistory(season.ID, "Submissions Closed", player.ID, membership.Role, player.ID != season.CommissionerPlayerID, fromStatus, season.Status)
	}
	return groupHome(group, membership, s.currentSeasonForGroup(season.GroupID), s.recentPerformedStuntsForGroup(season.GroupID), s.standingsForGroup(season.GroupID)), true, nil
}

func (s *MemoryStore) FinalizeSeason(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	season, ok := s.seasons[seasonID]
	if !ok {
		return GroupHomeResponse{}, false, ErrSeasonNotFound
	}
	group, ok := s.groups[season.GroupID]
	if !ok {
		return GroupHomeResponse{}, false, ErrSeasonNotFound
	}
	membership, ok := s.memberships[season.GroupID][player.ID]
	if !ok {
		return GroupHomeResponse{}, false, nil
	}
	if player.ID != season.CommissionerPlayerID && membership.Role != "Group Admin" {
		return GroupHomeResponse{}, false, nil
	}
	if season.Status != "Finalized" {
		fromStatus := season.Status
		season.Status = "Finalized"
		s.seasons[season.ID] = season
		s.finalizeSeasonStunts(season.ID)
		s.recordSeasonHistory(season.ID, "Season Finalized", player.ID, membership.Role, player.ID != season.CommissionerPlayerID, fromStatus, season.Status)
	}
	return groupHome(group, membership, s.currentSeasonForGroup(season.GroupID), s.recentPerformedStuntsForGroup(season.GroupID), s.standingsForGroup(season.GroupID)), true, nil
}

func (s *MemoryStore) SeasonHistory(ctx context.Context, player Player, seasonID string) (SeasonHistoryResponse, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	season, ok := s.seasons[seasonID]
	if !ok {
		return SeasonHistoryResponse{}, false, ErrSeasonNotFound
	}
	if _, ok := s.memberships[season.GroupID][player.ID]; !ok {
		return SeasonHistoryResponse{}, false, nil
	}
	entries := append([]SeasonHistoryEntry(nil), s.seasonEvents[seasonID]...)
	return SeasonHistoryResponse{Entries: entries}, true, nil
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
	if !s.submissionWindowOpen(stunt) {
		return EvidenceSubmission{}, false, ErrSubmissionWindowClosed
	}
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

func (s *MemoryStore) SubmitJudgment(ctx context.Context, player Player, stuntID string, difficulty int, transgression int, creativity int, documentation int) (Judgment, bool, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validJudgmentScores(difficulty, transgression, creativity, documentation) {
		return Judgment{}, false, false, ErrInvalidJudgmentScore
	}

	stunt, ok := s.stunts[stuntID]
	if !ok || !visiblePerformedStatus(stunt.Status) {
		return Judgment{}, false, false, ErrStuntNotFound
	}
	if stunt.Status != "Performed Stunt" {
		return Judgment{}, false, false, ErrJudgingWindowClosed
	}
	if _, ok := s.memberships[stunt.GroupID][player.ID]; !ok || stunt.PlayerID == player.ID {
		return Judgment{}, false, false, nil
	}
	if !s.judgingWindowOpen(stunt) {
		return Judgment{}, false, false, ErrJudgingWindowClosed
	}

	key := stuntID + ":" + player.ID
	_, existed := s.judgments[key]
	judgment := Judgment{
		ID:            stableID("judgment", key),
		StuntID:       stuntID,
		PlayerID:      player.ID,
		Difficulty:    difficulty,
		Transgression: transgression,
		Creativity:    creativity,
		Documentation: documentation,
	}
	s.judgments[key] = judgment
	return judgment, true, !existed, nil
}

func (s *MemoryStore) openSeasonForGroup(groupID string) *Season {
	s.ensureSeasonStatusesForGroup(groupID)
	for _, season := range s.seasons {
		if season.GroupID == groupID && isOpenSeasonStatus(season.Status) {
			result := season
			return &result
		}
	}
	return nil
}

func (s *MemoryStore) ensureSeasonStatusesForGroup(groupID string) {
	for id, season := range s.seasons {
		if season.GroupID != groupID {
			continue
		}
		wasFinalized := season.Status == "Finalized"
		s.ensureSeasonStatus(&season)
		s.seasons[id] = season
		if !wasFinalized && season.Status == "Finalized" {
			s.finalizeSeasonStunts(season.ID)
		}
	}
}

func (s *MemoryStore) activeSeasonForGroup(groupID string) *Season {
	season := s.openSeasonForGroup(groupID)
	if season != nil && season.Status == "Active" {
		return season
	}
	return nil
}

func (s *MemoryStore) currentSeasonForGroup(groupID string) *Season {
	return s.openSeasonForGroup(groupID)
}

func (s *MemoryStore) recentPerformedStuntsForGroup(groupID string) []PerformedStuntView {
	performed := []PerformedStuntView{}
	for _, stunt := range s.stunts {
		if stunt.GroupID != groupID || !visiblePerformedStatus(stunt.Status) {
			continue
		}
		evidence, ok := s.evidences[stunt.ID]
		if !ok {
			continue
		}
		performed = append(performed, PerformedStuntView{
			Stunt:     stunt,
			Performer: s.players[stunt.PlayerID],
			Evidence:  evidence,
		})
	}
	sort.Slice(performed, func(i, j int) bool {
		if performed[i].Evidence.CreatedAt.Equal(performed[j].Evidence.CreatedAt) {
			return performed[i].Stunt.ID > performed[j].Stunt.ID
		}
		return performed[i].Evidence.CreatedAt.After(performed[j].Evidence.CreatedAt)
	})
	return performed
}

func (s *MemoryStore) finalizeSeasonStunts(seasonID string) {
	for id, stunt := range s.stunts {
		if stunt.SeasonID == nil || *stunt.SeasonID != seasonID || stunt.Status != "Performed Stunt" {
			continue
		}
		if score, ok := s.finalScoreForStunt(stunt.ID); ok {
			stunt.Status = "Judged Stunt"
			stunt.FinalScore = &score
		} else {
			stunt.Status = "Unjudged Stunt"
			stunt.FinalScore = nil
		}
		s.stunts[id] = stunt
	}
}

func (s *MemoryStore) finalScoreForStunt(stuntID string) (int, bool) {
	total := 0
	count := 0
	for _, judgment := range s.judgments {
		if judgment.StuntID != stuntID {
			continue
		}
		total += judgment.Difficulty + judgment.Transgression + judgment.Creativity + judgment.Documentation
		count++
	}
	if count == 0 {
		return 0, false
	}
	return total / count, true
}

func (s *MemoryStore) standingsForGroup(groupID string) []StandingEntry {
	season := s.latestSeasonForGroup(groupID)
	if season == nil {
		return []StandingEntry{}
	}
	byPlayer := map[string]*StandingEntry{}
	for _, stunt := range s.stunts {
		if stunt.GroupID != groupID || stunt.SeasonID == nil || *stunt.SeasonID != season.ID || stunt.Status != "Judged Stunt" || stunt.FinalScore == nil {
			continue
		}
		entry := byPlayer[stunt.PlayerID]
		if entry == nil {
			entry = &StandingEntry{Player: s.players[stunt.PlayerID]}
			byPlayer[stunt.PlayerID] = entry
		}
		entry.SeasonScore += *stunt.FinalScore
		entry.JudgedStunts++
	}
	standings := []StandingEntry{}
	for _, entry := range byPlayer {
		standings = append(standings, *entry)
	}
	sort.Slice(standings, func(i, j int) bool {
		if standings[i].SeasonScore == standings[j].SeasonScore {
			return standings[i].Player.DisplayName < standings[j].Player.DisplayName
		}
		return standings[i].SeasonScore > standings[j].SeasonScore
	})
	return standings
}

func (s *MemoryStore) latestSeasonForGroup(groupID string) *Season {
	var latest *Season
	latestOrder := 0
	for _, season := range s.seasons {
		if season.GroupID != groupID {
			continue
		}
		candidate := season
		candidateOrder := s.seasonOrder[candidate.ID]
		if latest == nil || candidateOrder > latestOrder {
			latest = &candidate
			latestOrder = candidateOrder
		}
	}
	return latest
}

func (s *MemoryStore) judgingWindowOpen(stunt Stunt) bool {
	if stunt.SeasonID == nil {
		return true
	}
	season, ok := s.seasons[*stunt.SeasonID]
	return ok && isOpenSeasonStatus(season.Status)
}

func (s *MemoryStore) submissionWindowOpen(stunt Stunt) bool {
	if stunt.SeasonID == nil {
		return true
	}
	season, ok := s.seasons[*stunt.SeasonID]
	return ok && season.Status == "Active" && s.now().Before(season.SubmissionDeadline)
}

func (s *MemoryStore) ensureSeasonStatus(season *Season) {
	if season.Status == "Active" && s.now().After(season.SubmissionDeadline) {
		season.Status = "Judging Grace Period"
	}
	if season.Status == "Judging Grace Period" && s.now().After(season.JudgingDeadline) {
		season.Status = "Finalized"
	}
}

func (s *MemoryStore) recordSeasonHistory(seasonID string, action string, actorPlayerID string, actorRole string, override bool, fromStatus string, toStatus string) {
	s.seasonEventNumber++
	s.seasonEvents[seasonID] = append(s.seasonEvents[seasonID], SeasonHistoryEntry{
		ID:            stableID("season_history", seasonID+":"+strconv.Itoa(s.seasonEventNumber)),
		SeasonID:      seasonID,
		Action:        action,
		ActorPlayerID: actorPlayerID,
		ActorRole:     actorRole,
		Override:      override,
		FromStatus:    fromStatus,
		ToStatus:      toStatus,
		CreatedAt:     s.now().UTC(),
	})
}

func isOpenSeasonStatus(status string) bool {
	return status == "Active" || status == "Judging Grace Period"
}

func validJudgmentScores(scores ...int) bool {
	for _, score := range scores {
		if score < 0 || score > 10 {
			return false
		}
	}
	return true
}

func randomToken(kind string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate %s: %w", kind, err)
	}
	return kind + "_" + hex.EncodeToString(bytes), nil
}

const httpMethodPut = "PUT"

func groupHome(group Group, membership GroupMembership, activeSeason *Season, recentStunts []PerformedStuntView, standings []StandingEntry) GroupHomeResponse {
	return GroupHomeResponse{
		Group:        group,
		Membership:   membership,
		ActiveSeason: activeSeason,
		RecentStunts: recentStunts,
		Standings:    standings,
	}
}

func visiblePerformedStatus(status string) bool {
	return status == "Performed Stunt" || status == "Judged Stunt" || status == "Unjudged Stunt"
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
