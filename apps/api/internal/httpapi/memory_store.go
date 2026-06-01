package httpapi

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

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
	jumps             map[string]Jump
	uploads           map[string]EvidenceUploadAuthorization
	evidences         map[string]Evidence
	judgments         map[string]Judgment
	guestSessions     map[string]GuestSession
	disputes          map[string]Dispute
	now               func() time.Time
	groupNumber       int
	inviteNumber      int
	seasonNumber      int
	jumpNumber        int
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
		accounts:      map[string]MeResponse{},
		players:       map[string]Player{},
		groups:        map[string]Group{},
		memberships:   map[string]map[string]GroupMembership{},
		invites:       map[string]memoryInvite{},
		seasons:       map[string]Season{},
		seasonEvents:  map[string][]SeasonHistoryEntry{},
		seasonOrder:   map[string]int{},
		jumps:         map[string]Jump{},
		uploads:       map[string]EvidenceUploadAuthorization{},
		evidences:     map[string]Evidence{},
		judgments:     map[string]Judgment{},
		guestSessions: map[string]GuestSession{},
		disputes:      map[string]Dispute{},
		now:           now,
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
	recentJumps := s.recentPerformedJumpsForGroup(groupID)
	s.mu.Unlock()
	return groupHome(group, membership, season, recentJumps, standingsFromGame(standings)), true, nil
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
	recentJumps := s.recentPerformedJumpsForGroup(season.GroupID)
	s.mu.Unlock()
	return groupHome(group, membership, s.currentSeasonForGroup(season.GroupID), recentJumps, standingsFromGame(standings)), true, nil
}

