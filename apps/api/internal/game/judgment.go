package game

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidJudgmentScore = errors.New("Judgment scores must be between 0 and 10")
	ErrJumpNotFound        = errors.New("Jump not found")
	ErrJudgingWindowClosed  = errors.New("Judging Window closed")
	ErrForbidden            = errors.New("Judge must be a different Player than the performer")
)

// Judgment holds submitted scores for a Performed Jump.
type Judgment struct {
	ID            string
	JumpID       string
	PlayerID      string
	Difficulty    int
	Transgression int
	Creativity    int
	Presentation int
}

// JumpSnapshot is a read-only view of a Jump needed for game rules.
type JumpSnapshot struct {
	ID                  string
	GroupID             string
	PlayerID            string
	Status              string
	SeasonID            *string
	Source              string
	Destination         string
	Food                string
	FinalScore          *int
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

// MembershipSnapshot is a read-only view of a Group Membership.
type MembershipSnapshot struct {
	Role string
}

// JudgmentRepository defines persistence operations for the judgment flow.
type JudgmentRepository interface {
	// Jump returns the Jump for the given ID. The ok bool is true only when
	// the jump exists and is in a visible performed status.
	Jump(ctx context.Context, jumpID string) (JumpSnapshot, bool, error)
	// Season returns the Season for the given ID.
	Season(ctx context.Context, seasonID string) (SeasonSnapshot, error)
	// GroupMembership returns the membership for a player in a group.
	// ok is false when there is no membership.
	GroupMembership(ctx context.Context, playerID, groupID string) (MembershipSnapshot, bool, error)
	// UpsertJudgment creates or updates a judgment for a given jump and judge.
	// Returns true when the judgment was created (not updated).
	UpsertJudgment(ctx context.Context, jumpID, playerID string, difficulty, transgression, creativity, presentation int) (Judgment, bool, error)
}

// JudgmentInput bundles the parameters for a judgment submission.
type JudgmentInput struct {
	JumpID       string
	JudgePlayerID string
	Difficulty    int
	Transgression int
	Creativity    int
	Presentation int
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
// It returns a result with:
//   - Allowed = true and Created = true on first-time valid judgment
//   - Allowed = true and Created = false on edit of an existing judgment
//   - Allowed = false on self-judging (returns nil judgment, caller maps to 403)
//   - Err set for invalid input, jump not found, or closed judging window
func SubmitJudgment(ctx context.Context, repo JudgmentRepository, input JudgmentInput, now time.Time) JudgmentResult {
	// 1. Validate scores
	if !validScore(input.Difficulty) || !validScore(input.Transgression) || !validScore(input.Creativity) || !validScore(input.Presentation) {
		return JudgmentResult{Err: ErrInvalidJudgmentScore}
	}

	// 2. Look up the jump
	jump, ok, err := repo.Jump(ctx, input.JumpID)
	if err != nil {
		return JudgmentResult{Err: err}
	}
	if !ok {
		return JudgmentResult{Err: ErrJumpNotFound}
	}

	// 3. Jump must be in "Performed Jump" status to accept judgments.
	//    If it's any other visible status (Judged, Unjudged, Disqualified), window is closed.
	if jump.Status != "Performed Jump" {
		return JudgmentResult{Err: ErrJudgingWindowClosed}
	}

	// 4. Judge must be a member of the jump's group and not the performer
	membership, ok, err := repo.GroupMembership(ctx, input.JudgePlayerID, jump.GroupID)
	if err != nil {
		return JudgmentResult{Err: err}
	}
	if !ok || jump.PlayerID == input.JudgePlayerID {
		return JudgmentResult{Allowed: false}
	}
	_ = membership // role not needed for judgment, reserved for future

	// 5. Check judging window
	if jump.SeasonID != nil {
		season, err := repo.Season(ctx, *jump.SeasonID)
		if err != nil {
			return JudgmentResult{Err: err}
		}
		if !isOpenSeasonStatus(season.Status) {
			return JudgmentResult{Err: ErrJudgingWindowClosed}
		}
	}

	// 6. Persist the judgment (upsert)
	judgment, created, err := repo.UpsertJudgment(ctx, input.JumpID, input.JudgePlayerID, input.Difficulty, input.Transgression, input.Creativity, input.Presentation)
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
	return score >= 0 && score <= 10
}

func isOpenSeasonStatus(status string) bool {
	return status == "Active" || status == "Judging Grace Period"
}
