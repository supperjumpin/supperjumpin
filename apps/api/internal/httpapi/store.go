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

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

var ErrSeasonAlreadyOpen = errors.New("Group already has an active or closing Season")

var ErrSeasonNotFound = errors.New("Season not found")

var ErrJumpNotFound = errors.New("Jump not found")

var ErrEvidenceUploadAuthorizationNotFound = errors.New("Evidence upload authorization not found")

var ErrJudgingWindowClosed = errors.New("Judging Window closed")

var ErrSubmissionWindowClosed = errors.New("Submission Window closed")

var ErrInvalidJudgmentScore = errors.New("Judgment scores must be between 0 and 10")

var ErrInvalidDisputeConcern = errors.New("Dispute concern must be House Rules, Credibility, Source, Destination, Food, duplicate, or other")

var ErrDisputeNotFound = errors.New("Dispute not found")

var ErrInvalidDisputeResolution = errors.New("Dispute resolution must be No Action, Disqualified Jump, or Removed Jump")

// mapGameErr translates game-module typed errors into the corresponding
// httpapi sentinel errors so HTTP handlers can use their existing errors.Is checks.
func mapGameErr(err error) error {
	if errors.Is(err, game.ErrInvalidJudgmentScore) {
		return ErrInvalidJudgmentScore
	}
	if errors.Is(err, game.ErrStuntNotFound) {
		return ErrJumpNotFound
	}
	if errors.Is(err, game.ErrJudgingWindowClosed) {
		return ErrJudgingWindowClosed
	}
	if errors.Is(err, game.ErrEvidenceUploadAuthorizationNotFound) {
		return ErrEvidenceUploadAuthorizationNotFound
	}
	if errors.Is(err, game.ErrSubmissionWindowClosed) {
		return ErrSubmissionWindowClosed
	}
	if errors.Is(err, game.ErrSeasonAlreadyOpen) {
		return ErrSeasonAlreadyOpen
	}
	if errors.Is(err, game.ErrSeasonNotFound) {
		return ErrSeasonNotFound
	}
	if errors.Is(err, game.ErrInvalidDisputeConcern) {
		return ErrInvalidDisputeConcern
	}
	if errors.Is(err, game.ErrInvalidDisputeResolution) {
		return ErrInvalidDisputeResolution
	}
	if errors.Is(err, game.ErrDisputeNotFound) {
		return ErrDisputeNotFound
	}
	return err
}

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
	RecentJumps  []PerformedJumpView  `json:"recentJumps"`
	Standings    []StandingEntry      `json:"standings"`
}

type StandingEntry struct {
	Player      Player `json:"player"`
	SeasonScore int    `json:"seasonScore"`
	JudgedJumps int    `json:"judgedJumps"`
}

type PerformedJumpView struct {
	Jump      Jump      `json:"jump"`
	Performer Player    `json:"performer"`
	Evidence  Evidence  `json:"evidence"`
	Disputes  []Dispute `json:"disputes"`
}

type Jump struct {
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
	JumpID         string            `json:"jumpId"`
	UploadURL      string            `json:"uploadUrl"`
	UploadMethod   string            `json:"uploadMethod"`
	UploadHeaders  map[string]string `json:"uploadHeaders"`
	MediaObjectKey string            `json:"mediaObjectKey"`
	ExpiresAt      time.Time         `json:"expiresAt"`
}

type Evidence struct {
	ID             string    `json:"id"`
	JumpID         string    `json:"jumpId"`
	Caption        string    `json:"caption"`
	MediaObjectKey string    `json:"mediaObjectKey"`
	CreatedAt      time.Time `json:"createdAt"`
}

type EvidenceSubmission struct {
	Jump     Jump     `json:"jump"`
	Evidence Evidence `json:"evidence"`
}

type Judgment struct {
	ID            string `json:"id"`
	JumpID        string `json:"jumpId"`
	PlayerID      string `json:"playerId"`
	Difficulty    int    `json:"difficulty"`
	Transgression int    `json:"transgression"`
	Creativity    int    `json:"creativity"`
	Presentation  int    `json:"presentation"`
}

type Dispute struct {
	ID                 string  `json:"id"`
	JumpID             string  `json:"jumpId"`
	RaisedByPlayerID   string  `json:"raisedByPlayerId"`
	Concern            string  `json:"concern"`
	Details            string  `json:"details"`
	Status             string  `json:"status"`
	Resolution         *string `json:"resolution"`
	ResolutionReason   *string `json:"resolutionReason"`
	ResolvedByPlayerID *string `json:"resolvedByPlayerId"`
	OverrideResolution *string `json:"overrideResolution"`
	OverrideReason     *string `json:"overrideReason"`
	OverrideByPlayerID *string `json:"overrideByPlayerId"`
}

type DisputeResolution struct {
	Jump    Jump    `json:"jump"`
	Dispute Dispute `json:"dispute"`
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

// Store handles identity persistence.
type Store interface {
	BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error)
}

// Persistence combines game repository interfaces with transport-layer DTO
// assembly queries. Both MemoryStore and PostgresStore implement this interface.
type Persistence interface {
	game.GroupRepository
	game.StuntPlanningRepository
	game.EvidenceRepository
	game.JudgmentRepository
	game.SeasonRepository
	game.DisputeRepository

	GroupHomeForGroup(ctx context.Context, groupID string, player Player) (GroupHomeResponse, bool, error)
	GroupHomeForSeason(ctx context.Context, seasonID string, player Player) (GroupHomeResponse, bool, error)
	Now() time.Time
}

