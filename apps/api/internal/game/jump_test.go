package game_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// --- fake repos ---

type fakeCommitRepo struct {
	rounds  map[string]game.RoundSnapshot
	commits map[string]map[string]game.CommitSnapshot // roundID -> playerID -> commit
	createErr error
}

func newFakeCommitRepo() *fakeCommitRepo {
	return &fakeCommitRepo{
		rounds:  make(map[string]game.RoundSnapshot),
		commits: make(map[string]map[string]game.CommitSnapshot),
	}
}

func (f *fakeCommitRepo) FindRound(ctx context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	r, ok := f.rounds[roundID]
	return r, ok, nil
}

func (f *fakeCommitRepo) FindCommit(ctx context.Context, roundID, playerID string) (*game.CommitSnapshot, error) {
	if m, ok := f.commits[roundID]; ok {
		if c, ok2 := m[playerID]; ok2 {
			return &c, nil
		}
	}
	return nil, nil
}

func (f *fakeCommitRepo) CreateCommit(ctx context.Context, commit game.CommitSnapshot) error {
	if f.createErr != nil {
		return f.createErr
	}
	if _, ok := f.commits[commit.RoundID]; !ok {
		f.commits[commit.RoundID] = make(map[string]game.CommitSnapshot)
	}
	f.commits[commit.RoundID][commit.PlayerID] = commit
	return nil
}

type fakeSubmitRepo struct {
	rounds  map[string]game.RoundSnapshot
	commits map[string]map[string]game.CommitSnapshot
	jumps   map[string]map[string]game.JumpSnapshot
	evidence map[string][]string
}

func newFakeSubmitRepo() *fakeSubmitRepo {
	return &fakeSubmitRepo{
		rounds:   make(map[string]game.RoundSnapshot),
		commits:  make(map[string]map[string]game.CommitSnapshot),
		jumps:    make(map[string]map[string]game.JumpSnapshot),
		evidence: make(map[string][]string),
	}
}

func (f *fakeSubmitRepo) FindRound(ctx context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	r, ok := f.rounds[roundID]
	return r, ok, nil
}

func (f *fakeSubmitRepo) FindCommit(ctx context.Context, roundID, playerID string) (*game.CommitSnapshot, error) {
	if m, ok := f.commits[roundID]; ok {
		if c, ok2 := m[playerID]; ok2 {
			return &c, nil
		}
	}
	return nil, nil
}

func (f *fakeSubmitRepo) FindJump(ctx context.Context, roundID, playerID string) (*game.JumpSnapshot, error) {
	if m, ok := f.jumps[roundID]; ok {
		if j, ok2 := m[playerID]; ok2 {
			return &j, nil
		}
	}
	return nil, nil
}

func (f *fakeSubmitRepo) CreateJump(ctx context.Context, jump game.JumpSnapshot) error {
	if _, ok := f.jumps[jump.RoundID]; !ok {
		f.jumps[jump.RoundID] = make(map[string]game.JumpSnapshot)
	}
	f.jumps[jump.RoundID][jump.PlayerID] = jump
	return nil
}

func (f *fakeSubmitRepo) InsertEvidence(ctx context.Context, jumpID string, urls []string) error {
	f.evidence[jumpID] = urls
	return nil
}

type fakeListJumpsRepo struct {
	rounds  map[string]game.RoundSnapshot
	jumps   []game.JumpSnapshot
	evidence map[string][]string
	status  game.RoundStatus
	commits map[string]map[string]game.CommitSnapshot
	err     error
}

func newFakeListJumpsRepo() *fakeListJumpsRepo {
	return &fakeListJumpsRepo{
		rounds:   make(map[string]game.RoundSnapshot),
		evidence: make(map[string][]string),
		commits:  make(map[string]map[string]game.CommitSnapshot),
	}
}

func (f *fakeListJumpsRepo) FindRound(ctx context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	r, ok := f.rounds[roundID]
	return r, ok, nil
}

func (f *fakeListJumpsRepo) ListJumps(ctx context.Context, roundID string) ([]game.JumpSnapshot, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.jumps, nil
}

func (f *fakeListJumpsRepo) ListEvidence(ctx context.Context, jumpIDs []string) (map[string][]string, error) {
	return f.evidence, nil
}

func (f *fakeListJumpsRepo) GetRoundStatus(ctx context.Context, roundID string) (game.RoundStatus, error) {
	return f.status, nil
}

func (f *fakeListJumpsRepo) FindCommitForPlayer(ctx context.Context, roundID, playerID string) (*game.CommitSnapshot, error) {
	if m, ok := f.commits[roundID]; ok {
		if c, ok2 := m[playerID]; ok2 {
			return &c, nil
		}
	}
	return nil, nil
}

// --- CommitToRound tests ---

