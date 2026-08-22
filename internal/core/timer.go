package core

import (
	"sync"
	"time"
)

// Timer represents the core Pomodoro state machine.
type Timer struct {
	mu            sync.RWMutex
	config        Config
	state         TimerState
	session       SessionType
	remaining     time.Duration
	totalDuration time.Duration
	cycleCount    int
	activeTask    string
	onComplete    func(prevSession SessionType, nextSession SessionType)
}

// NewTimer constructs a new Pomodoro timer with the given configuration.
func NewTimer(cfg Config) *Timer {
	if cfg.WorkDuration <= 0 {
		cfg = DefaultConfig()
	}
	t := &Timer{
		config:        cfg,
		state:         StateStopped,
		session:       SessionWork,
		remaining:     cfg.WorkDuration,
		totalDuration: cfg.WorkDuration,
		cycleCount:    1,
		activeTask:    "General Focus",
	}
	return t
}

// UpdateConfig dynamically updates duration and settings, adjusting current timer if not running.
func (t *Timer) UpdateConfig(cfg Config) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.config = cfg
	if t.state == StateStopped {
		t.remaining = t.getDurationFor(t.session)
		t.totalDuration = t.remaining
	}
}

// GetConfig returns the current configuration copy.
func (t *Timer) GetConfig() Config {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.config
}

// SetOnComplete registers a hook triggered when a session finishes.
func (t *Timer) SetOnComplete(fn func(prevSession SessionType, nextSession SessionType)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onComplete = fn
}

// Start begins or resumes countdown.
func (t *Timer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state = StateRunning
}

// Pause suspends the countdown without resetting remaining time.
func (t *Timer) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state == StateRunning {
		t.state = StatePaused
	}
}

// Toggle pauses if running, or starts if paused/stopped.
func (t *Timer) Toggle() TimerState {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state == StateRunning {
		t.state = StatePaused
	} else {
		t.state = StateRunning
	}
	return t.state
}

// Reset restores the current session to its initial duration and stops it.
func (t *Timer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state = StateStopped
	t.remaining = t.getDurationFor(t.session)
	t.totalDuration = t.remaining
}

// Skip advances immediately to the next session.
func (t *Timer) Skip() (SessionType, SessionType) {
	t.mu.Lock()
	defer t.mu.Unlock()
	prev := t.session
	t.advanceSessionLocked()
	return prev, t.session
}

// SwitchSession directly sets the session type, resets the countdown, and stops it.
func (t *Timer) SwitchSession(s SessionType) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.session = s
	t.state = StateStopped
	t.remaining = t.getDurationFor(s)
	t.totalDuration = t.remaining
}

// SetTask sets the current active task title.
func (t *Timer) SetTask(task string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if task == "" {
		task = "General Focus"
	}
	t.activeTask = task
}

// ToggleSound switches sound enabled on/off.
func (t *Timer) ToggleSound() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.config.SoundEnabled = !t.config.SoundEnabled
	return t.config.SoundEnabled
}

// Tick decrements time by 1 second if running.
func (t *Timer) Tick() (bool, SessionSnapshot) {
	t.mu.Lock()
	if t.state != StateRunning {
		snap := t.snapshotLocked()
		t.mu.Unlock()
		return false, snap
	}

	t.remaining -= time.Second
	if t.remaining <= 0 {
		prev := t.session
		t.advanceSessionLocked()
		next := t.session
		snap := t.snapshotLocked()
		hook := t.onComplete
		t.mu.Unlock()

		if hook != nil {
			hook(prev, next)
		}
		return true, snap
	}

	snap := t.snapshotLocked()
	t.mu.Unlock()
	return false, snap
}

// Snapshot returns a copy of the current timer status.
func (t *Timer) Snapshot() SessionSnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.snapshotLocked()
}

func (t *Timer) snapshotLocked() SessionSnapshot {
	remSec := int(t.remaining.Seconds())
	if remSec < 0 {
		remSec = 0
	}
	totSec := int(t.totalDuration.Seconds())
	if totSec <= 0 {
		totSec = 1
	}
	ratio := 1.0 - (float64(remSec) / float64(totSec))
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	return SessionSnapshot{
		State:            t.state,
		Session:          t.session,
		RemainingSeconds: remSec,
		TotalSeconds:     totSec,
		CycleCount:       t.cycleCount,
		TargetCycles:     t.config.LongBreakInterval,
		ActiveTaskTitle:  t.activeTask,
		ProgressRatio:    ratio,
		SoundEnabled:     t.config.SoundEnabled,
	}
}

func (t *Timer) advanceSessionLocked() {
	if t.session == SessionWork {
		if t.cycleCount >= t.config.LongBreakInterval {
			t.session = SessionLongBreak
			t.cycleCount = 1
		} else {
			t.session = SessionShortBreak
			t.cycleCount++
		}
		t.state = StateStopped
		if t.config.AutoStartBreak {
			t.state = StateRunning
		}
	} else {
		t.session = SessionWork
		t.state = StateStopped
		if t.config.AutoStartWork {
			t.state = StateRunning
		}
	}

	t.remaining = t.getDurationFor(t.session)
	t.totalDuration = t.remaining
}

func (t *Timer) getDurationFor(s SessionType) time.Duration {
	switch s {
	case SessionShortBreak:
		return t.config.ShortBreakDuration
	case SessionLongBreak:
		return t.config.LongBreakDuration
	case SessionWork:
		fallthrough
	default:
		return t.config.WorkDuration
	}
}
