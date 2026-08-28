package app

import (
	"fmt"
	"strings"
	"zenpomo/internal/core"

	"github.com/charmbracelet/lipgloss"
)

const (
	// Minimum terminal dimensions required for standard tabbed layout.
	MinTerminalWidth  = 46
	MinTerminalHeight = 13
)

// View renders the entire TUI application matching terminal dimensions.
func (m Model) View() string {
	termW := m.width
	termH := m.height

	if termW <= 0 {
		termW = 80
	}
	if termH <= 0 {
		termH = 24
	}

	// 1. Minimum Dimensions Check
	if termW < MinTerminalWidth || termH < MinTerminalHeight {
		return m.renderTooSmallView(termW, termH)
	}

	// 2. Zen Mode View (Tab 1 only, pure minimalist clock)
	if m.zenMode && m.activeTab == TabTimer {
		return m.renderZenModeView(termW, termH)
	}

	// 3. Responsive Container Width
	contentWidth := termW - 4
	if contentWidth > 96 {
		contentWidth = 96
	}
	if contentWidth < MinTerminalWidth {
		contentWidth = MinTerminalWidth
	}

	// 4. Header with 4 Tabs & Session Status Badge
	header := m.renderHeader(contentWidth)

	// 5. Active Tab Body Content
	var tabContent string
	switch m.activeTab {
	case TabTimer:
		tabContent = m.renderTimerTab(contentWidth, termH)
	case TabTasks:
		tabContent = m.renderTasksTab(contentWidth, termH)
	case TabStats:
		tabContent = m.renderStatsTab(contentWidth, termH)
	case TabSettings:
		tabContent = m.renderSettingsTab(contentWidth, termH)
	default:
		tabContent = m.renderTimerTab(contentWidth, termH)
	}

	// 6. Footer / Status & Help Line
	footer := m.renderFooter(contentWidth)

	// Join all vertical sections
	mainView := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		tabContent,
		footer,
	)

	return lipgloss.Place(
		termW,
		termH,
		lipgloss.Center,
		lipgloss.Center,
		mainView,
	)
}

// renderHeader renders the top navigation bar with 4 clickable tab buttons and session badge.
func (m *Model) renderHeader(totalWidth int) string {
	sessionColor := m.getSessionColor()

	// Brand title
	brandStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Text).
		Background(m.theme.Border).
		Padding(0, 1)
	brand := brandStyle.Render("ZENPOMO")

	// Tab Definitions (Responsive tab names so they fit all terminal sizes)
	tabLabels := []string{"[1] Timer", "[2] Tasks", "[3] Stats", "[4] Config"}
	if totalWidth >= 84 {
		tabLabels = []string{"[1] Timer", "[2] Tasks", "[3] Analytics", "[4] Settings"}
	}

	var renderedTabs []string
	for i, label := range tabLabels {
		if Tab(i) == m.activeTab {
			activeStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(m.theme.Accent).
				Background(m.theme.TabActiveBg).
				Padding(0, 1)
			renderedTabs = append(renderedTabs, activeStyle.Render(label))
		} else {
			inactiveStyle := lipgloss.NewStyle().
				Foreground(m.theme.TextDim).
				Padding(0, 1)
			renderedTabs = append(renderedTabs, inactiveStyle.Render(label))
		}
	}

	tabBar := strings.Join(renderedTabs, "│")

	// Session Status Badge
	stateText := string(m.snapshot.State)
	if m.snapshot.State == core.StateRunning {
		stateText = "RUNNING"
	}
	dot := "●"
	if m.snapshot.State == core.StateStopped {
		dot = "■"
	} else if m.snapshot.State == core.StatePaused {
		dot = "○"
	}

	var badgeText string
	if totalWidth >= 78 {
		badgeText = fmt.Sprintf("%s %s • %s", dot, strings.ToUpper(string(m.snapshot.Session)), stateText)
	} else {
		badgeText = fmt.Sprintf("%s %s", dot, strings.ToUpper(string(m.snapshot.Session)))
	}

	sessionBadge := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1D2021")).
		Background(sessionColor).
		Padding(0, 1).
		Render(badgeText)

	// Available space calculation
	leftPart := lipgloss.JoinHorizontal(lipgloss.Center, brand, " ", tabBar)
	leftW := lipgloss.Width(leftPart)
	badgeW := lipgloss.Width(sessionBadge)
	space := totalWidth - leftW - badgeW
	if space < 1 {
		space = 1
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		leftPart,
		strings.Repeat(" ", space),
		sessionBadge,
	)
}

