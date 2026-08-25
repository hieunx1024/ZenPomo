package app

import (
	"fmt"
	"strings"
	"zenpomo/internal/clock"
	"zenpomo/internal/core"

	"github.com/charmbracelet/lipgloss"
)

const (
	// Minimum terminal dimensions required to render all components comfortably.
	MinTerminalWidth  = 48
	MinTerminalHeight = 14
)

var (
	// Tactile Unix Color Palette (Gruvbox / Nord minimal theme - Zero AI Neon)
	colorBorder     = lipgloss.Color("#504945") // Charcoal border
	colorBorderHigh = lipgloss.Color("#A89984") // Highlighted border
	colorTextLight  = lipgloss.Color("#EBDBB2") // Warm cream text
	colorTextDim    = lipgloss.Color("#928374") // Muted gray text
	colorWork       = lipgloss.Color("#FB4934") // Terracotta Red for Work
	colorBreak      = lipgloss.Color("#8EC07C") // Aqua Green for Break
	colorYellow     = lipgloss.Color("#FABD2F") // Warm Amber
	colorActiveTask = lipgloss.Color("#83A598") // Slate Blue

	baseStyle = lipgloss.NewStyle().
			Foreground(colorTextLight)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTextLight).
			Background(colorBorder).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTextLight)

	taskActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorActiveTask)

	taskDoneStyle = lipgloss.NewStyle().
			Foreground(colorTextDim)

	helpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorYellow)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorTextDim)

	configSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorYellow)
)

// makeBox creates a rounded border box whose TOTAL outer width matches totalWidth exactly.
func makeBox(totalWidth, totalHeight int, align lipgloss.Position, content string) string {
	if totalWidth < 10 {
		totalWidth = 10
	}
	innerWidth := totalWidth - 4
	if innerWidth < 1 {
		innerWidth = 1
	}

	st := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Width(innerWidth).
		Align(align)

	if totalHeight > 2 {
		innerHeight := totalHeight - 2
		if innerHeight < 1 {
			innerHeight = 1
		}
		st = st.Height(innerHeight)
	}

	return st.Render(content)
}

