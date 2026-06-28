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

// StandoutStampEntry is the dominant stance on a single jump in this Round —
// the reaction that jump is "known for". Ties are broken alphabetically by
// stance so derivation is deterministic.
type StandoutStampEntry struct {
	JumpID string
	Stance string
	Count  int
}

type GhostJumperEntry struct {
	PlayerID    string
	CommittedAt time.Time
}

type RecapSnapshot struct {
	RoundID          string
	CommunityID      string
	PromptID         string
	Status           string
	RevealBy         time.Time
	CreatedBy        string
	CreatedAt        time.Time
	Jumps            []RecapJumpEntry
	Comments         []CommentSnapshot
	GhostJumpers     []GhostJumperEntry
	Lore             []LoreEntrySnapshot
	NextRoundHook    NextRoundHookSnapshot
	StandoutStamps   []StandoutStampEntry
	StandoutComments []CommentSnapshot
}

// NextRoundHookSnapshot is the artifact slot the presentation layer uses to
// tease or set up the next Round. The domain surfaces whether an active Round
// already exists for the Community (and its Prompt if so); the wording/voice
// is a presentation concern, not domain state.
type NextRoundHookSnapshot struct {
	ActiveRoundID string // empty when no active Round exists for the Community
	PromptID      string // empty when no active Round exists
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
	FindActiveRound(ctx context.Context, communityID string) (*RoundSnapshot, error)
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

	// --- next-round hook ---
	// Surfaces whether an active Round already exists for the Community (set
	// up by a player after this one revealed). The wording/voice is a
	// presentation concern; the domain only produces the artifact slot.
	nextRoundHook := NextRoundHookSnapshot{}
	if next, err := repo.FindActiveRound(ctx, round.CommunityID); err != nil {
		return RecapResult{Allowed: false, Err: err}, nil
	} else if next != nil {
		nextRoundHook.ActiveRoundID = next.ID
		nextRoundHook.PromptID = next.PromptID
	}

	// --- standout stamps & comments ---
	// Standout Stamps: per jump with stamps, the dominant stance on that
	// jump (highest count; alphabetical tiebreak by stance for determinism).
	// Sorted by the jump's TotalStamps desc, then JumpID asc.
	// Standout Comments: comments posted on the jump(s) with the highest
	// TotalStamps in this Round (the Round's "moment(s)"). Round-level
	// comments (no JumpID) are part of the flat Comments list, not standouts.
	standoutStamps, standoutComments := deriveStandouts(recapJumps, comments)

	recap := RecapSnapshot{
		RoundID:          round.ID,
		CommunityID:      round.CommunityID,
		PromptID:         round.PromptID,
		Status:           round.Status,
		RevealBy:         round.RevealBy,
		CreatedBy:        round.CreatedBy,
		CreatedAt:        round.CreatedAt,
		Jumps:            recapJumps,
		Comments:         comments,
		GhostJumpers:     ghostJumpers,
		Lore:             lore,
		NextRoundHook:    nextRoundHook,
		StandoutStamps:   standoutStamps,
		StandoutComments: standoutComments,
	}

	return RecapResult{Recap: recap, Allowed: true}, nil
}

// deriveStandouts computes the standout stamps (dominant stance per jump with
// stamps) and standout comments (comments on the top-stamped jump(s)) from
// the assembled jump entries and comments. All ordering is deterministic:
//
//   - standoutStamps: sorted by jump TotalStamps desc, then JumpID asc;
//     per-jump dominant stance picked by count desc, then stance asc.
//   - standoutComments: grouped under the top TotalStamps (tie → multiple
//     jumps); within the result, comments preserve their input order.
func deriveStandouts(jumps []RecapJumpEntry, comments []CommentSnapshot) ([]StandoutStampEntry, []CommentSnapshot) {
	totalByJump := make(map[string]int, len(jumps))
	for _, j := range jumps {
		totalByJump[j.JumpID] = j.TotalStamps
	}

	// standout stamps
	stamps := make([]StandoutStampEntry, 0, len(jumps))
	for _, j := range jumps {
		if j.TotalStamps <= 0 {
			continue
		}
		bestStance := ""
		bestCount := 0
		for stance, count := range j.StampCounts {
			if count > bestCount || (count == bestCount && stance < bestStance) || bestStance == "" {
				bestStance = stance
				bestCount = count
			}
		}
		if bestStance == "" {
			continue
		}
		stamps = append(stamps, StandoutStampEntry{
			JumpID: j.JumpID,
			Stance: bestStance,
			Count:  bestCount,
		})
	}
	sort.Slice(stamps, func(i, j int) bool {
		ti, tj := totalByJump[stamps[i].JumpID], totalByJump[stamps[j].JumpID]
		if ti != tj {
			return ti > tj
		}
		return stamps[i].JumpID < stamps[j].JumpID
	})

	// standout comments: comments on the jump(s) with the highest TotalStamps
	maxTotal := 0
	for _, j := range jumps {
		if j.TotalStamps > maxTotal {
			maxTotal = j.TotalStamps
		}
	}
	topJumpIDs := make(map[string]bool)
	for _, j := range jumps {
		if maxTotal > 0 && j.TotalStamps == maxTotal {
			topJumpIDs[j.JumpID] = true
		}
	}
	standoutComments := make([]CommentSnapshot, 0)
	for _, c := range comments {
		if c.JumpID != "" && topJumpIDs[c.JumpID] {
			standoutComments = append(standoutComments, c)
		}
	}

	return stamps, standoutComments
}