// --- Transport-layer DTO helpers (game-command → DTO conversion) ---

func createGroup(ctx context.Context, db Persistence, player Player, name string) (GroupHomeResponse, error) {
	groupID := stableID("group", player.ID+":"+name+":"+strconv.FormatInt(time.Now().UnixNano(), 10))

	result := game.CreateGroup(ctx, db, game.CreateGroupInput{
		GroupID:         groupID,
		GroupName:       name,
		CreatorPlayerID: player.ID,
	})
	if result.Err != nil {
		return GroupHomeResponse{}, result.Err
	}
	return groupHome(
		Group{ID: result.Group.ID, Name: result.Group.Name},
		GroupMembership{GroupID: result.Membership.GroupID, PlayerID: result.Membership.PlayerID, Role: result.Membership.Role},
		nil, []PerformedJumpView{}, []StandingEntry{},
	), nil
}

func groupHomeHandler(ctx context.Context, db Persistence, player Player, groupID string) (GroupHomeResponse, bool, error) {
	ghResult := game.GroupHome(ctx, db, game.GroupHomeInput{
		PlayerID: player.ID,
		GroupID:  groupID,
	})
	if ghResult.Err != nil {
		return GroupHomeResponse{}, false, ghResult.Err
	}
	if !ghResult.Allowed {
		return GroupHomeResponse{}, false, nil
	}

	return db.GroupHomeForGroup(ctx, groupID, player)
}

func listGroups(ctx context.Context, db Persistence, player Player) (ListGroupsResponse, error) {
	memberships, err := game.ListGroups(ctx, db, player.ID)
	if err != nil {
		return ListGroupsResponse{}, err
	}

	result := make([]GroupMembershipSummary, len(memberships))
	for i, m := range memberships {
		result[i] = GroupMembershipSummary{
			Group:      Group{ID: m.Group.ID, Name: m.Group.Name},
			Membership: GroupMembership{GroupID: m.Membership.GroupID, PlayerID: m.Membership.PlayerID, Role: m.Membership.Role},
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Group.Name < result[j].Group.Name
	})
	return ListGroupsResponse{Memberships: result}, nil
}

func createInvite(ctx context.Context, db Persistence, player Player, groupID string) (Invite, bool, error) {
	result := game.CreateInvite(ctx, db, game.CreateInviteInput{
		GroupID:  groupID,
		PlayerID: player.ID,
		Now:      db.Now(),
	})
	if result.Err != nil {
		return Invite{}, false, result.Err
	}
	if !result.Allowed {
		return Invite{}, false, nil
	}
	return Invite{
		ID:        result.Invite.ID,
		GroupID:   result.Invite.GroupID,
		Token:     result.Invite.Token,
		CreatedBy: result.Invite.CreatedBy,
		ExpiresAt: time.Unix(result.Invite.ExpiresAt, 0).UTC(),
	}, true, nil
}

func acceptInvite(ctx context.Context, db Persistence, player Player, token string) (GroupHomeResponse, InviteAcceptStatus, error) {
	result := game.AcceptInvite(ctx, db, game.AcceptInviteInput{
		Token:    token,
		PlayerID: player.ID,
		Now:      db.Now(),
	})
	if result.Err != nil {
		return GroupHomeResponse{}, InviteInvalid, result.Err
	}
	if result.Status != game.InviteAccepted {
		switch result.Status {
		case game.InviteInvalid:
			return GroupHomeResponse{}, InviteInvalid, nil
		case game.InviteUsed:
			return GroupHomeResponse{}, InviteUsed, nil
		case game.InviteExpired:
			return GroupHomeResponse{}, InviteExpired, nil
		case game.InviteMember:
			return GroupHomeResponse{}, InviteMember, nil
		default:
			return GroupHomeResponse{}, InviteInvalid, nil
		}
	}

	ghResult, _, err := db.GroupHomeForGroup(ctx, result.Group.ID, player)
	if err != nil {
		return GroupHomeResponse{}, InviteInvalid, err
	}
	return ghResult, InviteAccepted, nil
}

func startSeason(ctx context.Context, db Persistence, player Player, groupID string, submissionDeadline time.Time, judgingDeadline time.Time) (GroupHomeResponse, bool, error) {
	result := game.StartSeason(ctx, db, game.StartSeasonInput{
		GroupID:            groupID,
		PlayerID:           player.ID,
		SubmissionDeadline: submissionDeadline,
		JudgingDeadline:    judgingDeadline,
	})
	if result.Err != nil {
		return GroupHomeResponse{}, true, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return GroupHomeResponse{}, false, nil
	}
	return db.GroupHomeForGroup(ctx, groupID, player)
}

func closeSeasonSubmissions(ctx context.Context, db Persistence, player Player, seasonID string) (GroupHomeResponse, bool, error) {
	result := game.CloseSeasonSubmissions(ctx, db, game.CloseSeasonSubmissionsInput{
		SeasonID: seasonID,
		PlayerID: player.ID,
	})
	if result.Err != nil {
		return GroupHomeResponse{}, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return GroupHomeResponse{}, false, nil
	}
	return db.GroupHomeForSeason(ctx, seasonID, player)
}

func finalizeSeason(ctx context.Context, db Persistence, player Player, seasonID string) (GroupHomeResponse, bool, error) {
	result := game.FinalizeSeason(ctx, db, game.FinalizeSeasonInput{
		SeasonID: seasonID,
		PlayerID: player.ID,
	})
	if result.Err != nil {
		return GroupHomeResponse{}, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return GroupHomeResponse{}, false, nil
	}
	return db.GroupHomeForSeason(ctx, seasonID, player)
}