func TestCommitToRoundSuccessfully(t *testing.T) {
	ctx := context.Background()
	repo := newFakeCommitRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "active"}

	now := frozenNow
	result, err := game.CommitToRound(ctx, repo, game.CommitToRoundInput{
		RoundID:  "round-1",
		PlayerID: "player-a",
	}, now)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if result.CommitID == "" {
		t.Fatal("expected non-empty commit ID")
	}

	// Verify it was persisted
	c, _ := repo.FindCommit(ctx, "round-1", "player-a")
	if c == nil {
		t.Fatal("expected commit to be persisted")
	}
	if c.RoundID != "round-1" {
		t.Fatalf("expected round round-1, got %s", c.RoundID)
	}
	if c.PlayerID != "player-a" {
		t.Fatalf("expected player player-a, got %s", c.PlayerID)
	}
}

func TestCommitToRoundFailsWhenRoundNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeCommitRepo()

	result, err := game.CommitToRound(ctx, repo, game.CommitToRoundInput{
		RoundID:  "nonexistent",
		PlayerID: "player-a",
	}, frozenNow)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for nonexistent round")
	}
	if !errors.Is(result.Err, game.ErrRoundNotFound) {
		t.Fatalf("expected ErrRoundNotFound, got %v", result.Err)
	}
}

func TestCommitToRoundFailsWhenAlreadyCommitted(t *testing.T) {
	ctx := context.Background()
	repo := newFakeCommitRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "active"}

	// First commit
	result1, _ := game.CommitToRound(ctx, repo, game.CommitToRoundInput{RoundID: "round-1", PlayerID: "player-a"}, frozenNow)
	if !result1.Allowed {
		t.Fatal("first commit should succeed")
	}

	// Second commit
	result2, err := game.CommitToRound(ctx, repo, game.CommitToRoundInput{RoundID: "round-1", PlayerID: "player-a"}, frozenNow)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result2.Allowed {
		t.Fatal("expected Allowed=false for duplicate commit")
	}
	if !errors.Is(result2.Err, game.ErrAlreadyCommitted) {
		t.Fatalf("expected ErrAlreadyCommitted, got %v", result2.Err)
	}
}

func TestCommitToRoundDifferentPlayers(t *testing.T) {
	ctx := context.Background()
	repo := newFakeCommitRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "active"}

	// Player A commits
	resultA, _ := game.CommitToRound(ctx, repo, game.CommitToRoundInput{RoundID: "round-1", PlayerID: "player-a"}, frozenNow)
	if !resultA.Allowed {
		t.Fatal("player A should commit successfully")
	}

	// Player B commits
	resultB, _ := game.CommitToRound(ctx, repo, game.CommitToRoundInput{RoundID: "round-1", PlayerID: "player-b"}, frozenNow)
	if !resultB.Allowed {
		t.Fatal("player B should commit successfully")
	}

	// Verify both commits
	cA, _ := repo.FindCommit(ctx, "round-1", "player-a")
	if cA == nil {
		t.Fatal("player A commit should exist")
	}
	cB, _ := repo.FindCommit(ctx, "round-1", "player-b")
	if cB == nil {
		t.Fatal("player B commit should exist")
	}
}

// --- SubmitJump tests ---

func TestSubmitJumpSuccessfully(t *testing.T) {
	ctx := context.Background()
	repo := newFakeSubmitRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "active"}
	repo.commits["round-1"] = map[string]game.CommitSnapshot{
		"player-a": {ID: "commit-1", RoundID: "round-1", PlayerID: "player-a", CommittedAt: frozenNow},
	}

	now := frozenNow.Add(time.Hour)
	result, err := game.SubmitJump(ctx, repo, game.SubmitJumpInput{
		RoundID:      "round-1",
		PlayerID:     "player-a",
		Caption:      "Check out this Crunchwrap in a wine glass",
		EvidenceURLs: []string{"https://example.com/photo1.jpg"},
	}, now)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if result.Jump.ID == "" {
		t.Fatal("expected non-empty jump ID")
	}
	if result.Jump.Caption != "Check out this Crunchwrap in a wine glass" {
		t.Fatalf("expected caption preserved, got %q", result.Jump.Caption)
	}
	if len(result.Jump.EvidenceURLs) != 1 {
		t.Fatalf("expected 1 evidence URL, got %d", len(result.Jump.EvidenceURLs))
	}
	if !result.Jump.SubmittedAt.Equal(now) {
		t.Fatalf("expected SubmittedAt=%v, got %v", now, result.Jump.SubmittedAt)
	}

	// Verify evidence persisted
	urls := repo.evidence[result.Jump.ID]
	if len(urls) != 1 {
		t.Fatalf("expected 1 evidence URL in repo, got %d", len(urls))
	}
	if urls[0] != "https://example.com/photo1.jpg" {
		t.Fatalf("expected evidence URL preserved, got %q", urls[0])
	}
}

