package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

var ErrInvalidFeedCursor = errors.New("invalid feed cursor")

type publicJumpDetailRead struct {
	Detail    *JumpDetail
	Tombstone *JumpTombstone
}

func loadPublicFeed(ctx context.Context, publicRead PublicReadFlow, judgment JudgmentFlow, viewer *MeResponse, cursorStr string, limit int, now time.Time) (PublicFeedResponse, error) {
	var cursorTS *time.Time
	var cursorID string
	if cursorStr != "" {
		ts, id, err := decodeCursor(cursorStr)
		if err != nil {
			return PublicFeedResponse{}, ErrInvalidFeedCursor
		}
		cursorTS = &ts
		cursorID = id
	}

	// Fetch limit+1 to detect whether there's a next page.
	cards, err := publicRead.FeedJumps(ctx, cursorTS, cursorID, limit+1)
	if err != nil {
		return PublicFeedResponse{}, err
	}

	if viewer != nil {
		cardIDs := make([]string, len(cards))
		for i, c := range cards {
			cardIDs[i] = c.ID
		}
		judged, err := judgment.HasJudgedJumps(ctx, viewer.Player.ID, cardIDs)
		if err != nil {
			return PublicFeedResponse{}, err
		}
		for i := range cards {
			hint := game.JudgmentEligibility(game.JumpSnapshot{
				ID:                   cards[i].ID,
				PlayerID:             cards[i].PerformerID,
				GracePeriodExpiresAt: cards[i].GracePeriodExpiresAt,
			}, viewer.Player.ID, judged[cards[i].ID], now)
			cards[i].ViewerContext = viewerContextFromHint(hint)
		}
	}

	var nextCursor *string
	if len(cards) > limit {
		cards = cards[:limit]
		last := cards[len(cards)-1]
		c := encodeCursor(last.CreatedAt, last.ID)
		nextCursor = &c
	}

	return PublicFeedResponse{Jumps: cards, NextCursor: nextCursor}, nil
}

func loadPublicJumpDetail(ctx context.Context, publicRead PublicReadFlow, judgment JudgmentFlow, viewer *MeResponse, jumpID string, now time.Time) (publicJumpDetailRead, bool, error) {
	detail, found, err := publicRead.JumpDetail(ctx, jumpID)
	if err != nil {
		return publicJumpDetailRead{}, false, err
	}
	if !found {
		return publicJumpDetailRead{}, false, nil
	}

	if detail.Status == "Removed Jump" {
		removedAt := detail.CreatedAt
		if detail.RemovedAt != nil {
			removedAt = *detail.RemovedAt
		}
		return publicJumpDetailRead{Tombstone: &JumpTombstone{
			ID:        detail.ID,
			Status:    "Removed Jump",
			Message:   "This Jump is no longer available",
			RemovedAt: removedAt.Format(time.RFC3339),
		}}, true, nil
	}

	if viewer != nil {
		hasJudged, err := judgment.HasJudgedJump(ctx, detail.ID, viewer.Player.ID)
		if err != nil {
			return publicJumpDetailRead{}, false, err
		}
		hint := game.JudgmentEligibility(game.JumpSnapshot{
			ID:                   detail.ID,
			PlayerID:             detail.PerformerID,
			GracePeriodExpiresAt: detail.GracePeriodExpiresAt,
		}, viewer.Player.ID, hasJudged, now)
		detail.ViewerContext = viewerContextFromHint(hint)
	} else {
		detail.ViewerContext = &ViewerContext{CanJudge: true}
	}

	return publicJumpDetailRead{Detail: &detail}, true, nil
}

// viewerContextFromHint converts a game.EligibilityHint into a transport-layer
// ViewerContext DTO.
func viewerContextFromHint(hint game.EligibilityHint) *ViewerContext {
	vc := &ViewerContext{CanJudge: hint.CanJudge}
	if hint.Reason != "" {
		r := hint.Reason
		vc.Reason = &r
	}
	if hint.GracePeriodEndsAt != nil {
		vc.GracePeriodEndsAt = hint.GracePeriodEndsAt
	}
	if !hint.CanJudge && hint.Reason == "already-judged" {
		vc.HasJudged = true
	}
	return vc
}
