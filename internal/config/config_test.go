package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/subut0n/mk/internal/i18n"
)

func TestDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Config.KeyScheme != KeySchemeArrows {
		t.Errorf("expected KeyScheme=%q, got %q", KeySchemeArrows, cfg.Config.KeyScheme)
	}
	if cfg.Config.Language != i18n.LangEN {
		t.Errorf("expected Language=%q, got %q", i18n.LangEN, cfg.Config.Language)
	}
	if cfg.Config.ColorScheme != ColorSchemeRainbow {
		t.Errorf("expected ColorScheme=%q, got %q", ColorSchemeRainbow, cfg.Config.ColorScheme)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := New()
	if err != nil {
		t.Fatal(err)
	}

	cfg.Config.KeyScheme = KeySchemeWASD
	cfg.Config.Language = i18n.LangFR
	cfg.Config.ColorScheme = ColorSchemeDeuteranopia
	cfg.Config.CustomUpKey = 'z'
	cfg.Config.CustomDownKey = 's'

	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	// Load into a new manager
	cfg2, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if cfg2.Config.KeyScheme != KeySchemeWASD {
		t.Errorf("expected KeyScheme=%q, got %q", KeySchemeWASD, cfg2.Config.KeyScheme)
	}
	if cfg2.Config.Language != i18n.LangFR {
		t.Errorf("expected Language=%q, got %q", i18n.LangFR, cfg2.Config.Language)
	}
	if cfg2.Config.ColorScheme != ColorSchemeDeuteranopia {
		t.Errorf("expected ColorScheme=%q, got %q", ColorSchemeDeuteranopia, cfg2.Config.ColorScheme)
	}
	if cfg2.Config.CustomUpKey != 'z' {
		t.Errorf("expected CustomUpKey='z', got %q", cfg2.Config.CustomUpKey)
	}
	if cfg2.Config.CustomDownKey != 's' {
		t.Errorf("expected CustomDownKey='s', got %q", cfg2.Config.CustomDownKey)
	}
}

func TestMigrationZQSD(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Create config dir and write legacy config with "zqsd" scheme
	configDir := filepath.Join(dir, "mk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	legacy := Config{
		KeyScheme:   "zqsd",
		Language:    i18n.LangFR,
		ColorScheme: ColorSchemeRainbow,
	}
	data, _ := json.MarshalIndent(legacy, "", "  ")
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Config.KeyScheme != KeySchemeCustom {
		t.Errorf("expected migration to 'custom', got %q", cfg.Config.KeyScheme)
	}
	if cfg.Config.CustomUpKey != 'z' {
		t.Errorf("expected CustomUpKey='z', got %d", cfg.Config.CustomUpKey)
	}
	if cfg.Config.CustomDownKey != 's' {
		t.Errorf("expected CustomDownKey='s', got %d", cfg.Config.CustomDownKey)
	}
}

func TestEmptyFieldsGetDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Create config with empty fields
	configDir := filepath.Join(dir, "mk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write minimal JSON with empty values
	data := []byte(`{"key_scheme":"arrows","language":"","color_scheme":""}`)
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Config.Language != i18n.LangEN {
		t.Errorf("expected Language default 'en', got %q", cfg.Config.Language)
	}
	if cfg.Config.ColorScheme != ColorSchemeRainbow {
		t.Errorf("expected ColorScheme default 'rainbow', got %q", cfg.Config.ColorScheme)
	}
}

func TestExists(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Exists() {
		t.Error("config should not exist before Save")
	}

	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	if !cfg.Exists() {
		t.Error("config should exist after Save")
	}
}