// View renders the TUI fitting 100% within the terminal dimensions both horizontally and vertically.
func (m Model) View() string {
	termW := m.width
	termH := m.height

	if termW <= 0 {
		termW = 80
	}
	if termH <= 0 {
		termH = 24
	}

	// 1. Check Minimum Viable Dimensions
	if termW < MinTerminalWidth || termH < MinTerminalHeight {
		return m.renderTooSmallView(termW, termH)
	}

	// 2. If in Configuration Mode, render the modal
	if m.inputMode == ModeConfig {
		return m.renderConfigModal(termW, termH)
	}

	var sessionColor lipgloss.Color
	if m.snapshot.Session == core.SessionWork {
		sessionColor = colorWork
	} else {
		sessionColor = colorBreak
	}

	// Calculate responsive content width (scales comfortably with terminal width, up to 104 on wide/fullscreen displays)
	contentWidth := termW - 4
	if contentWidth > 104 {
		contentWidth = 104
	}
	if contentWidth < MinTerminalWidth {
		contentWidth = MinTerminalWidth
	}

	// Header (Compact 1 line)
	stateText := string(m.snapshot.State)
	if m.snapshot.State == core.StateRunning {
		stateText = "RUNNING"
	}
	stateBadge := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1D2021")).
		Background(sessionColor).
		Padding(0, 1).
		Render(fmt.Sprintf("%s • %s", strings.ToUpper(string(m.snapshot.Session)), stateText))

	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		titleStyle.Render("ZENPOMO"),
		" ",
		stateBadge,
	)

	// Progress Bar
	barWidth := contentWidth - 24
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 56 {
		barWidth = 56
	}

	completedWidth := int(m.snapshot.ProgressRatio * float64(barWidth))
	if completedWidth > barWidth {
		completedWidth = barWidth
	}
	remainingWidth := barWidth - completedWidth

	filledBar := lipgloss.NewStyle().Foreground(sessionColor).Render(strings.Repeat("=", completedWidth))
	emptyBar := lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("-", remainingWidth))
	progressBar := fmt.Sprintf("[%s%s] %3.0f%%", filledBar, emptyBar, m.snapshot.ProgressRatio*100)

	// Cycle info
	var cycleBadges strings.Builder
	for i := 1; i <= m.snapshot.TargetCycles; i++ {
		if i <= m.snapshot.CycleCount {
			cycleBadges.WriteString("● ")
		} else {
			cycleBadges.WriteString("○ ")
		}
	}
	cycleInfo := fmt.Sprintf("Cycle: [ %s] (%d/%d)", cycleBadges.String(), m.snapshot.CycleCount, m.snapshot.TargetCycles)

	// Clock Section (Big ASCII if termH >= 20, else Compact)
	var clockBoxContent string
	if termH >= 20 && contentWidth >= 48 {
		clockLines := clock.RenderTime(m.snapshot.RemainingSeconds)
		clockRendered := lipgloss.NewStyle().
			Foreground(sessionColor).
			Bold(true).
			Render(strings.Join(clockLines, "\n"))

		clockBoxContent = lipgloss.JoinVertical(
			lipgloss.Center,
			clockRendered,
			progressBar,
			lipgloss.NewStyle().Foreground(colorTextDim).Render(cycleInfo),
		)
	} else {
		mins := m.snapshot.RemainingSeconds / 60
		secs := m.snapshot.RemainingSeconds % 60
		compactClock := lipgloss.NewStyle().
			Foreground(sessionColor).
			Bold(true).
			Render(fmt.Sprintf("[ %02d:%02d ]", mins, secs))

		clockBoxContent = lipgloss.JoinVertical(
			lipgloss.Center,
			compactClock,
			progressBar+"  "+lipgloss.NewStyle().Foreground(colorTextDim).Render(cycleInfo),
		)
	}
	clockBox := makeBox(contentWidth, 0, lipgloss.Center, clockBoxContent)

	// Lower Section: Responsive Task & Stats Layout
	var lowerSection string

	if contentWidth < 66 {
		// Single Column Stacked for narrower screens
		taskHeight := termH - 14
		if taskHeight < 4 {
			taskHeight = 4
		}
		if taskHeight > 7 {
			taskHeight = 7
		}
		taskBox := m.renderTaskBox(contentWidth, taskHeight)
		statsBox := m.renderStatsBox(contentWidth, 4)
		lowerSection = lipgloss.JoinVertical(lipgloss.Left, taskBox, statsBox)
	} else {
		// Two Columns Side-by-Side
		taskWidth := int(float64(contentWidth) * 0.58)
		statsWidth := contentWidth - taskWidth - 1

		clockHeight := lipgloss.Height(clockBox)
		calcHeight := termH - clockHeight - 4
		if calcHeight < 4 {
			calcHeight = 4
		}
		if calcHeight > 10 {
			calcHeight = 10
		}

		taskBox := m.renderTaskBox(taskWidth, calcHeight)
		statsBox := m.renderStatsBox(statsWidth, calcHeight)
		lowerSection = lipgloss.JoinHorizontal(lipgloss.Top, taskBox, " ", statsBox)
	}

	// Help Statusline
	helpItems := []string{
		fmt.Sprintf("%s %s", helpKeyStyle.Render("[Space]"), helpDescStyle.Render("Toggle")),
		fmt.Sprintf("%s %s", helpKeyStyle.Render("[n]"), helpDescStyle.Render("Next")),
		fmt.Sprintf("%s %s", helpKeyStyle.Render("[r]"), helpDescStyle.Render("Reset")),
		fmt.Sprintf("%s %s", helpKeyStyle.Render("[a]"), helpDescStyle.Render("Add")),
		fmt.Sprintf("%s %s", helpKeyStyle.Render("[d]"), helpDescStyle.Render("Done")),
		fmt.Sprintf("%s %s", helpKeyStyle.Render("[c]"), helpDescStyle.Render("Config")),
		fmt.Sprintf("%s %s", helpKeyStyle.Render("[q]"), helpDescStyle.Render("Quit")),
	}
	helpBar := lipgloss.NewStyle().Render(strings.Join(helpItems, " "))

	mainView := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		clockBox,
		lowerSection,
		helpBar,
	)

	return lipgloss.Place(
		termW,
		termH,
		lipgloss.Center,
		lipgloss.Center,
		mainView,
	)
}

func (m Model) renderTooSmallView(termW, termH int) string {
	boxWidth := termW - 2
	if boxWidth < 20 {
		boxWidth = 20
	}
	if boxWidth > 46 {
		boxWidth = 46
	}

	var content strings.Builder
	content.WriteString(headerStyle.Render("Terminal Too Small") + "\n\n")
	content.WriteString(fmt.Sprintf("Current:  %d x %d\n", termW, termH))
	content.WriteString(fmt.Sprintf("Required: %d x %d (min)\n\n", MinTerminalWidth, MinTerminalHeight))
	content.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render("Please enlarge your terminal window."))

	box := makeBox(boxWidth, 0, lipgloss.Center, content.String())

	return lipgloss.Place(
		termW,
		termH,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

func (m Model) renderTaskBox(width, height int) string {
	var taskListContent strings.Builder
	taskListContent.WriteString(headerStyle.Render("Tasks Queue") + "\n")

	if len(m.tasks) == 0 {
		taskListContent.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(" (No tasks. Press 'a')") + "\n")
	} else {
		maxItems := height - 3
		if maxItems < 1 {
			maxItems = 1
		}
		startIdx := 0
		if m.selectedTask >= maxItems {
			startIdx = m.selectedTask - maxItems + 1
		}
		endIdx := startIdx + maxItems
		if endIdx > len(m.tasks) {
			endIdx = len(m.tasks)
		}

		for i := startIdx; i < endIdx; i++ {
			t := m.tasks[i]
			prefix := " "
			if i == m.selectedTask {
				prefix = ">"
			}

			status := "[ ]"
			if t.IsDone {
				status = "[x]"
			}

			title := t.Title
			maxTitleLen := width - 15
			if maxTitleLen > 5 && len(title) > maxTitleLen {
				title = title[:maxTitleLen-3] + "..."
			}

			line := fmt.Sprintf("%s%s %s (%d/%d)", prefix, status, title, t.Completed, t.Target)
			if t.Title == m.snapshot.ActiveTaskTitle && !t.IsDone {
				line += " *"
			}

			if t.IsDone {
				taskListContent.WriteString(taskDoneStyle.Render(line) + "\n")
			} else if i == m.selectedTask {
				taskListContent.WriteString(taskActiveStyle.Render(line) + "\n")
			} else {
				taskListContent.WriteString(baseStyle.Render(line) + "\n")
			}
		}
	}

	if m.inputMode == ModeAddingTask {
		taskListContent.WriteString("\n" + lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Render("New: ") + m.textInput.View())
	}

	return makeBox(width, height, lipgloss.Left, taskListContent.String())
}