func seasonHistory(ctx context.Context, db Persistence, player Player, seasonID string) (SeasonHistoryResponse, bool, error) {
	result := game.SeasonHistory(ctx, db, game.SeasonHistoryInput{
		SeasonID: seasonID,
		PlayerID: player.ID,
	})
	if result.Err != nil {
		return SeasonHistoryResponse{}, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return SeasonHistoryResponse{}, false, nil
	}
	entries := make([]SeasonHistoryEntry, len(result.Entries))
	for i, e := range result.Entries {
		entries[i] = SeasonHistoryEntry{
			ID:            e.ID,
			SeasonID:      e.SeasonID,
			Action:        e.Action,
			ActorPlayerID: e.ActorPlayerID,
			ActorRole:     e.ActorRole,
			Override:      e.Override,
			FromStatus:    e.FromStatus,
			ToStatus:      e.ToStatus,
		}
	}
	return SeasonHistoryResponse{Entries: entries}, true, nil
}

func createIdea(ctx context.Context, db Persistence, player Player, groupID string, source string, destination string, food string) (Jump, bool, error) {
	result := game.CreateIdea(ctx, db, game.CreateIdeaInput{
		GroupID:     groupID,
		PlayerID:    player.ID,
		Source:      source,
		Destination: destination,
		Food:        food,
	})
	if result.Err != nil {
		return Jump{}, false, result.Err
	}
	if !result.Allowed {
		return Jump{}, false, nil
	}
	return stuntFromGame(result.Stunt), true, nil
}

func createPlannedStunt(ctx context.Context, db Persistence, player Player, ideaID string, offSeason bool) (Jump, bool, error) {
	result := game.CreatePlannedStunt(ctx, db, game.CreatePlannedStuntInput{
		IdeaID:    ideaID,
		PlayerID:  player.ID,
		OffSeason: offSeason,
	})
	if result.Err != nil {
		return Jump{}, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return Jump{}, false, nil
	}
	return stuntFromGame(result.Stunt), true, nil
}

func authorizeEvidenceUpload(ctx context.Context, db Persistence, player Player, stuntID string, contentType string) (EvidenceUploadAuthorization, bool, error) {
	result := game.AuthorizeEvidenceUpload(ctx, db, game.AuthorizeEvidenceUploadInput{
		StuntID:     stuntID,
		PlayerID:    player.ID,
		ContentType: contentType,
	})
	if result.Err != nil {
		return EvidenceUploadAuthorization{}, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return EvidenceUploadAuthorization{}, false, nil
	}
	return EvidenceUploadAuthorization{
		ID:             result.Authorization.ID,
		JumpID:         result.Authorization.StuntID,
		UploadURL:      "https://storage.supperjumpin.test/uploads/" + result.Authorization.MediaObjectKey,
		UploadMethod:   httpMethodPut,
		UploadHeaders:  map[string]string{"Content-Type": contentType},
		MediaObjectKey: result.Authorization.MediaObjectKey,
		ExpiresAt:      result.Authorization.ExpiresAt,
	}, true, nil
}

func submitEvidence(ctx context.Context, db Persistence, player Player, stuntID string, uploadAuthorizationID string, caption string) (EvidenceSubmission, bool, error) {
	result := game.SubmitEvidence(ctx, db, game.SubmitEvidenceInput{
		StuntID:               stuntID,
		PlayerID:              player.ID,
		UploadAuthorizationID: uploadAuthorizationID,
		Caption:               caption,
	}, db.Now())
	if result.Err != nil {
		return EvidenceSubmission{}, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return EvidenceSubmission{}, false, nil
	}
	return EvidenceSubmission{
		Jump: Jump{
			ID:          result.Stunt.ID,
			GroupID:     result.Stunt.GroupID,
			PlayerID:    result.Stunt.PlayerID,
			SeasonID:    result.Stunt.SeasonID,
			Status:      stuntStatusToJumpStatus(result.Stunt.Status),
			Source:      result.Stunt.Source,
			Destination: result.Stunt.Destination,
			Food:        result.Stunt.Food,
			OffSeason:   result.Stunt.SeasonID == nil,
		},
		Evidence: Evidence{
			ID:             result.Evidence.ID,
			JumpID:         result.Evidence.StuntID,
			Caption:        result.Evidence.Caption,
			MediaObjectKey: result.Evidence.MediaObjectKey,
			CreatedAt:      result.Evidence.CreatedAt,
		},
	}, true, nil
}

func submitJudgment(ctx context.Context, db Persistence, player Player, stuntID string, difficulty int, transgression int, creativity int, presentation int) (Judgment, bool, bool, error) {
	result := game.SubmitJudgment(ctx, db, game.JudgmentInput{
		StuntID:       stuntID,
		JudgePlayerID: player.ID,
		Difficulty:    difficulty,
		Transgression: transgression,
		Creativity:    creativity,
		Documentation: presentation,
	}, db.Now())
	if result.Err != nil {
		return Judgment{}, false, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return Judgment{}, false, false, nil
	}
	return Judgment{
		ID:            result.Judgment.ID,
		JumpID:        result.Judgment.StuntID,
		PlayerID:      result.Judgment.PlayerID,
		Difficulty:    result.Judgment.Difficulty,
		Transgression: result.Judgment.Transgression,
		Creativity:    result.Judgment.Creativity,
		Presentation:  result.Judgment.Documentation,
	}, true, result.Created, nil
}

