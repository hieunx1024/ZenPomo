package analytics

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"zenpomo/internal/storage"
)

func TestComputeSummaryAndStreak(t *testing.T) {
	now := time.Now()
	todayStr := now.Format("2006-01-02")
	yesterdayStr := now.AddDate(0, 0, -1).Format("2006-01-02")
	twoDaysAgoStr := now.AddDate(0, 0, -2).Format("2006-01-02")

	stats := map[string]storage.DailyStats{
		yesterdayStr: {
			Date:           yesterdayStr,
			FocusMinutes:   50,
			CompletedPomos: 2,
			CompletedTasks: 1,
		},
		twoDaysAgoStr: {
			Date:           twoDaysAgoStr,
			FocusMinutes:   75,
			CompletedPomos: 3,
			CompletedTasks: 2,
		},
	}

	todayStat := storage.DailyStats{
		Date:           todayStr,
		FocusMinutes:   25,
		CompletedPomos: 1,
		CompletedTasks: 1,
	}

	summary := ComputeSummary(stats, todayStat)
	if summary.TotalPomos != 6 {
		t.Errorf("TotalPomos = %d; want 6", summary.TotalPomos)
	}
	if summary.TotalMinutes != 150 {
		t.Errorf("TotalMinutes = %d; want 150", summary.TotalMinutes)
	}
	if summary.TotalTasks != 4 {
		t.Errorf("TotalTasks = %d; want 4", summary.TotalTasks)
	}
	if summary.CurrentStreak != 3 {
		t.Errorf("CurrentStreak = %d; want 3", summary.CurrentStreak)
	}
	if summary.LongestStreak < 3 {
		t.Errorf("LongestStreak = %d; want >= 3", summary.LongestStreak)
	}
	if len(summary.RecentDays) != 7 {
		t.Errorf("RecentDays length = %d; want 7", len(summary.RecentDays))
	}
}

func TestExportMarkdown(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zenpomo-export-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	targetFile := filepath.Join(tmpDir, "zenpomo-stats.md")
	summary := Summary{
		TotalPomos:          10,
		TotalMinutes:        250,
		TotalTasks:          5,
		CurrentStreak:       4,
		LongestStreak:       7,
		DailyAverageMinutes: 50.0,
		RecentDays: []DayStat{
			{Date: "2026-08-28", FocusMinutes: 50, CompletedPomos: 2, CompletedTasks: 1},
		},
	}
	tasks := []storage.Task{
		{ID: "1", Title: "Build TUI", Tags: []string{"frontend", "tui"}, Target: 3, Completed: 3, IsDone: true},
		{ID: "2", Title: "Setup IPC", Tags: []string{"backend"}, Target: 2, Completed: 1, IsDone: false},
	}

	dailyStats := map[string]storage.DailyStats{
		"2026-08-28": {Date: "2026-08-28", FocusMinutes: 50, CompletedPomos: 2, CompletedTasks: 1},
	}
	err = ExportMarkdown(targetFile, summary, tasks, dailyStats, storage.DailyStats{})
	if err != nil {
		t.Fatalf("ExportMarkdown failed: %v", err)
	}

	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read exported markdown: %v", err)
	}
	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Fatalf("Exported markdown file is empty")
	}
	if !testing.Short() {
		if !contains(contentStr, "ZenPomo Productivity Report") || !contains(contentStr, "Build TUI") {
			t.Errorf("Exported file missing expected contents: %s", contentStr)
		}
	}
}

func TestComputeWeeklyStats(t *testing.T) {
	now := time.Now()
	todayStr := now.Format("2006-01-02")
	yesterdayStr := now.AddDate(0, 0, -1).Format("2006-01-02")
	lastWeekStr := now.AddDate(0, 0, -8).Format("2006-01-02")

	stats := map[string]storage.DailyStats{
		yesterdayStr: {Date: yesterdayStr, FocusMinutes: 25, CompletedPomos: 1, CompletedTasks: 1},
		lastWeekStr:  {Date: lastWeekStr, FocusMinutes: 100, CompletedPomos: 4, CompletedTasks: 2},
	}
	todayStat := storage.DailyStats{Date: todayStr, FocusMinutes: 25, CompletedPomos: 1}

	weeks := ComputeWeeklyStats(stats, todayStat, 4)
	if len(weeks) != 4 {
		t.Fatalf("len(weeks) = %d; want 4", len(weeks))
	}

	currentWeek := weeks[len(weeks)-1]
	if currentWeek.FocusMinutes != 50 {
		t.Errorf("current week FocusMinutes = %d; want 50", currentWeek.FocusMinutes)
	}
	if currentWeek.CompletedPomos != 2 {
		t.Errorf("current week CompletedPomos = %d; want 2", currentWeek.CompletedPomos)
	}

	totalMinutes := 0
	for _, w := range weeks {
		totalMinutes += w.FocusMinutes
	}
	if totalMinutes != 150 {
		t.Errorf("total minutes across weeks = %d; want 150", totalMinutes)
	}
}

func TestComputeMonthlyStats(t *testing.T) {
	now := time.Now()
	todayStr := now.Format("2006-01-02")
	todayStat := storage.DailyStats{Date: todayStr, FocusMinutes: 30, CompletedPomos: 1, CompletedTasks: 1}

	months := ComputeMonthlyStats(nil, todayStat, 3)
	if len(months) != 3 {
		t.Fatalf("len(months) = %d; want 3", len(months))
	}

	currentMonth := months[len(months)-1]
	wantLabel := now.Format("2006-01")
	if currentMonth.Label != wantLabel {
		t.Errorf("current month Label = %q; want %q", currentMonth.Label, wantLabel)
	}
	if currentMonth.FocusMinutes != 30 {
		t.Errorf("current month FocusMinutes = %d; want 30", currentMonth.FocusMinutes)
	}
}

func TestExportCSV(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zenpomo-csv-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	targetFile := filepath.Join(tmpDir, "zenpomo-stats.csv")
	stats := map[string]storage.DailyStats{
		"2026-08-28": {Date: "2026-08-28", FocusMinutes: 50, CompletedPomos: 2, CompletedTasks: 1},
	}
	todayStat := storage.DailyStats{Date: "2026-08-29", FocusMinutes: 25, CompletedPomos: 1, CompletedTasks: 0}

	if err := ExportCSV(targetFile, stats, todayStat); err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read exported CSV: %v", err)
	}
	contentStr := string(content)
	if !contains(contentStr, "date,focus_minutes,completed_pomos,completed_tasks") {
		t.Errorf("CSV missing header: %s", contentStr)
	}
	if !contains(contentStr, "2026-08-28,50,2,1") || !contains(contentStr, "2026-08-29,25,1,0") {
		t.Errorf("CSV missing expected rows: %s", contentStr)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && len(substr) > 0 && searchSubstr(s, substr)))
}

func searchSubstr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
