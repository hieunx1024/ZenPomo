package analytics

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	"zenpomo/internal/storage"
)

// DayStat holds summarized focus metrics for a single date.
type DayStat struct {
	Date           string
	FocusMinutes   int
	CompletedPomos int
	CompletedTasks int
}

// Summary contains aggregated metrics over the user's entire history.
type Summary struct {
	TotalPomos          int
	TotalMinutes        int
	TotalTasks          int
	CurrentStreak       int
	LongestStreak       int
	DailyAverageMinutes float64
	DaysTracked         int
	RecentDays          []DayStat // Last 7 days in chronological order
}

// ComputeSummary calculates all aggregate analytics from daily stats and tasks.
func ComputeSummary(stats map[string]storage.DailyStats, todayStats storage.DailyStats) Summary {
	merged := make(map[string]storage.DailyStats, len(stats)+1)
	for k, v := range stats {
		merged[k] = v
	}
	if todayStats.Date != "" {
		merged[todayStats.Date] = todayStats
	}

	summary := Summary{}
	if len(merged) == 0 {
		return summary
	}

	var dates []string
	for d, s := range merged {
		dates = append(dates, d)
		summary.TotalPomos += s.CompletedPomos
		summary.TotalMinutes += s.FocusMinutes
		summary.TotalTasks += s.CompletedTasks
	}

	sort.Strings(dates)
	summary.DaysTracked = len(dates)
	if summary.DaysTracked > 0 {
		summary.DailyAverageMinutes = float64(summary.TotalMinutes) / float64(summary.DaysTracked)
	}

	// Calculate streaks
	now := time.Now()
	currentStreak := 0
	for i := 0; i < 365; i++ {
		dStr := now.AddDate(0, 0, -i).Format("2006-01-02")
		if s, ok := merged[dStr]; ok && s.CompletedPomos > 0 {
			currentStreak++
		} else if i == 0 {
			// Today not done yet is fine if yesterday was active
			continue
		} else {
			break
		}
	}
	summary.CurrentStreak = currentStreak

	// Calculate longest streak across history
	longestStreak := 0
	tempStreak := 0
	if len(dates) > 0 {
		firstDate, _ := time.Parse("2006-01-02", dates[0])
		lastDate, _ := time.Parse("2006-01-02", dates[len(dates)-1])
		for curr := firstDate; !curr.After(lastDate); curr = curr.AddDate(0, 0, 1) {
			dStr := curr.Format("2006-01-02")
			if s, ok := merged[dStr]; ok && s.CompletedPomos > 0 {
				tempStreak++
				if tempStreak > longestStreak {
					longestStreak = tempStreak
				}
			} else {
				tempStreak = 0
			}
		}
	}
	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}
	summary.LongestStreak = longestStreak

	// Last 7 days stats
	for i := 6; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		dStr := d.Format("2006-01-02")
		stat := DayStat{Date: dStr}
		if s, ok := merged[dStr]; ok {
			stat.FocusMinutes = s.FocusMinutes
			stat.CompletedPomos = s.CompletedPomos
			stat.CompletedTasks = s.CompletedTasks
		}
		summary.RecentDays = append(summary.RecentDays, stat)
	}

	return summary
}

// ExportMarkdown generates a Markdown report formatted for Obsidian, Notion, or local storage.
func ExportMarkdown(targetPath string, summary Summary, tasks []storage.Task) error {
	var sb strings.Builder
	now := time.Now().Format("2006-01-02 15:04:05")

	sb.WriteString("# ZenPomo Productivity Report\n\n")
	sb.WriteString(fmt.Sprintf("> Generated on: `%s`\n\n", now))

	sb.WriteString("## Overview Metrics\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("| :--- | :--- |\n")
	sb.WriteString(fmt.Sprintf("| **Current Streak** | %d Days |\n", summary.CurrentStreak))
	sb.WriteString(fmt.Sprintf("| **Longest Streak** | %d Days |\n", summary.LongestStreak))
	sb.WriteString(fmt.Sprintf("| **Total Focus Time** | %d hrs %d min (%d minutes) |\n", summary.TotalMinutes/60, summary.TotalMinutes%60, summary.TotalMinutes))
	sb.WriteString(fmt.Sprintf("| **Total Completed Pomos** | %d |\n", summary.TotalPomos))
	sb.WriteString(fmt.Sprintf("| **Total Tasks Completed** | %d |\n", summary.TotalTasks))
	sb.WriteString(fmt.Sprintf("| **Daily Average Focus** | %.1f min/day |\n\n", summary.DailyAverageMinutes))

	sb.WriteString("## Past 7 Days Activity\n\n")
	sb.WriteString("| Date | Focus Time | Pomodoros | Tasks Done |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- |\n")
	for _, day := range summary.RecentDays {
		sb.WriteString(fmt.Sprintf("| %s | %d min | %d pomos | %d |\n", day.Date, day.FocusMinutes, day.CompletedPomos, day.CompletedTasks))
	}
	sb.WriteString("\n")

	sb.WriteString("## Task Queue & Completion Status\n\n")
	if len(tasks) == 0 {
		sb.WriteString("*No tasks recorded.*\n")
	} else {
		sb.WriteString("| Status | Title | Tags | Progress |\n")
		sb.WriteString("| :--- | :--- | :--- | :--- |\n")
		for _, t := range tasks {
			status := "[ ] Pending"
			if t.IsDone {
				status = "[x] Done"
			}
			tagsStr := "-"
			if len(t.Tags) > 0 {
				var formattedTags []string
				for _, tag := range t.Tags {
					formattedTags = append(formattedTags, "`#"+tag+"`")
				}
				tagsStr = strings.Join(formattedTags, " ")
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d/%d pomos |\n", status, t.Title, tagsStr, t.Completed, t.Target))
		}
	}
	sb.WriteString("\n---\n*Report exported automatically by ZenPomo TUI*\n")

	return os.WriteFile(targetPath, []byte(sb.String()), 0644)
}