func createDispute(ctx context.Context, db Persistence, player Player, stuntID string, concern string, details string) (Dispute, bool, error) {
	result := game.CreateDispute(ctx, db, game.CreateDisputeInput{
		PlayerID: player.ID,
		StuntID:  stuntID,
		Concern:  concern,
		Details:  details,
	})
	if result.Err != nil {
		return Dispute{}, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return Dispute{}, false, nil
	}
	dispute := disputeFromSnapshot(result.Dispute)
	dispute.RaisedByPlayerID = player.ID
	return dispute, true, nil
}

func resolveDispute(ctx context.Context, db Persistence, player Player, disputeID string, resolution string, resolutionReason string) (DisputeResolution, bool, error) {
	result := game.ResolveDispute(ctx, db, game.ResolveDisputeInput{
		PlayerID:         player.ID,
		DisputeID:        disputeID,
		Resolution:       jumpStatusToStuntStatus(resolution),
		ResolutionReason: resolutionReason,
	})
	if result.Err != nil {
		return DisputeResolution{}, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return DisputeResolution{}, false, nil
	}
	return DisputeResolution{
		Jump:    stuntFromGame(result.Stunt),
		Dispute: disputeFromSnapshot(result.Dispute),
	}, true, nil
}

// --- MemoryStore ---

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
	stunts            map[string]Jump
	uploads           map[string]EvidenceUploadAuthorization
	evidences         map[string]Evidence
	judgments         map[string]Judgment
	disputes          map[string]Dispute
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
		stunts:       map[string]Jump{},
		uploads:      map[string]EvidenceUploadAuthorization{},
		evidences:    map[string]Evidence{},
		judgments:    map[string]Judgment{},
		disputes:     map[string]Dispute{},
		now:          now,
	}
}

func (s *MemoryStore) Now() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.now()
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

func (s *MemoryStore) GroupHomeForGroup(ctx context.Context, groupID string, player Player) (GroupHomeResponse, bool, error) {
	s.mu.Lock()

	group, ok := s.groups[groupID]
	if !ok {
		s.mu.Unlock()
		return GroupHomeResponse{}, false, nil
	}
	membership, ok := s.memberships[groupID][player.ID]
	if !ok {
		s.mu.Unlock()
		return GroupHomeResponse{}, false, nil
	}
	newlyFinalized := s.ensureSeasonStatusesForGroup(groupID)
	season := s.currentSeasonForGroup(groupID)
	s.mu.Unlock()

	for _, id := range newlyFinalized {
		game.AutoFinalizeSeason(context.Background(), s, id)
	}

	standings, err := game.Standings(context.Background(), s, groupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}

	s.mu.Lock()
	recentStunts := s.recentPerformedStuntsForGroup(groupID)
	s.mu.Unlock()
	return groupHome(group, membership, season, recentStunts, standingsFromGame(standings)), true, nil
}

func (s *MemoryStore) GroupHomeForSeason(ctx context.Context, seasonID string, player Player) (GroupHomeResponse, bool, error) {
	s.mu.Lock()

	season, ok := s.seasons[seasonID]
	if !ok {
		s.mu.Unlock()
		return GroupHomeResponse{}, false, nil
	}
	group, ok := s.groups[season.GroupID]
	if !ok {
		s.mu.Unlock()
		return GroupHomeResponse{}, false, nil
	}
	membership, ok := s.memberships[season.GroupID][player.ID]
	if !ok {
		s.mu.Unlock()
		return GroupHomeResponse{}, false, nil
	}
	newlyFinalized := s.ensureSeasonStatusesForGroup(season.GroupID)
	s.mu.Unlock()

	for _, id := range newlyFinalized {
		game.AutoFinalizeSeason(context.Background(), s, id)
	}

	standings, err := game.Standings(context.Background(), s, season.GroupID)
	if err != nil {
		return GroupHomeResponse{}, false, err
	}

	s.mu.Lock()
	recentStunts := s.recentPerformedStuntsForGroup(season.GroupID)
	s.mu.Unlock()
	return groupHome(group, membership, s.currentSeasonForGroup(season.GroupID), recentStunts, standingsFromGame(standings)), true, nil
}

// game.JudgmentRepository adapter

func (s *MemoryStore) Stunt(ctx context.Context, stuntID string) (game.StuntSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stunt, ok := s.stunts[stuntID]
	if !ok || !visiblePerformedStatus(stunt.Status) {
		return game.StuntSnapshot{}, false, nil
	}
	return stuntToSnapshot(stunt), true, nil
}

func (s *MemoryStore) Season(ctx context.Context, seasonID string) (game.SeasonSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	season, ok := s.seasons[seasonID]
	if !ok {
		return game.SeasonSnapshot{}, nil
	}
	return seasonToSnapshot(season), nil
}

func seasonToSnapshot(season Season) game.SeasonSnapshot {
	return game.SeasonSnapshot{
		ID:                   season.ID,
		GroupID:              season.GroupID,
		CommissionerPlayerID: season.CommissionerPlayerID,
		Status:               season.Status,
		SubmissionDeadline:   season.SubmissionDeadline,
		JudgingDeadline:      season.JudgingDeadline,
	}
}

func (s *MemoryStore) GroupMembership(ctx context.Context, playerID, groupID string) (game.MembershipSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, ok := s.memberships[groupID][playerID]
	if !ok {
		return game.MembershipSnapshot{}, false, nil
	}
	return game.MembershipSnapshot{Role: m.Role}, true, nil
}

