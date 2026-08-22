package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_CRUD(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zenpomo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "data.json")
	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}

	task := store.AddTask("Implement TUI", 2)
	if task.Title != "Implement TUI" || task.Target != 2 {
		t.Fatalf("unexpected task data: %+v", task)
	}

	tasks := store.GetTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	toggled := store.ToggleTask(task.ID)
	if !toggled {
		t.Fatalf("expected task to be marked done")
	}

	store.IncrementPomo("Implement TUI", 25)
	stats := store.GetTodayStats()
	if stats.CompletedPomos != 1 || stats.FocusMinutes != 25 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	store.DeleteTask(task.ID)
	if len(store.GetTasks()) != 0 {
		t.Fatalf("expected 0 tasks after delete")
	}
}
