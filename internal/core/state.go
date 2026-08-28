package core

import "time"

// SessionType represents the type of current pomodoro session.
type SessionType string

const (
	SessionWork       SessionType = "Work"
	SessionShortBreak SessionType = "Short Break"
	SessionLongBreak  SessionType = "Long Break"
)

// TimerState represents the running state of the timer.
type TimerState string

const (
	StateStopped TimerState = "Stopped"
	StateRunning TimerState = "Running"
	StatePaused  TimerState = "Paused"
)

// Config holds the customizable duration and environment settings for pomodoro cycles.
type Config struct {
	WorkDuration       time.Duration `json:"work_duration"`
	ShortBreakDuration time.Duration `json:"short_break_duration"`
	LongBreakDuration  time.Duration `json:"long_break_duration"`
	LongBreakInterval  int           `json:"long_break_interval"` // default: 4 cycles
	AutoStartBreak     bool          `json:"auto_start_break"`
	AutoStartWork      bool          `json:"auto_start_work"`
	SoundEnabled       bool          `json:"sound_enabled"`
	NotificationEnable bool          `json:"notification_enabled"`
	Theme              string        `json:"theme"`         // gruvbox, catppuccin, tokyonight, nord, dracula, rosepine, monochrome
	AmbientSound       string        `json:"ambient_sound"` // none, rain, whitenoise, waves, coffee
	OnWorkStart        string        `json:"on_work_start"` // shell command executed when work session begins
	OnBreakStart       string        `json:"on_break_start"`// shell command executed when break begins
}

// DefaultConfig returns clean, standard Pomodoro settings.
func DefaultConfig() Config {
	return Config{
		WorkDuration:       25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
		LongBreakInterval:  4,
		AutoStartBreak:     true,
		AutoStartWork:      true,
		SoundEnabled:       true,
		NotificationEnable: true,
		Theme:              "gruvbox",
		AmbientSound:       "none",
		OnWorkStart:        "",
		OnBreakStart:       "",
	}
}

// SessionSnapshot contains the immutable current state of the timer for UI/IPC serialization.
type SessionSnapshot struct {
	State            TimerState  `json:"state"`
	Session          SessionType `json:"session"`
	RemainingSeconds int         `json:"remaining_seconds"`
	TotalSeconds     int         `json:"total_seconds"`
	CycleCount       int         `json:"cycle_count"`
	TargetCycles     int         `json:"target_cycles"`
	ActiveTaskTitle  string      `json:"active_task_title"`
	ProgressRatio    float64     `json:"progress_ratio"`
	SoundEnabled     bool        `json:"sound_enabled"`
}
