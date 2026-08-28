package app

import (
	"fmt"
	"strings"
	"zenpomo/internal/clock"

	"github.com/charmbracelet/lipgloss"
)

// renderTimerTab renders Tab 1: Single cohesive focus card with balanced vertical rhythm.
func (m Model) renderTimerTab(contentWidth, termH int) string {
	sessionColor := m.getSessionColor()

	// Progress Bar (scaled to fit nicely in the frame)
	barWidth := contentWidth - 28
	if barWidth < 12 {
		barWidth = 12
	}
	if barWidth > 48 {
		barWidth = 48
	}

	completedWidth := int(m.snapshot.ProgressRatio * float64(barWidth))
	if completedWidth > barWidth {
		completedWidth = barWidth
	}
	remainingWidth := barWidth - completedWidth

	filledBar := lipgloss.NewStyle().Foreground(sessionColor).Render(strings.Repeat("=", completedWidth))
	emptyBar := lipgloss.NewStyle().Foreground(m.theme.Border).Render(strings.Repeat("-", remainingWidth))
	progressBar := fmt.Sprintf("[%s%s] %3.0f%%", filledBar, emptyBar, m.snapshot.ProgressRatio*100)

	// Cycle info badges
	var cycleBadges strings.Builder
	for i := 1; i <= m.snapshot.TargetCycles; i++ {
		if i <= m.snapshot.CycleCount {
			cycleBadges.WriteString("● ")
		} else {
			cycleBadges.WriteString("○ ")
		}
	}
	cycleInfo := fmt.Sprintf("Cycle: [ %s] (%d/%d)", cycleBadges.String(), m.snapshot.CycleCount, m.snapshot.TargetCycles)

	// Clock Section (Big ASCII when height >= 18 and width >= 48, else compact)
	var clockRendered string
	if termH >= 18 && contentWidth >= 48 {
		clockLines := clock.RenderTime(m.snapshot.RemainingSeconds)
		clockRendered = lipgloss.NewStyle().
			Foreground(sessionColor).
			Bold(true).
			Render(strings.Join(clockLines, "\n"))
	} else {
		mins := m.snapshot.RemainingSeconds / 60
		secs := m.snapshot.RemainingSeconds % 60
		clockRendered = lipgloss.NewStyle().
			Foreground(sessionColor).
			Bold(true).
			Render(fmt.Sprintf("[ %02d:%02d ]", mins, secs))
	}

	// Bottom Focus Info (Single-line 3-column layout without line wrapping)
	activeTask := m.snapshot.ActiveTaskTitle
	if activeTask == "" {
		activeTask = "General Focus"
	}
	maxTaskLen := contentWidth/3 - 4
	if maxTaskLen < 10 {
		maxTaskLen = 10
	}
	if len(activeTask) > maxTaskLen {
		activeTask = activeTask[:maxTaskLen-3] + "..."
	}

	soundStatus := "Off"
	if m.snapshot.SoundEnabled {
		soundStatus = "Chimes On"
	}
	if m.config.AmbientSound != "" && m.config.AmbientSound != "none" {
		soundStatus = fmt.Sprintf("Ambient: %s", strings.Title(m.config.AmbientSound))
	}

	todayPomos := m.todayStats.CompletedPomos
	focusMins := m.todayStats.FocusMinutes

	metaCols := fmt.Sprintf("Task: %s   │   Sound: %s   │   Today: %d pomos (%dm)",
		lipgloss.NewStyle().Bold(true).Foreground(m.theme.Accent).Render(activeTask),
		lipgloss.NewStyle().Foreground(m.theme.TextDim).Render(soundStatus),
		todayPomos,
		focusMins,
	)

	dividerWidth := contentWidth - 6
	if dividerWidth < 10 {
		dividerWidth = 10
	}
	divider := lipgloss.NewStyle().
		Foreground(m.theme.Border).
		Render(strings.Repeat("─", dividerWidth))

	cardContent := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		clockRendered,
		"",
		progressBar,
		lipgloss.NewStyle().Foreground(m.theme.TextDim).Render(cycleInfo),
		"",
		divider,
		metaCols,
	)

	calcHeight := termH - 6
	if calcHeight < 12 {
		calcHeight = 12
	}
	if calcHeight > 18 {
		calcHeight = 18
	}

	return m.makeBox(contentWidth, calcHeight, lipgloss.Center, cardContent)
}