// game.GroupRepository adapter methods for MemoryStore

func (s *MemoryStore) InsertGroup(ctx context.Context, groupID, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.groups[groupID] = Group{ID: groupID, Name: name}
	if s.memberships[groupID] == nil {
		s.memberships[groupID] = map[string]GroupMembership{}
	}
	return nil
}

func (s *MemoryStore) InsertMembership(_ context.Context, groupID, playerID, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.memberships[groupID] == nil {
		s.memberships[groupID] = map[string]GroupMembership{}
	}
	s.memberships[groupID][playerID] = GroupMembership{GroupID: groupID, PlayerID: playerID, Role: role}
	return nil
}

func (s *MemoryStore) Group(ctx context.Context, groupID string) (game.GroupSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.groups[groupID]
	if !ok {
		return game.GroupSnapshot{}, false, nil
	}
	return game.GroupSnapshot{ID: g.ID, Name: g.Name}, true, nil
}

func (s *MemoryStore) Membership(ctx context.Context, playerID, groupID string) (game.GroupMembershipSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, ok := s.memberships[groupID][playerID]
	if !ok {
		return game.GroupMembershipSnapshot{}, false, nil
	}
	return game.GroupMembershipSnapshot{GroupID: m.GroupID, PlayerID: m.PlayerID, Role: m.Role}, true, nil
}

func (s *MemoryStore) MembershipsForPlayer(ctx context.Context, playerID string) ([]game.MembershipWithGroupSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []game.MembershipWithGroupSnapshot
	for groupID, groupMemberships := range s.memberships {
		m, ok := groupMemberships[playerID]
		if !ok {
			continue
		}
		g := s.groups[groupID]
		result = append(result, game.MembershipWithGroupSnapshot{
			Group:      game.GroupSnapshot{ID: g.ID, Name: g.Name},
			Membership: game.GroupMembershipSnapshot{GroupID: m.GroupID, PlayerID: m.PlayerID, Role: m.Role},
		})
	}
	return result, nil
}

func (s *MemoryStore) InsertInvite(ctx context.Context, groupID, createdByPlayerID string, expiresAt int64) (game.InviteSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.inviteNumber++
	token, err := randomToken("invite_token")
	if err != nil {
		return game.InviteSnapshot{}, err
	}
	invite := Invite{
		ID:        stableID("invite", groupID+":"+strconv.Itoa(s.inviteNumber)),
		GroupID:   groupID,
		Token:     token,
		CreatedBy: createdByPlayerID,
		ExpiresAt: time.Unix(expiresAt, 0).UTC(),
	}
	s.invites[invite.Token] = memoryInvite{Invite: invite}
	return game.InviteSnapshot{
		ID:        invite.ID,
		GroupID:   invite.GroupID,
		Token:     invite.Token,
		CreatedBy: invite.CreatedBy,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *MemoryStore) InviteByToken(ctx context.Context, token string) (game.InviteSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	invite, ok := s.invites[token]
	if !ok {
		return game.InviteSnapshot{}, false, nil
	}
	var usedBy *string
	if invite.UsedBy != "" {
		usedBy = &invite.UsedBy
	}
	return game.InviteSnapshot{
		ID:        invite.ID,
		GroupID:   invite.GroupID,
		Token:     invite.Token,
		CreatedBy: invite.CreatedBy,
		ExpiresAt: invite.ExpiresAt.Unix(),
		UsedBy:    usedBy,
	}, true, nil
}

func (s *MemoryStore) MarkInviteUsed(ctx context.Context, token, playerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	invite := s.invites[token]
	invite.UsedBy = playerID
	s.invites[token] = invite
	return nil
}

func (s *MemoryStore) StuntByID(ctx context.Context, stuntID string) (game.StuntSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stunt, ok := s.stunts[stuntID]
	if !ok {
		return game.StuntSnapshot{}, false, nil
	}
	return stuntToSnapshot(stunt), true, nil
}

func (s *MemoryStore) UpsertJudgment(ctx context.Context, stuntID, playerID string, difficulty, transgression, creativity, documentation int) (game.Judgment, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := stuntID + ":" + playerID
	_, existed := s.judgments[key]
	httpJudgment := Judgment{
		ID:            stableID("judgment", key),
		JumpID:        stuntID,
		PlayerID:      playerID,
		Difficulty:    difficulty,
		Transgression: transgression,
		Creativity:    creativity,
		Presentation:  documentation,
	}
	s.judgments[key] = httpJudgment
	return game.Judgment{
		ID:            httpJudgment.ID,
		StuntID:       httpJudgment.JumpID,
		PlayerID:      httpJudgment.PlayerID,
		Difficulty:    httpJudgment.Difficulty,
		Transgression: httpJudgment.Transgression,
		Creativity:    httpJudgment.Creativity,
		Documentation: httpJudgment.Presentation,
	}, !existed, nil
}

// stuntFromGame converts a game.StuntSnapshot to the httpapi Jump type.
func stuntFromGame(snap game.StuntSnapshot) Jump {
	return Jump{
		ID:          snap.ID,
		GroupID:     snap.GroupID,
		PlayerID:    snap.PlayerID,
		Status:      stuntStatusToJumpStatus(snap.Status),
		SeasonID:    snap.SeasonID,
		Source:      snap.Source,
		Destination: snap.Destination,
		Food:        snap.Food,
		OffSeason:   snap.SeasonID == nil,
	}
}

func stuntToSnapshot(stunt Jump) game.StuntSnapshot {
	return game.StuntSnapshot{
		ID:          stunt.ID,
		GroupID:     stunt.GroupID,
		PlayerID:    stunt.PlayerID,
		Status:      jumpStatusToStuntStatus(stunt.Status),
		SeasonID:    stunt.SeasonID,
		Source:      stunt.Source,
		Destination: stunt.Destination,
		Food:        stunt.Food,
		FinalScore:  stunt.FinalScore,
	}
}

// game.StuntPlanningRepository adapter methods

func (s *MemoryStore) InsertIdea(ctx context.Context, groupID, playerID, source, destination, food string) (game.StuntSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stuntNumber++
	stunt := Jump{
		ID:          stableID("stunt", groupID+":"+playerID+":"+strconv.Itoa(s.stuntNumber)),
		GroupID:     groupID,
		PlayerID:    playerID,
		Status:      "Idea",
		Source:      source,
		Destination: destination,
		Food:        food,
		OffSeason:   true,
	}
	s.stunts[stunt.ID] = stunt
	return stuntToSnapshot(stunt), nil
}

func (s *MemoryStore) Idea(ctx context.Context, stuntID string) (game.StuntSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stunt, ok := s.stunts[stuntID]
	if !ok {
		return game.StuntSnapshot{}, false, nil
	}
	return stuntToSnapshot(stunt), true, nil
}

func (s *MemoryStore) ActiveSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	season := s.activeSeasonForGroup(groupID)
	if season == nil {
		return game.SeasonSnapshot{}, nil
	}
	return game.SeasonSnapshot{ID: season.ID, Status: season.Status}, nil
}

