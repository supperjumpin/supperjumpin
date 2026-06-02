package httpapi

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

var ErrSeasonAlreadyOpen = errors.New("Group already has an active or closing Season")

var ErrSeasonNotFound = errors.New("Season not found")

var ErrJumpNotFound = errors.New("Jump not found")

var ErrEvidenceUploadAuthorizationNotFound = errors.New("Evidence upload authorization not found")

var ErrJudgingWindowClosed = errors.New("Judging Window closed")

var ErrSubmissionWindowClosed = errors.New("Submission Window closed")

var ErrInvalidJudgmentScore = errors.New("Judgment scores must be between 1 and 4")

var ErrAuthorGracePeriodActive = errors.New("Author Grace Period is still active")

var ErrInvalidDisputeConcern = errors.New("Dispute concern must be House Rules, Credibility, Source, Destination, Food, duplicate, or other")

var ErrDisputeNotFound = errors.New("Dispute not found")

var ErrInvalidDisputeResolution = errors.New("Dispute resolution must be No Action, Disqualified Jump, or Removed Jump")

var ErrGuestCapReached = errors.New("Guest Judgment cap reached")

var ErrInvalidJudgeIdentity = errors.New("Judgment must have exactly one judge identity: player or guest session")

var ErrOpenMonthNotClosed = errors.New("Open month has not soft-closed yet")

// mapGameErr translates game-module typed errors into the corresponding
// httpapi sentinel errors so HTTP handlers can use their existing errors.Is checks.
func mapGameErr(err error) error {
	if errors.Is(err, game.ErrInvalidJudgmentScore) {
		return ErrInvalidJudgmentScore
	}
	if errors.Is(err, game.ErrJumpNotFound) {
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
	if errors.Is(err, game.ErrAuthorGracePeriodActive) {
		return ErrAuthorGracePeriodActive
	}
	if errors.Is(err, game.ErrGuestCapReached) {
		return ErrGuestCapReached
	}
	if errors.Is(err, game.ErrInvalidJudgeIdentity) {
		return ErrInvalidJudgeIdentity
	}
	if errors.Is(err, game.ErrOpenMonthNotClosed) {
		return ErrOpenMonthNotClosed
	}
	return err
}

// Store handles identity persistence.
type Store interface {
	BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error)
}

// Persistence combines game repository interfaces with transport-layer DTO
// assembly queries. Both MemoryStore and PostgresStore implement this interface.
// TODO: As non-Group/non-Season flows are touched, narrow helper parameters to
// flow-specific interfaces instead of passing this broad composed seam around.
type Persistence interface {
	game.GroupRepository
	game.JumpRepository
	game.EvidenceRepository
	game.JudgmentRepository
	game.SeasonRepository
	game.DisputeRepository
	game.OpenRepository

	GroupHomeForGroup(ctx context.Context, groupID string, player Player) (GroupHomeResponse, bool, error)
	GroupHomeForSeason(ctx context.Context, seasonID string, player Player) (GroupHomeResponse, bool, error)
	CreateGuestSession(ctx context.Context, id string) error
	Now() time.Time

	// Public read path
	FeedJumps(ctx context.Context, cursorTS *time.Time, cursorID string, limit int) ([]JumpCard, error)
	JumpDetail(ctx context.Context, jumpID string) (JumpDetail, bool, error)
	HasJudgedJump(ctx context.Context, jumpID, playerID string) (bool, error)
	// HasJudgedJumps returns a map of jumpID → true only for jumps the player
	// has judged. Absent keys mean "not judged by this player" — Do not
	// distinguish "queried and false" from "not in result set"; it's always
	// absent for not-judged. Callers check with judgedMap[id] (yields false
	// for absent keys via Go zero-value semantics, which is correct here).
	HasJudgedJumps(ctx context.Context, playerID string, jumpIDs []string) (map[string]bool, error)
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
	// TODO(#119): Public Feed/Jump Detail work should define its own read model
	// instead of reusing Group Home's membership-gated shape.
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
	return jumpFromGame(result.Jump), true, nil
}

