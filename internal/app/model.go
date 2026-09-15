package app

import (
	"time"
	"zenpomo/internal/core"
	"zenpomo/internal/daemon"
	"zenpomo/internal/storage"
	"zenpomo/internal/theme"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Tab represents the active top-level screen in the TUI.
type Tab int

const (
	TabTimer Tab = iota
	TabTasks
	TabStats
	TabSettings
)

// StatsView selects the aggregation granularity shown on the Stats tab.
type StatsView int

const (
	StatsViewDaily StatsView = iota
	StatsViewWeekly
	StatsViewMonthly
)

// Next cycles to the next stats view (Daily -> Weekly -> Monthly -> Daily).
func (v StatsView) Next() StatsView {
	return (v + 1) % 3
}

// Label returns the human-readable name of the stats view.
func (v StatsView) Label() string {
	switch v {
	case StatsViewWeekly:
		return "Weekly"
	case StatsViewMonthly:
		return "Monthly"
	default:
		return "Daily"
	}
}

// InputMode indicates whether user is in normal navigation mode, adding a task, or editing a task.
type InputMode int

const (
	ModeNormal InputMode = iota
	ModeAddingTask
	ModeEditingTask
	ModeConfig // For CLI backwards-compatibility
)

// TabBounds records the screen X-coordinates of each tab header button for mouse clicking,
// relative to the start of the rendered header line (column 0 = first character of the header).
type TabBounds struct {
	StartX int
	EndX   int
}

// layoutOffsets records where the last rendered frame placed the main view on screen, so mouse
// events (given in absolute terminal coordinates) can be translated into frame-relative coordinates.
//
// Model is copied by value between View() and Update() calls (bubbletea semantics), but this
// field holds a pointer, so both copies share the same backing struct: writes made during View()
// are visible to Update() on the next input event.
type layoutOffsets struct {
	left int
	top  int
}

// Model holds the entire TUI application state.
type Model struct {
	client            *daemon.Client
	store             *storage.Store
	snapshot          core.SessionSnapshot
	todayStats        storage.DailyStats
	allStats          map[string]storage.DailyStats
	config            core.Config
	theme             theme.Palette
	activeTab         Tab
	zenMode           bool
	showHelp          bool
	tasks             []storage.Task
	selectedTask      int
	editingTaskID     string
	inputMode         InputMode
	textInput         textinput.Model
	configCursor      int
	tabBounds         *[4]TabBounds
	layout            *layoutOffsets
	width             int
	height            int
	statusMessage     string
	statusExpiry      time.Time
	statsView         StatsView
}

// TickMsg is delivered periodically to poll/sync timer state.
type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// NewModel creates and initializes the TUI model with standard tab layout, backed by the
// user's real config-dir store and background daemon.
func NewModel(initialMode ...InputMode) Model {
	store, _ := storage.NewStore()
	return newModelWithStore(store, initialMode...)
}

// newModelForTest builds a Model against an isolated, caller-provided store so tests never
// read from or write to the user's real ~/.config/zenpomo/data.json.
func newModelForTest(store *storage.Store, initialMode ...InputMode) Model {
	return newModelWithStore(store, initialMode...)
}

func newModelWithStore(store *storage.Store, initialMode ...InputMode) Model {
	client := daemon.NewClient()
	_ = client.EnsureDaemon()

	ti := textinput.New()
	ti.Placeholder = "Task title (e.g. Refactor API #backend)"
	ti.CharLimit = 64
	ti.Width = 48

	snap, _ := client.GetSnapshot()
	cfg := core.DefaultConfig()
	if store != nil {
		cfg = store.GetConfig()
	}
	if client != nil {
		if remoteCfg, err := client.GetConfig(); err == nil {
			cfg = remoteCfg
		}
	}

	currentTheme := theme.GetTheme(cfg.Theme)

	tasks := []storage.Task{}
	var stats storage.DailyStats
	var allStats map[string]storage.DailyStats
	if store != nil {
		tasks = store.GetTasks()
		stats = store.GetTodayStats()
		allStats = store.GetAllStats()
	}

	mode := ModeNormal
	activeTab := TabTimer
	if len(initialMode) > 0 {
		mode = initialMode[0]
		if mode == ModeConfig {
			activeTab = TabSettings
			mode = ModeNormal
		}
	}

	return Model{
		client:        client,
		store:         store,
		snapshot:      snap,
		todayStats:    stats,
		allStats:      allStats,
		config:        cfg,
		theme:         currentTheme,
		activeTab:     activeTab,
		zenMode:       false,
		showHelp:      false,
		tasks:         tasks,
		selectedTask:  0,
		inputMode:     mode,
		textInput:     ti,
		configCursor:  0,
		tabBounds:     &[4]TabBounds{},
		layout:        &layoutOffsets{},
		width:         80,
		height:        24,
	}
}

// Init initializes the tea program loop.
func (m Model) Init() tea.Cmd {
	return doTick()
}
