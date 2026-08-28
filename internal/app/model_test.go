package app

import (
	"testing"
	"zenpomo/internal/theme"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelInitAndTabSwitch(t *testing.T) {
	m := NewModel()
	if m.activeTab != TabTimer {
		t.Errorf("Initial activeTab = %v; want %v (TabTimer)", m.activeTab, TabTimer)
	}

	// Switch to TabTasks via key "2"
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model := newM.(Model)
	if model.activeTab != TabTasks {
		t.Errorf("activeTab after '2' = %v; want %v (TabTasks)", model.activeTab, TabTasks)
	}

	// Switch to TabStats via key "3"
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model = newM.(Model)
	if model.activeTab != TabStats {
		t.Errorf("activeTab after '3' = %v; want %v (TabStats)", model.activeTab, TabStats)
	}

	// Switch to TabSettings via key "4"
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model = newM.(Model)
	if model.activeTab != TabSettings {
		t.Errorf("activeTab after '4' = %v; want %v (TabSettings)", model.activeTab, TabSettings)
	}

	// Cycle forward via Tab key
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = newM.(Model)
	if model.activeTab != TabTimer {
		t.Errorf("activeTab after 'tab' from 4 = %v; want %v (TabTimer)", model.activeTab, TabTimer)
	}
}

func TestZenModeToggle(t *testing.T) {
	m := NewModel()
	if m.zenMode {
		t.Errorf("Initial zenMode should be false")
	}

	// Press 'z' on TabTimer -> Zen mode on
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	model := newM.(Model)
	if !model.zenMode {
		t.Errorf("zenMode after 'z' = false; want true")
	}

	// Press 'z' again -> Zen mode off
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	model = newM.(Model)
	if model.zenMode {
		t.Errorf("zenMode after second 'z' = true; want false")
	}
}

func TestThemeCycling(t *testing.T) {
	m := NewModel()
	initialTheme := m.config.Theme

	// Press 't' to cycle theme
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model := newM.(Model)
	if model.config.Theme == initialTheme {
		t.Errorf("Theme should have cycled after 't'; remained %s", initialTheme)
	}
	expectedTheme := theme.NextTheme(initialTheme)
	if model.config.Theme != expectedTheme {
		t.Errorf("Theme after 't' = %s; want %s", model.config.Theme, expectedTheme)
	}
}

func TestTasksTabOperations(t *testing.T) {
	m := NewModel()
	// Switch to TabTasks
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model := newM.(Model)

	// Press 'a' to enter adding mode
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = newM.(Model)
	if model.inputMode != ModeAddingTask {
		t.Errorf("inputMode after 'a' = %v; want ModeAddingTask", model.inputMode)
	}

	// Type task title with tags and estimation
	model.textInput.SetValue("Refactor Auth #security est:3")
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = newM.(Model)

	if model.inputMode != ModeNormal {
		t.Errorf("inputMode after Enter = %v; want ModeNormal", model.inputMode)
	}

	// Add second task
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = newM.(Model)
	model.textInput.SetValue("Write Docs #docs est:2")
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = newM.(Model)

	if len(model.tasks) < 2 {
		t.Fatalf("expected at least 2 tasks, got %d", len(model.tasks))
	}

	// Select first task and edit it ('e')
	model.selectedTask = 0
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = newM.(Model)
	if model.inputMode != ModeEditingTask {
		t.Errorf("inputMode after 'e' = %v; want ModeEditingTask", model.inputMode)
	}

	model.textInput.SetValue("Refactor Auth Module #security #core est:4")
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = newM.(Model)

	if model.tasks[0].Title != "Refactor Auth Module" || model.tasks[0].Target != 4 {
		t.Errorf("Edited task mismatch: %+v", model.tasks[0])
	}

	// Reorder tasks down using 'J'
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	model = newM.(Model)
	if model.selectedTask != 1 {
		t.Errorf("selectedTask after 'J' = %d; want 1", model.selectedTask)
	}

	// Reorder back up using 'K'
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	model = newM.(Model)
	if model.selectedTask != 0 {
		t.Errorf("selectedTask after 'K' = %d; want 0", model.selectedTask)
	}
}

func TestAnalyticsTabExport(t *testing.T) {
	m := NewModel()
	// Switch to TabStats ('3')
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model := newM.(Model)
	if model.activeTab != TabStats {
		t.Fatalf("activeTab after '3' = %v; want TabStats", model.activeTab)
	}

	// Press 'x' to export markdown report
	newM, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model = newM.(Model)

	if model.statusMessage == "" {
		t.Errorf("Expected statusMessage after export 'x', got empty")
	}
}


