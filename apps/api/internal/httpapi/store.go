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
// assembly queries. PostgresStore implements this interface; unit tests should
// use small per-test fakes or mocks instead of a shared in-memory store.
type Persistence interface {
	game.JumpRepository
	game.JudgmentRepository
	game.OpenRepository

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

func createPerformedJump(ctx context.Context, db Persistence, player Player, source, destination, food, caption, mediaObjectKey string) (Jump, error) {
	result := game.CreatePerformedJump(ctx, db, game.CreatePerformedJumpInput{
		PlayerID:       player.ID,
		Source:         source,
		Destination:    destination,
		Food:           food,
		Caption:        caption,
		MediaObjectKey: mediaObjectKey,
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
