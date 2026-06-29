package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestScheduler_WatchdogFiresMissedTimers(t *testing.T) {
	mock := clock.NewMock()
	revealBy := mock.Now().Add(30 * time.Minute)

	var (
		mu       sync.Mutex
		firedIDs []string
	)
	onFire := func(roundID string) error {
		mu.Lock()
		defer mu.Unlock()
		firedIDs = append(firedIDs, roundID)
		return nil
	}

	sched := New(Config{Clock: mock, OnFire: onFire})
	if err := sched.Schedule(t.Context(), "round-1", revealBy); err != nil {
		t.Fatalf("Schedule: %v", err)
	}

	sched.mu.Lock()
	entry := sched.active["round-1"]
	entry.timer.Stop()
	sched.active["round-1"] = entry
	sched.mu.Unlock()

	sched.StartWatchdog(1 * time.Minute)
	defer sched.StopWatchdog()
	<-sched.watchdogReady

	mock.Add(31 * time.Minute)

	mu.Lock()
	defer mu.Unlock()
	if got, want := len(firedIDs), 1; got != want {
		t.Fatalf("watchdog-fired callbacks: got %d, want %d", got, want)
	}
	if firedIDs[0] != "round-1" {
		t.Errorf("fired roundID: got %q, want %q", firedIDs[0], "round-1")
	}
}
