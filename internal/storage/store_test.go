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

func TestStore_MultiInstanceSync(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zenpomo-sync-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "data.json")
	storeA, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to init storeA: %v", err)
	}

	storeB, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to init storeB: %v", err)
	}

	// Instance A adds a task
	task := storeA.AddTask("Synced Task", 4)

	// Instance B increments pomo
	storeB.IncrementPomo(task.Title, 25)

	// Instance A checks today's stats - should reflect Instance B's change immediately
	statsA := storeA.GetTodayStats()
	if statsA.CompletedPomos != 1 || statsA.FocusMinutes != 25 {
		t.Fatalf("expected storeA to see stats from storeB, got %+v", statsA)
	}

	// Instance A adds another task without overwriting Instance B's stats
	storeA.AddTask("Another Task", 1)

	// Instance B checks stats again - should NOT have been wiped out by storeA's AddTask
	statsB := storeB.GetTodayStats()
	if statsB.CompletedPomos != 1 || statsB.FocusMinutes != 25 {
		t.Fatalf("expected storeB stats preserved after storeA AddTask, got %+v", statsB)
	}
}

func TestStore_TaskSorting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zenpomo-sort-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "data.json")
	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}

	task1 := store.AddTask("Task 1 (Done)", 1)
	task2 := store.AddTask("Task 2 (Pending)", 2)
	task3 := store.AddTask("Task 3 (Pending)", 3)

	// Mark Task 1 as Done
	store.ToggleTask(task1.ID)

	tasks := store.GetTasks()
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	// Task 2 and Task 3 should come first (uncompleted)
	if tasks[0].ID != task2.ID || tasks[1].ID != task3.ID {
		t.Fatalf("expected uncompleted tasks first, got: %+v", tasks)
	}

	// Task 1 should come last (completed)
	if tasks[2].ID != task1.ID || !tasks[2].IsDone {
		t.Fatalf("expected completed task at the end, got: %+v", tasks[2])
	}
}
