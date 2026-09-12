package formatter

import "strings"

// FrontmatterFields carries the values rendered into a session note's YAML
// frontmatter. Title and Date drive the always-present keys; the rest are
// omitted when empty rather than written blank.
type FrontmatterFields struct {
	Title     string
	Date      string // YYYY-MM-DD
	Project   string
	Model     string
	SessionID string
	Version   string
}

// FormatFrontmatter renders an Obsidian-compatible YAML frontmatter block,
// terminated by a newline. It is hand-encoded rather than marshalled so the
// module keeps a zero-dependency footprint: adding a YAML library would force a
// vendorHash update in the consuming Nix flake.
//
// Title is always double-quoted — real first messages begin with '#' or contain
// ':', both of which break unquoted YAML. Date is emitted bare so Obsidian
// Bases parses it as a date and can sort on it; journal repeats it as a
// wikilink for navigation to the daily note.
func FormatFrontmatter(f FrontmatterFields) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("type: session\n")
	b.WriteString("title: " + quoteYAML(f.Title) + "\n")
	if f.Date != "" {
		b.WriteString("date: " + f.Date + "\n")
		b.WriteString("journal: " + quoteYAML("[["+f.Date+"]]") + "\n")
	}
	writeIfSet(&b, "project", f.Project)
	// parser.ExtractMetadata defaults Model to "unknown", which is a value, not
	// an absence — omit it so Bases sees a missing property rather than a
	// literal "unknown" to group and filter on.
	if f.Model != "" && f.Model != "unknown" {
		b.WriteString("model: " + f.Model + "\n")
	}
	writeIfSet(&b, "session", f.SessionID)
	if f.Version != "" {
		b.WriteString("cc_version: v" + f.Version + "\n")
	}
	b.WriteString("---\n")
	return b.String()
}

func writeIfSet(b *strings.Builder, key, value string) {
	if value != "" {
		b.WriteString(key + ": " + value + "\n")
	}
}

// quoteYAML wraps s in double quotes, escaping backslashes and double quotes.
func quoteYAML(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
