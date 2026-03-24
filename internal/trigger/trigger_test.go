package trigger

import (
	"testing"
	"time"
)

func TestRuleFires(t *testing.T) {
	var fired []string
	callback := func(event string) {
		fired = append(fired, event)
	}

	engine := New(callback)
	engine.AddRule(Rule{
		Event:     "git.commit",
		Threshold: 1,
		Cooldown:  1 * time.Second,
	})

	engine.Fire("git.commit")
	if len(fired) != 1 {
		t.Errorf("expected 1 fire, got %d", len(fired))
	}
}

func TestRuleCooldown(t *testing.T) {
	var fired []string
	callback := func(event string) {
		fired = append(fired, event)
	}

	engine := New(callback)
	engine.AddRule(Rule{
		Event:     "git.commit",
		Threshold: 1,
		Cooldown:  5 * time.Minute,
	})

	engine.Fire("git.commit")
	engine.Fire("git.commit")
	if len(fired) != 1 {
		t.Errorf("expected 1 fire (cooldown), got %d", len(fired))
	}
}

func TestRuleThreshold(t *testing.T) {
	var fired []string
	callback := func(event string) {
		fired = append(fired, event)
	}

	engine := New(callback)
	engine.AddRule(Rule{
		Event:     "terminal.error_repeat",
		Threshold: 3,
		Cooldown:  1 * time.Second,
	})

	engine.Fire("terminal.error_repeat")
	engine.Fire("terminal.error_repeat")
	if len(fired) != 0 {
		t.Errorf("expected 0 fires before threshold, got %d", len(fired))
	}

	engine.Fire("terminal.error_repeat")
	if len(fired) != 1 {
		t.Errorf("expected 1 fire at threshold, got %d", len(fired))
	}
}