func (s *MemoryStore) UpdateStuntToPlanned(ctx context.Context, stuntID, playerID string, seasonID *string) (game.StuntSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stunt := s.stunts[stuntID]
	stunt.Status = "Planned Stunt"
	stunt.SeasonID = seasonID
	stunt.OffSeason = seasonID == nil
	s.stunts[stuntID] = stunt
	return stuntToSnapshot(stunt), nil
}

// game.EvidenceRepository adapter methods for MemoryStore

func (s *MemoryStore) PlannedStunt(ctx context.Context, stuntID string) (game.StuntSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stunt, ok := s.stunts[stuntID]
	if !ok || stunt.Status != "Planned Stunt" {
		return game.StuntSnapshot{}, false, nil
	}
	return stuntToSnapshot(stunt), true, nil
}

func (s *MemoryStore) CreateAuthorization(ctx context.Context, stuntID, playerID, contentType string) (game.AuthorizationSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.uploadNumber++
	now := s.now()
	id := stableID("evidence_upload", stuntID+":"+strconv.Itoa(s.uploadNumber))
	mediaObjectKey := "uploads/" + stuntID + "/" + strconv.Itoa(s.uploadNumber)
	auth := EvidenceUploadAuthorization{
		ID:             id,
		JumpID:         stuntID,
		UploadURL:      "https://storage.supperjumpin.test/uploads/" + mediaObjectKey,
		UploadMethod:   httpMethodPut,
		UploadHeaders:  map[string]string{"Content-Type": contentType},
		MediaObjectKey: mediaObjectKey,
		ExpiresAt:      now.Add(15 * time.Minute).UTC(),
	}
	s.uploads[id] = auth
	return game.AuthorizationSnapshot{
		ID:             id,
		StuntID:        stuntID,
		MediaObjectKey: mediaObjectKey,
		ExpiresAt:      auth.ExpiresAt,
	}, nil
}

func (s *MemoryStore) ClaimAndAdvance(ctx context.Context, authorizationID, stuntID, playerID, caption string) (game.EvidenceCreateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	auth, ok := s.uploads[authorizationID]
	if !ok || auth.JumpID != stuntID || s.now().After(auth.ExpiresAt) {
		return game.EvidenceCreateResult{}, game.ErrEvidenceUploadAuthorizationNotFound
	}
	delete(s.uploads, authorizationID)

	stunt := s.stunts[stuntID]
	stunt.Status = "Performed Stunt"
	s.stunts[stunt.ID] = stunt

	evidenceID := stableID("evidence", stuntID+":"+authorizationID)
	s.evidences[stuntID] = Evidence{
		ID:             evidenceID,
		JumpID:         stuntID,
		Caption:        caption,
		MediaObjectKey: auth.MediaObjectKey,
		CreatedAt:      s.now().UTC(),
	}
	return game.EvidenceCreateResult{
		EvidenceID:     evidenceID,
		MediaObjectKey: auth.MediaObjectKey,
	}, nil
}

// game.SeasonRepository adapter methods for MemoryStore

func (s *MemoryStore) OpenSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	season := s.openSeasonForGroup(groupID)
	if season == nil {
		return game.SeasonSnapshot{}, nil
	}
	return seasonToSnapshot(*season), nil
}

func (s *MemoryStore) InsertSeason(ctx context.Context, groupID, commissionerPlayerID string, submissionDeadline, judgingDeadline time.Time) (game.SeasonSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.seasonNumber++
	season := Season{
		ID:                   stableID("season", groupID+":"+strconv.Itoa(s.seasonNumber)),
		GroupID:              groupID,
		CommissionerPlayerID: commissionerPlayerID,
		Status:               "Active",
		SubmissionDeadline:   submissionDeadline.UTC(),
		JudgingDeadline:      judgingDeadline.UTC(),
	}
	s.seasons[season.ID] = season
	s.seasonOrder[season.ID] = s.seasonNumber
	return seasonToSnapshot(season), nil
}

