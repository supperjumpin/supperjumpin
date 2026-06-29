package bot

import (
	"context"
	"net/http"
	"testing"
	"time"
)

type fakeRevealScheduler struct {
	scheduled map[string]time.Time
}

func (f *fakeRevealScheduler) Schedule(ctx context.Context, roundID string, revealBy time.Time) error {
	if f.scheduled == nil {
		f.scheduled = map[string]time.Time{}
	}
	f.scheduled[roundID] = revealBy
	return nil
}

func TestRoundStartHandler_SchedulesRevealFromResponse(t *testing.T) {
	h := newTestHarness(t)
	revealBy := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	h.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"round":{"id":"round-7","communityId":"community-1","promptId":"prompt-1","status":"open","revealBy":"2026-06-01T12:00:00Z","createdBy":"user-1","createdAt":"2026-06-01T11:00:00Z","prompt":{"id":"prompt-1","copy":"Test","theme":"x","costTier":"low"}}}`))
	})
	sched := &fakeRevealScheduler{}
	h.handler.SetRevealScheduler(sched)

	interaction := IncomingInteraction{
		Type:    InteractionApplicationCommand,
		GuildID: "guild-1",
		UserID:  "user-1",
		Command: CommandRoute{Name: "round", Subcommand: "start"},
		Options: map[string]string{
			"communityId":       "community-1",
			"revealTimeframeId": "tf-1",
		},
	}

	if _, err := h.handleInteraction(t, interaction); err != nil {
		t.Fatalf("handle: %v", err)
	}

	got, ok := sched.scheduled["round-7"]
	if !ok {
		t.Fatalf("sched.scheduled: missing round-7, got %v", sched.scheduled)
	}
	if !got.Equal(revealBy) {
		t.Errorf("scheduled time: got %v, want %v", got, revealBy)
	}
}
