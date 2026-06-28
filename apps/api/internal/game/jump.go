package game

import (
	"context"
	"errors"
	"time"
)

var (
	ErrRoundNotFound     = errors.New("round not found")
	ErrAlreadyCommitted  = errors.New("player already committed to this round")
	ErrCannotSubmit      = errors.New("player must commit before submitting")
	ErrAlreadySubmitted  = errors.New("player already submitted to this round")
	ErrJumpNotFound      = errors.New("jump not found")
)

// --- commit types ---

type CommitToRoundInput struct {
	RoundID  string
	PlayerID string
}

type CommitResult struct {
	CommitID string
	Allowed  bool
	Err      error
}

// --- submit types ---

type SubmitJumpInput struct {
	RoundID      string
	PlayerID     string
	Caption      string
	EvidenceURLs []string
}

type SubmitJumpResult struct {
	Jump    JumpSnapshot
	Allowed bool
	Err     error
}

// --- list types ---

type ListJumpsResult struct {
	Jumps   []JumpSnapshot
	Allowed bool
	Err     error
}

// --- get jump types ---

type GetJumpResult struct {
	Jump    JumpSnapshot
	Allowed bool
	Err     error
}

// --- snapshots ---

type JumpSnapshot struct {
	ID            string
	RoundID       string
	PlayerID      string
	Caption       string
	EvidenceURLs  []string
	SubmittedAt   time.Time
	SealedViewer  bool // when true, Caption and EvidenceURLs are empty (sealed)
	PlayerHasCommitted  bool
	PlayerHasSubmitted  bool
}

type RoundStatus struct {
	ID             string
	Status         string
	RevealBy       time.Time
	CommitCount    int
	SubmissionCount int
}

// --- commit repo ---

type CommitRepo interface {
	FindRound(ctx context.Context, roundID string) (RoundSnapshot, bool, error)
	FindCommit(ctx context.Context, roundID, playerID string) (*CommitSnapshot, error)
	CreateCommit(ctx context.Context, commit CommitSnapshot) error
}

type CommitSnapshot struct {
	ID          string
	RoundID     string
	PlayerID    string
	CommittedAt time.Time
}

// --- submit repo ---

type SubmitRepo interface {
	FindRound(ctx context.Context, roundID string) (RoundSnapshot, bool, error)
	FindCommit(ctx context.Context, roundID, playerID string) (*CommitSnapshot, error)
	FindJump(ctx context.Context, roundID, playerID string) (*JumpSnapshot, error)
	CreateJump(ctx context.Context, jump JumpSnapshot) error
	InsertEvidence(ctx context.Context, jumpID string, urls []string) error
}

// --- list repo ---

type ListJumpsRepo interface {
	FindRound(ctx context.Context, roundID string) (RoundSnapshot, bool, error)
	ListJumps(ctx context.Context, roundID string) ([]JumpSnapshot, error)
	ListEvidence(ctx context.Context, jumpIDs []string) (map[string][]string, error)
	GetRoundStatus(ctx context.Context, roundID string) (RoundStatus, error)
	FindCommitForPlayer(ctx context.Context, roundID, playerID string) (*CommitSnapshot, error)
}

// --- get jump repo ---

type GetJumpRepo interface {
	FindRound(ctx context.Context, roundID string) (RoundSnapshot, bool, error)
	GetJumpByID(ctx context.Context, jumpID string) (JumpSnapshot, error)
	ListEvidenceForJump(ctx context.Context, jumpID string) ([]string, error)
}

// --- domain functions ---

func CommitToRound(ctx context.Context, repo CommitRepo, input CommitToRoundInput, now time.Time) (CommitResult, error) {
	_, roundExists, err := repo.FindRound(ctx, input.RoundID)
	if err != nil {
		return CommitResult{Allowed: false, Err: err}, nil
	}
	if !roundExists {
		return CommitResult{Allowed: false, Err: ErrRoundNotFound}, nil
	}

	existing, err := repo.FindCommit(ctx, input.RoundID, input.PlayerID)
	if err != nil {
		return CommitResult{Allowed: false, Err: err}, nil
	}
	if existing != nil {
		return CommitResult{Allowed: false, Err: ErrAlreadyCommitted}, nil
	}

	commitID := domainStableID("commit", input.RoundID+":"+input.PlayerID)
	commit := CommitSnapshot{
		ID:          commitID,
		RoundID:     input.RoundID,
		PlayerID:    input.PlayerID,
		CommittedAt: now,
	}

	if err := repo.CreateCommit(ctx, commit); err != nil {
		return CommitResult{Allowed: false, Err: err}, nil
	}

	return CommitResult{CommitID: commitID, Allowed: true}, nil
}