func (s *MemoryStore) UpdateSeasonStatus(ctx context.Context, seasonID, action, actorPlayerID, actorRole string, override bool, fromStatus, toStatus string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	season := s.seasons[seasonID]
	season.Status = toStatus
	s.seasons[seasonID] = season
	s.recordSeasonHistory(seasonID, action, actorPlayerID, actorRole, override, fromStatus, toStatus)
	return nil
}

func (s *MemoryStore) StuntsForSeason(ctx context.Context, seasonID string) ([]game.StuntSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []game.StuntSnapshot
	for _, stunt := range s.stunts {
		if stunt.SeasonID != nil && *stunt.SeasonID == seasonID {
			result = append(result, stuntToSnapshot(stunt))
		}
	}
	return result, nil
}

func (s *MemoryStore) JudgmentsForStunt(ctx context.Context, stuntID string) ([]game.Judgment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []game.Judgment
	for _, j := range s.judgments {
		if j.JumpID == stuntID {
			result = append(result, game.Judgment{
				ID:            j.ID,
				StuntID:       j.JumpID,
				PlayerID:      j.PlayerID,
				Difficulty:    j.Difficulty,
				Transgression: j.Transgression,
				Creativity:    j.Creativity,
				Documentation: j.Presentation,
			})
		}
	}
	return result, nil
}

func (s *MemoryStore) UpdateStuntFinalization(ctx context.Context, stuntID string, status string, finalScore *int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stunt := s.stunts[stuntID]
	stunt.Status = status
	stunt.FinalScore = finalScore
	s.stunts[stuntID] = stunt
	return nil
}

func (s *MemoryStore) LatestSeasonForGroup(ctx context.Context, groupID string) (game.SeasonSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	season := s.latestSeasonForGroup(groupID)
	if season == nil {
		return game.SeasonSnapshot{}, nil
	}
	return seasonToSnapshot(*season), nil
}

func (s *MemoryStore) GroupPlayers(ctx context.Context, groupID string) ([]game.PlayerSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []game.PlayerSnapshot
	for _, membership := range s.memberships[groupID] {
		player := s.players[membership.PlayerID]
		result = append(result, game.PlayerSnapshot{
			ID:          player.ID,
			DisplayName: player.DisplayName,
		})
	}
	return result, nil
}

func (s *MemoryStore) SeasonHistoryEntries(ctx context.Context, seasonID string) ([]game.SeasonHistoryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := s.seasonEvents[seasonID]
	result := make([]game.SeasonHistoryEntry, len(entries))
	for i, e := range entries {
		result[i] = game.SeasonHistoryEntry{
			ID:            e.ID,
			SeasonID:      e.SeasonID,
			Action:        e.Action,
			ActorPlayerID: e.ActorPlayerID,
			ActorRole:     e.ActorRole,
			Override:      e.Override,
			FromStatus:    e.FromStatus,
			ToStatus:      e.ToStatus,
		}
	}
	return result, nil
}

// game.DisputeRepository adapter methods for MemoryStore

func (s *MemoryStore) InsertDispute(ctx context.Context, stuntID, raisedByPlayerID, concern, details string) (game.DisputeSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := stableID("dispute", stuntID+":"+raisedByPlayerID+":"+strconv.Itoa(len(s.disputes)+1))
	dispute := Dispute{
		ID:               id,
		JumpID:           stuntID,
		RaisedByPlayerID: raisedByPlayerID,
		Concern:          concern,
		Details:          details,
		Status:           "Open",
	}
	s.disputes[dispute.ID] = dispute
	return disputeToSnapshot(dispute), nil
}

func (s *MemoryStore) Dispute(ctx context.Context, disputeID string) (game.DisputeSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dispute, ok := s.disputes[disputeID]
	if !ok {
		return game.DisputeSnapshot{}, nil
	}
	return disputeToSnapshot(dispute), nil
}

func (s *MemoryStore) UpdateDisputeResolution(ctx context.Context, disputeID, resolution, resolutionReason, resolvedByPlayerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dispute := s.disputes[disputeID]
	dispute.Status = "Resolved"
	dispute.Resolution = &resolution
	dispute.ResolutionReason = &resolutionReason
	dispute.ResolvedByPlayerID = &resolvedByPlayerID
	s.disputes[disputeID] = dispute
	return nil
}

func (s *MemoryStore) UpdateDisputeOverride(ctx context.Context, disputeID, overrideResolution, overrideReason, overrideByPlayerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dispute := s.disputes[disputeID]
	dispute.Status = "Overridden"
	dispute.OverrideResolution = &overrideResolution
	dispute.OverrideReason = &overrideReason
	dispute.OverrideByPlayerID = &overrideByPlayerID
	s.disputes[disputeID] = dispute
	return nil
}

func (s *MemoryStore) UpdateStuntStatusAfterDispute(ctx context.Context, stuntID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stunt := s.stunts[stuntID]
	stunt.Status = status
	stunt.FinalScore = nil
	s.stunts[stunt.ID] = stunt
	return nil
}

// --- MemoryStore private helpers ---

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

