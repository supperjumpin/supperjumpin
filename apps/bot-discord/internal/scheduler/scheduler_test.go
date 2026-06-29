package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestScheduler_FiresCallbackAtRevealBy(t *testing.T) {
	mock := clock.NewMock()
	revealBy := mock.Now().Add(1 * time.Hour)

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

	mock.Add(1*time.Hour + time.Second)

	mu.Lock()
	defer mu.Unlock()
	if got, want := len(firedIDs), 1; got != want {
		t.Fatalf("fired callbacks: got %d, want %d", got, want)
	}
	if firedIDs[0] != "round-1" {
		t.Errorf("fired roundID: got %q, want %q", firedIDs[0], "round-1")
	}
}

func TestScheduler_FiresImmediatelyWhenRevealByIsPast(t *testing.T) {
	mock := clock.NewMock()
	revealBy := mock.Now().Add(-1 * time.Hour)

	fired := make(chan string, 1)
	onFire := func(roundID string) error {
		fired <- roundID
		return nil
	}

	sched := New(Config{Clock: mock, OnFire: onFire})

	if err := sched.Schedule(t.Context(), "round-1", revealBy); err != nil {
		t.Fatalf("Schedule: %v", err)
	}

	select {
	case id := <-fired:
		if id != "round-1" {
			t.Errorf("fired roundID: got %q, want %q", id, "round-1")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("did not fire for past revealBy")
	}
}
