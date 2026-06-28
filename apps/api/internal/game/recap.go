package game

import (
	"context"
	"sort"
	"time"
)

// --- recap row types ---

type RecapReactionRow struct {
	JumpID      string
	StampStance string
}

type RecapGhostJumperRow struct {
	PlayerID    string
	CommittedAt time.Time
}

// --- recap snapshots ---

type RecapJumpEntry struct {
	JumpID       string
	PlayerID     string
	Caption      string
	EvidenceURLs []string
	SubmittedAt  time.Time
	StampCounts  map[string]int
	TotalStamps  int
}

type GhostJumperEntry struct {
	PlayerID    string
	CommittedAt time.Time
}

type RecapSnapshot struct {
	RoundID      string
	CommunityID  string
	PromptID     string
	Status       string
	RevealBy     time.Time
	CreatedBy    string
	CreatedAt    time.Time
	Jumps        []RecapJumpEntry
	Comments     []CommentSnapshot
	GhostJumpers []GhostJumperEntry
	Lore         []LoreEntrySnapshot
}

type RecapResult struct {
	Recap   RecapSnapshot
	Allowed bool
	Err     error
}

// --- recap repo ---

type RecapRepo interface {
	GetRound(ctx context.Context, roundID string) (RoundSnapshot, bool, error)
	ListJumpsWithContent(ctx context.Context, roundID string) ([]JumpSnapshot, error)
	ListEvidence(ctx context.Context, jumpIDs []string) (map[string][]string, error)
	ListReactionsForRound(ctx context.Context, roundID string) ([]RecapReactionRow, error)
	ListAllCommentsForRound(ctx context.Context, roundID string) ([]CommentSnapshot, error)
	ListGhostJumpers(ctx context.Context, roundID string) ([]RecapGhostJumperRow, error)
	ListRevealedReactionsForCommunity(ctx context.Context, communityID string) ([]LoreReactionRow, error)
}

// --- domain function ---

func AssembleRecap(ctx context.Context, repo RecapRepo, roundID string) (RecapResult, error) {
	round, ok, err := repo.GetRound(ctx, roundID)
	if err != nil {
		return RecapResult{Allowed: false, Err: err}, nil
	}
	if !ok {
		return RecapResult{Allowed: false, Err: ErrRoundNotFound}, nil
	}
	if round.Status != "revealed" {
		return RecapResult{Allowed: false, Err: ErrRoundNotRevealed}, nil
	}

	// --- jumps with evidence ---
	jumps, err := repo.ListJumpsWithContent(ctx, roundID)
	if err != nil {
		return RecapResult{Allowed: false, Err: err}, nil
	}

	jumpIDs := make([]string, 0, len(jumps))
	for _, j := range jumps {
		jumpIDs = append(jumpIDs, j.ID)
	}
	evidenceMap, err := repo.ListEvidence(ctx, jumpIDs)
	if err != nil {
		return RecapResult{Allowed: false, Err: err}, nil
	}

	// --- reaction summaries per jump ---
	reactions, err := repo.ListReactionsForRound(ctx, roundID)
	if err != nil {
		return RecapResult{Allowed: false, Err: err}, nil
	}

	reactionByJump := make(map[string][]RecapReactionRow)
	for _, rxn := range reactions {
		reactionByJump[rxn.JumpID] = append(reactionByJump[rxn.JumpID], rxn)
	}

	// --- comments ---
	comments, err := repo.ListAllCommentsForRound(ctx, roundID)
	if err != nil {
		return RecapResult{Allowed: false, Err: err}, nil
	}

	// --- ghost jumpers ---
	ghostRows, err := repo.ListGhostJumpers(ctx, roundID)
	if err != nil {
		return RecapResult{Allowed: false, Err: err}, nil
	}
	ghostJumpers := make([]GhostJumperEntry, 0, len(ghostRows))
	for _, g := range ghostRows {
		ghostJumpers = append(ghostJumpers, GhostJumperEntry{
			PlayerID:    g.PlayerID,
			CommittedAt: g.CommittedAt,
		})
	}

	// --- assemble jumps with stamp counts ---
	recapJumps := make([]RecapJumpEntry, 0, len(jumps))
	for _, j := range jumps {
		entry := RecapJumpEntry{
			JumpID:       j.ID,
			PlayerID:     j.PlayerID,
			Caption:      j.Caption,
			EvidenceURLs: evidenceMap[j.ID],
			SubmittedAt:  j.SubmittedAt,
			StampCounts:  make(map[string]int),
		}

		for _, rxn := range reactionByJump[j.ID] {
			entry.StampCounts[rxn.StampStance]++
			entry.TotalStamps++
		}

		recapJumps = append(recapJumps, entry)
	}

	// --- lore ---
	loreRows, err := repo.ListRevealedReactionsForCommunity(ctx, round.CommunityID)
	if err != nil {
		return RecapResult{Allowed: false, Err: err}, nil
	}

	type jumpGroup struct {
		jumpID       string
		roundID      string
		jumpCaption  string
		jumpPlayerID string
		stanceCounts map[string]int
	}
	groups := make(map[string]*jumpGroup)
	for _, row := range loreRows {
		g, ok2 := groups[row.JumpID]
		if !ok2 {
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

	lore := make([]LoreEntrySnapshot, 0, len(groups))
	for _, g := range groups {
		total := 0
		for _, c := range g.stanceCounts {
			total += c
		}
		lore = append(lore, LoreEntrySnapshot{
			JumpID:       g.jumpID,
			RoundID:      g.roundID,
			JumpCaption:  g.jumpCaption,
			JumpPlayerID: g.jumpPlayerID,
			StampCounts:  g.stanceCounts,
			TotalStamps:  total,
		})
	}
	sort.Slice(lore, func(i, j int) bool {
		return lore[i].TotalStamps > lore[j].TotalStamps
	})

	recap := RecapSnapshot{
		RoundID:      round.ID,
		CommunityID:  round.CommunityID,
		PromptID:     round.PromptID,
		Status:       round.Status,
		RevealBy:     round.RevealBy,
		CreatedBy:    round.CreatedBy,
		CreatedAt:    round.CreatedAt,
		Jumps:        recapJumps,
		Comments:     comments,
		GhostJumpers: ghostJumpers,
		Lore:         lore,
	}

	return RecapResult{Recap: recap, Allowed: true}, nil
}