func (s *MemoryStore) ensureSeasonStatusesForGroup(groupID string) []string {
	var newlyFinalized []string
	for id, season := range s.seasons {
		if season.GroupID != groupID {
			continue
		}
		wasFinalized := season.Status == "Finalized"
		s.ensureSeasonStatus(&season)
		s.seasons[id] = season
		if !wasFinalized && season.Status == "Finalized" {
			newlyFinalized = append(newlyFinalized, season.ID)
		}
	}
	return newlyFinalized
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

func (s *MemoryStore) recentPerformedStuntsForGroup(groupID string) []PerformedJumpView {
	performed := []PerformedJumpView{}
	for _, stunt := range s.stunts {
		if stunt.GroupID != groupID || !visiblePerformedStatus(stunt.Status) {
			continue
		}
		evidence, ok := s.evidences[stunt.ID]
		if !ok {
			continue
		}
		performed = append(performed, PerformedJumpView{
			Jump:      jumpForResponse(stunt),
			Performer: s.players[stunt.PlayerID],
			Evidence:  evidence,
			Disputes:  s.disputesForStunt(stunt.ID),
		})
	}
	sort.Slice(performed, func(i, j int) bool {
		if performed[i].Evidence.CreatedAt.Equal(performed[j].Evidence.CreatedAt) {
			return performed[i].Jump.ID > performed[j].Jump.ID
		}
		return performed[i].Evidence.CreatedAt.After(performed[j].Evidence.CreatedAt)
	})
	return performed
}

func (s *MemoryStore) disputesForStunt(stuntID string) []Dispute {
	disputes := []Dispute{}
	for _, dispute := range s.disputes {
		if dispute.JumpID != stuntID {
			continue
		}
		disputes = append(disputes, dispute)
	}
	sort.Slice(disputes, func(i, j int) bool {
		return disputes[i].ID < disputes[j].ID
	})
	return disputes
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

// --- Package-level helpers ---

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

func standingsFromGame(entries []game.StandingEntry) []StandingEntry {
	result := make([]StandingEntry, len(entries))
	for i, e := range entries {
		result[i] = StandingEntry{
			Player:      Player{ID: e.PlayerID, DisplayName: e.DisplayName},
			SeasonScore: e.SeasonScore,
			JudgedJumps: e.JudgedStunts,
		}
	}
	return result
}

func groupHome(group Group, membership GroupMembership, activeSeason *Season, recentJumps []PerformedJumpView, standings []StandingEntry) GroupHomeResponse {
	return GroupHomeResponse{
		Group:        group,
		Membership:   membership,
		ActiveSeason: activeSeason,
		RecentJumps:  recentJumps,
		Standings:    standings,
	}
}

func disputeToSnapshot(d Dispute) game.DisputeSnapshot {
	return game.DisputeSnapshot{
		ID:                   d.ID,
		StuntID:              d.JumpID,
		RaisedByPlayerID:     d.RaisedByPlayerID,
		Concern:              d.Concern,
		Details:              d.Details,
		Status:               d.Status,
		Resolution:           mapOptionalStatus(d.Resolution, jumpStatusToStuntStatus),
		ResolutionReason:     d.ResolutionReason,
		ResolvedByPlayerID:   d.ResolvedByPlayerID,
		OverrideResolution:   mapOptionalStatus(d.OverrideResolution, jumpStatusToStuntStatus),
		OverrideReason:       d.OverrideReason,
		OverrideByPlayerID:   d.OverrideByPlayerID,
	}
}

func disputeFromSnapshot(snap game.DisputeSnapshot) Dispute {
	return Dispute{
		ID:                   snap.ID,
		JumpID:               snap.StuntID,
		RaisedByPlayerID:     snap.RaisedByPlayerID,
		Concern:              snap.Concern,
		Details:              snap.Details,
		Status:               snap.Status,
		Resolution:           mapOptionalStatus(snap.Resolution, stuntStatusToJumpStatus),
		ResolutionReason:     snap.ResolutionReason,
		ResolvedByPlayerID:   snap.ResolvedByPlayerID,
		OverrideResolution:   mapOptionalStatus(snap.OverrideResolution, stuntStatusToJumpStatus),
		OverrideReason:       snap.OverrideReason,
		OverrideByPlayerID:   snap.OverrideByPlayerID,
	}
}

func visiblePerformedStatus(status string) bool {
	return status == "Performed Stunt" || status == "Judged Stunt" || status == "Unjudged Stunt" || status == "Disqualified Stunt"
}

func jumpForResponse(jump Jump) Jump {
	jump.Status = stuntStatusToJumpStatus(jump.Status)
	return jump
}

func stuntStatusToJumpStatus(status string) string {
	switch status {
	case "Planned Stunt":
		return "Planned Jump"
	case "Performed Stunt":
		return "Performed Jump"
	case "Judged Stunt":
		return "Judged Jump"
	case "Unjudged Stunt":
		return "Unjudged Jump"
	case "Disqualified Stunt":
		return "Disqualified Jump"
	case "Removed Stunt":
		return "Removed Jump"
	default:
		return status
	}
}

func jumpStatusToStuntStatus(status string) string {
	switch status {
	case "Planned Jump":
		return "Planned Stunt"
	case "Performed Jump":
		return "Performed Stunt"
	case "Judged Jump":
		return "Judged Stunt"
	case "Unjudged Jump":
		return "Unjudged Stunt"
	case "Disqualified Jump":
		return "Disqualified Stunt"
	case "Removed Jump":
		return "Removed Stunt"
	default:
		return status
	}
}

func mapOptionalStatus(value *string, mapper func(string) string) *string {
	if value == nil {
		return nil
	}
	mapped := mapper(*value)
	return &mapped
}

func stringPointer(value string) *string {
	return &value
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