func (m Model) renderStatsBox(width, height int) string {
	todayStats := m.todayStats
	if todayStats.Date == "" && m.store != nil {
		todayStats = m.store.GetTodayStats()
	}
	soundStatus := "ON"
	if !m.snapshot.SoundEnabled {
		soundStatus = "OFF"
	}

	var statsContent strings.Builder
	statsContent.WriteString(headerStyle.Render("Today Focus") + "\n")
	statsContent.WriteString(fmt.Sprintf("Pomos Done: %s\n", lipgloss.NewStyle().Bold(true).Foreground(colorYellow).Render(fmt.Sprintf("%d", todayStats.CompletedPomos))))
	statsContent.WriteString(fmt.Sprintf("Focus Time: %s\n", lipgloss.NewStyle().Bold(true).Foreground(colorTextLight).Render(fmt.Sprintf("%d m", todayStats.FocusMinutes))))
	statsContent.WriteString(fmt.Sprintf("Tasks Done: %s\n", lipgloss.NewStyle().Bold(true).Foreground(colorBreak).Render(fmt.Sprintf("%d", todayStats.CompletedTasks))))
	statsContent.WriteString(fmt.Sprintf("Sound [m]:  %s\n", lipgloss.NewStyle().Foreground(colorTextDim).Render(soundStatus)))

	activeName := m.snapshot.ActiveTaskTitle
	maxActiveLen := width - 13
	if maxActiveLen < 5 {
		maxActiveLen = 5
	}
	if len(activeName) > maxActiveLen {
		activeName = activeName[:maxActiveLen-3] + "..."
	}
	statsContent.WriteString(fmt.Sprintf("Active:     %s", lipgloss.NewStyle().Foreground(colorActiveTask).Render(activeName)))

	return makeBox(width, height, lipgloss.Left, statsContent.String())
}

func (m Model) renderConfigModal(termW, termH int) string {
	modalWidth := 54
	if termW > 0 && termW-4 < modalWidth {
		modalWidth = termW - 4
	}

	var content strings.Builder
	content.WriteString(headerStyle.Render("Configuration Settings") + "\n\n")

	fields := []struct {
		label string
		value string
	}{
		{"Work Duration", fmt.Sprintf("[ %2d ] min", int(m.config.WorkDuration.Minutes()))},
		{"Short Break", fmt.Sprintf("[ %2d ] min", int(m.config.ShortBreakDuration.Minutes()))},
		{"Long Break", fmt.Sprintf("[ %2d ] min", int(m.config.LongBreakDuration.Minutes()))},
		{"Long Break Interval", fmt.Sprintf("[  %d ] cycles", m.config.LongBreakInterval)},
		{"Auto-start Breaks", boolString(m.config.AutoStartBreak)},
		{"Auto-start Focus", boolString(m.config.AutoStartWork)},
		{"Sound Cues", boolString(m.config.SoundEnabled)},
		{"Desktop Notifications", boolString(m.config.NotificationEnable)},
	}

	for i, f := range fields {
		prefix := "  "
		style := baseStyle
		if i == m.configCursor {
			prefix = "> "
			style = configSelectedStyle
		}

		line := fmt.Sprintf("%s%-24s %s", prefix, f.label+":", f.value)
		content.WriteString(style.Render(line) + "\n")
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("-", modalWidth-6)) + "\n")
	content.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render("[j/k] Select   [h/l hoặc +/-] Adjust   [Space] Toggle\n[Enter / Esc / c] Save & Close"))

	modal := makeBox(modalWidth, 0, lipgloss.Left, content.String())

	return lipgloss.Place(
		termW,
		termH,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}

func boolString(v bool) string {
	if v {
		return "[ ON  ]"
	}
	return "[ OFF ]"
}