// renderFooter renders quick action keybindings and status notifications.
func (m Model) renderFooter(width int) string {
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.Accent)
	descStyle := lipgloss.NewStyle().Foreground(m.theme.TextDim)

	var items []string
	switch m.activeTab {
	case TabTimer:
		items = []string{
			fmt.Sprintf("%s %s", keyStyle.Render("[Space]"), descStyle.Render("Toggle")),
			fmt.Sprintf("%s %s", keyStyle.Render("[n]"), descStyle.Render("Next")),
			fmt.Sprintf("%s %s", keyStyle.Render("[r]"), descStyle.Render("Reset")),
			fmt.Sprintf("%s %s", keyStyle.Render("[z]"), descStyle.Render("Zen")),
			fmt.Sprintf("%s %s", keyStyle.Render("[1-4]"), descStyle.Render("Tabs")),
			fmt.Sprintf("%s %s", keyStyle.Render("[t]"), descStyle.Render("Theme")),
			fmt.Sprintf("%s %s", keyStyle.Render("[q]"), descStyle.Render("Quit")),
		}
	case TabTasks:
		items = []string{
			fmt.Sprintf("%s %s", keyStyle.Render("[j/k]"), descStyle.Render("Select")),
			fmt.Sprintf("%s %s", keyStyle.Render("[J/K]"), descStyle.Render("Move")),
			fmt.Sprintf("%s %s", keyStyle.Render("[Space]"), descStyle.Render("Active")),
			fmt.Sprintf("%s %s", keyStyle.Render("[a]"), descStyle.Render("Add")),
			fmt.Sprintf("%s %s", keyStyle.Render("[e]"), descStyle.Render("Edit")),
			fmt.Sprintf("%s %s", keyStyle.Render("[d]"), descStyle.Render("Done")),
			fmt.Sprintf("%s %s", keyStyle.Render("[x]"), descStyle.Render("Del")),
			fmt.Sprintf("%s %s", keyStyle.Render("[1-4]"), descStyle.Render("Tabs")),
			fmt.Sprintf("%s %s", keyStyle.Render("[q]"), descStyle.Render("Quit")),
		}
	case TabStats:
		items = []string{
			fmt.Sprintf("%s %s", keyStyle.Render("[x]"), descStyle.Render("Export")),
			fmt.Sprintf("%s %s", keyStyle.Render("[r]"), descStyle.Render("Refresh")),
			fmt.Sprintf("%s %s", keyStyle.Render("[1-4]"), descStyle.Render("Tabs")),
			fmt.Sprintf("%s %s", keyStyle.Render("[t]"), descStyle.Render("Theme")),
			fmt.Sprintf("%s %s", keyStyle.Render("[?]"), descStyle.Render("Help")),
			fmt.Sprintf("%s %s", keyStyle.Render("[q]"), descStyle.Render("Quit")),
		}
	case TabSettings:
		items = []string{
			fmt.Sprintf("%s %s", keyStyle.Render("[j/k]"), descStyle.Render("Navigate")),
			fmt.Sprintf("%s %s", keyStyle.Render("[h/l]"), descStyle.Render("Change")),
			fmt.Sprintf("%s %s", keyStyle.Render("[Space]"), descStyle.Render("Toggle")),
			fmt.Sprintf("%s %s", keyStyle.Render("[1-4]"), descStyle.Render("Tabs")),
			fmt.Sprintf("%s %s", keyStyle.Render("[q]"), descStyle.Render("Quit")),
		}
	}

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(strings.Join(items, "  "))
}

// renderZenModeView renders the ultra-minimalist focus screen.
func (m Model) renderZenModeView(termW, termH int) string {
	sessionColor := m.getSessionColor()

	mins := m.snapshot.RemainingSeconds / 60
	secs := m.snapshot.RemainingSeconds % 60
	clockStr := fmt.Sprintf("%02d:%02d", mins, secs)

	clockRendered := lipgloss.NewStyle().
		Bold(true).
		Foreground(sessionColor).
		Render(clockStr)

	taskName := m.snapshot.ActiveTaskTitle
	if taskName == "" {
		taskName = "General Focus"
	}
	taskRendered := lipgloss.NewStyle().
		Foreground(m.theme.TextDim).
		Render(taskName)

	hint := lipgloss.NewStyle().
		Foreground(m.theme.Border).
		Render("press [z] to exit zen mode • [Space] pause")

	zenContent := lipgloss.JoinVertical(
		lipgloss.Center,
		clockRendered,
		"\n",
		taskRendered,
		"\n\n",
		hint,
	)

	return lipgloss.Place(
		termW,
		termH,
		lipgloss.Center,
		lipgloss.Center,
		zenContent,
	)
}

// renderTooSmallView displays a polite message when terminal window is too constrained.
func (m Model) renderTooSmallView(termW, termH int) string {
	boxWidth := termW - 2
	if boxWidth < 20 {
		boxWidth = 20
	}
	if boxWidth > 46 {
		boxWidth = 46
	}

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(m.theme.Text).Render("Terminal Too Small") + "\n\n")
	content.WriteString(fmt.Sprintf("Current:  %d x %d\n", termW, termH))
	content.WriteString(fmt.Sprintf("Required: %d x %d (min)\n\n", MinTerminalWidth, MinTerminalHeight))
	content.WriteString(lipgloss.NewStyle().Foreground(m.theme.TextDim).Render("Please enlarge your terminal window."))

	box := m.makeBox(boxWidth, 0, lipgloss.Center, content.String())

	return lipgloss.Place(
		termW,
		termH,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// makeBox creates a themed rounded border box.
func (m Model) makeBox(totalWidth, totalHeight int, align lipgloss.Position, content string) string {
	if totalWidth < 10 {
		totalWidth = 10
	}
	innerWidth := totalWidth - 4
	if innerWidth < 1 {
		innerWidth = 1
	}

	st := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
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

// getSessionColor returns the current active theme color according to session type.
func (m Model) getSessionColor() lipgloss.Color {
	switch m.snapshot.Session {
	case core.SessionWork:
		return m.theme.Work
	case core.SessionShortBreak:
		return m.theme.Break
	case core.SessionLongBreak:
		return m.theme.LongBreak
	default:
		return m.theme.Work
	}
}
