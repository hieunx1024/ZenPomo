package app

import (
	"fmt"
	"strings"
	"time"
	"zenpomo/internal/analytics"
	"zenpomo/internal/daemon"
	"zenpomo/internal/theme"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const totalSettingsRows = 10

// Update handles user input, mouse clicks, window resizing, and background timer ticks.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		if m.client != nil {
			if resp, err := m.client.SendTUICommand(daemon.CmdGetStatus); err == nil {
				m.snapshot = resp.Snapshot
				m.todayStats = resp.Stats
				m.tasks = resp.Tasks
				switch resp.RequestedMode {
				case "config":
					m.activeTab = TabSettings
				case "timer":
					m.activeTab = TabTimer
				case "tasks":
					m.activeTab = TabTasks
				case "stats":
					m.activeTab = TabStats
				}
			} else if m.store != nil {
				m.todayStats = m.store.GetTodayStats()
				m.tasks = m.store.GetTasks()
				m.allStats = m.store.GetAllStats()
			}
		} else if m.store != nil {
			m.todayStats = m.store.GetTodayStats()
			m.tasks = m.store.GetTasks()
			m.allStats = m.store.GetAllStats()
		}
		return m, doTick()

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Click near top header (Y <= 3) to switch tabs based on X position
			if msg.Y <= 3 {
				centerX := m.width / 2
				offset := msg.X - (centerX - 35)
				switch {
				case offset >= 10 && offset < 22:
					m.activeTab = TabTimer
				case offset >= 22 && offset < 34:
					m.activeTab = TabTasks
				case offset >= 34 && offset < 50:
					m.activeTab = TabStats
				case offset >= 50 && offset < 66:
					m.activeTab = TabSettings
				}
				return m, nil
			}
		}

	case tea.KeyMsg:
		// 1. Adding Task Mode
		if m.inputMode == ModeAddingTask {
			switch msg.String() {
			case "esc":
				m.inputMode = ModeNormal
				m.textInput.Blur()
				return m, nil
			case "enter":
				val := strings.TrimSpace(m.textInput.Value())
				if val != "" && m.store != nil {
					task := m.store.AddTask(val, 1)
					m.tasks = m.store.GetTasks()
					m.todayStats = m.store.GetTodayStats()
					if m.snapshot.ActiveTaskTitle == "General Focus" || m.snapshot.ActiveTaskTitle == "" {
						_, _ = m.client.SendCommand(daemon.CmdSetTask, task.Title)
					}
				}
				m.inputMode = ModeNormal
				m.textInput.Blur()
				return m, nil
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		// 2. Editing Task Mode
		if m.inputMode == ModeEditingTask {
			switch msg.String() {
			case "esc":
				m.inputMode = ModeNormal
				m.textInput.Blur()
				return m, nil
			case "enter":
				val := strings.TrimSpace(m.textInput.Value())
				if val != "" && m.store != nil && m.editingTaskID != "" {
					m.store.EditTask(m.editingTaskID, val)
					m.tasks = m.store.GetTasks()
					m.todayStats = m.store.GetTodayStats()
				}
				m.inputMode = ModeNormal
				m.textInput.Blur()
				return m, nil
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		// 3. Global Navigation Keys
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "1":
			m.activeTab = TabTimer
			m.zenMode = false
			return m, nil

		case "2":
			m.activeTab = TabTasks
			m.zenMode = false
			return m, nil

		case "3":
			m.activeTab = TabStats
			m.zenMode = false
			return m, nil

		case "4":
			m.activeTab = TabSettings
			m.zenMode = false
			return m, nil

		case "tab", "]":
			m.activeTab = (m.activeTab + 1) % 4
			m.zenMode = false
			return m, nil

		case "shift+tab", "[":
			m.activeTab = (m.activeTab + 3) % 4
			m.zenMode = false
			return m, nil

		case "z":
			if m.activeTab == TabTimer {
				m.zenMode = !m.zenMode
			}
			return m, nil

		case "t":
			m.config.Theme = theme.NextTheme(m.config.Theme)
			m.theme = theme.GetTheme(m.config.Theme)
			m.saveConfig()
			return m, nil
		}

		// 4. Tab-Specific Handlers
		switch m.activeTab {
		case TabTimer:
			switch msg.String() {
			case " ":
				if m.client != nil {
					_, _ = m.client.SendCommand(daemon.CmdToggle)
				}
				return m, nil

			case "n":
				if m.client != nil {
					_, _ = m.client.SendCommand(daemon.CmdSkip)
				}
				return m, nil

			case "r":
				if m.client != nil {
					_, _ = m.client.SendCommand(daemon.CmdReset)
				}
				return m, nil

			case "m":
				if m.client != nil {
					_, _ = m.client.SendCommand(daemon.CmdToggleSound)
					if cfg, err := m.client.GetConfig(); err == nil {
						m.config = cfg
					}
				}
				return m, nil

			case "c":
				m.activeTab = TabSettings
				return m, nil

			case "a":
				m.activeTab = TabTasks
				m.inputMode = ModeAddingTask
				m.textInput.Reset()
				m.textInput.Placeholder = "Task title (e.g. Refactor API #backend est:3)"
				m.textInput.Focus()
				return m, textinput.Blink
			}

		case TabTasks:
			switch msg.String() {
			case "j", "down":
				if len(m.tasks) > 0 && m.selectedTask < len(m.tasks)-1 {
					m.selectedTask++
				}
				return m, nil

			case "k", "up":
				if m.selectedTask > 0 {
					m.selectedTask--
				}
				return m, nil

			case "J", "shift+down":
				// Reorder task down in queue
				if len(m.tasks) > 0 && m.selectedTask < len(m.tasks)-1 && m.store != nil {
					if m.store.ReorderTask(m.selectedTask, m.selectedTask+1) {
						m.tasks = m.store.GetTasks()
						m.selectedTask++
					}
				}
				return m, nil

			case "K", "shift+up":
				// Reorder task up in queue
				if m.selectedTask > 0 && m.store != nil {
					if m.store.ReorderTask(m.selectedTask, m.selectedTask-1) {
						m.tasks = m.store.GetTasks()
						m.selectedTask--
					}
				}
				return m, nil

			case " ", "enter":
				if len(m.tasks) > 0 && m.selectedTask < len(m.tasks) {
					task := m.tasks[m.selectedTask]
					if m.client != nil {
						_, _ = m.client.SendCommand(daemon.CmdSetTask, task.Title)
					}
				}
				return m, nil

			case "a":
				m.inputMode = ModeAddingTask
				m.textInput.Reset()
				m.textInput.Placeholder = "Task title (e.g. Fix DB index #perf est:2)"
				m.textInput.Focus()
				return m, textinput.Blink

			case "e":
				if len(m.tasks) > 0 && m.selectedTask < len(m.tasks) {
					task := m.tasks[m.selectedTask]
					m.editingTaskID = task.ID
					m.inputMode = ModeEditingTask

					// Pre-fill input value
					var prefilled strings.Builder
					prefilled.WriteString(task.Title)
					for _, tag := range task.Tags {
						prefilled.WriteString(" #" + tag)
					}
					if task.Target > 1 {
						prefilled.WriteString(fmt.Sprintf(" est:%d", task.Target))
					}

					m.textInput.SetValue(prefilled.String())
					m.textInput.Focus()
					return m, textinput.Blink
				}
				return m, nil

			case "d":
				if len(m.tasks) > 0 && m.selectedTask < len(m.tasks) {
					task := m.tasks[m.selectedTask]
					if m.store != nil {
						m.store.ToggleTask(task.ID)
						m.tasks = m.store.GetTasks()
						m.todayStats = m.store.GetTodayStats()
					}
				}
				return m, nil

			case "x":
				if len(m.tasks) > 0 && m.selectedTask < len(m.tasks) {
					task := m.tasks[m.selectedTask]
					if m.store != nil {
						m.store.DeleteTask(task.ID)
						m.tasks = m.store.GetTasks()
						if m.selectedTask >= len(m.tasks) && m.selectedTask > 0 {
							m.selectedTask--
						}
					}
				}
				return m, nil

			case "C":
				// Clear completed tasks
				if m.store != nil {
					m.store.ClearCompleted()
					m.tasks = m.store.GetTasks()
					if m.selectedTask >= len(m.tasks) && m.selectedTask > 0 {
						m.selectedTask = len(m.tasks) - 1
					}
				}
				return m, nil
			}

		case TabStats:
			switch msg.String() {
			case "r":
				if m.store != nil {
					m.allStats = m.store.GetAllStats()
					m.todayStats = m.store.GetTodayStats()
				}
				return m, nil

			case "x":
				summary := analytics.ComputeSummary(m.allStats, m.todayStats)
				exportErr := analytics.ExportMarkdown("zenpomo-stats.md", summary, m.tasks)
				if exportErr == nil {
					m.statusMessage = "Report exported to zenpomo-stats.md"
				} else {
					m.statusMessage = "Export failed: " + exportErr.Error()
				}
				m.statusExpiry = time.Now().Add(4 * time.Second)
				return m, nil
			}

		case TabSettings:
			switch msg.String() {
			case "j", "down":
				m.configCursor = (m.configCursor + 1) % totalSettingsRows
				return m, nil

			case "k", "up":
				m.configCursor = (m.configCursor + totalSettingsRows - 1) % totalSettingsRows
				return m, nil

			case "l", "right", "+", "=":
				m.adjustSettings(1)
				return m, nil

			case "h", "left", "-":
				m.adjustSettings(-1)
				return m, nil

			case " ", "enter":
				m.toggleSetting()
				return m, nil
			}
		}
	}

	return m, nil
}

func (m *Model) saveConfig() {
	if m.client != nil {
		_ = m.client.UpdateConfig(m.config)
	}
	if m.store != nil {
		m.store.UpdateConfig(m.config)
	}
}

func (m *Model) adjustSettings(delta int) {
	switch m.configCursor {
	case 0: // Theme
		m.config.Theme = theme.NextTheme(m.config.Theme)
		m.theme = theme.GetTheme(m.config.Theme)
	case 1: // Work Duration
		newMins := int(m.config.WorkDuration.Minutes()) + delta
		if newMins >= 1 && newMins <= 120 {
			m.config.WorkDuration = time.Duration(newMins) * time.Minute
		}
	case 2: // Short Break
		newMins := int(m.config.ShortBreakDuration.Minutes()) + delta
		if newMins >= 1 && newMins <= 60 {
			m.config.ShortBreakDuration = time.Duration(newMins) * time.Minute
		}
	case 3: // Long Break
		newMins := int(m.config.LongBreakDuration.Minutes()) + delta
		if newMins >= 1 && newMins <= 90 {
			m.config.LongBreakDuration = time.Duration(newMins) * time.Minute
		}
	case 4: // Long Break Interval
		newVal := m.config.LongBreakInterval + delta
		if newVal >= 1 && newVal <= 12 {
			m.config.LongBreakInterval = newVal
		}
	case 5: // AutoStartBreak
		m.config.AutoStartBreak = !m.config.AutoStartBreak
	case 6: // AutoStartWork
		m.config.AutoStartWork = !m.config.AutoStartWork
	case 7: // Sound Enabled
		m.config.SoundEnabled = !m.config.SoundEnabled
	case 8: // Notification Enabled
		m.config.NotificationEnable = !m.config.NotificationEnable
	case 9: // Ambient Sound
		ambientOptions := []string{"none", "rain", "whitenoise", "waves", "coffee"}
		curIdx := 0
		for i, opt := range ambientOptions {
			if opt == m.config.AmbientSound {
				curIdx = i
				break
			}
		}
		nextIdx := (curIdx + delta + len(ambientOptions)) % len(ambientOptions)
		m.config.AmbientSound = ambientOptions[nextIdx]
	}
	m.saveConfig()
}

func (m *Model) toggleSetting() {
	switch m.configCursor {
	case 0: // Cycle theme
		m.config.Theme = theme.NextTheme(m.config.Theme)
		m.theme = theme.GetTheme(m.config.Theme)
	case 5: // AutoStartBreak
		m.config.AutoStartBreak = !m.config.AutoStartBreak
	case 6: // AutoStartWork
		m.config.AutoStartWork = !m.config.AutoStartWork
	case 7: // Sound Enabled
		m.config.SoundEnabled = !m.config.SoundEnabled
	case 8: // Notification Enabled
		m.config.NotificationEnable = !m.config.NotificationEnable
	case 9: // Ambient Sound
		m.adjustSettings(1)
	default:
		m.adjustSettings(1)
	}
	m.saveConfig()
}
