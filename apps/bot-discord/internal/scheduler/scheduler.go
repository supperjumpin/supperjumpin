package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
)

type OnFire func(roundID string) error

type Config struct {
	Clock  clock.Clock
	OnFire OnFire
	State  StateStore
}

type StateStore interface {
	Read() (map[string]time.Time, error)
	Write(map[string]time.Time) error
	Remove(roundID string) error
}

type entry struct {
	timer    *clock.Timer
	revealBy time.Time
}

type Scheduler struct {
	clk    clock.Clock
	onFire OnFire
	state  StateStore

	mu              sync.Mutex
	active          map[string]entry
	watchdogRunning bool
	watchdogStop    chan struct{}
	watchdogReady   chan struct{}
}

func New(cfg Config) *Scheduler {
	return &Scheduler{
		clk:    cfg.Clock,
		onFire: cfg.OnFire,
		state:  cfg.State,
		active: map[string]entry{},
	}
}

func (s *Scheduler) Schedule(ctx context.Context, roundID string, revealBy time.Time) error {
	s.mu.Lock()
	if prev, ok := s.active[roundID]; ok {
		prev.timer.Stop()
	}
	s.mu.Unlock()

	delay := revealBy.Sub(s.clk.Now())
	if delay <= 0 {
		go s.fire(roundID)
		return nil
	}

	t := s.clk.AfterFunc(delay, func() {
		s.fire(roundID)
	})

	s.mu.Lock()
	s.active[roundID] = entry{timer: t, revealBy: revealBy}
	s.mu.Unlock()

	if s.state != nil {
		if err := s.persistLocked(); err != nil {
			return fmt.Errorf("scheduler: persist state: %w", err)
		}
	}
	return nil
}

func (s *Scheduler) Cancel(roundID string) {
	s.mu.Lock()
	if e, ok := s.active[roundID]; ok {
		e.timer.Stop()
		delete(s.active, roundID)
	}
	s.mu.Unlock()
	if s.state != nil {
		_ = s.state.Remove(roundID)
	}
}

func (s *Scheduler) fire(roundID string) {
	s.mu.Lock()
	delete(s.active, roundID)
	s.mu.Unlock()
	if s.state != nil {
		_ = s.state.Remove(roundID)
	}
	if s.onFire != nil {
		_ = s.onFire(roundID)
	}
}

func (s *Scheduler) persistLocked() error {
	s.mu.Lock()
	snapshot := make(map[string]time.Time, len(s.active))
	for id, e := range s.active {
		snapshot[id] = e.revealBy
	}
	s.mu.Unlock()
	return s.state.Write(snapshot)
}

func (s *Scheduler) Recover(ctx context.Context) error {
	if s.state == nil {
		return nil
	}
	stored, err := s.state.Read()
	if err != nil {
		return fmt.Errorf("scheduler: read state: %w", err)
	}
	for id, revealBy := range stored {
		if err := s.Schedule(ctx, id, revealBy); err != nil {
			return fmt.Errorf("scheduler: recover %q: %w", id, err)
		}
	}
	return nil
}

func (s *Scheduler) StartWatchdog(interval time.Duration) {
	s.mu.Lock()
	if s.watchdogRunning {
		s.mu.Unlock()
		return
	}
	s.watchdogRunning = true
	s.watchdogReady = make(chan struct{})
	ready := s.watchdogReady
	s.mu.Unlock()

	go s.watchdogLoop(interval, ready)
}

func (s *Scheduler) StopWatchdog() {
	s.mu.Lock()
	if !s.watchdogRunning {
		s.mu.Unlock()
		return
	}
	s.watchdogRunning = false
	if s.watchdogStop != nil {
		close(s.watchdogStop)
		s.watchdogStop = nil
	}
	s.mu.Unlock()
}

func (s *Scheduler) watchdogLoop(interval time.Duration, ready chan struct{}) {
	close(ready)
	for {
		s.mu.Lock()
		if !s.watchdogRunning {
			s.mu.Unlock()
			return
		}
		stopCh := make(chan struct{})
		s.watchdogStop = stopCh
		s.mu.Unlock()

		select {
		case <-s.clk.After(interval):
			s.checkMissed()
		case <-stopCh:
			return
		}
	}
}

func (s *Scheduler) checkMissed() {
	s.mu.Lock()
	now := s.clk.Now()
	missed := []string{}
	for id, e := range s.active {
		if !e.revealBy.After(now) {
			missed = append(missed, id)
		}
	}
	s.mu.Unlock()

	for _, id := range missed {
		s.fire(id)
	}
}
