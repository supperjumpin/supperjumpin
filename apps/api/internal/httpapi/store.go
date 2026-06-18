package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

var ErrJumpNotFound = errors.New("Jump not found")

var ErrJudgingWindowClosed = errors.New("Judging Window closed")

var ErrInvalidJudgmentScore = errors.New("Judgment scores must be between 1 and 4")

var ErrAuthorGracePeriodActive = errors.New("Author Grace Period is still active")

var ErrGuestCapReached = errors.New("Guest Judgment cap reached")

var ErrInvalidJudgeIdentity = errors.New("Judgment must have exactly one judge identity: player or guest session")

var ErrAlreadyJudged = errors.New("Judge has already submitted a Judgment for this Jump")

var ErrOpenMonthNotClosed = errors.New("Open month has not soft-closed yet")

var ErrInvalidCaption = errors.New("Caption is required")

var ErrAuthorGracePeriodExpired = errors.New("Author Grace Period has expired")

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
	if errors.Is(err, game.ErrAuthorGracePeriodActive) {
		return ErrAuthorGracePeriodActive
	}
	if errors.Is(err, game.ErrGuestCapReached) {
		return ErrGuestCapReached
	}
	if errors.Is(err, game.ErrInvalidJudgeIdentity) {
		return ErrInvalidJudgeIdentity
	}
	if errors.Is(err, game.ErrAlreadyJudged) {
		return ErrAlreadyJudged
	}
	if errors.Is(err, game.ErrOpenMonthNotClosed) {
		return ErrOpenMonthNotClosed
	}
	if errors.Is(err, game.ErrAuthorGracePeriodExpired) {
		return ErrAuthorGracePeriodExpired
	}
	if errors.Is(err, game.ErrInvalidCaption) {
		return ErrInvalidCaption
	}
	return err
}

// UpdateDisplayNameResponse is the response from PATCH /v1/me/display-name.
type UpdateDisplayNameResponse struct {
	Player Player `json:"player"`
}

// Store handles identity persistence.
type Store interface {
	BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error)
	UpdateDisplayName(ctx context.Context, playerID string, displayName string) (Player, error)
}

// JumpPlanningFlow is the narrow interface needed by the jump planning transport
// helper. It embeds only the game repository methods required for creating jumps.
type JumpPlanningFlow interface {
	game.JumpRepository
}

// JudgmentFlow is the narrow interface needed by the judgment transport helper.
// It embeds only the game repository and guest-session methods required for
// submitting judgments.
type JudgmentFlow interface {
	game.JudgmentRepository
	CreateGuestSession(ctx context.Context, id string) error
}

// PublicReadFlow is the narrow interface needed by the public feed and jump
// detail handlers. It provides DTO assembly.
type PublicReadFlow interface {
	FeedJumps(ctx context.Context, cursorTS *time.Time, cursorID string, limit int) ([]JumpCard, error)
	JumpDetail(ctx context.Context, jumpID string) (JumpDetail, bool, error)
}

// OpenFlow is the narrow interface needed by the Open scoring handler.
type OpenFlow interface {
	game.OpenRepository
}

// CaptionEditFlow is the narrow interface needed by the caption edit transport
// helper.
type CaptionEditFlow interface {
	game.CaptionEditRepository
}

// JumpRetractFlow is the narrow interface needed by the retract transport
// helper.
type JumpRetractFlow interface {
	game.RetractJumpRepository
}

// --- Transport-layer DTO helpers (game-command → DTO conversion) ---

func createPerformedJump(ctx context.Context, db JumpPlanningFlow, player Player, source, destination, food, caption, mediaObjectKey string, now time.Time) (Jump, error) {
	result := game.CreatePerformedJump(ctx, db, game.CreatePerformedJumpInput{
		PlayerID:       player.ID,
		Source:         source,
		Destination:    destination,
		Food:           food,
		Caption:        caption,
		MediaObjectKey: mediaObjectKey,
	}, now)
	if result.Err != nil {
		return Jump{}, mapGameErr(result.Err)
	}
	return jumpFromGame(result.Jump), nil
}

func submitJudgment(ctx context.Context, db JudgmentFlow, playerID, guestSessionID, provenance, jumpID string, commitment int, transgression int, creativity int, presentation int, now time.Time, guestCap int) (Judgment, bool, error) {
	var opts []game.SubmitJudgmentOption
	if guestCap > 0 {
		opts = append(opts, game.WithGuestCap(guestCap))
	}
	judgment, err := game.SubmitJudgment(ctx, db, game.JudgmentInput{
		JumpID:         jumpID,
		JudgePlayerID:  playerID,
		GuestSessionID: guestSessionID,
		Provenance:     provenance,
		Commitment:     commitment,
		Transgression:  transgression,
		Creativity:     creativity,
		Presentation:   presentation,
	}, now, opts...)
	if err != nil {
		return Judgment{}, false, mapGameErr(err)
	}
	return Judgment{
		ID:             judgment.ID,
		JumpID:         judgment.JumpID,
		PlayerID:       judgment.PlayerID,
		GuestSessionID: judgment.GuestSessionID,
		Provenance:     judgment.Provenance,
		Commitment:     judgment.Commitment,
		Transgression:  judgment.Transgression,
		Creativity:     judgment.Creativity,
		Presentation:   judgment.Presentation,
	}, true, nil
}

func editCaption(ctx context.Context, db CaptionEditFlow, jumpID string, playerID string, caption string, now time.Time) (bool, error) {
	result := game.EditCaption(ctx, db, game.EditCaptionInput{
		JumpID:   jumpID,
		PlayerID: playerID,
		Caption:  caption,
	}, now)
	if result.Err != nil {
		return false, mapGameErr(result.Err)
	}
	return result.Allowed, nil
}

func retractJump(ctx context.Context, db JumpRetractFlow, jumpID string, playerID string, now time.Time) (bool, error) {
	result := game.RetractJump(ctx, db, game.RetractJumpInput{
		JumpID:   jumpID,
		PlayerID: playerID,
	}, now)
	if result.Err != nil {
		return false, mapGameErr(result.Err)
	}
	return result.Allowed, nil
}
