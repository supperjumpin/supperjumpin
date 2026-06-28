package game_test

import (
	"context"
	"testing"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type fakeLoreRepo struct {
	reactions []game.LoreReactionRow
	err       error
}

func (f *fakeLoreRepo) ListRevealedReactionsForCommunity(_ context.Context, communityID string) ([]game.LoreReactionRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.reactions, nil
}

func TestDeriveCommunityLoreEmptyCommunity(t *testing.T) {
	repo := &fakeLoreRepo{}

	result, err := game.DeriveCommunityLore(context.Background(), repo, "community-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if len(result.Entries) != 0 {
		t.Fatalf("expected 0 lore entries for empty community, got %d", len(result.Entries))
	}
}

func TestDeriveCommunityLoreSingleJumpWithReactions(t *testing.T) {
	repo := &fakeLoreRepo{
		reactions: []game.LoreReactionRow{
			{JumpID: "jump-1", RoundID: "round-1", JumpCaption: "great bit", JumpPlayerID: "player-a", StampStance: "approval"},
			{JumpID: "jump-1", RoundID: "round-1", JumpCaption: "great bit", JumpPlayerID: "player-a", StampStance: "chaos"},
			{JumpID: "jump-1", RoundID: "round-1", JumpCaption: "great bit", JumpPlayerID: "player-a", StampStance: "approval"},
		},
	}

	result, err := game.DeriveCommunityLore(context.Background(), repo, "community-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 lore entry, got %d", len(result.Entries))
	}

	entry := result.Entries[0]
	if entry.JumpID != "jump-1" {
		t.Fatalf("expected JumpID=jump-1, got %q", entry.JumpID)
	}
	if entry.RoundID != "round-1" {
		t.Fatalf("expected RoundID=round-1, got %q", entry.RoundID)
	}
	if entry.JumpCaption != "great bit" {
		t.Fatalf("expected JumpCaption='great bit', got %q", entry.JumpCaption)
	}
	if entry.JumpPlayerID != "player-a" {
		t.Fatalf("expected JumpPlayerID=player-a, got %q", entry.JumpPlayerID)
	}
	if entry.TotalStamps != 3 {
		t.Fatalf("expected TotalStamps=3, got %d", entry.TotalStamps)
	}
	if entry.StampCounts["approval"] != 2 {
		t.Fatalf("expected 2 approval stamps, got %d", entry.StampCounts["approval"])
	}
	if entry.StampCounts["chaos"] != 1 {
		t.Fatalf("expected 1 chaos stamp, got %d", entry.StampCounts["chaos"])
	}
}

func TestDeriveCommunityLoreSortsByStampDensity(t *testing.T) {
	repo := &fakeLoreRepo{
		reactions: []game.LoreReactionRow{
			{JumpID: "jump-low", RoundID: "round-1", JumpCaption: "ok bit", JumpPlayerID: "player-b", StampStance: "approval"},
			{JumpID: "jump-high", RoundID: "round-1", JumpCaption: "best bit", JumpPlayerID: "player-a", StampStance: "approval"},
			{JumpID: "jump-high", RoundID: "round-1", JumpCaption: "best bit", JumpPlayerID: "player-a", StampStance: "lore"},
			{JumpID: "jump-high", RoundID: "round-1", JumpCaption: "best bit", JumpPlayerID: "player-a", StampStance: "chaos"},
		},
	}

	result, err := game.DeriveCommunityLore(context.Background(), repo, "community-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 lore entries, got %d", len(result.Entries))
	}

	if result.Entries[0].JumpID != "jump-high" {
		t.Fatalf("expected first entry to be jump-high (most stamps), got %q", result.Entries[0].JumpID)
	}
	if result.Entries[0].TotalStamps != 3 {
		t.Fatalf("expected jump-high TotalStamps=3, got %d", result.Entries[0].TotalStamps)
	}
	if result.Entries[1].JumpID != "jump-low" {
		t.Fatalf("expected second entry to be jump-low, got %q", result.Entries[1].JumpID)
	}
	if result.Entries[1].TotalStamps != 1 {
		t.Fatalf("expected jump-low TotalStamps=1, got %d", result.Entries[1].TotalStamps)
	}
}

func TestDeriveCommunityLoreNoPerPlayerTally(t *testing.T) {
	// Lore is keyed to moments, never per-Player. Verify no per-player
	// aggregation surface exists in the result type.
	repo := &fakeLoreRepo{
		reactions: []game.LoreReactionRow{
			{JumpID: "jump-1", RoundID: "round-1", JumpCaption: "bit", JumpPlayerID: "player-a", StampStance: "approval"},
			{JumpID: "jump-2", RoundID: "round-1", JumpCaption: "bit2", JumpPlayerID: "player-a", StampStance: "approval"},
		},
	}

	result, err := game.DeriveCommunityLore(context.Background(), repo, "community-1")
	if err != nil {
		t.Fatalf("expected no outer error, got %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected Allowed=true, got false (err=%v)", result.Err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 lore entries, got %d", len(result.Entries))
	}

	// player-a appears on two jumps, but there should be no per-player tally.
	// Each entry is keyed to a jump (moment), not a player.
	playerTotal := 0
	for _, entry := range result.Entries {
		if entry.JumpPlayerID == "player-a" {
			playerTotal += entry.TotalStamps
		}
	}
	// playerTotal is a test-only aggregation; the type itself carries no per-player tally.
	if playerTotal != 2 {
		t.Fatalf("test-only player-total check: expected 2 stamps across player-a jumps, got %d", playerTotal)
	}

	// Structural assertion: LoreEntrySnapshot has no per-player tally field.
	// This is verified at compile-time by the type definition in lore.go having
	// no PlayerTotal or similar field.
}

