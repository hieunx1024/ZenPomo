package core

import (
	"testing"
	"time"
)

func TestTimer_StateTransitions(t *testing.T) {
	cfg := Config{
		WorkDuration:       3 * time.Second,
		ShortBreakDuration: 2 * time.Second,
		LongBreakDuration:  4 * time.Second,
		LongBreakInterval:  2,
		AutoStartBreak:     true,
		AutoStartWork:      true,
		SoundEnabled:       true,
	}

	timer := NewTimer(cfg)
	snap := timer.Snapshot()
	if snap.State != StateStopped || snap.Session != SessionWork {
		t.Fatalf("expected initial state Stopped/Work, got %v/%v", snap.State, snap.Session)
	}

	timer.Start()
	snap = timer.Snapshot()
	if snap.State != StateRunning {
		t.Fatalf("expected state Running, got %v", snap.State)
	}

	// Tick 1
	completed, snap := timer.Tick()
	if completed || snap.RemainingSeconds != 2 {
		t.Fatalf("expected 2s remaining, got %v (completed=%v)", snap.RemainingSeconds, completed)
	}

	// Tick 2
	completed, snap = timer.Tick()
	if completed || snap.RemainingSeconds != 1 {
		t.Fatalf("expected 1s remaining, got %v", snap.RemainingSeconds)
	}

	// Tick 3: should complete Work session and auto-advance to Short Break
	completed, snap = timer.Tick()
	if !completed {
		t.Fatalf("expected session completion on 3rd tick")
	}
	if snap.Session != SessionShortBreak {
		t.Fatalf("expected Short Break, got %v", snap.Session)
	}
	if snap.RemainingSeconds != 2 {
		t.Fatalf("expected 2s remaining for Short Break, got %v", snap.RemainingSeconds)
	}
}

func TestTimer_ToggleAndSkip(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)

	st := timer.Toggle()
	if st != StateRunning {
		t.Fatalf("expected Running, got %v", st)
	}

	st = timer.Toggle()
	if st != StatePaused {
		t.Fatalf("expected Paused, got %v", st)
	}

	prev, next := timer.Skip()
	if prev != SessionWork || next != SessionShortBreak {
		t.Fatalf("expected Skip from Work to Short Break, got %v -> %v", prev, next)
	}
}
