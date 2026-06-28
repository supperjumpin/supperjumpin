package game_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

var commentFrozenNow = time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)

// --- fake PostCommentRepo ---

type fakePostCommentRepo struct {
	rounds       map[string]game.RoundSnapshot
	players      map[string]game.PlayerSnapshot
	jumps        map[string]game.JumpSnapshot
	comments     []game.CommentSnapshot
	getRoundErr  error
	getJumpErr   error
	createErr    error
}

func newFakePostCommentRepo() *fakePostCommentRepo {
	return &fakePostCommentRepo{
		rounds:  make(map[string]game.RoundSnapshot),
		players: make(map[string]game.PlayerSnapshot),
		jumps:   make(map[string]game.JumpSnapshot),
	}
}

func (f *fakePostCommentRepo) GetRound(_ context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	if f.getRoundErr != nil {
		return game.RoundSnapshot{}, false, f.getRoundErr
	}
	r, ok := f.rounds[roundID]
	return r, ok, nil
}

func (f *fakePostCommentRepo) FindPlayer(_ context.Context, playerID string) (game.PlayerSnapshot, bool, error) {
	p, ok := f.players[playerID]
	return p, ok, nil
}

func (f *fakePostCommentRepo) GetJump(_ context.Context, jumpID string) (game.JumpSnapshot, error) {
	if f.getJumpErr != nil {
		return game.JumpSnapshot{}, f.getJumpErr
	}
	j, ok := f.jumps[jumpID]
	if !ok {
		return game.JumpSnapshot{}, game.ErrJumpNotFound
	}
	return j, nil
}

func (f *fakePostCommentRepo) CreateComment(_ context.Context, c game.CommentSnapshot) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.comments = append(f.comments, c)
	return nil
}

func TestPostCommentOnRevealedJump(t *testing.T) {
	ctx := context.Background()
	repo := newFakePostCommentRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}
	repo.jumps["jump-1"] = game.JumpSnapshot{ID: "jump-1", RoundID: "round-1"}

	result, err := game.PostComment(ctx, repo, game.PostCommentInput{
		RoundID:  "round-1",
		JumpID:   "jump-1",
		PlayerID: "player-a",
		Body:     "Great jump!",
	}, commentFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if result.Comment.RoundID != "round-1" {
		t.Fatalf("expected RoundID round-1, got %s", result.Comment.RoundID)
	}
	if result.Comment.JumpID != "jump-1" {
		t.Fatalf("expected JumpID jump-1, got %s", result.Comment.JumpID)
	}
	if result.Comment.PlayerID != "player-a" {
		t.Fatalf("expected PlayerID player-a, got %s", result.Comment.PlayerID)
	}
	if result.Comment.Body != "Great jump!" {
		t.Fatalf("expected Body 'Great jump!', got %q", result.Comment.Body)
	}
	if result.Comment.CreatedAt != commentFrozenNow {
		t.Fatalf("expected CreatedAt %v, got %v", commentFrozenNow, result.Comment.CreatedAt)
	}
	if len(repo.comments) != 1 {
		t.Fatalf("expected 1 comment created, got %d", len(repo.comments))
	}
}

func TestPostCommentOnRevealedRound(t *testing.T) {
	ctx := context.Background()
	repo := newFakePostCommentRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}

	result, err := game.PostComment(ctx, repo, game.PostCommentInput{
		RoundID:  "round-1",
		JumpID:   "",
		PlayerID: "player-a",
		Body:     "What a round!",
	}, commentFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if result.Comment.RoundID != "round-1" {
		t.Fatalf("expected RoundID round-1, got %s", result.Comment.RoundID)
	}
	if result.Comment.JumpID != "" {
		t.Fatalf("expected empty JumpID for round-level comment, got %s", result.Comment.JumpID)
	}
	if result.Comment.Body != "What a round!" {
		t.Fatalf("expected Body 'What a round!', got %q", result.Comment.Body)
	}
	if len(repo.comments) != 1 {
		t.Fatalf("expected 1 comment created, got %d", len(repo.comments))
	}
}

func TestPostCommentFailsOnNonRevealedRound(t *testing.T) {
	ctx := context.Background()
	repo := newFakePostCommentRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "active"}
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}
	repo.jumps["jump-1"] = game.JumpSnapshot{ID: "jump-1", RoundID: "round-1"}

	result, err := game.PostComment(ctx, repo, game.PostCommentInput{
		RoundID:  "round-1",
		JumpID:   "jump-1",
		PlayerID: "player-a",
		Body:     "Too early!",
	}, commentFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when round not revealed")
	}
	if !errors.Is(result.Err, game.ErrRoundNotRevealed) {
		t.Fatalf("expected ErrRoundNotRevealed, got %v", result.Err)
	}
}

func TestPostCommentFailsOnUnknownJump(t *testing.T) {
	ctx := context.Background()
	repo := newFakePostCommentRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}

	result, err := game.PostComment(ctx, repo, game.PostCommentInput{
		RoundID:  "round-1",
		JumpID:   "jump-missing",
		PlayerID: "player-a",
		Body:     "Where's the jump?",
	}, commentFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when jump not found")
	}
	if !errors.Is(result.Err, game.ErrJumpNotFound) {
		t.Fatalf("expected ErrJumpNotFound, got %v", result.Err)
	}
}

