package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"zenpomo/internal/core"
)

// Task represents a todo item with pomodoro count targets.
type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Target    int       `json:"target"`    // target pomodoros (e.g. 3)
	Completed int       `json:"completed"` // pomodoros completed
	IsDone    bool      `json:"is_done"`
	CreatedAt time.Time `json:"created_at"`
}

// DailyStats tracks productivity metrics per day.
type DailyStats struct {
	Date           string `json:"date"` // YYYY-MM-DD
	FocusMinutes   int    `json:"focus_minutes"`
	CompletedPomos int    `json:"completed_pomos"`
	CompletedTasks int    `json:"completed_tasks"`
}

// StoreData represents the root JSON database file structure.
type StoreData struct {
	Config core.Config           `json:"config"`
	Tasks  []Task                `json:"tasks"`
	Stats  map[string]DailyStats `json:"stats"` // date -> DailyStats
}

// Store manages thread-safe JSON file persistence.
type Store struct {
	mu       sync.RWMutex
	filePath string
	data     StoreData
}

// NewStore initializes or loads the storage from the OS config directory.
func NewStore(customPath ...string) (*Store, error) {
	var path string
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			configDir = os.TempDir()
		}
		appDir := filepath.Join(configDir, "zenpomo")
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create config dir: %w", err)
		}
		path = filepath.Join(appDir, "data.json")
	}

	s := &Store{
		filePath: path,
		data: StoreData{
			Config: core.DefaultConfig(),
			Tasks:  []Task{},
			Stats:  make(map[string]DailyStats),
		},
	}

	_ = s.Reload() // if file doesn't exist yet, default is used
	return s, nil
}

func (s *Store) loadLocked() error {
	b, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var data StoreData
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	if data.Stats == nil {
		data.Stats = make(map[string]DailyStats)
	}
	s.data = data
	return nil
}

func (s *Store) saveLocked() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.filePath)
}

// Reload forces a fresh read of the JSON file from disk into memory.
func (s *Store) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

// Save writes current in-memory state to the JSON file atomically.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

// UpdateConfig updates and persists new configuration.
func (s *Store) UpdateConfig(cfg core.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.loadLocked()
	s.data.Config = cfg
	_ = s.saveLocked()
}

// GetTasks returns a copy of all tasks, with uncompleted tasks first and completed tasks last.
func (s *Store) GetTasks() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.loadLocked()
	res := make([]Task, 0, len(s.data.Tasks))
	for _, t := range s.data.Tasks {
		if !t.IsDone {
			res = append(res, t)
		}
	}
	for _, t := range s.data.Tasks {
		if t.IsDone {
			res = append(res, t)
		}
	}
	return res
}

// AddTask appends a new task to the queue.
func (s *Store) AddTask(title string, target int) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.loadLocked()
	if target <= 0 {
		target = 1
	}
	task := Task{
		ID:        fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Title:     title,
		Target:    target,
		Completed: 0,
		IsDone:    false,
		CreatedAt: time.Now(),
	}
	s.data.Tasks = append(s.data.Tasks, task)
	_ = s.saveLocked()
	return task
}

// ToggleTask flips the completed status of a task.
func (s *Store) ToggleTask(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.loadLocked()
	today := time.Now().Format("2006-01-02")
	stat := s.data.Stats[today]
	stat.Date = today

	var toggled bool
	for i := range s.data.Tasks {
		if s.data.Tasks[i].ID == id {
			s.data.Tasks[i].IsDone = !s.data.Tasks[i].IsDone
			toggled = s.data.Tasks[i].IsDone
			if toggled {
				stat.CompletedTasks++
			} else if stat.CompletedTasks > 0 {
				stat.CompletedTasks--
			}
			break
		}
	}
	s.data.Stats[today] = stat
	_ = s.saveLocked()
	return toggled
}

// DeleteTask removes a task by ID.
func (s *Store) DeleteTask(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.loadLocked()
	newList := make([]Task, 0, len(s.data.Tasks))
	for _, t := range s.data.Tasks {
		if t.ID != id {
			newList = append(newList, t)
		}
	}
	s.data.Tasks = newList
	_ = s.saveLocked()
}

// IncrementPomo increments pomo count for today and active task if any.
func (s *Store) IncrementPomo(taskTitle string, minutes int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.loadLocked()
	today := time.Now().Format("2006-01-02")
	stat := s.data.Stats[today]
	stat.Date = today
	stat.CompletedPomos++
	stat.FocusMinutes += minutes

	for i := range s.data.Tasks {
		if s.data.Tasks[i].Title == taskTitle && !s.data.Tasks[i].IsDone {
			s.data.Tasks[i].Completed++
			if s.data.Tasks[i].Completed >= s.data.Tasks[i].Target {
				s.data.Tasks[i].IsDone = true
				stat.CompletedTasks++
			}
			break
		}
	}

	s.data.Stats[today] = stat
	_ = s.saveLocked()
}

// GetTodayStats returns metrics for the current day.
func (s *Store) GetTodayStats() DailyStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.loadLocked()
	today := time.Now().Format("2006-01-02")
	if stat, ok := s.data.Stats[today]; ok {
		return stat
	}
	return DailyStats{Date: today}
}

// GetConfig returns a copy of the timer configuration.
func (s *Store) GetConfig() core.Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.loadLocked()
	return s.data.Config
}
