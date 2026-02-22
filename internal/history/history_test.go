package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupTestHistory(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	configDir := filepath.Join(dir, "mk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	return &Manager{
		filePath: filepath.Join(configDir, "history.json"),
	}
}

func TestAddAndRecent(t *testing.T) {
	m := setupTestHistory(t)

	if err := m.Add("build"); err != nil {
		t.Fatal(err)
	}
	if err := m.Add("test"); err != nil {
		t.Fatal(err)
	}

	entries := m.Recent(10)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Most recent first
	if entries[0].Target != "test" {
		t.Errorf("expected first entry 'test', got %q", entries[0].Target)
	}
	if entries[1].Target != "build" {
		t.Errorf("expected second entry 'build', got %q", entries[1].Target)
	}
}

func TestCap50(t *testing.T) {
	m := setupTestHistory(t)

	for i := 0; i < 60; i++ {
		if err := m.Add("target"); err != nil {
			t.Fatal(err)
		}
	}

	entries := m.Recent(100)
	if len(entries) != 50 {
		t.Fatalf("expected cap of 50 entries, got %d", len(entries))
	}
}

func TestRecentOrder(t *testing.T) {
	m := setupTestHistory(t)

	targets := []string{"first", "second", "third"}
	for _, tgt := range targets {
		if err := m.Add(tgt); err != nil {
			t.Fatal(err)
		}
	}

	entries := m.Recent(3)
	if entries[0].Target != "third" {
		t.Errorf("expected 'third' first, got %q", entries[0].Target)
	}
	if entries[2].Target != "first" {
		t.Errorf("expected 'first' last, got %q", entries[2].Target)
	}
}

func TestCorruptedFile(t *testing.T) {
	m := setupTestHistory(t)

	// Write corrupted JSON
	if err := os.WriteFile(m.filePath, []byte("{invalid json"), 0600); err != nil {
		t.Fatal(err)
	}

	// load should return an error but not panic
	err := m.load()
	if err == nil {
		t.Error("expected error for corrupted JSON")
	}

	// Should still be able to add new entries
	if err := m.Add("build"); err != nil {
		t.Fatal(err)
	}
	entries := m.Recent(10)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after corrupted file, got %d", len(entries))
	}
}

func TestEmptyHistory(t *testing.T) {
	m := setupTestHistory(t)

	entries := m.Recent(10)
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestPersistence(t *testing.T) {
	m := setupTestHistory(t)

	if err := m.Add("build"); err != nil {
		t.Fatal(err)
	}

	// Read back from disk
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		t.Fatal(err)
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 persisted entry, got %d", len(entries))
	}
	if entries[0].Target != "build" {
		t.Errorf("expected 'build', got %q", entries[0].Target)
	}
}
