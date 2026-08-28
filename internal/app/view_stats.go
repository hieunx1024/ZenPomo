package app

import (
	"fmt"
	"strings"
	"time"
	"zenpomo/internal/analytics"

	"github.com/charmbracelet/lipgloss"
)

// renderStatsTab renders Tab 3: Visual Analytics with Contribution Heatmap, Sparklines, and Markdown Export.
func (m Model) renderStatsTab(contentWidth, termH int) string {
	calcHeight := termH - 7
	if calcHeight < 8 {
		calcHeight = 8
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Text)

	metricValStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Accent)

	metricLabelStyle := lipgloss.NewStyle().
		Foreground(m.theme.TextDim)

	var content strings.Builder
	content.WriteString(headerStyle.Render("ANALYTICS & INSIGHTS") + "\n\n")

	// Calculate comprehensive summary
	summary := analytics.ComputeSummary(m.allStats, m.todayStats)

	// 1. Metric Cards Row (Pure Unix typography)
	metricsRow := fmt.Sprintf("Streak: %s (Best: %s)  │  Focus: %s  │  Pomos: %s  │  Avg: %s",
		metricValStyle.Render(fmt.Sprintf("%d Days", summary.CurrentStreak)),
		metricValStyle.Render(fmt.Sprintf("%d Days", summary.LongestStreak)),
		metricValStyle.Render(fmt.Sprintf("%dh %02dm", summary.TotalMinutes/60, summary.TotalMinutes%60)),
		metricValStyle.Render(fmt.Sprintf("%d", summary.TotalPomos)),
		metricValStyle.Render(fmt.Sprintf("%.0f m/day", summary.DailyAverageMinutes)),
	)
	content.WriteString(metricsRow + "\n\n")

	// 2. 28-Day Contribution Heatmap (4 weeks x 7 days)
	content.WriteString(headerStyle.Render("4-Week Focus Heatmap") + "  " + metricLabelStyle.Render("(░ 0m, ▒ 1-50m, ▓ 50-125m, █ >125m)") + "\n")
	heatmap := m.renderHeatmap(28)
	content.WriteString(heatmap + "\n\n")

	// 3. 7-Day Sparkline Distribution
	content.WriteString(headerStyle.Render("Past 7 Days Focus Distribution") + "\n")
	sparkline := m.renderSparklines(summary.RecentDays)
	content.WriteString(sparkline + "\n")

	// 4. Status Notification (e.g. Exported report)
	if m.statusMessage != "" && time.Now().Before(m.statusExpiry) {
		content.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(m.theme.Break).Render("[OK] "+m.statusMessage) + "\n")
	} else {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.BorderActive).Render(strings.Repeat("-", contentWidth-6)) + "\n")
		content.WriteString(metricLabelStyle.Render("[x] Export Markdown (zenpomo-stats.md)   [r] Refresh Stats   [t] Cycle Theme"))
	}

	return m.makeBox(contentWidth, calcHeight, lipgloss.Left, content.String())
}

// renderHeatmap produces a 4-week grid of Unicode blocks representing focus density.
func (m Model) renderHeatmap(days int) string {
	now := time.Now()
	var blocks []string

	for i := days - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		dateKey := d.Format("2006-01-02")
		minutes := 0
		if stat, ok := m.allStats[dateKey]; ok {
			minutes = stat.FocusMinutes
		} else if dateKey == m.todayStats.Date {
			minutes = m.todayStats.FocusMinutes
		}

		var char string
		var style lipgloss.Style

		switch {
		case minutes == 0:
			char = "░"
			style = lipgloss.NewStyle().Foreground(m.theme.Border)
		case minutes <= 50:
			char = "▒"
			style = lipgloss.NewStyle().Foreground(m.theme.Break)
		case minutes <= 125:
			char = "▓"
			style = lipgloss.NewStyle().Foreground(m.theme.Highlight)
		default:
			char = "█"
			style = lipgloss.NewStyle().Foreground(m.theme.Work)
		}

		blocks = append(blocks, style.Render(char+" "))
		if (days-i)%7 == 0 {
			blocks = append(blocks, " ")
		}
	}

	return strings.Join(blocks, "")
}

// renderSparklines produces a 7-day sparkline bar chart.
func (m Model) renderSparklines(recentDays []analytics.DayStat) string {
	sparkChars := []string{" ", " ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	var lines []string
	var labels []string

	maxMinutes := 1
	for _, d := range recentDays {
		if d.FocusMinutes > maxMinutes {
			maxMinutes = d.FocusMinutes
		}
		// Format label MM/DD
		parts := strings.Split(d.Date, "-")
		if len(parts) == 3 {
			labels = append(labels, parts[1]+"/"+parts[2])
		} else {
			labels = append(labels, d.Date)
		}
	}

	for _, d := range recentDays {
		level := int((float64(d.FocusMinutes) / float64(maxMinutes)) * 8)
		if level < 0 {
			level = 0
		}
		if level > 8 {
			level = 8
		}
		char := sparkChars[level]
		lines = append(lines, lipgloss.NewStyle().Foreground(m.theme.Accent).Render(char+"    "))
	}

	return strings.Join(lines, "") + "\n" + lipgloss.NewStyle().Foreground(m.theme.TextDim).Render(strings.Join(labels, " "))
}
