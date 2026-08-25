package app

import (
	"time"
	"zenpomo/internal/core"
	"zenpomo/internal/daemon"
	"zenpomo/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// InputMode indicates whether user is in normal navigation mode, typing a task, or configuring settings.
type InputMode int

const (
	ModeNormal InputMode = iota
	ModeAddingTask
	ModeConfig
)

// Model holds the entire TUI application state.
type Model struct {
	client       *daemon.Client
	store        *storage.Store
	snapshot     core.SessionSnapshot
	todayStats   storage.DailyStats
	config       core.Config
	configCursor int
	tasks        []storage.Task
	selectedTask int
	inputMode    InputMode
	textInput    textinput.Model
	width        int
	height       int
	err          error
}

// TickMsg is delivered every 250ms to poll/sync timer state.
type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// NewModel creates and initializes the TUI model.
func NewModel(initialMode ...InputMode) Model {
	store, _ := storage.NewStore()
	client := daemon.NewClient()
	_ = client.EnsureDaemon()

	ti := textinput.New()
	ti.Placeholder = "Enter task title..."
	ti.CharLimit = 50
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

	tasks := []storage.Task{}
	var stats storage.DailyStats
	if store != nil {
		tasks = store.GetTasks()
		stats = store.GetTodayStats()
	}

	mode := ModeNormal
	if len(initialMode) > 0 {
		mode = initialMode[0]
	}

	return Model{
		client:       client,
		store:        store,
		snapshot:     snap,
		todayStats:   stats,
		config:       cfg,
		configCursor: 0,
		tasks:        tasks,
		selectedTask: 0,
		inputMode:    mode,
		textInput:    ti,
		width:        80,
		height:       24,
	}
}

// Init initializes the tea program loop.
func (m Model) Init() tea.Cmd {
	return doTick()
}
