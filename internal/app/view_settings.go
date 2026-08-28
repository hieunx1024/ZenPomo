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

	for i, f := range fields {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(m.theme.Text)
		if i == m.configCursor {
			prefix = "> "
			style = cursorStyle
		}

		line := fmt.Sprintf("%s%-22s %-16s %s", prefix, f.label+":", f.value, dimStyle.Render("("+f.hint+")"))
		content.WriteString(style.Render(line) + "\n")
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.BorderActive).Render(strings.Repeat("-", contentWidth-6)) + "\n")
	content.WriteString(dimStyle.Render("[j/k] Select setting   [h/l hoặc +/-] Adjust value   [Space/Enter] Toggle option"))

	return m.makeBox(contentWidth, calcHeight, lipgloss.Left, content.String())
}

func boolBadge(v bool) string {
	if v {
		return "[ ON  ]"
	}
	return "[ OFF ]"
}
