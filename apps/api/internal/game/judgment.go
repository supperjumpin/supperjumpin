package game

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidJudgmentScore    = errors.New("Judgment scores must be between 1 and 4")
	ErrJumpNotFound            = errors.New("Jump not found")
	ErrJudgingWindowClosed     = errors.New("Judging Window closed")
	ErrForbidden               = errors.New("Judge must be a different Player than the performer")
	ErrAuthorGracePeriodActive = errors.New("Author Grace Period is still active")
	ErrGuestCapReached         = errors.New("Guest Judgment cap reached")
	ErrAlreadyJudged           = errors.New("Judge has already submitted a Judgment for this Jump")
	ErrInvalidJudgeIdentity    = errors.New("Judgment must have exactly one judge identity: player or guest session")
)

// Judgment holds submitted scores for a Performed Jump.
type Judgment struct {
	ID             string
	JumpID         string
	PlayerID       string
	GuestSessionID string
	Provenance     string
	OpenMonthID    *string
	Commitment     int
	Transgression  int
	Creativity     int
	Presentation   int
}

// JumpSnapshot is a read-only view of a Jump needed for game rules.
type JumpSnapshot struct {
	ID                   string
	PlayerID             string
	Status               string
	SeasonID             *string
	Source               string
	Destination          string
	Food                 string
	FinalScore           *int
	GracePeriodExpiresAt time.Time
}

// SeasonSnapshot is a read-only view of a Season needed for game rules.
type SeasonSnapshot struct {
	ID                   string
	GroupID              string
	CommissionerPlayerID string
	Status               string
	SubmissionDeadline   time.Time
	JudgingDeadline      time.Time
}

