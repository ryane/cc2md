package formatter

import (
	"regexp"
	"strings"
)

var setextHrRegex = regexp.MustCompile(`(?m)^-{3,}$`)

func FormatAssistantText(texts []string) string {
	escaped := make([]string, len(texts))
	for i, t := range texts {
		escaped[i] = EscapeAngleBrackets(EscapeSetextHrs(t))
	}
	// Render Claude prose as plain paragraphs (NOT a blockquote). Claude
	// responses routinely contain fenced code blocks, and fenced code nested
	// inside a `>` block renders unreliably in Obsidian (the quote visually
	// breaks where the code begins). Plain prose under the "**Claude**"
	// header reads cleanly and avoids that nesting.
	return strings.Join(escaped, "\n\n")
}

// EscapeSetextHrs replaces bare --- lines with <hr> to prevent setext H2 headings.
func EscapeSetextHrs(text string) string {
	return setextHrRegex.ReplaceAllString(text, "<hr>")
}

// EscapeAngleBrackets replaces bare "<" / ">" with HTML entities so stray
// pseudo-tags in prose (e.g. <task>, <path>) are not parsed as inline HTML,
// which corrupts rendering of everything that follows. Angle brackets inside
// inline (`...`) or fenced (```...```) code spans are left untouched.
func EscapeAngleBrackets(text string) string {
	var b strings.Builder
	b.Grow(len(text))

	var inFence, inInlineCode bool
	runes := []rune(text)
	atLineStart := true

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Detect a fenced code delimiter (``` or more) at the start of a line.
		if atLineStart && r == '`' && fenceAt(runes, i) {
			j := i
			for j < len(runes) && runes[j] == '`' {
				j++
			}
			b.WriteString(string(runes[i:j]))
			inFence = !inFence
			inInlineCode = false
			i = j - 1
			atLineStart = false
			continue
		}

		switch {
		case inFence:
			// pass everything through verbatim until the closing fence
		case r == '`':
			inInlineCode = !inInlineCode
		case !inInlineCode && r == '<':
			b.WriteString("&lt;")
			atLineStart = false
			continue
		case !inInlineCode && r == '>':
			b.WriteString("&gt;")
			atLineStart = false
			continue
		}

		b.WriteRune(r)
		atLineStart = r == '\n'
	}

	return b.String()
}

// fenceAt reports whether position i begins a fenced-code delimiter of three
// or more backticks.
func fenceAt(runes []rune, i int) bool {
	n := 0
	for i < len(runes) && runes[i] == '`' {
		n++
		i++
	}
	return n >= 3
}

func FormatThinking(blocks []string, collapse bool, flavor MarkdownFlavor) string {
	if len(blocks) == 0 {
		return ""
	}

	escaped := make([]string, len(blocks))
	for i, b := range blocks {
		escaped[i] = EscapeSetextHrs(b)
	}
	combined := strings.Join(escaped, "\n\n<hr>\n\n")

	if flavor == FlavorCommonMark || !collapse {
		return "**Thinking:**\n\n" + combined
	}

	if flavor == FlavorObsidian {
		return "> [!note]- Thinking\n" + prefixLinesObsidian(combined)
	}

	return strings.Join([]string{
		"<details>",
		"<summary>Thinking</summary>",
		"",
		combined,
		"",
		"</details>",
	}, "\n")
}
