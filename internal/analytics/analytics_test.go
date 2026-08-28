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

	err = ExportMarkdown(targetFile, summary, tasks)
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
