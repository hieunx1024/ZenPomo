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
	content.WriteString(headerStyle.Render("ANALYTICS & INSIGHTS") + "  " + metricLabelStyle.Render("[ View: "+m.statsView.Label()+" ]") + "\n\n")

	// Calculate comprehensive summary
	summary := analytics.ComputeSummary(m.allStats, m.todayStats)

	usable := boxContentWidth(contentWidth)

	// 1. Metric Cards Row: chips that wrap onto their own lines (never mid-word) on narrow
	// terminals instead of getting clipped mid-value.
	metricsRow := wrapItems([]string{
		fmt.Sprintf("Streak: %s (Best: %s)", metricValStyle.Render(fmt.Sprintf("%d Days", summary.CurrentStreak)), metricValStyle.Render(fmt.Sprintf("%d Days", summary.LongestStreak))),
		"Focus: " + metricValStyle.Render(fmt.Sprintf("%dh %02dm", summary.TotalMinutes/60, summary.TotalMinutes%60)),
		"Pomos: " + metricValStyle.Render(fmt.Sprintf("%d", summary.TotalPomos)),
		"Avg: " + metricValStyle.Render(fmt.Sprintf("%.0f m/day", summary.DailyAverageMinutes)),
	}, usable, "  │  ")
	content.WriteString(metricsRow + "\n\n")

	switch m.statsView {
	case StatsViewWeekly:
		content.WriteString(headerStyle.Render("Focus Time — Last 8 Weeks") + "\n")
		content.WriteString(m.renderPeriodBars(contentWidth, analytics.ComputeWeeklyStats(m.allStats, m.todayStats, 8)) + "\n")
	case StatsViewMonthly:
		content.WriteString(headerStyle.Render("Focus Time — Last 6 Months") + "\n")
		content.WriteString(m.renderPeriodBars(contentWidth, analytics.ComputeMonthlyStats(m.allStats, m.todayStats, 6)) + "\n")
	default:
		// 2. 28-Day Contribution Heatmap (4 weeks x 7 days)
		legend := wrapItems([]string{
			headerStyle.Render("4-Week Focus Heatmap"),
			metricLabelStyle.Render("(░ 0m, ▒ 1-50m, ▓ 50-125m, █ >125m)"),
		}, usable, "  ")
		content.WriteString(legend + "\n")
		heatmap := m.renderHeatmap(28)
		content.WriteString(heatmap + "\n\n")

		// 3. 7-Day Sparkline Distribution
		content.WriteString(headerStyle.Render("Past 7 Days Focus Distribution") + "\n")
		sparkline := m.renderSparklines(summary.RecentDays)
		content.WriteString(sparkline + "\n")
	}

	// 4. Status Notification (e.g. Exported report)
	if m.statusMessage != "" && time.Now().Before(m.statusExpiry) {
		content.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(m.theme.Break).Render("[OK] "+m.statusMessage) + "\n")
	} else {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.BorderActive).Render(strings.Repeat("-", contentWidth-6)) + "\n")
		hints := []string{"[v] Cycle View", "[x] Export Markdown", "[c] Export CSV", "[r] Refresh", "[t] Theme"}
		for i, h := range hints {
			hints[i] = metricLabelStyle.Render(h)
		}
		content.WriteString(wrapItems(hints, usable, "   "))
	}

	return m.makeBox(contentWidth, calcHeight, lipgloss.Left, content.String())
}

// renderPeriodBars renders a horizontal bar chart for week/month aggregated stats. Column
// widths are fixed (and the bar shrinks on narrow terminals) so every row is exactly one line
// and the bars/values all line up underneath each other.
func (m Model) renderPeriodBars(contentWidth int, stats []analytics.PeriodStat) string {
	labelStyle := lipgloss.NewStyle().Foreground(m.theme.TextDim)
	valueStyle := lipgloss.NewStyle().Foreground(m.theme.Accent)
	barStyle := lipgloss.NewStyle().Foreground(m.theme.Work)

	maxMinutes := 1
	labelColWidth := 0
	for _, s := range stats {
		if s.FocusMinutes > maxMinutes {
			maxMinutes = s.FocusMinutes
		}
		if l := len([]rune(s.Label)); l > labelColWidth {
			labelColWidth = l
		}
	}

	const valueColWidth = 20 // "23h 59m  (999 pomos)" fits comfortably within this
	usable := boxContentWidth(contentWidth)
	barWidth := usable - labelColWidth - 2 - valueColWidth - 2
	if barWidth > 24 {
		barWidth = 24 // no need for a bar wider than this even when there's room to spare
	}
	if barWidth < 6 {
		barWidth = 6 // always show a minimally useful bar, even if the row must be clipped
	}

	var lines []string
	for _, s := range stats {
		filled := int((float64(s.FocusMinutes) / float64(maxMinutes)) * float64(barWidth))
		if filled < 0 {
			filled = 0
		}
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		label := fitRunes(s.Label, labelColWidth)
		valueText := fitRunes(fmt.Sprintf("%dh %02dm  (%d pomos)", s.FocusMinutes/60, s.FocusMinutes%60, s.CompletedPomos), valueColWidth)

		line := fmt.Sprintf("%s  %s  %s",
			labelStyle.Render(label),
			barStyle.Render(bar),
			valueStyle.Render(valueText),
		)
		if w := lipgloss.Width(line); w > usable {
			line = lipgloss.NewStyle().MaxWidth(usable).Render(line)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
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
