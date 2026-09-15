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

	// 2. Help Overlay (available from any tab, takes priority over everything but the size check)
	if m.showHelp {
		return m.renderHelpOverlay(termW, termH)
	}

	// 3. Zen Mode View (Tab 1 only, pure minimalist clock)
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

	// Record exactly where this frame places the main view on screen (same formula lipgloss.Place
	// uses internally for Center/Center) so mouse clicks can be mapped back to header tab bounds.
	if m.layout != nil {
		left := (termW - lipgloss.Width(mainView)) / 2
		if left < 0 {
			left = 0
		}
		top := (termH - lipgloss.Height(mainView)) / 2
		if top < 0 {
			top = 0
		}
		m.layout.left = left
		m.layout.top = top
	}

	return lipgloss.Place(
		termW,
		termH,
		lipgloss.Center,
		lipgloss.Center,
		mainView,
	)
}

// renderHeader renders the top navigation bar with 4 clickable tab buttons and session badge.
// It also records each tab's clickable column range into m.tabBounds (if set), and always
// returns a string of exactly totalWidth visible columns so it stays aligned with the rest of
// the frame and so those recorded bounds stay meaningful.
func (m Model) renderHeader(totalWidth int) string {
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
	badgeW := lipgloss.Width(sessionBadge)

	// Lay out brand + tabs, tracking each tab's [start,end) column range as we go. Drop the
	// brand tag first when it's too narrow to fit everything (small/split terminal panes),
	// since the tab bar and session badge carry the essential information.
	showBrand := true
	leftW, tabBounds := layoutTabBar(brand, renderedTabs, showBrand)
	if leftW+badgeW > totalWidth {
		showBrand = false
		leftW, tabBounds = layoutTabBar(brand, renderedTabs, showBrand)
	}
	if m.tabBounds != nil {
		*m.tabBounds = tabBounds
	}

	var leftPart string
	if showBrand {
		leftPart = brand + " " + strings.Join(renderedTabs, "│")
	} else {
		leftPart = strings.Join(renderedTabs, "│")
	}

	space := totalWidth - leftW - badgeW
	if space < 1 {
		space = 1
	}

	line := lipgloss.JoinHorizontal(
		lipgloss.Center,
		leftPart,
		strings.Repeat(" ", space),
		sessionBadge,
	)

	// Safety net: guarantee the header is always exactly totalWidth wide (pad if short,
	// truncate if a very narrow terminal still can't fit everything) so it stays aligned
	// with the box and footer below it, and so tabBounds above stay accurate. Note: this
	// uses MaxWidth (hard truncate) rather than Style.Width, which would word-wrap an
	// overlong line onto a second row instead of clipping it.
	if lineWidth := lipgloss.Width(line); lineWidth > totalWidth {
		line = lipgloss.NewStyle().MaxWidth(totalWidth).Render(line)
	} else if lineWidth < totalWidth {
		line += strings.Repeat(" ", totalWidth-lineWidth)
	}
	return line
}

// layoutTabBar computes the total rendered width of the brand+tab-bar segment and the
// [start,end) column range of each of the 4 tabs within it, without the trailing session badge.
func layoutTabBar(brand string, renderedTabs []string, showBrand bool) (totalWidth int, bounds [4]TabBounds) {
	col := 0
	if showBrand {
		col = lipgloss.Width(brand) + 1 // brand + single space separator
	}
	for i, tab := range renderedTabs {
		w := lipgloss.Width(tab)
		bounds[i] = TabBounds{StartX: col, EndX: col + w}
		col += w
		if i < len(renderedTabs)-1 {
			col++ // "│" separator
		}
	}
	return col, bounds
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
			fmt.Sprintf("%s %s", keyStyle.Render("[a]"), descStyle.Render("Add Task")),
			fmt.Sprintf("%s %s", keyStyle.Render("[t]"), descStyle.Render("Theme")),
			fmt.Sprintf("%s %s", keyStyle.Render("[?]"), descStyle.Render("Help")),
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
			fmt.Sprintf("%s %s", keyStyle.Render("[C]"), descStyle.Render("Clear Done")),
			fmt.Sprintf("%s %s", keyStyle.Render("[?]"), descStyle.Render("Help")),
			fmt.Sprintf("%s %s", keyStyle.Render("[q]"), descStyle.Render("Quit")),
		}
	case TabStats:
		items = []string{
			fmt.Sprintf("%s %s", keyStyle.Render("[v]"), descStyle.Render("View")),
			fmt.Sprintf("%s %s", keyStyle.Render("[x]"), descStyle.Render("Export MD")),
			fmt.Sprintf("%s %s", keyStyle.Render("[c]"), descStyle.Render("Export CSV")),
			fmt.Sprintf("%s %s", keyStyle.Render("[r]"), descStyle.Render("Refresh")),
			fmt.Sprintf("%s %s", keyStyle.Render("[t]"), descStyle.Render("Theme")),
			fmt.Sprintf("%s %s", keyStyle.Render("[?]"), descStyle.Render("Help")),
			fmt.Sprintf("%s %s", keyStyle.Render("[q]"), descStyle.Render("Quit")),
		}
	case TabSettings:
		items = []string{
			fmt.Sprintf("%s %s", keyStyle.Render("[j/k]"), descStyle.Render("Navigate")),
			fmt.Sprintf("%s %s", keyStyle.Render("[h/l]"), descStyle.Render("Change")),
			fmt.Sprintf("%s %s", keyStyle.Render("[Space]"), descStyle.Render("Toggle")),
			fmt.Sprintf("%s %s", keyStyle.Render("[?]"), descStyle.Render("Help")),
			fmt.Sprintf("%s %s", keyStyle.Render("[q]"), descStyle.Render("Quit")),
		}
	}

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(wrapItems(items, width, "  "))
}

