package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

type fakeState struct {
	mu      sync.Mutex
	stored  map[string]time.Time
	readErr error
}

func (f *fakeState) Read() (map[string]time.Time, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.readErr != nil {
		return nil, f.readErr
	}
	out := make(map[string]time.Time, len(f.stored))
	for k, v := range f.stored {
		out[k] = v
	}
	return out, nil
}

func (f *fakeState) Write(m map[string]time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.stored == nil {
		f.stored = map[string]time.Time{}
	}
	for k, v := range m {
		f.stored[k] = v
	}
	return nil
}

func (f *fakeState) Remove(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.stored, id)
	return nil
}

func TestScheduler_RecoverReArmsFutureReveals(t *testing.T) {
	mock := clock.NewMock()
	state := &fakeState{}
	state.stored = map[string]time.Time{
		"round-1": mock.Now().Add(1 * time.Hour),
	}

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

	sched := New(Config{Clock: mock, OnFire: onFire, State: state})
	if err := sched.Recover(t.Context()); err != nil {
		t.Fatalf("Recover: %v", err)
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

func TestScheduler_RecoverFiresLateForPastReveals(t *testing.T) {
	mock := clock.NewMock()
	state := &fakeState{}
	state.stored = map[string]time.Time{
		"round-stale": mock.Now().Add(-1 * time.Hour),
	}

	fired := make(chan string, 1)
	onFire := func(roundID string) error {
		fired <- roundID
		return nil
	}

	sched := New(Config{Clock: mock, OnFire: onFire, State: state})
	if err := sched.Recover(t.Context()); err != nil {
		t.Fatalf("Recover: %v", err)
	}

	select {
	case id := <-fired:
		if id != "round-stale" {
			t.Errorf("fired roundID: got %q, want %q", id, "round-stale")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("did not fire for past revealBy on recovery")
	}
}

func TestScheduler_SchedulePersistsToState(t *testing.T) {
	mock := clock.NewMock()
	revealBy := mock.Now().Add(1 * time.Hour)
	state := &fakeState{}

	sched := New(Config{Clock: mock, State: state})
	if err := sched.Schedule(t.Context(), "round-1", revealBy); err != nil {
		t.Fatalf("Schedule: %v", err)
	}

	state.mu.Lock()
	defer state.mu.Unlock()
	if len(state.stored) != 1 {
		t.Fatalf("state.stored: got %d entries, want 1", len(state.stored))
	}
	got, ok := state.stored["round-1"]
	if !ok {
		t.Fatal("state.stored: missing round-1")
	}
	if !got.Equal(revealBy) {
		t.Errorf("state.stored[round-1]: got %v, want %v", got, revealBy)
	}
}

func TestScheduler_FireRemovesFromState(t *testing.T) {
	mock := clock.NewMock()
	revealBy := mock.Now().Add(1 * time.Hour)
	state := &fakeState{}

	var fired bool
	onFire := func(roundID string) error {
		fired = true
		return nil
	}

	sched := New(Config{Clock: mock, OnFire: onFire, State: state})
	if err := sched.Schedule(t.Context(), "round-1", revealBy); err != nil {
		t.Fatalf("Schedule: %v", err)
	}

	mock.Add(1*time.Hour + time.Second)

	if !fired {
		t.Fatal("onFire: not called")
	}

	state.mu.Lock()
	defer state.mu.Unlock()
	if _, ok := state.stored["round-1"]; ok {
		t.Errorf("state.stored[round-1]: still present after fire, want removed")
	}
}