func TestSubmitJumpFailsWhenNotCommitted(t *testing.T) {
	ctx := context.Background()
	repo := newFakeSubmitRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "active"}

	result, err := game.SubmitJump(ctx, repo, game.SubmitJumpInput{
		RoundID:  "round-1",
		PlayerID: "player-a",
		Caption:  "Trying to submit without committing",
	}, frozenNow)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when not committed")
	}
	if !errors.Is(result.Err, game.ErrCannotSubmit) {
		t.Fatalf("expected ErrCannotSubmit, got %v", result.Err)
	}
}

func TestSubmitJumpFailsWhenAlreadySubmitted(t *testing.T) {
	ctx := context.Background()
	repo := newFakeSubmitRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "active"}
	repo.commits["round-1"] = map[string]game.CommitSnapshot{
		"player-a": {ID: "commit-1", RoundID: "round-1", PlayerID: "player-a", CommittedAt: frozenNow},
	}

	// First submit
	result1, _ := game.SubmitJump(ctx, repo, game.SubmitJumpInput{
		RoundID:  "round-1",
		PlayerID: "player-a",
		Caption:  "First submission",
	}, frozenNow)
	if !result1.Allowed {
		t.Fatal("first submit should succeed")
	}

	// Second submit
	result2, err := game.SubmitJump(ctx, repo, game.SubmitJumpInput{
		RoundID:  "round-1",
		PlayerID: "player-a",
		Caption:  "Second submission",
	}, frozenNow)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result2.Allowed {
		t.Fatal("expected Allowed=false for duplicate submit")
	}
	if !errors.Is(result2.Err, game.ErrAlreadySubmitted) {
		t.Fatalf("expected ErrAlreadySubmitted, got %v", result2.Err)
	}
}

func TestSubmitJumpFailsWhenRoundNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeSubmitRepo()

	result, err := game.SubmitJump(ctx, repo, game.SubmitJumpInput{
		RoundID:  "nonexistent",
		PlayerID: "player-a",
		Caption:  "Test",
	}, frozenNow)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for nonexistent round")
	}
	if !errors.Is(result.Err, game.ErrRoundNotFound) {
		t.Fatalf("expected ErrRoundNotFound, got %v", result.Err)
	}
}

// --- ListJumpsForRound tests ---

func TestListJumpsSealedPreReveal(t *testing.T) {
	ctx := context.Background()
	repo := newFakeListJumpsRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "active"}
	repo.status = game.RoundStatus{ID: "round-1", Status: "active", SubmissionCount: 2, CommitCount: 2}
	repo.jumps = []game.JumpSnapshot{
		{ID: "jump-a", RoundID: "round-1", PlayerID: "player-a", Caption: "Alice's caption", SubmittedAt: frozenNow},
		{ID: "jump-b", RoundID: "round-1", PlayerID: "player-b", Caption: "Bob's caption", SubmittedAt: frozenNow},
	}
	repo.evidence = map[string][]string{
		"jump-a": {"https://example.com/alice.jpg"},
		"jump-b": {"https://example.com/bob.jpg"},
	}

	// Viewer = player-a (author of jump-a)
	result, err := game.ListJumpsForRound(ctx, repo, "round-1", "player-a")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if len(result.Jumps) != 2 {
		t.Fatalf("expected 2 jumps, got %d", len(result.Jumps))
	}

	// player-a should see their own caption (author)
	var jumpA game.JumpSnapshot
	var jumpB game.JumpSnapshot
	for _, j := range result.Jumps {
		if j.PlayerID == "player-a" {
			jumpA = j
		} else {
			jumpB = j
		}
	}

	if jumpA.SealedViewer {
		t.Fatal("author should see their own jump content (unsealed)")
	}
	if jumpA.Caption != "Alice's caption" {
		t.Fatalf("expected author's caption, got %q", jumpA.Caption)
	}
	if len(jumpA.EvidenceURLs) != 1 {
		t.Fatalf("expected author's evidence URLs, got %d", len(jumpA.EvidenceURLs))
	}

	if !jumpB.SealedViewer {
		t.Fatal("non-author should see sealed jump")
	}
	if jumpB.Caption != "" {
		t.Fatalf("expected sealed caption empty, got %q", jumpB.Caption)
	}
	if len(jumpB.EvidenceURLs) != 0 {
		t.Fatalf("expected sealed evidence URLs empty, got %d", len(jumpB.EvidenceURLs))
	}
}