func TestPostCommentFailsOnUnknownPlayer(t *testing.T) {
	ctx := context.Background()
	repo := newFakePostCommentRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.jumps["jump-1"] = game.JumpSnapshot{ID: "jump-1", RoundID: "round-1"}

	result, err := game.PostComment(ctx, repo, game.PostCommentInput{
		RoundID:  "round-1",
		JumpID:   "jump-1",
		PlayerID: "player-missing",
		Body:     "Who am I?",
	}, commentFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when player not found")
	}
	if !errors.Is(result.Err, game.ErrPlayerNotFound) {
		t.Fatalf("expected ErrPlayerNotFound, got %v", result.Err)
	}
}

func TestPostCommentFailsOnEmptyBody(t *testing.T) {
	ctx := context.Background()
	repo := newFakePostCommentRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.players["player-a"] = game.PlayerSnapshot{ID: "player-a", DisplayName: "Alice"}
	repo.jumps["jump-1"] = game.JumpSnapshot{ID: "jump-1", RoundID: "round-1"}

	result, err := game.PostComment(ctx, repo, game.PostCommentInput{
		RoundID:  "round-1",
		JumpID:   "jump-1",
		PlayerID: "player-a",
		Body:     "",
	}, commentFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when body is empty")
	}
	if !errors.Is(result.Err, game.ErrCommentBodyEmpty) {
		t.Fatalf("expected ErrCommentBodyEmpty, got %v", result.Err)
	}
}

func TestPostCommentNonJumperCanComment(t *testing.T) {
	ctx := context.Background()
	repo := newFakePostCommentRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.players["player-b"] = game.PlayerSnapshot{ID: "player-b", DisplayName: "Bob"}
	repo.jumps["jump-1"] = game.JumpSnapshot{ID: "jump-1", RoundID: "round-1", PlayerID: "player-a"}

	result, err := game.PostComment(ctx, repo, game.PostCommentInput{
		RoundID:  "round-1",
		JumpID:   "jump-1",
		PlayerID: "player-b",
		Body:     "Nice work!",
	}, commentFrozenNow)

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected non-jumper to be allowed to comment, got err=%v", result.Err)
	}
	if len(repo.comments) != 1 {
		t.Fatalf("expected 1 comment created, got %d", len(repo.comments))
	}
}

// --- fake ListCommentsRepo ---

type fakeListCommentsRepo struct {
	rounds   map[string]game.RoundSnapshot
	comments []game.CommentSnapshot
	listErr  error
}

func newFakeListCommentsRepo() *fakeListCommentsRepo {
	return &fakeListCommentsRepo{
		rounds: make(map[string]game.RoundSnapshot),
	}
}

func (f *fakeListCommentsRepo) GetRound(_ context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	r, ok := f.rounds[roundID]
	return r, ok, nil
}

func (f *fakeListCommentsRepo) ListComments(_ context.Context, roundID, jumpID string) ([]game.CommentSnapshot, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	var result []game.CommentSnapshot
	for _, c := range f.comments {
		if c.RoundID == roundID && c.JumpID == jumpID {
			result = append(result, c)
		}
	}
	return result, nil
}

func TestListCommentsForJump(t *testing.T) {
	ctx := context.Background()
	repo := newFakeListCommentsRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.comments = []game.CommentSnapshot{
		{ID: "comment-1", RoundID: "round-1", JumpID: "jump-1", PlayerID: "player-a", Body: "First!"},
		{ID: "comment-2", RoundID: "round-1", JumpID: "jump-1", PlayerID: "player-b", Body: "Second!"},
		{ID: "comment-3", RoundID: "round-1", JumpID: "", PlayerID: "player-a", Body: "Round chat"},
	}

	result, err := game.ListComments(ctx, repo, game.ListCommentsInput{
		RoundID: "round-1",
		JumpID:  "jump-1",
	})

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if len(result.Comments) != 2 {
		t.Fatalf("expected 2 jump comments, got %d", len(result.Comments))
	}
}

func TestListCommentsForRound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeListCommentsRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.comments = []game.CommentSnapshot{
		{ID: "comment-1", RoundID: "round-1", JumpID: "jump-1", PlayerID: "player-a", Body: "Jump comment"},
		{ID: "comment-2", RoundID: "round-1", JumpID: "", PlayerID: "player-a", Body: "Round chat"},
	}

	result, err := game.ListComments(ctx, repo, game.ListCommentsInput{
		RoundID: "round-1",
		JumpID:  "",
	})

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if len(result.Comments) != 1 {
		t.Fatalf("expected 1 round-level comment, got %d", len(result.Comments))
	}
	if result.Comments[0].Body != "Round chat" {
		t.Fatalf("expected 'Round chat', got %q", result.Comments[0].Body)
	}
}

func TestListCommentsRoundNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeListCommentsRepo()

	result, err := game.ListComments(ctx, repo, game.ListCommentsInput{
		RoundID: "round-missing",
	})

	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when round not found")
	}
	if !errors.Is(result.Err, game.ErrRoundNotFound) {
		t.Fatalf("expected ErrRoundNotFound, got %v", result.Err)
	}
}
