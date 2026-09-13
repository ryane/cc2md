package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTitle_DerivesFromCommandArgs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.jsonl")
	line := `{"type":"user","message":{"content":"<command-name>/org-task</command-name>\n<command-args>Areas/dotfiles/Retire cesium atticd SQLite database.md</command-args>"}}`
	if err := os.WriteFile(path, []byte(line+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var sb strings.Builder
	if err := runTitleTo(&sb, []string{path}); err != nil {
		t.Fatalf("runTitleTo: %v", err)
	}
	got := strings.TrimRight(sb.String(), "\n")
	if got != "Retire cesium atticd SQLite database" {
		t.Errorf("got %q", got)
	}
}

func TestRunTitle_EmptyTranscriptYieldsSentinel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jsonl")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	var sb strings.Builder
	if err := runTitleTo(&sb, []string{path}); err != nil {
		t.Fatalf("runTitleTo: %v", err)
	}
	if got := strings.TrimRight(sb.String(), "\n"); got != "(no title)" {
		t.Errorf("got %q, want %q", got, "(no title)")
	}
}

func TestRunTitle_MissingFileErrors(t *testing.T) {
	var sb strings.Builder
	err := runTitleTo(&sb, []string{filepath.Join(t.TempDir(), "nope.jsonl")})
	if err == nil {
		t.Fatal("expected an error for a missing transcript, got nil")
	}
}
