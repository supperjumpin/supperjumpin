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
	ErrInvalidJudgeIdentity    = errors.New("Judgment must have exactly one judge identity: player or guest session")
)

// Judgment holds submitted scores for a Performed Jump.
type Judgment struct {
	ID             string
	JumpID         string
	PlayerID       string
	GuestSessionID string
	Provenance     string
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
	// Jump returns the Jump for the given ID. The ok bool is true only when
	// the jump exists and is in a visible performed status.
	Jump(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	// Season returns the Season for the given ID.
	Season(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	// SubmitAcceptedJudgment atomically persists an accepted judgment.
	SubmitAcceptedJudgment(ctx context.Context, input JudgmentInput) (Judgment, bool, error)
	// HasJudgedJump returns true if the player has already submitted a Judgment for this Jump.
	HasJudgedJump(ctx context.Context, jumpID, playerID string) (bool, error)
	// HasJudgedJumps returns a map of jumpID → true only for jumps the player has judged.
	HasJudgedJumps(ctx context.Context, playerID string, jumpIDs []string) (map[string]bool, error)
}

// JudgmentInput bundles the parameters for a judgment submission.
type JudgmentInput struct {
	JumpID         string
	JudgePlayerID  string
	GuestSessionID string
	Provenance     string
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

// SubmitJudgment evaluates judgment rules.
//
// Submitting a Judgment requires:
//   - Valid scores (1–4 forced-choice)
//   - The Jump exists and is in "Performed Jump" status
//   - The Author Grace Period has expired
//   - The Judge is not the performer
//   - The judging window is open (for season-linked Jumps)
//
// On the first valid Judgment, the Jump transitions to "Judged Jump".
//
// It returns a result with:
//   - Allowed = true and Created = true on first-time valid judgment with transition
//   - Allowed = true and Created = false on edit of an existing judgment
//   - Allowed = false on self-judging (returns nil judgment, caller maps to 403)
//   - Err set for invalid input, jump not found, grace period active, or closed judging window
func SubmitJudgment(ctx context.Context, repo JudgmentRepository, input JudgmentInput, now time.Time) JudgmentResult {
	// 1. Exactly one judge identity must be provided.
	playerSet := input.JudgePlayerID != ""
	if playerSet == (input.GuestSessionID != "") { // both or neither
		return JudgmentResult{Err: ErrInvalidJudgeIdentity}
	}

	// 2. Validate scores
	if !validScore(input.Commitment) || !validScore(input.Transgression) || !validScore(input.Creativity) || !validScore(input.Presentation) {
		return JudgmentResult{Err: ErrInvalidJudgmentScore}
	}

	// 3. Look up the jump
	jump, ok, err := repo.Jump(ctx, input.JumpID)
	if err != nil {
		return JudgmentResult{Err: err}
	}
	if !ok {
		return JudgmentResult{Err: ErrJumpNotFound}
	}

	// 4. Jump must be in "Performed Jump" or "Judged Jump" status to accept judgments.
	//    "Performed Jump" is the initial state; after the first judgment it becomes "Judged Jump".
	//    "Unjudged Jump" and "Disqualified Jump" are terminal states — window is closed.
	if jump.Status != "Performed Jump" && jump.Status != "Judged Jump" {
		return JudgmentResult{Err: ErrJudgingWindowClosed}
	}

	// 5. Author Grace Period must have expired.
	if !now.After(jump.GracePeriodExpiresAt) {
		return JudgmentResult{Err: ErrAuthorGracePeriodActive}
	}

	// 6. Judge must not be the performer. Group membership is NOT required.
	if playerSet && jump.PlayerID == input.JudgePlayerID {
		return JudgmentResult{Allowed: false}
	}

	// 7. Check judging window for season-linked jumps
	if jump.SeasonID != nil {
		season, err := repo.Season(ctx, *jump.SeasonID)
		if err != nil {
			return JudgmentResult{Err: err}
		}
		if !isOpenSeasonStatus(season.Status) {
			return JudgmentResult{Err: ErrJudgingWindowClosed}
		}
	}

	// 8. Persist the accepted judgment atomically.
	judgment, created, err := repo.SubmitAcceptedJudgment(ctx, input)
	if err != nil {
		return JudgmentResult{Err: err}
	}

	return JudgmentResult{
		Judgment: judgment,
		Allowed:  true,
		Created:  created,
	}
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
