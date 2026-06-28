package game_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

var recapFrozenNow = time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)

// --- fake RecapRepo ---

type fakeRecapRepo struct {
	round          game.RoundSnapshot
	roundExists    bool
	jumps          []game.JumpSnapshot
	evidence       map[string][]string
	reactions      []game.RecapReactionRow
	comments       []game.CommentSnapshot
	ghostJumpers   []game.RecapGhostJumperRow
	loreReactions  []game.LoreReactionRow
	err            error
}

func (f *fakeRecapRepo) GetRound(_ context.Context, roundID string) (game.RoundSnapshot, bool, error) {
	if f.err != nil {
		return game.RoundSnapshot{}, false, f.err
	}
	return f.round, f.roundExists, nil
}

func (f *fakeRecapRepo) ListJumpsWithContent(_ context.Context, roundID string) ([]game.JumpSnapshot, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.jumps, nil
}

func (f *fakeRecapRepo) ListEvidence(_ context.Context, jumpIDs []string) (map[string][]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.evidence, nil
}

func (f *fakeRecapRepo) ListReactionsForRound(_ context.Context, roundID string) ([]game.RecapReactionRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.reactions, nil
}

func (f *fakeRecapRepo) ListAllCommentsForRound(_ context.Context, roundID string) ([]game.CommentSnapshot, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.comments, nil
}

func (f *fakeRecapRepo) ListGhostJumpers(_ context.Context, roundID string) ([]game.RecapGhostJumperRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.ghostJumpers, nil
}

func (f *fakeRecapRepo) ListRevealedReactionsForCommunity(_ context.Context, communityID string) ([]game.LoreReactionRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.loreReactions, nil
}

func revealedRound() game.RoundSnapshot {
	return game.RoundSnapshot{
		ID:          "round-1",
		CommunityID: "community-1",
		PromptID:    "prompt-1",
		Status:      "revealed",
		RevealBy:    recapFrozenNow,
		CreatedBy:   "player-initiator",
		CreatedAt:   recapFrozenNow,
	}
}

func jumpSnap(id, playerID, caption string) game.JumpSnapshot {
	return game.JumpSnapshot{
		ID:          id,
		RoundID:     "round-1",
		PlayerID:    playerID,
		Caption:     caption,
		SubmittedAt: recapFrozenNow,
	}
}

func TestAssembleRecapRoundNotFound(t *testing.T) {
	repo := &fakeRecapRepo{roundExists: false}

	result, err := game.AssembleRecap(context.Background(), repo, "nonexistent")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for nonexistent round")
	}
	if result.Err != game.ErrRoundNotFound {
		t.Fatalf("expected ErrRoundNotFound, got %v", result.Err)
	}
}

func TestAssembleRecapRoundNotRevealed(t *testing.T) {
	repo := &fakeRecapRepo{
		roundExists: true,
		round: game.RoundSnapshot{
			ID:          "round-1",
			CommunityID: "community-1",
			Status:      "active",
		},
	}

	result, err := game.AssembleRecap(context.Background(), repo, "round-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false for non-revealed round")
	}
	if result.Err != game.ErrRoundNotRevealed {
		t.Fatalf("expected ErrRoundNotRevealed, got %v", result.Err)
	}
}

