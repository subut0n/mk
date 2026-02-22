package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempMakefile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "Makefile")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBlockComment(t *testing.T) {
	path := writeTempMakefile(t, `## Build the project
build:
	go build .
`)
	targets, err := ParseMakefile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Name != "build" {
		t.Errorf("expected name 'build', got %q", targets[0].Name)
	}
	if targets[0].Description != "Build the project" {
		t.Errorf("expected description 'Build the project', got %q", targets[0].Description)
	}
}

func TestInlineComment(t *testing.T) {
	path := writeTempMakefile(t, `test: ## Run tests
	go test ./...
`)
	targets, err := ParseMakefile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Name != "test" {
		t.Errorf("expected name 'test', got %q", targets[0].Name)
	}
	if targets[0].Description != "Run tests" {
		t.Errorf("expected description 'Run tests', got %q", targets[0].Description)
	}
}

func TestVariablesIgnored(t *testing.T) {
	path := writeTempMakefile(t, `
GO := go
VERSION ?= dev
LDFLAGS += -s
CC != which gcc

## Build
build:
	$(GO) build .
`)
	targets, err := ParseMakefile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Name != "build" {
		t.Errorf("expected name 'build', got %q", targets[0].Name)
	}
}

func TestHiddenTargetsFiltered(t *testing.T) {
	path := writeTempMakefile(t, `## Hidden
.PHONY: build
## Visible
build:
	echo ok
`)
	targets, err := ParseMakefile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Name != "build" {
		t.Errorf("expected 'build', got %q", targets[0].Name)
	}
}

func TestEmptyFile(t *testing.T) {
	path := writeTempMakefile(t, "")
	targets, err := ParseMakefile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 0 {
		t.Fatalf("expected 0 targets, got %d", len(targets))
	}
}

func TestNonexistentFile(t *testing.T) {
	_, err := ParseMakefile("/nonexistent/Makefile")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestDescriptionResetByNonCommentLine(t *testing.T) {
	path := writeTempMakefile(t, `## This description should not carry over
echo "hello"
build:
	go build .
`)
	targets, err := ParseMakefile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Description != "" {
		t.Errorf("expected empty description, got %q", targets[0].Description)
	}
}

func TestMultipleTargets(t *testing.T) {
	path := writeTempMakefile(t, `## Build
build:
	go build .

test: ## Test
	go test ./...

## Lint the code
lint:
	golangci-lint run
`)
	targets, err := ParseMakefile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	expected := []struct{ name, desc string }{
		{"build", "Build"},
		{"test", "Test"},
		{"lint", "Lint the code"},
	}
	for i, e := range expected {
		if targets[i].Name != e.name {
			t.Errorf("target %d: expected name %q, got %q", i, e.name, targets[i].Name)
		}
		if targets[i].Description != e.desc {
			t.Errorf("target %d: expected desc %q, got %q", i, e.desc, targets[i].Description)
		}
	}
}

func TestColonEqualNotTarget(t *testing.T) {
	path := writeTempMakefile(t, `
FOO := bar
BAR ::= baz

## Real target
build:
	echo build
`)
	targets, err := ParseMakefile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Name != "build" {
		t.Errorf("expected 'build', got %q", targets[0].Name)
	}
}
