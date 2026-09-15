package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderSettingsTab renders Tab 4: Theme, Durations, Ambient Sound, and Automation Settings.
func (m Model) renderSettingsTab(contentWidth, termH int) string {
	calcHeight := termH - 7
	if calcHeight < 8 {
		calcHeight = 8
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Text)

	cursorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Accent)

	dimStyle := lipgloss.NewStyle().
		Foreground(m.theme.TextDim)

	var content strings.Builder
	content.WriteString(headerStyle.Render("SETTINGS & ENVIRONMENT") + "\n\n")

	ambientDisplay := m.config.AmbientSound
	if ambientDisplay == "" {
		ambientDisplay = "none"
	}

	fields := []struct {
		label string
		value string
		hint  string
	}{
		{"Active Theme", fmt.Sprintf("[ %-12s ]", strings.ToUpper(m.theme.Name)), "Gruvbox, Catppuccin, TokyoNight, Nord, Dracula, RosePine, Monochrome"},
		{"Work Duration", fmt.Sprintf("[ %2d min ]", int(m.config.WorkDuration.Minutes())), "Focus period length (1-120 min)"},
		{"Short Break", fmt.Sprintf("[ %2d min ]", int(m.config.ShortBreakDuration.Minutes())), "Rest duration between pomodoros (1-60 min)"},
		{"Long Break", fmt.Sprintf("[ %2d min ]", int(m.config.LongBreakDuration.Minutes())), "Extended recovery rest (1-90 min)"},
		{"Long Break Interval", fmt.Sprintf("[ %2d cycles ]", m.config.LongBreakInterval), "Cycles before a long break (1-12)"},
		{"Auto-start Breaks", boolBadge(m.config.AutoStartBreak), "Automatically begin rest timer after work"},
		{"Auto-start Focus", boolBadge(m.config.AutoStartWork), "Automatically begin next pomo after rest"},
		{"Sound Chimes", boolBadge(m.config.SoundEnabled), "Play audio chimes on session transition"},
		{"Desktop Notifications", boolBadge(m.config.NotificationEnable), "Send OS native banner notifications"},
		{"Ambient Sound", fmt.Sprintf("[ %-10s ]", strings.Title(ambientDisplay)), "None, Rain, Whitenoise, Waves, Coffee"},
	}

	// Fixed-width columns (prefix/label/value) plus a hint column that fills whatever space is
	// left and is truncated to fit, so a long hint (e.g. the theme list) can never wrap a row
	// onto a second line and break alignment with the rows around it.
	const valueColWidth = 16 // exactly fits the widest rendered value, e.g. "[ GRUVBOX      ]"
	usable := boxContentWidth(contentWidth)
	const prefixWidth = 2
	const labelSep = 1
	labelColWidth := 22
	if maxLabel := usable - prefixWidth - labelSep - valueColWidth; maxLabel < labelColWidth {
		labelColWidth = maxLabel // shrink the label column before ever letting a row overflow
	}
	if labelColWidth < 6 {
		labelColWidth = 6
	}
	hintColWidth := usable - prefixWidth - labelColWidth - labelSep - valueColWidth - 3 // " (" + ")"
	if hintColWidth < 0 {
		hintColWidth = 0
	}

	for i, f := range fields {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(m.theme.Text)
		if i == m.configCursor {
			prefix = "> "
			style = cursorStyle
		}

		hint := ""
		if hintColWidth > 0 {
			hint = " " + dimStyle.Render("("+truncateRunes(f.hint, hintColWidth)+")")
		}
		line := prefix + fitRunes(f.label+":", labelColWidth) + " " + fitRunes(f.value, valueColWidth) + hint

		// Safety net: a styled value/hint substring could measure a hair wider than intended;
		// clip rather than let the outer box word-wrap the row onto a second line.
		if w := lipgloss.Width(line); w > usable {
			line = lipgloss.NewStyle().MaxWidth(usable).Render(line)
		}
		content.WriteString(style.Render(line) + "\n")
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.BorderActive).Render(strings.Repeat("-", contentWidth-6)) + "\n")
	hints := []string{"[j/k] Select setting", "[h/l hoặc +/-] Adjust value", "[Space/Enter] Toggle option"}
	for i, h := range hints {
		hints[i] = dimStyle.Render(h)
	}
	content.WriteString(wrapItems(hints, usable, "   "))

	return m.makeBox(contentWidth, calcHeight, lipgloss.Left, content.String())
}

func boolBadge(v bool) string {
	if v {
		return "[ ON  ]"
	}
	return "[ OFF ]"
}
