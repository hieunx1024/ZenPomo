package app

import (
	"time"
	"zenpomo/internal/daemon"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles user input, window resizing, and background timer ticks.
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
				if resp.RequestedMode == "config" {
					if cfg, err := m.client.GetConfig(); err == nil {
						m.config = cfg
					}
					m.inputMode = ModeConfig
				}
			} else if m.store != nil {
				m.todayStats = m.store.GetTodayStats()
				m.tasks = m.store.GetTasks()
			}
		} else if m.store != nil {
			m.todayStats = m.store.GetTodayStats()
			m.tasks = m.store.GetTasks()
		}
		return m, doTick()

	case tea.KeyMsg:
		// 1. Adding Task Mode
		if m.inputMode == ModeAddingTask {
			switch msg.String() {
			case "esc":
				m.inputMode = ModeNormal
				m.textInput.Blur()
				return m, nil
			case "enter":
				val := m.textInput.Value()
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

		// 2. In-App Configuration Mode
		if m.inputMode == ModeConfig {
			switch msg.String() {
			case "esc", "enter", "c":
				// Save & Exit Config Mode
				if m.client != nil {
					_ = m.client.UpdateConfig(m.config)
				}
				if m.store != nil {
					m.store.UpdateConfig(m.config)
				}
				m.inputMode = ModeNormal
				return m, nil

			case "j", "down", "tab":
				m.configCursor = (m.configCursor + 1) % 8
				return m, nil

			case "k", "up", "shift+tab":
				m.configCursor = (m.configCursor + 7) % 8
				return m, nil

			case "l", "right", "+", "=":
				m.adjustConfig(1)
				return m, nil

			case "h", "left", "-":
				m.adjustConfig(-1)
				return m, nil

			case " ":
				switch m.configCursor {
				case 4:
					m.config.AutoStartBreak = !m.config.AutoStartBreak
				case 5:
					m.config.AutoStartWork = !m.config.AutoStartWork
				case 6:
					m.config.SoundEnabled = !m.config.SoundEnabled
				case 7:
					m.config.NotificationEnable = !m.config.NotificationEnable
				default:
					m.adjustConfig(1)
				}
				return m, nil
			}
			return m, nil
		}

		// 3. Normal Mode Navigation
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

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
			if m.client != nil {
				if cfg, err := m.client.GetConfig(); err == nil {
					m.config = cfg
				}
			}
			m.inputMode = ModeConfig
			return m, nil

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

		case "enter":
			if len(m.tasks) > 0 && m.selectedTask < len(m.tasks) {
				task := m.tasks[m.selectedTask]
				if m.client != nil {
					_, _ = m.client.SendCommand(daemon.CmdSetTask, task.Title)
				}
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

		case "a":
			m.inputMode = ModeAddingTask
			m.textInput.Reset()
			m.textInput.Focus()
			return m, textinput.Blink
		}
	}

	return m, nil
}

func (m *Model) adjustConfig(delta int) {
	switch m.configCursor {
	case 0: // Work Duration
		newMins := int(m.config.WorkDuration.Minutes()) + delta
		if newMins >= 1 && newMins <= 120 {
			m.config.WorkDuration = time.Duration(newMins) * time.Minute
		}
	case 1: // Short Break
		newMins := int(m.config.ShortBreakDuration.Minutes()) + delta
		if newMins >= 1 && newMins <= 60 {
			m.config.ShortBreakDuration = time.Duration(newMins) * time.Minute
		}
	case 2: // Long Break
		newMins := int(m.config.LongBreakDuration.Minutes()) + delta
		if newMins >= 1 && newMins <= 90 {
			m.config.LongBreakDuration = time.Duration(newMins) * time.Minute
		}
	case 3: // Long Break Interval
		newVal := m.config.LongBreakInterval + delta
		if newVal >= 1 && newVal <= 12 {
			m.config.LongBreakInterval = newVal
		}
	case 4: // AutoStartBreak
		m.config.AutoStartBreak = !m.config.AutoStartBreak
	case 5: // AutoStartWork
		m.config.AutoStartWork = !m.config.AutoStartWork
	case 6: // Sound
		m.config.SoundEnabled = !m.config.SoundEnabled
	case 7: // Notifications
		m.config.NotificationEnable = !m.config.NotificationEnable
	}
}
