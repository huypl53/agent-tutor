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

	state, ok := e.rules[event]
	if !ok {
		e.mu.Unlock()
		return
	}

	state.count++

	if state.count < state.rule.Threshold {
		e.mu.Unlock()
		return
	}

	if !state.lastFire.IsZero() && time.Since(state.lastFire) < state.rule.Cooldown {
		e.mu.Unlock()
		return
	}

	state.count = 0
	state.lastFire = time.Now()
	cb := e.callback

	e.mu.Unlock()

	cb(event)
}