func createPlannedJump(ctx context.Context, db Persistence, player Player, ideaID string, offSeason bool) (Jump, bool, error) {
	result := game.CreatePlannedJump(ctx, db, game.CreatePlannedJumpInput{
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
	return jumpFromGame(result.Jump), true, nil
}

func authorizeEvidenceUpload(ctx context.Context, db Persistence, player Player, jumpID string, contentType string) (EvidenceUploadAuthorization, bool, error) {
	result := game.AuthorizeEvidenceUpload(ctx, db, game.AuthorizeEvidenceUploadInput{
		JumpID:      jumpID,
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
		JumpID:         result.Authorization.JumpID,
		UploadURL:      "https://storage.supperjumpin.test/uploads/" + result.Authorization.MediaObjectKey,
		UploadMethod:   httpMethodPut,
		UploadHeaders:  map[string]string{"Content-Type": contentType},
		MediaObjectKey: result.Authorization.MediaObjectKey,
		ExpiresAt:      result.Authorization.ExpiresAt,
	}, true, nil
}

func submitEvidence(ctx context.Context, db Persistence, player Player, jumpID string, uploadAuthorizationID string, caption string) (EvidenceSubmission, bool, error) {
	result := game.SubmitEvidence(ctx, db, game.SubmitEvidenceInput{
		JumpID:                jumpID,
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
			ID:          result.Jump.ID,
			GroupID:     result.Jump.GroupID,
			PlayerID:    result.Jump.PlayerID,
			SeasonID:    result.Jump.SeasonID,
			Status:      result.Jump.Status,
			Source:      result.Jump.Source,
			Destination: result.Jump.Destination,
			Food:        result.Jump.Food,
			OffSeason:   result.Jump.SeasonID == nil,
		},
		Evidence: Evidence{
			ID:             result.Evidence.ID,
			JumpID:         result.Evidence.JumpID,
			Caption:        result.Evidence.Caption,
			MediaObjectKey: result.Evidence.MediaObjectKey,
			CreatedAt:      result.Evidence.CreatedAt,
		},
	}, true, nil
}

func createPerformedJump(ctx context.Context, db Persistence, player Player, source, destination, food, caption, mediaObjectKey string, groupID string) (Jump, error) {
	result := game.CreatePerformedJump(ctx, db, game.CreatePerformedJumpInput{
		PlayerID:       player.ID,
		Source:         source,
		Destination:    destination,
		Food:           food,
		Caption:        caption,
		MediaObjectKey: mediaObjectKey,
		GroupID:        groupID,
	}, db.Now())
	if result.Err != nil {
		return Jump{}, mapGameErr(result.Err)
	}
	return jumpFromGame(result.Jump), nil
}

func submitJudgment(ctx context.Context, db Persistence, playerID, guestSessionID, provenance, jumpID string, commitment int, transgression int, creativity int, presentation int) (Judgment, bool, bool, error) {
	result := game.SubmitJudgment(ctx, db, game.JudgmentInput{
		JumpID:         jumpID,
		JudgePlayerID:  playerID,
		GuestSessionID: guestSessionID,
		Provenance:     provenance,
		Commitment:     commitment,
		Transgression:  transgression,
		Creativity:     creativity,
		Presentation:   presentation,
	}, db.Now())
	if result.Err != nil {
		return Judgment{}, false, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return Judgment{}, false, false, nil
	}
	return Judgment{
		ID:             result.Judgment.ID,
		JumpID:         result.Judgment.JumpID,
		PlayerID:       result.Judgment.PlayerID,
		GuestSessionID: result.Judgment.GuestSessionID,
		Provenance:     result.Judgment.Provenance,
		Commitment:     result.Judgment.Commitment,
		Transgression:  result.Judgment.Transgression,
		Creativity:     result.Judgment.Creativity,
		Presentation:   result.Judgment.Presentation,
	}, true, result.Created, nil
}

func createDispute(ctx context.Context, db Persistence, player Player, jumpID string, concern string, details string) (Dispute, bool, error) {
	result := game.CreateDispute(ctx, db, game.CreateDisputeInput{
		PlayerID: player.ID,
		JumpID:   jumpID,
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
		Resolution:       resolution,
		ResolutionReason: resolutionReason,
	})
	if result.Err != nil {
		return DisputeResolution{}, false, mapGameErr(result.Err)
	}
	if !result.Allowed {
		return DisputeResolution{}, false, nil
	}
	return DisputeResolution{
		Jump:    jumpFromGame(result.Jump),
		Dispute: disputeFromSnapshot(result.Dispute),
	}, true, nil
}
