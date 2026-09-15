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

// PeriodStat holds aggregated focus metrics for a week or month bucket.
type PeriodStat struct {
	Label          string // e.g. "09/08-09/14" for a week, "2026-09" for a month
	FocusMinutes   int
	CompletedPomos int
	CompletedTasks int
}

func mergeStats(stats map[string]storage.DailyStats, todayStats storage.DailyStats) map[string]storage.DailyStats {
	merged := make(map[string]storage.DailyStats, len(stats)+1)
	for k, v := range stats {
		merged[k] = v
	}
	if todayStats.Date != "" {
		merged[todayStats.Date] = todayStats
	}
	return merged
}

// weekStartMonday returns the Monday 00:00 of the week containing t.
func weekStartMonday(t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7 // treat Sunday as day 7 so weeks start on Monday
	}
	return t.AddDate(0, 0, -(wd - 1))
}

// ComputeWeeklyStats aggregates daily stats into Monday-start week buckets,
// returning the most recent `weeks` weeks in chronological order.
func ComputeWeeklyStats(stats map[string]storage.DailyStats, todayStats storage.DailyStats, weeks int) []PeriodStat {
	merged := mergeStats(stats, todayStats)
	now := time.Now()
	currentWeekStart := weekStartMonday(now)

	result := make([]PeriodStat, 0, weeks)
	for i := weeks - 1; i >= 0; i-- {
		weekStart := currentWeekStart.AddDate(0, 0, -7*i)
		weekEnd := weekStart.AddDate(0, 0, 6)

		p := PeriodStat{Label: fmt.Sprintf("%s-%s", weekStart.Format("01/02"), weekEnd.Format("01/02"))}
		for d := weekStart; !d.After(weekEnd); d = d.AddDate(0, 0, 1) {
			if s, ok := merged[d.Format("2006-01-02")]; ok {
				p.FocusMinutes += s.FocusMinutes
				p.CompletedPomos += s.CompletedPomos
				p.CompletedTasks += s.CompletedTasks
			}
		}
		result = append(result, p)
	}
	return result
}

// ComputeMonthlyStats aggregates daily stats into calendar-month buckets,
// returning the most recent `months` months in chronological order.
func ComputeMonthlyStats(stats map[string]storage.DailyStats, todayStats storage.DailyStats, months int) []PeriodStat {
	merged := mergeStats(stats, todayStats)
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	result := make([]PeriodStat, 0, months)
	for i := months - 1; i >= 0; i-- {
		monthStart := currentMonth.AddDate(0, -i, 0)
		p := PeriodStat{Label: monthStart.Format("2006-01")}
		for dStr, s := range merged {
			d, err := time.Parse("2006-01-02", dStr)
			if err != nil {
				continue
			}
			if d.Year() == monthStart.Year() && d.Month() == monthStart.Month() {
				p.FocusMinutes += s.FocusMinutes
				p.CompletedPomos += s.CompletedPomos
				p.CompletedTasks += s.CompletedTasks
			}
		}
		result = append(result, p)
	}
	return result
}

// ExportCSV writes daily stats history as a spreadsheet-friendly CSV file.
func ExportCSV(targetPath string, stats map[string]storage.DailyStats, todayStats storage.DailyStats) error {
	merged := mergeStats(stats, todayStats)

	dates := make([]string, 0, len(merged))
	for d := range merged {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	var sb strings.Builder
	sb.WriteString("date,focus_minutes,completed_pomos,completed_tasks\n")
	for _, d := range dates {
		s := merged[d]
		sb.WriteString(fmt.Sprintf("%s,%d,%d,%d\n", d, s.FocusMinutes, s.CompletedPomos, s.CompletedTasks))
	}

	return os.WriteFile(targetPath, []byte(sb.String()), 0644)
}

// ExportMarkdown generates a Markdown report formatted for Obsidian, Notion, or local storage.
// dailyStats/todayStats are used to additionally break down focus time by week and month.
func ExportMarkdown(targetPath string, summary Summary, tasks []storage.Task, dailyStats map[string]storage.DailyStats, todayStats storage.DailyStats) error {
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

	sb.WriteString("## Weekly Activity (Last 8 Weeks)\n\n")
	sb.WriteString("| Week | Focus Time | Pomodoros | Tasks Done |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- |\n")
	for _, wk := range ComputeWeeklyStats(dailyStats, todayStats, 8) {
		sb.WriteString(fmt.Sprintf("| %s | %d min | %d pomos | %d |\n", wk.Label, wk.FocusMinutes, wk.CompletedPomos, wk.CompletedTasks))
	}
	sb.WriteString("\n")

	sb.WriteString("## Monthly Activity (Last 6 Months)\n\n")
	sb.WriteString("| Month | Focus Time | Pomodoros | Tasks Done |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- |\n")
	for _, mo := range ComputeMonthlyStats(dailyStats, todayStats, 6) {
		sb.WriteString(fmt.Sprintf("| %s | %d min | %d pomos | %d |\n", mo.Label, mo.FocusMinutes, mo.CompletedPomos, mo.CompletedTasks))
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