func (s *MemoryStore) Jump(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jump, ok := s.jumps[jumpID]
	if !ok || !visiblePerformedStatus(jump.Status) {
		return game.JumpSnapshot{}, false, nil
	}
	return jumpToSnapshot(jump), true, nil
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

func (s *MemoryStore) GroupMembership(ctx context.Context, playerID, groupID string) (game.MembershipSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, ok := s.memberships[groupID][playerID]
	if !ok {
		return game.MembershipSnapshot{}, false, nil
	}
	return game.MembershipSnapshot{Role: m.Role}, true, nil
}

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

func (s *MemoryStore) JumpByID(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jump, ok := s.jumps[jumpID]
	if !ok {
		return game.JumpSnapshot{}, false, nil
	}
	return jumpToSnapshot(jump), true, nil
}

func (s *MemoryStore) UpsertJudgment(ctx context.Context, jumpID, playerID, guestSessionID, provenance string, commitment, transgression, creativity, presentation int) (game.Judgment, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var key string
	if playerID != "" {
		key = jumpID + ":" + playerID
	} else {
		key = jumpID + ":guest:" + guestSessionID
	}

	_, existed := s.judgments[key]
	httpJudgment := Judgment{
		ID:             stableID("judgment", key),
		JumpID:         jumpID,
		PlayerID:       playerID,
		GuestSessionID: guestSessionID,
		Provenance:     provenance,
		Commitment:     commitment,
		Transgression:  transgression,
		Creativity:     creativity,
		Presentation:   presentation,
	}
	s.judgments[key] = httpJudgment
	return game.Judgment{
		ID:             httpJudgment.ID,
		JumpID:         httpJudgment.JumpID,
		PlayerID:       httpJudgment.PlayerID,
		GuestSessionID: httpJudgment.GuestSessionID,
		Provenance:     httpJudgment.Provenance,
		Commitment:     httpJudgment.Commitment,
		Transgression:  httpJudgment.Transgression,
		Creativity:     httpJudgment.Creativity,
		Presentation:   httpJudgment.Presentation,
	}, !existed, nil
}

func (s *MemoryStore) AdvanceJumpToJudged(ctx context.Context, jumpID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jump, ok := s.jumps[jumpID]
	if !ok {
		return game.ErrJumpNotFound
	}
	if jump.Status != "Performed Jump" {
		return nil
	}
	jump.Status = "Judged Jump"
	s.jumps[jumpID] = jump
	return nil
}

func (s *MemoryStore) GuestSessionJudgmentCount(ctx context.Context, guestSessionID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gs, ok := s.guestSessions[guestSessionID]
	if !ok {
		return 0, nil
	}
	return gs.JudgmentCount, nil
}

func (s *MemoryStore) IncrementGuestSessionJudgmentCount(ctx context.Context, guestSessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	gs := s.guestSessions[guestSessionID]
	gs.JudgmentCount++
	s.guestSessions[guestSessionID] = gs
	return nil
}

func (s *MemoryStore) CreateGuestSession(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.guestSessions[id] = GuestSession{
		ID:            id,
		JudgmentCount: 0,
		CreatedAt:     s.now().Unix(),
	}
	return nil
}

func (s *MemoryStore) InsertIdea(ctx context.Context, groupID, playerID, source, destination, food string) (game.JumpSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jumpNumber++
	jump := Jump{
		ID:          stableID("jump", groupID+":"+playerID+":"+strconv.Itoa(s.jumpNumber)),
		GroupID:     groupID,
		PlayerID:    playerID,
		Status:      "Idea",
		Source:      source,
		Destination: destination,
		Food:        food,
		OffSeason:   true,
		CreatedAt:   s.now().UTC(),
	}
	s.jumps[jump.ID] = jump
	return jumpToSnapshot(jump), nil
}

func (s *MemoryStore) Idea(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jump, ok := s.jumps[jumpID]
	if !ok {
		return game.JumpSnapshot{}, false, nil
	}
	return jumpToSnapshot(jump), true, nil
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

func (s *MemoryStore) UpdateJumpToPlanned(ctx context.Context, jumpID, playerID string, seasonID *string) (game.JumpSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jump := s.jumps[jumpID]
	jump.Status = "Planned Jump"
	jump.SeasonID = seasonID
	jump.OffSeason = seasonID == nil
	s.jumps[jumpID] = jump
	return jumpToSnapshot(jump), nil
}

func (s *MemoryStore) InsertPerformedJump(ctx context.Context, params game.InsertPerformedJumpParams) (game.JumpSnapshot, game.EvidenceSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jumpNumber++
	id := stableID("jump", params.PlayerID+":"+strconv.Itoa(s.jumpNumber))
	jump := Jump{
		ID:                   id,
		GroupID:              params.GroupID,
		PlayerID:             params.PlayerID,
		Status:               "Performed Jump",
		SeasonID:             params.SeasonID,
		Source:               params.Source,
		Destination:          params.Destination,
		Food:                 params.Food,
		OffSeason:            params.SeasonID == nil,
		GracePeriodExpiresAt: params.GracePeriodExpiresAt,
		CreatedAt:            s.now().UTC(),
	}
	s.jumps[id] = jump

	evidenceID := stableID("evidence", id)
	now := s.now().UTC()
	evidence := Evidence{
		ID:             evidenceID,
		JumpID:         id,
		Caption:        params.Caption,
		MediaObjectKey: params.MediaObjectKey,
		CreatedAt:      now,
	}
	s.evidences[id] = evidence

	return jumpToSnapshot(jump), game.EvidenceSnapshot{
		ID:             evidenceID,
		JumpID:         id,
		PlayerID:       params.PlayerID,
		MediaObjectKey: params.MediaObjectKey,
		Caption:        params.Caption,
		CreatedAt:      now,
	}, nil
}

func (s *MemoryStore) PlannedJump(ctx context.Context, jumpID string) (game.JumpSnapshot, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jump, ok := s.jumps[jumpID]
	if !ok || jump.Status != "Planned Jump" {
		return game.JumpSnapshot{}, false, nil
	}
	return jumpToSnapshot(jump), true, nil
}

func (s *MemoryStore) CreateAuthorization(ctx context.Context, jumpID, playerID, contentType string) (game.AuthorizationSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.uploadNumber++
	now := s.now()
	id := stableID("evidence_upload", jumpID+":"+strconv.Itoa(s.uploadNumber))
	mediaObjectKey := "uploads/" + jumpID + "/" + strconv.Itoa(s.uploadNumber)
	auth := EvidenceUploadAuthorization{
		ID:             id,
		JumpID:         jumpID,
		UploadURL:      "https://storage.supperjumpin.test/uploads/" + mediaObjectKey,
		UploadMethod:   httpMethodPut,
		UploadHeaders:  map[string]string{"Content-Type": contentType},
		MediaObjectKey: mediaObjectKey,
		ExpiresAt:      now.Add(15 * time.Minute).UTC(),
	}
	s.uploads[id] = auth
	return game.AuthorizationSnapshot{
		ID:             id,
		JumpID:         jumpID,
		MediaObjectKey: mediaObjectKey,
		ExpiresAt:      auth.ExpiresAt,
	}, nil
}

func (s *MemoryStore) ClaimAndAdvance(ctx context.Context, authorizationID, jumpID, playerID, caption string) (game.EvidenceCreateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	auth, ok := s.uploads[authorizationID]
	if !ok || auth.JumpID != jumpID || s.now().After(auth.ExpiresAt) {
		return game.EvidenceCreateResult{}, game.ErrEvidenceUploadAuthorizationNotFound
	}
	delete(s.uploads, authorizationID)

	jump := s.jumps[jumpID]
	jump.Status = "Performed Jump"
	s.jumps[jump.ID] = jump

	evidenceID := stableID("evidence", jumpID+":"+authorizationID)
	s.evidences[jumpID] = Evidence{
		ID:             evidenceID,
		JumpID:         jumpID,
		Caption:        caption,
		MediaObjectKey: auth.MediaObjectKey,
		CreatedAt:      s.now().UTC(),
	}
	return game.EvidenceCreateResult{
		EvidenceID:     evidenceID,
		MediaObjectKey: auth.MediaObjectKey,
	}, nil
}

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

func (s *MemoryStore) JumpsForSeason(ctx context.Context, seasonID string) ([]game.JumpSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []game.JumpSnapshot
	for _, jump := range s.jumps {
		if jump.SeasonID != nil && *jump.SeasonID == seasonID {
			result = append(result, jumpToSnapshot(jump))
		}
	}
	return result, nil
}

func (s *MemoryStore) JudgmentsForJump(ctx context.Context, jumpID string) ([]game.Judgment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []game.Judgment
	for _, j := range s.judgments {
		if j.JumpID == jumpID {
			result = append(result, game.Judgment{
				ID:            j.ID,
				JumpID:        j.JumpID,
				PlayerID:      j.PlayerID,
				Commitment:    j.Commitment,
				Transgression: j.Transgression,
				Creativity:    j.Creativity,
				Presentation:  j.Presentation,
			})
		}
	}
	return result, nil
}

func (s *MemoryStore) UpdateJumpFinalization(ctx context.Context, jumpID string, status string, finalScore *int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jump := s.jumps[jumpID]
	jump.Status = status
	jump.FinalScore = finalScore
	s.jumps[jumpID] = jump
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

func (s *MemoryStore) InsertDispute(ctx context.Context, jumpID, raisedByPlayerID, concern, details string) (game.DisputeSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := stableID("dispute", jumpID+":"+raisedByPlayerID+":"+strconv.Itoa(len(s.disputes)+1))
	dispute := Dispute{
		ID:               id,
		JumpID:           jumpID,
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

func (s *MemoryStore) UpdateJumpStatusAfterDispute(ctx context.Context, jumpID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jump := s.jumps[jumpID]
	jump.Status = status
	jump.FinalScore = nil
	s.jumps[jump.ID] = jump
	return nil
}

func (s *MemoryStore) JumpsForOpenMonth(ctx context.Context, year, month int) ([]game.JumpSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	var result []game.JumpSnapshot
	for _, jump := range s.jumps {
		if jump.CreatedAt.Equal(start) || jump.CreatedAt.After(start) && jump.CreatedAt.Before(end) {
			if jump.Status == "Performed Jump" || jump.Status == "Judged Jump" || jump.Status == "Unjudged Jump" || jump.Status == "Disqualified Jump" {
				result = append(result, jumpToSnapshot(jump))
			}
		}
	}
	return result, nil
}

func (s *MemoryStore) UpdateJumpOpenFinalScore(ctx context.Context, jumpID string, score *int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jump, ok := s.jumps[jumpID]
	if !ok {
		return nil
	}
	jump.OpenFinalScore = score
	s.jumps[jumpID] = jump
	return nil
}

func (s *MemoryStore) PlayersForOpenMonth(ctx context.Context, year, month int) ([]game.PlayerSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	seen := map[string]bool{}
	var result []game.PlayerSnapshot
	for _, jump := range s.jumps {
		if jump.CreatedAt.Equal(start) || jump.CreatedAt.After(start) && jump.CreatedAt.Before(end) {
			if jump.Status == "Performed Jump" || jump.Status == "Judged Jump" || jump.Status == "Unjudged Jump" || jump.Status == "Disqualified Jump" {
				if !seen[jump.PlayerID] {
					seen[jump.PlayerID] = true
					player := s.players[jump.PlayerID]
					result = append(result, game.PlayerSnapshot{ID: player.ID, DisplayName: player.DisplayName})
				}
			}
		}
	}
	return result, nil
}

func (s *MemoryStore) UpsertOpenStanding(ctx context.Context, year, month int, playerID string, score, judgedJumps int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// MemoryStore doesn't need to persist open standings for basic tests.
	return nil
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

func (s *MemoryStore) ensureSeasonStatusesForGroup(groupID string) []string {
	// TODO: If Groups/Seasons become active again, move deadline-based status
	// progression into a pure game helper and let adapters only persist changes.
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

func (s *MemoryStore) recentPerformedJumpsForGroup(groupID string) []PerformedJumpView {
	performed := []PerformedJumpView{}
	for _, jump := range s.jumps {
		if jump.GroupID != groupID || !visiblePerformedStatus(jump.Status) {
			continue
		}
		evidence, ok := s.evidences[jump.ID]
		if !ok {
			continue
		}
		performed = append(performed, PerformedJumpView{
			Jump:      jump,
			Performer: s.players[jump.PlayerID],
			Evidence:  evidence,
			Disputes:  s.disputesForJump(jump.ID),
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

func (s *MemoryStore) disputesForJump(jumpID string) []Dispute {
	disputes := []Dispute{}
	for _, dispute := range s.disputes {
		if dispute.JumpID != jumpID {
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

// FeedJumps returns a page of public Feed Jumps from MemoryStore.
func (s *MemoryStore) FeedJumps(ctx context.Context, cursorTS *time.Time, cursorID string, limit int) ([]JumpCard, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var visible []Jump
	for _, jump := range s.jumps {
		if !visiblePerformedStatus(jump.Status) {
			continue
		}
		visible = append(visible, jump)
	}

	sort.Slice(visible, func(i, j int) bool {
		if !visible[i].CreatedAt.Equal(visible[j].CreatedAt) {
			return visible[i].CreatedAt.After(visible[j].CreatedAt)
		}
		return visible[i].ID > visible[j].ID
	})

	startIdx := 0
	if cursorTS != nil {
		for idx, j := range visible {
			if j.CreatedAt.Before(*cursorTS) || (j.CreatedAt.Equal(*cursorTS) && j.ID < cursorID) {
				startIdx = idx
				break
			}
		}
	}

	if startIdx >= len(visible) {
		return []JumpCard{}, nil
	}

	endIdx := startIdx + limit
	if endIdx > len(visible) {
		endIdx = len(visible)
	}

	cards := make([]JumpCard, 0, endIdx-startIdx)
	for _, jump := range visible[startIdx:endIdx] {
		card := s.jumpToCard(jump)
		cards = append(cards, card)
	}
	return cards, nil
}

// JumpDetail returns the full detail view of a Jump from MemoryStore.
func (s *MemoryStore) JumpDetail(ctx context.Context, jumpID string) (JumpDetail, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jump, ok := s.jumps[jumpID]
	if !ok {
		return JumpDetail{}, false, nil
	}

	return s.jumpToDetail(jump), true, nil
}

// HasJudgedJump returns true if the player has already judged this Jump in MemoryStore.
func (s *MemoryStore) HasJudgedJump(ctx context.Context, jumpID, playerID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, j := range s.judgments {
		if j.JumpID == jumpID && j.PlayerID == playerID {
			return true, nil
		}
	}
	return false, nil
}

// jumpToCard converts a Jump into a JumpCard for the public feed.
func (s *MemoryStore) jumpToCard(jump Jump) JumpCard {
	var ra float64
	var jc int

	// Compute running average from stored judgments
	var total float64
	var count int
	for _, j := range s.judgments {
		if j.JumpID == jump.ID {
			composite := float64(j.Commitment+j.Transgression+j.Creativity+j.Presentation) / 4.0
			total += composite
			count++
		}
	}
	if count > 0 {
		ra = total / float64(count)
	}
	jc = count

	_ = jc

	pc := Player{}
	for _, p := range s.players {
		if p.ID == jump.PlayerID {
			pc = p
			break
		}
	}

	ev := Evidence{}
	for _, e := range s.evidences {
		if e.JumpID == jump.ID {
			ev = e
			break
		}
	}

	return JumpCard{
		ID:                   jump.ID,
		PerformerName:        pc.DisplayName,
		PerformerID:          jump.PlayerID,
		Source:               jump.Source,
		Destination:          jump.Destination,
		Food:                 jump.Food,
		Caption:              ev.Caption,
		MediaObjectKey:       ev.MediaObjectKey,
		Status:               jump.Status,
		GracePeriodExpiresAt: jump.GracePeriodExpiresAt,
		RunningAverage:       ra,
		JudgmentCount:        count,
		CreatedAt:            jump.CreatedAt,
	}
}

// jumpToDetail converts a Jump into a JumpDetail.
func (s *MemoryStore) jumpToDetail(jump Jump) JumpDetail {
	card := s.jumpToCard(jump)

	var disputes []Dispute
	for _, d := range s.disputes {
		if d.JumpID == jump.ID {
			disputes = append(disputes, d)
		}
	}

	return JumpDetail{
		ID:                   card.ID,
		PerformerName:        card.PerformerName,
		PerformerID:          card.PerformerID,
		Source:               card.Source,
		Destination:          card.Destination,
		Food:                 card.Food,
		Caption:              card.Caption,
		MediaObjectKey:       card.MediaObjectKey,
		Status:               card.Status,
		GracePeriodExpiresAt: card.GracePeriodExpiresAt,
		RunningAverage:       card.RunningAverage,
		JudgmentCount:        card.JudgmentCount,
		CreatedAt:            card.CreatedAt,
		FinalScore:           jump.FinalScore,
		Disputes:             disputes,
	}
}