func TestAssembleRecapSuccessWithNoData(t *testing.T) {
	repo := &fakeRecapRepo{
		roundExists: true,
		round:       revealedRound(),
	}

	result, err := game.AssembleRecap(context.Background(), repo, "round-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}

	recap := result.Recap
	if recap.RoundID != "round-1" {
		t.Fatalf("expected RoundID=round-1, got %q", recap.RoundID)
	}
	if recap.CommunityID != "community-1" {
		t.Fatalf("expected CommunityID=community-1, got %q", recap.CommunityID)
	}
	if recap.PromptID != "prompt-1" {
		t.Fatalf("expected PromptID=prompt-1, got %q", recap.PromptID)
	}
	if recap.Status != "revealed" {
		t.Fatalf("expected Status=revealed, got %q", recap.Status)
	}
	if len(recap.Jumps) != 0 {
		t.Fatalf("expected 0 jumps, got %d", len(recap.Jumps))
	}
	if len(recap.Comments) != 0 {
		t.Fatalf("expected 0 comments, got %d", len(recap.Comments))
	}
	if len(recap.GhostJumpers) != 0 {
		t.Fatalf("expected 0 ghost jumpers, got %d", len(recap.GhostJumpers))
	}
	if len(recap.Lore) != 0 {
		t.Fatalf("expected 0 lore entries, got %d", len(recap.Lore))
	}
}

func TestAssembleRecapJumpsWithStampsAndEvidence(t *testing.T) {
	repo := &fakeRecapRepo{
		roundExists: true,
		round:       revealedRound(),
		jumps: []game.JumpSnapshot{
			jumpSnap("jump-1", "player-a", "great bit"),
		},
		evidence: map[string][]string{
			"jump-1": {"url-1", "url-2"},
		},
		reactions: []game.RecapReactionRow{
			{JumpID: "jump-1", StampStance: "approval"},
			{JumpID: "jump-1", StampStance: "chaos"},
			{JumpID: "jump-1", StampStance: "approval"},
		},
	}

	result, err := game.AssembleRecap(context.Background(), repo, "round-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}

	recap := result.Recap
	if len(recap.Jumps) != 1 {
		t.Fatalf("expected 1 jump, got %d", len(recap.Jumps))
	}

	j := recap.Jumps[0]
	if j.JumpID != "jump-1" {
		t.Fatalf("expected JumpID=jump-1, got %q", j.JumpID)
	}
	if j.PlayerID != "player-a" {
		t.Fatalf("expected PlayerID=player-a, got %q", j.PlayerID)
	}
	if j.Caption != "great bit" {
		t.Fatalf("expected Caption='great bit', got %q", j.Caption)
	}
	if len(j.EvidenceURLs) != 2 {
		t.Fatalf("expected 2 evidence URLs, got %d", len(j.EvidenceURLs))
	}
	if j.EvidenceURLs[0] != "url-1" {
		t.Fatalf("expected evidence URL 'url-1', got %q", j.EvidenceURLs[0])
	}
	if j.TotalStamps != 3 {
		t.Fatalf("expected TotalStamps=3, got %d", j.TotalStamps)
	}
	if j.StampCounts["approval"] != 2 {
		t.Fatalf("expected 2 approval stamps, got %d", j.StampCounts["approval"])
	}
	if j.StampCounts["chaos"] != 1 {
		t.Fatalf("expected 1 chaos stamp, got %d", j.StampCounts["chaos"])
	}
}

func TestAssembleRecapComments(t *testing.T) {
	repo := &fakeRecapRepo{
		roundExists: true,
		round:       revealedRound(),
		comments: []game.CommentSnapshot{
			{ID: "c1", RoundID: "round-1", PlayerID: "player-a", Body: "lol", CreatedAt: recapFrozenNow},
			{ID: "c2", RoundID: "round-1", JumpID: "jump-1", PlayerID: "player-b", Body: "nice!", CreatedAt: recapFrozenNow},
		},
	}

	result, err := game.AssembleRecap(context.Background(), repo, "round-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}

	if len(result.Recap.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(result.Recap.Comments))
	}
	if result.Recap.Comments[0].Body != "lol" {
		t.Fatalf("expected first comment body='lol', got %q", result.Recap.Comments[0].Body)
	}
	if result.Recap.Comments[1].Body != "nice!" {
		t.Fatalf("expected second comment body='nice!', got %q", result.Recap.Comments[1].Body)
	}
	if result.Recap.Comments[1].JumpID != "jump-1" {
		t.Fatalf("expected second comment JumpID='jump-1', got %q", result.Recap.Comments[1].JumpID)
	}
}

func TestAssembleRecapGhostJumpers(t *testing.T) {
	repo := &fakeRecapRepo{
		roundExists: true,
		round:       revealedRound(),
		ghostJumpers: []game.RecapGhostJumperRow{
			{PlayerID: "player-ghost", CommittedAt: recapFrozenNow},
		},
	}

	result, err := game.AssembleRecap(context.Background(), repo, "round-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}

	if len(result.Recap.GhostJumpers) != 1 {
		t.Fatalf("expected 1 ghost jumper, got %d", len(result.Recap.GhostJumpers))
	}
	if result.Recap.GhostJumpers[0].PlayerID != "player-ghost" {
		t.Fatalf("expected ghost jumper playerID='player-ghost', got %q", result.Recap.GhostJumpers[0].PlayerID)
	}
}

func TestAssembleRecapLore(t *testing.T) {
	repo := &fakeRecapRepo{
		roundExists: true,
		round:       revealedRound(),
		loreReactions: []game.LoreReactionRow{
			{JumpID: "jump-old", RoundID: "round-0", JumpCaption: "classic", JumpPlayerID: "player-a", StampStance: "approval"},
			{JumpID: "jump-old", RoundID: "round-0", JumpCaption: "classic", JumpPlayerID: "player-a", StampStance: "lore"},
		},
	}

	result, err := game.AssembleRecap(context.Background(), repo, "round-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}

	if len(result.Recap.Lore) != 1 {
		t.Fatalf("expected 1 lore entry, got %d", len(result.Recap.Lore))
	}
	lore := result.Recap.Lore[0]
	if lore.JumpID != "jump-old" {
		t.Fatalf("expected lore JumpID=jump-old, got %q", lore.JumpID)
	}
	if lore.JumpCaption != "classic" {
		t.Fatalf("expected lore JumpCaption=classic, got %q", lore.JumpCaption)
	}
	if lore.TotalStamps != 2 {
		t.Fatalf("expected lore TotalStamps=2, got %d", lore.TotalStamps)
	}
}

func TestAssembleRecapFullAssembly(t *testing.T) {
	repo := &fakeRecapRepo{
		roundExists: true,
		round:       revealedRound(),
		jumps: []game.JumpSnapshot{
			jumpSnap("jump-1", "player-a", "first"),
			jumpSnap("jump-2", "player-b", "second"),
		},
		evidence: map[string][]string{
			"jump-1": {"ev1"},
			"jump-2": {"ev2"},
		},
		reactions: []game.RecapReactionRow{
			{JumpID: "jump-1", StampStance: "approval"},
			{JumpID: "jump-2", StampStance: "approval"},
			{JumpID: "jump-2", StampStance: "chaos"},
		},
		comments: []game.CommentSnapshot{
			{ID: "c1", RoundID: "round-1", PlayerID: "player-c", Body: "round comment", CreatedAt: recapFrozenNow},
		},
		ghostJumpers: []game.RecapGhostJumperRow{
			{PlayerID: "player-ghost", CommittedAt: recapFrozenNow},
		},
		loreReactions: []game.LoreReactionRow{
			{JumpID: "jump-lore", RoundID: "round-0", JumpCaption: "lore bit", JumpPlayerID: "player-z", StampStance: "lore"},
		},
	}

	result, err := game.AssembleRecap(context.Background(), repo, "round-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}

	recap := result.Recap

	// Jumps
	if len(recap.Jumps) != 2 {
		t.Fatalf("expected 2 jumps, got %d", len(recap.Jumps))
	}
	if recap.Jumps[0].TotalStamps != 1 {
		t.Fatalf("expected jump-1 TotalStamps=1, got %d", recap.Jumps[0].TotalStamps)
	}
	if recap.Jumps[1].TotalStamps != 2 {
		t.Fatalf("expected jump-2 TotalStamps=2, got %d", recap.Jumps[1].TotalStamps)
	}

	// Comments
	if len(recap.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(recap.Comments))
	}

	// Ghost jumpers
	if len(recap.GhostJumpers) != 1 {
		t.Fatalf("expected 1 ghost jumper, got %d", len(recap.GhostJumpers))
	}

	// Lore
	if len(recap.Lore) != 1 {
		t.Fatalf("expected 1 lore entry, got %d", len(recap.Lore))
	}
}

func TestAssembleRecapRepoErrorPropagated(t *testing.T) {
	repo := &fakeRecapRepo{err: errors.New("db down")}

	result, err := game.AssembleRecap(context.Background(), repo, "round-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if result.Allowed {
		t.Fatal("expected Allowed=false on repo error")
	}
	if result.Err == nil || result.Err.Error() != "db down" {
		t.Fatalf("expected db down error, got %v", result.Err)
	}
}
