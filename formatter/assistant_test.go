package formatter

import (
	"strings"
	"testing"
)

func TestFormatAssistantText_BlockquotePrefixed(t *testing.T) {
	got := FormatAssistantText([]string{"first line\nsecond line"})
	want := "> first line\n> second line"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatAssistantText_MultipleBlocksSeparated(t *testing.T) {
	got := FormatAssistantText([]string{"block one", "block two"})
	want := "> block one\n>\n> block two"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatAssistantText_EscapesBareAngleTags(t *testing.T) {
	// A bare <task> in prose must not pass through as raw inline HTML.
	got := FormatAssistantText([]string{`the trigger ("let's work on @<task>")`})
	if strings.Contains(got, "<task>") {
		t.Errorf("expected bare <task> to be escaped, got: %s", got)
	}
	if !strings.Contains(got, "&lt;task&gt;") {
		t.Errorf("expected &lt;task&gt;, got: %s", got)
	}
}

func TestFormatAssistantText_PreservesAngleBracketsInCodeSpans(t *testing.T) {
	// Inside inline code, <path> renders fine and must be left untouched.
	got := FormatAssistantText([]string{"resolve `<path>` against root"})
	if !strings.Contains(got, "`<path>`") {
		t.Errorf("expected `<path>` preserved inside code span, got: %s", got)
	}
	if strings.Contains(got, "&lt;path&gt;") {
		t.Errorf("did not expect escaping inside code span, got: %s", got)
	}
}

func TestFormatAssistantText_PreservesAngleBracketsInFencedCode(t *testing.T) {
	in := "before\n```\n<task> not html here\n```\nafter <task> here"
	got := FormatAssistantText([]string{in})
	if !strings.Contains(got, "<task> not html here") {
		t.Errorf("expected <task> preserved inside fenced code, got: %s", got)
	}
	if !strings.Contains(got, "after &lt;task&gt; here") {
		t.Errorf("expected bare <task> outside fence to be escaped, got: %s", got)
	}
}

func TestFormatThinking_EmptyBlocks(t *testing.T) {
	got := FormatThinking([]string{}, false, FlavorGFM)
	if got != "" {
		t.Errorf("expected empty string for empty blocks, got: %q", got)
	}
}

func TestFormatThinking_GFM_NotCollapsed(t *testing.T) {
	got := FormatThinking([]string{"I think..."}, false, FlavorGFM)
	want := "**Thinking:**\n\nI think..."
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatThinking_GFM_Collapsed(t *testing.T) {
	got := FormatThinking([]string{"I think..."}, true, FlavorGFM)
	if !strings.Contains(got, "<details>") {
		t.Errorf("expected <details>, got: %s", got)
	}
	if !strings.Contains(got, "<summary>Thinking</summary>") {
		t.Errorf("expected <summary>Thinking</summary>, got: %s", got)
	}
	if !strings.Contains(got, "I think...") {
		t.Errorf("expected content, got: %s", got)
	}
	if !strings.Contains(got, "</details>") {
		t.Errorf("expected </details>, got: %s", got)
	}
}

func TestFormatThinking_GFM_MultipleBlocks(t *testing.T) {
	got := FormatThinking([]string{"block one", "block two"}, false, FlavorGFM)
	if !strings.Contains(got, "block one") {
		t.Errorf("expected block one, got: %s", got)
	}
	if !strings.Contains(got, "block two") {
		t.Errorf("expected block two, got: %s", got)
	}
	if !strings.Contains(got, "<hr>") {
		t.Errorf("expected <hr> separator, got: %s", got)
	}
}

func TestFormatThinking_GFM_EscapesSetextHrs(t *testing.T) {
	got := FormatThinking([]string{"above\n---\nbelow"}, false, FlavorGFM)
	if strings.Contains(got, "\n---\n") {
		t.Errorf("expected --- to be escaped, got: %s", got)
	}
	if !strings.Contains(got, "<hr>") {
		t.Errorf("expected <hr> in place of ---, got: %s", got)
	}
}

func TestFormatThinking_CommonMark_CollapseIgnored(t *testing.T) {
	got := FormatThinking([]string{"I think..."}, true, FlavorCommonMark)
	if strings.Contains(got, "<details>") {
		t.Errorf("CommonMark should not produce <details>, got: %s", got)
	}
	want := "**Thinking:**\n\nI think..."
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatThinking_CommonMark_NotCollapsed(t *testing.T) {
	got := FormatThinking([]string{"thought"}, false, FlavorCommonMark)
	want := "**Thinking:**\n\nthought"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatThinking_Obsidian_Collapsed(t *testing.T) {
	got := FormatThinking([]string{"I think..."}, true, FlavorObsidian)
	if strings.Contains(got, "<details>") {
		t.Errorf("Obsidian should not produce <details>, got: %s", got)
	}
	want := "> [!note]- Thinking\n> I think..."
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatThinking_Obsidian_NotCollapsed(t *testing.T) {
	got := FormatThinking([]string{"thought"}, false, FlavorObsidian)
	want := "**Thinking:**\n\nthought"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