// JudgmentRepository defines persistence operations for the judgment flow.
type JudgmentRepository interface {
	// Jump returns the Jump for the given ID. The ok bool is true if the
	// jump exists and is in a visible performed status.
	Jump(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	// Season returns the Season for the given ID.
	Season(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	// ActiveOpen returns the current active competition window based on the provided clock.
	ActiveOpen(ctx context.Context, now time.Time) (*OpenMonth, error)
	// SubmitAcceptedJudgment atomically persists an accepted judgment.
	SubmitAcceptedJudgment(ctx context.Context, input JudgmentInput) (Judgment, error)
	// HasJudgedJump returns true if the player has already submitted a Judgment for this Jump.
	HasJudgedJump(ctx context.Context, jumpID, playerID string) (bool, error)
	// HasJudgedJumps returns a map of jumpID → true only for jumps the player has judged.
	HasJudgedJumps(ctx context.Context, playerID string, jumpIDs []string) (map[string]bool, error)
	// HasGuestJudgedJump returns true if the guest session has already submitted a Judgment for this Jump.
	HasGuestJudgedJump(ctx context.Context, jumpID, guestSessionID string) (bool, error)
	// GuestSessionJudgmentCount returns the number of Judgments already filed by a guest session.
	GuestSessionJudgmentCount(ctx context.Context, guestSessionID string) (int, error)
	// IncrementGuestSessionJudgmentCount increments a guest session's judgment cap count.
	IncrementGuestSessionJudgmentCount(ctx context.Context, guestSessionID string) error
	// CreateGuestSession creates a new guest session with the given ID.
	CreateGuestSession(ctx context.Context, id string) error
}

// JudgmentInput bundles the parameters for a judgment submission.
type JudgmentInput struct {
	JumpID         string
	JudgePlayerID  string
	GuestSessionID string
	Provenance     string
	OpenMonthID    *string
	Commitment     int
	Transgression  int
	Creativity     int
	Presentation   int
}

// JudgmentResult is the outcome of a judgment submission.
type JudgmentResult struct {
	Judgment Judgment
	Allowed  bool
	Created  bool
	Err      error
}

// SubmitJudgment evaluates judgment rules and, if allowed, persists an immutable
// first-time Judgment for the given Jump.
//
// The function evaluates guards in deterministic PRD order:
//   1. Exactly one judge identity
//   2. Jump exists
//   3. Jump is in an open performed/judged status
//   4. Self-judging (player is the performer)
//   5. Author Grace Period expired
//   6. Judge has not already submitted a Judgment for this Jump
//   7. Guest cap not reached
//   8. Scores are valid (1–4)
//   9. Season judging window is open (for season-linked Jumps)
//
// On success the Jump is advanced to "Judged Jump" as part of the same atomic
// persistence operation. Judgments are immutable; a duplicate submission returns
// ErrAlreadyJudged.
//
// The default guest cap is 5. Use WithGuestCap(n) to override.
func SubmitJudgment(ctx context.Context, repo JudgmentRepository, input JudgmentInput, now time.Time, opts ...SubmitJudgmentOption) (Judgment, error) {
	cfg := submitJudgmentConfig{guestCap: defaultGuestCap}
	for _, opt := range opts {
		opt(&cfg)
	}
	// 1. Exactly one judge identity must be provided.
	playerSet := input.JudgePlayerID != ""
	if playerSet == (input.GuestSessionID != "") { // both or neither
		return Judgment{}, ErrInvalidJudgeIdentity
	}

	// 2. Look up the jump.
	jump, ok, err := repo.Jump(ctx, input.JumpID)
	if err != nil {
		return Judgment{}, err
	}
	if !ok {
		return Judgment{}, ErrJumpNotFound
	}

	// 3. Jump must be in "Performed Jump" or "Judged Jump" status to accept judgments.
	if jump.Status != "Performed Jump" && jump.Status != "Judged Jump" {
		return Judgment{}, ErrJudgingWindowClosed
	}

	// 4. Judge must not be the performer. Group membership is NOT required.
	if playerSet && jump.PlayerID == input.JudgePlayerID {
		return Judgment{}, ErrForbidden
	}

	// 5. Author Grace Period must have expired.
	if !now.After(jump.GracePeriodExpiresAt) {
		return Judgment{}, ErrAuthorGracePeriodActive
	}

	// 6. Judge must not have already submitted a Judgment for this Jump.
	if playerSet {
		already, err := repo.HasJudgedJump(ctx, input.JumpID, input.JudgePlayerID)
		if err != nil {
			return Judgment{}, err
		}
		if already {
			return Judgment{}, ErrAlreadyJudged
		}
	} else {
		already, err := repo.HasGuestJudgedJump(ctx, input.JumpID, input.GuestSessionID)
		if err != nil {
			return Judgment{}, err
		}
		if already {
			return Judgment{}, ErrAlreadyJudged
		}
	}

	// 7. Guest cap must not be reached.
	if input.GuestSessionID != "" {
		count, err := repo.GuestSessionJudgmentCount(ctx, input.GuestSessionID)
		if err != nil {
			return Judgment{}, err
		}
		if count >= cfg.guestCap {
			return Judgment{}, ErrGuestCapReached
		}
	}

	// 8. Validate scores.
	if !validScore(input.Commitment) || !validScore(input.Transgression) || !validScore(input.Creativity) || !validScore(input.Presentation) {
		return Judgment{}, ErrInvalidJudgmentScore
	}

	// 9. Check judging window for season-linked jumps.
	if jump.SeasonID != nil {
		season, err := repo.Season(ctx, *jump.SeasonID)
		if err != nil {
			return Judgment{}, err
		}
		if !isOpenSeasonStatus(season.Status) {
			return Judgment{}, ErrJudgingWindowClosed
		}
	}

	// 10. Capture Open Month provenance at insertion time.
	open, err := repo.ActiveOpen(ctx, now)
	if err == nil && open != nil {
		input.OpenMonthID = &open.ID
	}

	// 11. Persist the accepted judgment atomically.
	judgment, err := repo.SubmitAcceptedJudgment(ctx, input)
	if err != nil {
		return Judgment{}, err
	}

	return judgment, nil
}

func validScore(score int) bool {
	return score >= 1 && score <= 4
}

// EligibilityHint is the structured result of JudgmentEligibility.
type EligibilityHint struct {
	CanJudge          bool
	Reason            string
	GracePeriodEndsAt *time.Time
}

// JudgmentEligibility determines whether a viewer can judge a Jump.
// It checks, in order:
//  1. Empty viewerID — always eligible (unauthenticated/guest prompt path)
//  2. Self-judging — viewer is the performer → not eligible
//  3. Grace period active — now < GracePeriodExpiresAt → not eligible
//  4. Already judged — hasJudged == true → not eligible
//
// The caller is responsible for fetching hasJudged (via HasJudgedJump or
// HasJudgedJumps) so the transport layer can batch the query on the feed path.
func JudgmentEligibility(jump JumpSnapshot, viewerID string, hasJudged bool, now time.Time) EligibilityHint {
	hint := EligibilityHint{CanJudge: true}

	if viewerID == "" {
		return hint
	}

	// 1. Self-judging
	if viewerID == jump.PlayerID {
		hint.CanJudge = false
		hint.Reason = "self-judging"
		return hint
	}

	// 2. Grace period active
	if now.Before(jump.GracePeriodExpiresAt) {
		hint.CanJudge = false
		hint.Reason = "grace-period"
		hint.GracePeriodEndsAt = &jump.GracePeriodExpiresAt
		return hint
	}

	// 3. Already judged
	if hasJudged {
		hint.CanJudge = false
		hint.Reason = "already-judged"
		return hint
	}

	return hint
}

func isOpenSeasonStatus(status string) bool {
	return status == "Active" || status == "Judging Grace Period"
}

// defaultGuestCap is the maximum number of Judgments a Guest Judge can submit
// before being asked to create an Account.
const defaultGuestCap = 5

// submitJudgmentConfig holds optional configuration for SubmitJudgment.
type submitJudgmentConfig struct {
	guestCap int
}

// SubmitJudgmentOption configures SubmitJudgment behavior.
type SubmitJudgmentOption func(*submitJudgmentConfig)

// WithGuestCap overrides the default guest judgment cap (5).
// A value of 0 means no cap (unlimited guest judgments).
func WithGuestCap(cap int) SubmitJudgmentOption {
	return func(c *submitJudgmentConfig) {
		c.guestCap = cap
	}
}
