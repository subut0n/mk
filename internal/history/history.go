package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Entry represents a single command history record.
type Entry struct {
	Target    string    `json:"target"`
	Directory string    `json:"directory"`
	ExecutedAt time.Time `json:"executed_at"`
}

// Manager handles persistent command history.
type Manager struct {
	filePath string
	entries  []Entry
}

// New creates a new history Manager.
func New() (*Manager, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}

	configDir := filepath.Join(dir, "mk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	m := &Manager{
		filePath: filepath.Join(configDir, "history.json"),
	}

	// Load existing history
	_ = m.load()

	return m, nil
}

// Add records a target execution in the history.
func (m *Manager) Add(target string) error {
	dir, _ := os.Getwd()
	m.entries = append([]Entry{{
		Target:     target,
		Directory:  dir,
		ExecutedAt: time.Now(),
	}}, m.entries...)

	// Keep only the 50 most recent entries
	if len(m.entries) > 50 {
		m.entries = m.entries[:50]
	}

	return m.save()
}

// Recent returns the n most recent history entries.
func (m *Manager) Recent(n int) []Entry {
	if n > len(m.entries) {
		n = len(m.entries)
	}
	return m.entries[:n]
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.entries)
}

func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, data, 0600)
}