// wrapItems joins hint/metric items with sep, wrapping onto additional lines at item
// boundaries (never mid-item) so a row of chips never overflows or gets clipped mid-word on
// narrower terminals.
func wrapItems(items []string, width int, sep string) string {
	sepWidth := lipgloss.Width(sep)

	var lines []string
	var current []string
	currentWidth := 0

	for _, item := range items {
		itemWidth := lipgloss.Width(item)
		addWidth := itemWidth
		if len(current) > 0 {
			addWidth += sepWidth
		}
		if len(current) > 0 && currentWidth+addWidth > width {
			lines = append(lines, strings.Join(current, sep))
			current = []string{item}
			currentWidth = itemWidth
		} else {
			current = append(current, item)
			currentWidth += addWidth
		}
	}
	if len(current) > 0 {
		lines = append(lines, strings.Join(current, sep))
	}

	return strings.Join(lines, "\n")
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

// renderHelpOverlay renders the full keybinding cheatsheet, grouped by tab context.
func (m Model) renderHelpOverlay(termW, termH int) string {
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.Accent)
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.Text)
	descStyle := lipgloss.NewStyle().Foreground(m.theme.TextDim)

	row := func(key, desc string) string {
		return fmt.Sprintf("  %-13s %s", keyStyle.Render(key), descStyle.Render(desc))
	}

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(m.theme.Text).Render("ZENPOMO — KEYBINDINGS") + "\n\n")

	content.WriteString(sectionStyle.Render("Global") + "\n")
	content.WriteString(row("1-4 / Tab", "Switch tabs") + "\n")
	content.WriteString(row("t", "Cycle color theme") + "\n")
	content.WriteString(row("?", "Toggle this help") + "\n")
	content.WriteString(row("q / Ctrl+C", "Quit (daemon keeps running)") + "\n\n")

	content.WriteString(sectionStyle.Render("Timer") + "\n")
	content.WriteString(row("Space", "Start / Pause") + "\n")
	content.WriteString(row("n", "Skip to next session") + "\n")
	content.WriteString(row("r", "Reset current session") + "\n")
	content.WriteString(row("z", "Toggle Zen mode") + "\n")
	content.WriteString(row("m", "Mute / unmute audio cues") + "\n")
	content.WriteString(row("a", "Quick-add a task") + "\n\n")

	content.WriteString(sectionStyle.Render("Tasks") + "\n")
	content.WriteString(row("j/k", "Move selection") + "\n")
	content.WriteString(row("J/K", "Reorder task in queue") + "\n")
	content.WriteString(row("Space/Enter", "Set as active task") + "\n")
	content.WriteString(row("a", "Add task") + "\n")
	content.WriteString(row("e", "Edit task") + "\n")
	content.WriteString(row("d", "Toggle done") + "\n")
	content.WriteString(row("x", "Delete task") + "\n")
	content.WriteString(row("C", "Clear completed") + "\n\n")

	content.WriteString(sectionStyle.Render("Stats") + "\n")
	content.WriteString(row("v", "Cycle Daily/Weekly/Monthly view") + "\n")
	content.WriteString(row("x", "Export Markdown report") + "\n")
	content.WriteString(row("c", "Export CSV") + "\n")
	content.WriteString(row("r", "Refresh") + "\n\n")

	content.WriteString(sectionStyle.Render("Settings") + "\n")
	content.WriteString(row("j/k", "Navigate options") + "\n")
	content.WriteString(row("h/l", "Change value") + "\n")
	content.WriteString(row("Space", "Toggle option") + "\n\n")

	content.WriteString(lipgloss.NewStyle().Foreground(m.theme.Border).Render("Press any key to close"))

	boxWidth := 56
	if boxWidth > termW-2 {
		boxWidth = termW - 2
	}
	box := m.makeBox(boxWidth, 0, lipgloss.Left, content.String())

	return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, box)
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
// boxContentWidth returns how many visible columns of plain text fit on one line inside a
// makeBox of the given total width, after its border and padding are accounted for.
func boxContentWidth(totalWidth int) int {
	w := totalWidth - 6
	if w < 1 {
		w = 1
	}
	return w
}

// truncateRunes trims s to at most n runes, replacing the tail with an ellipsis when it doesn't
// fit. Operates on runes (not bytes), so it never splits a multi-byte UTF-8 character (e.g. a
// Vietnamese task title) in the middle.
func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n == 1 {
		return string(r[:1])
	}
	return string(r[:n-1]) + "…"
}

// padRunes right-pads s with spaces to exactly n runes wide. Use after truncateRunes so a
// column always occupies exactly n columns, keeping rows aligned like a table.
func padRunes(s string, n int) string {
	l := len([]rune(s))
	if l >= n {
		return s
	}
	return s + strings.Repeat(" ", n-l)
}

// fitRunes truncates-or-pads s to exactly n runes, so it always occupies exactly one fixed-width
// table column.
func fitRunes(s string, n int) string {
	return padRunes(truncateRunes(s, n), n)
}

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
