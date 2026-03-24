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

func TestFireFromCallbackDoesNotDeadlock(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)

		var fired []string
		var engine *Engine
		engine = New(func(event string) {
			fired = append(fired, event)
			if event == "a" {
				engine.Fire("b")
			}
		})
		engine.AddRule(Rule{Event: "a", Threshold: 1, Cooldown: 0})
		engine.AddRule(Rule{Event: "b", Threshold: 1, Cooldown: 0})

		engine.Fire("a")

		if len(fired) != 2 || fired[0] != "a" || fired[1] != "b" {
			t.Errorf("expected [a b], got %v", fired)
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock: Fire from callback did not complete within 2s")
	}
}
