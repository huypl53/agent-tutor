package trigger

import (
	"sync"
	"time"
)

type Rule struct {
	Event     string
	Threshold int
	Cooldown  time.Duration
}

type ruleState struct {
	rule     Rule
	count    int
	lastFire time.Time
}

type Engine struct {
	mu       sync.Mutex
	rules    map[string]*ruleState
	callback func(event string)
}

func New(callback func(event string)) *Engine {
	return &Engine{
		rules:    make(map[string]*ruleState),
		callback: callback,
	}
}

func (e *Engine) AddRule(r Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules[r.Event] = &ruleState{rule: r}
}

func (e *Engine) Fire(event string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	state, ok := e.rules[event]
	if !ok {
		return
	}

	state.count++

	if state.count < state.rule.Threshold {
		return
	}

	if !state.lastFire.IsZero() && time.Since(state.lastFire) < state.rule.Cooldown {
		return
	}

	state.count = 0
	state.lastFire = time.Now()

	// Call callback synchronously so test assertions work immediately.
	// The caller can wrap in a goroutine if needed.
	e.callback(event)
}
