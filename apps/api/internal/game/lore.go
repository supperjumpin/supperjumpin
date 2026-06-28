package game

import (
	"context"
	"sort"
)

type LoreReactionRow struct {
	JumpID       string
	StampStance  string
	RoundID      string
	JumpCaption  string
	JumpPlayerID string
}

type LoreEntrySnapshot struct {
	JumpID       string
	RoundID      string
	JumpCaption  string
	JumpPlayerID string
	StampCounts  map[string]int
	TotalStamps  int
}

type CommunityLoreResult struct {
	Entries []LoreEntrySnapshot
	Allowed bool
	Err     error
}

type LoreRepo interface {
	ListRevealedReactionsForCommunity(ctx context.Context, communityID string) ([]LoreReactionRow, error)
}

func DeriveCommunityLore(ctx context.Context, repo LoreRepo, communityID string) (CommunityLoreResult, error) {
	rows, err := repo.ListRevealedReactionsForCommunity(ctx, communityID)
	if err != nil {
		return CommunityLoreResult{Allowed: false, Err: err}, nil
	}

	type jumpGroup struct {
		jumpID       string
		roundID      string
		jumpCaption  string
		jumpPlayerID string
		stanceCounts map[string]int
	}

	groups := make(map[string]*jumpGroup)

	for _, row := range rows {
		g, ok := groups[row.JumpID]
		if !ok {
			g = &jumpGroup{
				jumpID:       row.JumpID,
				roundID:      row.RoundID,
				jumpCaption:  row.JumpCaption,
				jumpPlayerID: row.JumpPlayerID,
				stanceCounts: make(map[string]int),
			}
			groups[row.JumpID] = g
		}
		g.stanceCounts[row.StampStance]++
	}

	entries := make([]LoreEntrySnapshot, 0, len(groups))
	for _, g := range groups {
		total := 0
		for _, c := range g.stanceCounts {
			total += c
		}
		entries = append(entries, LoreEntrySnapshot{
			JumpID:       g.jumpID,
			RoundID:      g.roundID,
			JumpCaption:  g.jumpCaption,
			JumpPlayerID: g.jumpPlayerID,
			StampCounts:  g.stanceCounts,
			TotalStamps:  total,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].TotalStamps > entries[j].TotalStamps
	})

	return CommunityLoreResult{Entries: entries, Allowed: true}, nil
}
