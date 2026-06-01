package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

// --- Package-level helpers ---

func isOpenSeasonStatus(status string) bool {
	return status == "Active" || status == "Judging Grace Period"
}

func randomToken(kind string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate %s: %w", kind, err)
	}
	return kind + "_" + hex.EncodeToString(bytes), nil
}

const httpMethodPut = "PUT"

func standingsFromGame(entries []game.StandingEntry) []StandingEntry {
	result := make([]StandingEntry, len(entries))
	for i, e := range entries {
		result[i] = StandingEntry{
			Player:      Player{ID: e.PlayerID, DisplayName: e.DisplayName},
			SeasonScore: e.SeasonScore,
			JudgedJumps: e.JudgedJumps,
		}
	}
	return result
}

func groupHome(group Group, membership GroupMembership, activeSeason *Season, recentJumps []PerformedJumpView, standings []StandingEntry) GroupHomeResponse {
	return GroupHomeResponse{
		Group:        group,
		Membership:   membership,
		ActiveSeason: activeSeason,
		RecentJumps:  recentJumps,
		Standings:    standings,
	}
}

func jumpFromGame(snap game.JumpSnapshot) Jump {
	return Jump{
		ID:                   snap.ID,
		GroupID:              snap.GroupID,
		PlayerID:             snap.PlayerID,
		Status:               snap.Status,
		SeasonID:             snap.SeasonID,
		Source:               snap.Source,
		Destination:          snap.Destination,
		Food:                 snap.Food,
		OffSeason:            snap.SeasonID == nil,
		GracePeriodExpiresAt: snap.GracePeriodExpiresAt,
	}
}

func jumpToSnapshot(jump Jump) game.JumpSnapshot {
	return game.JumpSnapshot{
		ID:                   jump.ID,
		GroupID:              jump.GroupID,
		PlayerID:             jump.PlayerID,
		Status:               jump.Status,
		SeasonID:             jump.SeasonID,
		Source:               jump.Source,
		Destination:          jump.Destination,
		Food:                 jump.Food,
		FinalScore:           jump.FinalScore,
		GracePeriodExpiresAt: jump.GracePeriodExpiresAt,
	}
}

func seasonToSnapshot(season Season) game.SeasonSnapshot {
	return game.SeasonSnapshot{
		ID:                   season.ID,
		GroupID:              season.GroupID,
		CommissionerPlayerID: season.CommissionerPlayerID,
		Status:               season.Status,
		SubmissionDeadline:   season.SubmissionDeadline,
		JudgingDeadline:      season.JudgingDeadline,
	}
}

func disputeToSnapshot(d Dispute) game.DisputeSnapshot {
	return game.DisputeSnapshot{
		ID:                 d.ID,
		JumpID:             d.JumpID,
		RaisedByPlayerID:   d.RaisedByPlayerID,
		Concern:            d.Concern,
		Details:            d.Details,
		Status:             d.Status,
		Resolution:         d.Resolution,
		ResolutionReason:   d.ResolutionReason,
		ResolvedByPlayerID: d.ResolvedByPlayerID,
		OverrideResolution: d.OverrideResolution,
		OverrideReason:     d.OverrideReason,
		OverrideByPlayerID: d.OverrideByPlayerID,
	}
}

func disputeFromSnapshot(snap game.DisputeSnapshot) Dispute {
	return Dispute{
		ID:                 snap.ID,
		JumpID:             snap.JumpID,
		RaisedByPlayerID:   snap.RaisedByPlayerID,
		Concern:            snap.Concern,
		Details:            snap.Details,
		Status:             snap.Status,
		Resolution:         snap.Resolution,
		ResolutionReason:   snap.ResolutionReason,
		ResolvedByPlayerID: snap.ResolvedByPlayerID,
		OverrideResolution: snap.OverrideResolution,
		OverrideReason:     snap.OverrideReason,
		OverrideByPlayerID: snap.OverrideByPlayerID,
	}
}

func visiblePerformedStatus(status string) bool {
	return status == "Performed Jump" || status == "Judged Jump" || status == "Unjudged Jump" || status == "Disqualified Jump"
}

func stableID(kind string, value string) string {
	sum := sha256.Sum256([]byte(kind + ":" + value))
	return kind + "_" + hex.EncodeToString(sum[:])[:12]
}

func displayName(email string) string {
	name, _, ok := strings.Cut(email, "@")
	if !ok || name == "" {
		return "player"
	}
	return name
}