func SubmitJump(ctx context.Context, repo SubmitRepo, input SubmitJumpInput, now time.Time) (SubmitJumpResult, error) {
	_, roundExists, err := repo.FindRound(ctx, input.RoundID)
	if err != nil {
		return SubmitJumpResult{Allowed: false, Err: err}, nil
	}
	if !roundExists {
		return SubmitJumpResult{Allowed: false, Err: ErrRoundNotFound}, nil
	}

	commit, err := repo.FindCommit(ctx, input.RoundID, input.PlayerID)
	if err != nil {
		return SubmitJumpResult{Allowed: false, Err: err}, nil
	}
	if commit == nil {
		return SubmitJumpResult{Allowed: false, Err: ErrCannotSubmit}, nil
	}

	existing, err := repo.FindJump(ctx, input.RoundID, input.PlayerID)
	if err != nil {
		return SubmitJumpResult{Allowed: false, Err: err}, nil
	}
	if existing != nil {
		return SubmitJumpResult{Allowed: false, Err: ErrAlreadySubmitted}, nil
	}

	jumpID := domainStableID("jump", input.RoundID+":"+input.PlayerID)
	jump := JumpSnapshot{
		ID:           jumpID,
		RoundID:      input.RoundID,
		PlayerID:     input.PlayerID,
		Caption:      input.Caption,
		EvidenceURLs: input.EvidenceURLs,
		SubmittedAt:  now,
	}

	if err := repo.CreateJump(ctx, jump); err != nil {
		return SubmitJumpResult{Allowed: false, Err: err}, nil
	}

	if len(input.EvidenceURLs) > 0 {
		if err := repo.InsertEvidence(ctx, jumpID, input.EvidenceURLs); err != nil {
			return SubmitJumpResult{Allowed: false, Err: err}, nil
		}
	}

	return SubmitJumpResult{Jump: jump, Allowed: true}, nil
}

func ListJumpsForRound(ctx context.Context, repo ListJumpsRepo, roundID, viewerPlayerID string) (ListJumpsResult, error) {
	_, roundExists, err := repo.FindRound(ctx, roundID)
	if err != nil {
		return ListJumpsResult{Allowed: false, Err: err}, nil
	}
	if !roundExists {
		return ListJumpsResult{Allowed: false, Err: ErrRoundNotFound}, nil
	}

	jumps, err := repo.ListJumps(ctx, roundID)
	if err != nil {
		return ListJumpsResult{Allowed: false, Err: err}, nil
	}

	status, err := repo.GetRoundStatus(ctx, roundID)
	if err != nil {
		return ListJumpsResult{Allowed: false, Err: err}, nil
	}

	isRevealed := status.Status == "revealed"

	committedByPlayer := make(map[string]bool)
	for _, j := range jumps {
		if j.PlayerID == viewerPlayerID {
			committedByPlayer[j.PlayerID] = true
		}
	}

	viewerCommit, _ := repo.FindCommitForPlayer(ctx, roundID, viewerPlayerID)
	viewerHasCommitted := viewerCommit != nil

	jumpIDs := make([]string, 0, len(jumps))
	for _, j := range jumps {
		jumpIDs = append(jumpIDs, j.ID)
	}
	evidenceMap, err := repo.ListEvidence(ctx, jumpIDs)
	if err != nil {
		return ListJumpsResult{Allowed: false, Err: err}, nil
	}

	result := make([]JumpSnapshot, 0, len(jumps))
	for _, j := range jumps {
		isAuthor := j.PlayerID == viewerPlayerID

		snapshot := JumpSnapshot{
			ID:                 j.ID,
			RoundID:            j.RoundID,
			PlayerID:           j.PlayerID,
			SubmittedAt:        j.SubmittedAt,
			PlayerHasCommitted: true,
			PlayerHasSubmitted: true,
		}

		if isRevealed || isAuthor {
			snapshot.Caption = j.Caption
			snapshot.EvidenceURLs = evidenceMap[j.ID]
			snapshot.SealedViewer = false
		} else {
			snapshot.SealedViewer = true
		}

		result = append(result, snapshot)
	}

	// Also include committed-but-not-submitted players
	if !isRevealed && viewerHasCommitted {
		// Add the viewer's own commit status if they haven't submitted
		foundSelf := false
		for _, j := range result {
			if j.PlayerID == viewerPlayerID {
				foundSelf = true
				break
			}
		}
		if !foundSelf && viewerCommit != nil {
			// Viewer committed but hasn't submitted
			result = append(result, JumpSnapshot{
				ID:                 "",
				RoundID:            roundID,
				PlayerID:           viewerPlayerID,
				SealedViewer:       false,
				PlayerHasCommitted: true,
				PlayerHasSubmitted: false,
			})
		}
	}

	return ListJumpsResult{Jumps: result, Allowed: true}, nil
}

func GetJump(ctx context.Context, repo GetJumpRepo, jumpID, viewerPlayerID string) (GetJumpResult, error) {
	jump, err := repo.GetJumpByID(ctx, jumpID)
	if err != nil {
		if errors.Is(err, ErrJumpNotFound) {
			return GetJumpResult{Allowed: false, Err: ErrJumpNotFound}, nil
		}
		return GetJumpResult{Allowed: false, Err: err}, nil
	}

	roundData, roundExists, err := repo.FindRound(ctx, jump.RoundID)
	if err != nil {
		return GetJumpResult{Allowed: false, Err: err}, nil
	}
	if !roundExists {
		return GetJumpResult{Allowed: false, Err: ErrRoundNotFound}, nil
	}

	isRevealed := roundData.Status == "revealed"
	isAuthor := jump.PlayerID == viewerPlayerID

	result := JumpSnapshot{
		ID:                 jump.ID,
		RoundID:            jump.RoundID,
		PlayerID:           jump.PlayerID,
		SubmittedAt:        jump.SubmittedAt,
		PlayerHasCommitted: true,
		PlayerHasSubmitted: true,
	}

	if isRevealed || isAuthor {
		result.Caption = jump.Caption
		urls, err := repo.ListEvidenceForJump(ctx, jumpID)
		if err != nil {
			return GetJumpResult{Allowed: false, Err: err}, nil
		}
		result.EvidenceURLs = urls
		result.SealedViewer = false
	} else {
		result.SealedViewer = true
	}

	return GetJumpResult{Jump: result, Allowed: true}, nil
}
