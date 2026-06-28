package game

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCommentBodyEmpty = errors.New("comment body must not be empty")
)

type CommentSnapshot struct {
	ID        string
	RoundID   string
	JumpID    string // empty if round-level comment
	PlayerID  string
	Body      string
	CreatedAt time.Time
}

type PostCommentInput struct {
	RoundID  string
	JumpID   string // empty for round-level comment
	PlayerID string
	Body     string
}

type PostCommentResult struct {
	Comment CommentSnapshot
	Allowed bool
	Err     error
}

type ListCommentsInput struct {
	RoundID string
	JumpID  string // empty to list round-level comments
}

type ListCommentsResult struct {
	Comments []CommentSnapshot
	Allowed  bool
	Err      error
}

type PostCommentRepo interface {
	GetRound(ctx context.Context, roundID string) (RoundSnapshot, bool, error)
	FindPlayer(ctx context.Context, playerID string) (PlayerSnapshot, bool, error)
	GetJump(ctx context.Context, jumpID string) (JumpSnapshot, error)
	CreateComment(ctx context.Context, comment CommentSnapshot) error
}

type ListCommentsRepo interface {
	GetRound(ctx context.Context, roundID string) (RoundSnapshot, bool, error)
	ListComments(ctx context.Context, roundID, jumpID string) ([]CommentSnapshot, error)
}

func PostComment(ctx context.Context, repo PostCommentRepo, input PostCommentInput, now time.Time) (PostCommentResult, error) {
	round, roundExists, err := repo.GetRound(ctx, input.RoundID)
	if err != nil {
		return PostCommentResult{Allowed: false, Err: err}, nil
	}
	if !roundExists {
		return PostCommentResult{Allowed: false, Err: ErrRoundNotFound}, nil
	}
	if round.Status != "revealed" {
		return PostCommentResult{Allowed: false, Err: ErrRoundNotRevealed}, nil
	}

	_, playerExists, err := repo.FindPlayer(ctx, input.PlayerID)
	if err != nil {
		return PostCommentResult{Allowed: false, Err: err}, nil
	}
	if !playerExists {
		return PostCommentResult{Allowed: false, Err: ErrPlayerNotFound}, nil
	}

	if input.JumpID != "" {
		_, err := repo.GetJump(ctx, input.JumpID)
		if err != nil {
			if errors.Is(err, ErrJumpNotFound) {
				return PostCommentResult{Allowed: false, Err: ErrJumpNotFound}, nil
			}
			return PostCommentResult{Allowed: false, Err: err}, nil
		}
	}

	if input.Body == "" {
		return PostCommentResult{Allowed: false, Err: ErrCommentBodyEmpty}, nil
	}

	var commentID string
	if input.JumpID != "" {
		commentID = domainStableID("comment", input.RoundID+":"+input.JumpID+":"+input.PlayerID+":"+now.Format(time.RFC3339Nano))
	} else {
		commentID = domainStableID("comment", input.RoundID+":"+input.PlayerID+":"+now.Format(time.RFC3339Nano))
	}

	comment := CommentSnapshot{
		ID:        commentID,
		RoundID:   input.RoundID,
		JumpID:    input.JumpID,
		PlayerID:  input.PlayerID,
		Body:      input.Body,
		CreatedAt: now,
	}

	if err := repo.CreateComment(ctx, comment); err != nil {
		return PostCommentResult{Allowed: false, Err: err}, nil
	}

	return PostCommentResult{Comment: comment, Allowed: true}, nil
}

func ListComments(ctx context.Context, repo ListCommentsRepo, input ListCommentsInput) (ListCommentsResult, error) {
	_, roundExists, err := repo.GetRound(ctx, input.RoundID)
	if err != nil {
		return ListCommentsResult{Allowed: false, Err: err}, nil
	}
	if !roundExists {
		return ListCommentsResult{Allowed: false, Err: ErrRoundNotFound}, nil
	}

	comments, err := repo.ListComments(ctx, input.RoundID, input.JumpID)
	if err != nil {
		return ListCommentsResult{Allowed: false, Err: err}, nil
	}

	return ListCommentsResult{Comments: comments, Allowed: true}, nil
}
