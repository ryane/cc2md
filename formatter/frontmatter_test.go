package formatter

import (
	"strings"
	"testing"

	"github.com/magarcia/ccsession-viewer/parser"
)

func TestFormatFrontmatter_AllFields(t *testing.T) {
	got := FormatFrontmatter(FrontmatterFields{
		Title:     "Retire cesium atticd SQLite database",
		Date:      "2026-09-12",
		Project:   "dotfiles",
		Model:     "claude-opus-5",
		SessionID: "5c3ef91f-ed30-4413-a676-8f4961e4b459",
		Version:   "2.1.263",
	})

	want := `---
type: session
title: "Retire cesium atticd SQLite database"
date: 2026-09-12
journal: "[[2026-09-12]]"
project: dotfiles
model: claude-opus-5
session: 5c3ef91f-ed30-4413-a676-8f4961e4b459
cc_version: v2.1.263
---
`
	if got != want {
		t.Errorf("frontmatter mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatFrontmatter_EscapesTitle(t *testing.T) {
	got := FormatFrontmatter(FrontmatterFields{
		Title: `# Execute: the "hard" one \ here`,
		Date:  "2026-09-12",
	})
	if !strings.Contains(got, `title: "# Execute: the \"hard\" one \\ here"`) {
		t.Errorf("title not escaped correctly: %s", got)
	}
}

func TestFormatFrontmatter_OmitsEmptyOptionalFields(t *testing.T) {
	got := FormatFrontmatter(FrontmatterFields{
		Title: "something",
		Date:  "2026-09-12",
	})
	for _, absent := range []string{"project:", "model:", "session:", "cc_version:"} {
		if strings.Contains(got, absent) {
			t.Errorf("expected %q to be omitted when empty, got:\n%s", absent, got)
		}
	}
	// type, title, date and journal are always present.
	for _, present := range []string{"type: session", "title:", "date:", "journal:"} {
		if !strings.Contains(got, present) {
			t.Errorf("expected %q to be present, got:\n%s", present, got)
		}
	}
}

func TestFormatFrontmatter_OmitsUnknownModel(t *testing.T) {
	// parser.ExtractMetadata defaults Model to "unknown"; that is an absence,
	// not a value worth indexing.
	got := FormatFrontmatter(FrontmatterFields{
		Title: "something",
		Date:  "2026-09-12",
		Model: "unknown",
	})
	if strings.Contains(got, "model:") {
		t.Errorf("expected model to be omitted when %q, got:\n%s", "unknown", got)
	}
}

func TestFormatFrontmatter_NoDateOmitsJournal(t *testing.T) {
	got := FormatFrontmatter(FrontmatterFields{Title: "something"})
	if strings.Contains(got, "journal:") {
		t.Errorf("journal should be omitted when date is empty, got:\n%s", got)
	}
	if strings.Contains(got, "date:") {
		t.Errorf("date should be omitted when empty, got:\n%s", got)
	}
}

func TestFormatSession_IncludesFrontmatterWhenRequested(t *testing.T) {
	opts := FormatOptions{
		Flavor: FlavorObsidian,
		Frontmatter: &FrontmatterFields{
			Title: "a test session",
			Date:  "2026-09-12",
		},
	}
	got := FormatSession(parser.SessionMetadata{}, nil, opts)

	if !strings.HasPrefix(got, "---\ntype: session\n") {
		t.Errorf("expected frontmatter at the very start, got:\n%.120s", got)
	}
	if !strings.Contains(got, "\n# Session\n") {
		t.Errorf("expected the Session heading to survive, got:\n%.200s", got)
	}
}

func TestFormatSession_OmitsFrontmatterByDefault(t *testing.T) {
	got := FormatSession(parser.SessionMetadata{}, nil, FormatOptions{Flavor: FlavorObsidian})
	if strings.HasPrefix(got, "---") {
		t.Errorf("frontmatter should be opt-in, got:\n%.120s", got)
	}
}
