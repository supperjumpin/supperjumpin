package bot

import (
	"context"
	"sync"
)

type RoundInfo struct {
	RoundID    string
	ChannelID  string
	PromptCopy string
}

type RoundRegistry struct {
	mu      sync.RWMutex
	byRound map[string]RoundInfo
}

func NewRoundRegistry() *RoundRegistry {
	return &RoundRegistry{byRound: map[string]RoundInfo{}}
}

func (r *RoundRegistry) Remember(info RoundInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byRound[info.RoundID] = info
}

func (r *RoundRegistry) Get(roundID string) (RoundInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.byRound[roundID]
	return info, ok
}

func (r *RoundRegistry) Forget(roundID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byRound, roundID)
}

type RecapMessage struct {
	RoundID    string
	PromptCopy string
	Jumps      []RecapJump
}

type RecapJump struct {
	ID          string
	Caption     string
	TotalStamps int
	StampCounts map[string]int
}

type RecapPoster interface {
	PostReveal(ctx context.Context, channelID string, recap RecapMessage) error
}
