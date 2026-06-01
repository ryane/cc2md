package hook

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSlugifyTitle(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain", "hello world", "hello-world"},
		{"mixed punctuation", "Fix the bug — now!", "fix-the-bug-now"},
		{"already-slug", "already-a-slug", "already-a-slug"},
		{"trim hyphens", "---foo---bar---", "foo-bar"},
		{"unicode dropped", "café résumé", "caf-r-sum"},
		{"emoji dropped", "ship it 🚢🚢🚢", "ship-it"},
		{"all symbols → empty", "!@#$%^&*()", ""},
		{"cap at 50 boundary", "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda", "alpha-beta-gamma-delta-epsilon-zeta-eta-theta"},
		{"kelvin sign treated as non-ASCII", "K K K", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SlugifyTitle(tt.in)
			if got != tt.want {
				t.Errorf("SlugifyTitle(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestProjectSlug(t *testing.T) {
	tests := []struct {
		name           string
		transcriptPath string
		want           string
	}{
		{
			"encoded ~/.claude/projects path",
			"/Users/ryan/.claude/projects/-Users-ryan-Projects-cc2md/abc.jsonl",
			"cc2md",
		},
		{
			"deep encoded path",
			"/Users/ryan/.claude/projects/-Users-ryan-Projects-some-app/abc.jsonl",
			"app",
		},
		{
			"non-encoded plain dir",
			"/tmp/sandbox/abc.jsonl",
			"sandbox",
		},
		{
			"transcript at root (no parent dir name)",
			"/abc.jsonl",
			"unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProjectSlug(tt.transcriptPath)
			if got != tt.want {
				t.Errorf("ProjectSlug(%q) = %q, want %q", tt.transcriptPath, got, tt.want)
			}
		})
	}
}

func TestIDShort(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "unknown"},
		{"a", "a"},
		{"abcdef12", "abcdef12"},
		{"0b9c1f3a-7e4d-4f2a-b8c1-3d4e5f6a7b8c", "0b9c1f3a"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := IDShort(tt.in); got != tt.want {
				t.Errorf("IDShort(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBuildFilename(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	got := BuildFilename(date, "exporting-session-via-hook", "0b9c1f3a")
	want := "2026-06-01-exporting-session-via-hook-0b9c1f3a.md"
	if got != want {
		t.Errorf("with title: got %q, want %q", got, want)
	}

	got = BuildFilename(date, "", "0b9c1f3a")
	want = "2026-06-01-0b9c1f3a.md"
	if got != want {
		t.Errorf("no title: got %q, want %q", got, want)
	}
}

func TestResolveDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home dir: %v", err)
	}

	tests := []struct {
		in, want string
	}{
		{"/tmp/archive", "/tmp/archive"},
		{"~/claude-code-logs", filepath.Join(home, "claude-code-logs")},
		{"~/sub/dir", filepath.Join(home, "sub/dir")},
		{"~", home},
		{"./relative", "./relative"}, // no expansion
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := ResolveDir(tt.in); got != tt.want {
				t.Errorf("ResolveDir(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestOutputPath(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	// transcriptPath drives ProjectSlug; sessionID drives IDShort; title is
	// passed in (caller has already extracted + slugified).
	got := OutputPath(
		"/tmp/archive",
		"/Users/ryan/.claude/projects/-Users-ryan-Projects-cc2md/abc.jsonl",
		"0b9c1f3a-7e4d-4f2a-b8c1-3d4e5f6a7b8c",
		"my-session",
		date,
	)
	want := "/tmp/archive/cc2md/2026-06-01-my-session-0b9c1f3a.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