func TestListJumpsRevealedShowsAll(t *testing.T) {
	ctx := context.Background()
	repo := newFakeListJumpsRepo()
	repo.rounds["round-1"] = game.RoundSnapshot{ID: "round-1", Status: "revealed"}
	repo.status = game.RoundStatus{ID: "round-1", Status: "revealed", SubmissionCount: 2, CommitCount: 2}
	repo.jumps = []game.JumpSnapshot{
		{ID: "jump-a", RoundID: "round-1", PlayerID: "player-a", Caption: "Alice's caption", SubmittedAt: frozenNow},
		{ID: "jump-b", RoundID: "round-1", PlayerID: "player-b", Caption: "Bob's caption", SubmittedAt: frozenNow},
	}
	repo.evidence = map[string][]string{
		"jump-a": {"https://example.com/alice.jpg"},
		"jump-b": {"https://example.com/bob.jpg"},
	}

	// Viewer = player-c (not author of either)
	result, err := game.ListJumpsForRound(ctx, repo, "round-1", "player-c")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if len(result.Jumps) != 2 {
		t.Fatalf("expected 2 jumps, got %d", len(result.Jumps))
	}

	for _, j := range result.Jumps {
		if j.SealedViewer {
			t.Fatalf("after reveal, all jumps should be unsealed (player %s still sealed)", j.PlayerID)
		}
		if j.Caption == "" {
			t.Fatalf("after reveal, caption should be visible for player %s", j.PlayerID)
		}
	}
}

func TestListJumpsRoundNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeListJumpsRepo()

	result, err := game.ListJumpsForRound(ctx, repo, "nonexistent", "player-a")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for nonexistent round")
	}
	if !errors.Is(result.Err, game.ErrRoundNotFound) {
		t.Fatalf("expected ErrRoundNotFound, got %v", result.Err)
	}
}

// --- GetJump tests ---

func TestGetJumpAuthorSeesFullContentPreReveal(t *testing.T) {
	repo := &fakeGetJumpRepo{
		round:       game.RoundSnapshot{ID: "round-1", Status: "active"},
		roundExists: true,
		jump:        game.JumpSnapshot{ID: "jump-1", RoundID: "round-1", PlayerID: "player-a", Caption: "My jump", SubmittedAt: frozenNow},
		evidence:    []string{"https://example.com/pic.jpg"},
	}

	result, err := game.GetJump(context.Background(), repo, "jump-1", "player-a")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if result.Jump.SealedViewer {
		t.Fatal("author should see full content pre-reveal")
	}
	if result.Jump.Caption != "My jump" {
		t.Fatalf("expected caption, got %q", result.Jump.Caption)
	}
	if len(result.Jump.EvidenceURLs) != 1 {
		t.Fatalf("expected evidence URLs, got %d", len(result.Jump.EvidenceURLs))
	}
}

func TestGetJumpNonAuthorSeesSealedContentPreReveal(t *testing.T) {
	repo := &fakeGetJumpRepo{
		round:       game.RoundSnapshot{ID: "round-1", Status: "active"},
		roundExists: true,
		jump:        game.JumpSnapshot{ID: "jump-1", RoundID: "round-1", PlayerID: "player-a", Caption: "My jump", SubmittedAt: frozenNow},
		evidence:    []string{"https://example.com/pic.jpg"},
	}

	result, err := game.GetJump(context.Background(), repo, "jump-1", "player-b")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if !result.Jump.SealedViewer {
		t.Fatal("non-author should see sealed content pre-reveal")
	}
	if result.Jump.Caption != "" {
		t.Fatalf("expected empty caption for sealed viewer, got %q", result.Jump.Caption)
	}
}

type fakeGetJumpRepo struct {
	round       game.RoundSnapshot
	roundExists bool
	jump        game.JumpSnapshot
	evidence    []string
	getErr      error
}

func (f *fakeGetJumpRepo) FindRound(ctx context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	return f.round, f.roundExists, nil
}

func (f *fakeGetJumpRepo) GetJumpByID(ctx context.Context, jumpID string) (game.JumpSnapshot, error) {
	if f.getErr != nil {
		return game.JumpSnapshot{}, f.getErr
	}
	return f.jump, nil
}

func (f *fakeGetJumpRepo) ListEvidenceForJump(ctx context.Context, jumpID string) ([]string, error) {
	return f.evidence, nil
}

func TestGetJumpFailsWhenRoundMissing(t *testing.T) {
	// Defensive: an orphaned Jump (its Round no longer exists) should return a
	// clean ErrRoundNotFound instead of silently treating the round as
	// un-revealed (which would surface as opaque sealed content).
	repo := &fakeGetJumpRepo{
		roundExists: false,
		jump:        game.JumpSnapshot{ID: "jump-1", RoundID: "round-missing", PlayerID: "player-a", Caption: "orphan", SubmittedAt: frozenNow},
	}

	result, err := game.GetJump(context.Background(), repo, "jump-1", "player-a")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false when round is missing")
	}
	if !errors.Is(result.Err, game.ErrRoundNotFound) {
		t.Fatalf("expected ErrRoundNotFound, got %v", result.Err)
	}
}
