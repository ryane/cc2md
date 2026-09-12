package internal

import (
	"strings"
	"unicode/utf8"
)

type TruncateResult struct {
	Text       string
	Truncated  bool
	TotalLines int
}

func TruncateLines(text string, maxLines int) TruncateResult {
	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return TruncateResult{Text: text, Truncated: false, TotalLines: len(lines)}
	}
	return TruncateResult{
		Text:       strings.Join(lines[:maxLines], "\n") + "\n... (truncated)",
		Truncated:  true,
		TotalLines: len(lines),
	}
}

// TruncateString truncates text to at most maxLen bytes, appending "..." when
// it truncates. The cut never splits a multi-byte UTF-8 rune: if maxLen falls
// inside one, the whole rune is dropped. A byte-wise cut here was the source of
// invalid UTF-8 in archived transcripts.
func TruncateString(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	cut := maxLen
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return text[:cut] + "..."
}
