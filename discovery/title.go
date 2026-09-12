package discovery

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// maxTitleLen caps a derived title. Titles live in YAML frontmatter, not in a
// filename, so this is a readability limit rather than a filesystem one.
const maxTitleLen = 80

// noTitle is the sentinel used when nothing usable can be derived.
const noTitle = "(no title)"

var (
	reTitleCommandArgs = regexp.MustCompile(`(?s)<command-args>(.*?)</command-args>`)
	// A markdown note path. Vault paths contain spaces ("Areas/dotfiles/Retire
	// cesium atticd SQLite database.md"), so this anchors on the .md suffix
	// rather than on non-whitespace runs — an `\S`-based pattern matches none
	// of the real paths in this vault.
	reTitleNotePath = regexp.MustCompile(`^[^\n]+\.md$`)
)

// DeriveTitle reads a session's first user message and derives a human-readable
// title. See DeriveTitleFromText for the rules.
func DeriveTitle(transcriptPath string) string {
	return DeriveTitleFromText(ExtractFirstUserMessageRaw(transcriptPath))
}

// DeriveTitleFromText derives a human-readable title from a raw first user
// message. Rules produce a candidate; an empty candidate falls through:
//
//  1. Slash-command — keep the <command-args> text as the candidate. A command
//     with no args yields an empty candidate and falls through.
//  2. Vault-note path — if the candidate looks like a path to a .md note, use
//     its basename without the extension. This is a pattern match, not a vault
//     existence check, so it works for notes since renamed or deleted.
//  3. Fallback — the message itself, tags stripped and whitespace collapsed.
//
// The result reproduces what was typed, typos included; rule 2 is not a
// spell-checker. Returns noTitle when nothing usable remains.
func DeriveTitleFromText(raw string) string {
	candidate := ""
	if m := reTitleCommandArgs.FindStringSubmatch(raw); m != nil {
		candidate = strings.TrimSpace(m[1])
	}
	if candidate == "" {
		// Reuse the package's existing tag stripper rather than a bare
		// `<[^>]*>` strip: stripXMLTags removes <command-message> *with its
		// content*, which a tag-only strip would leave behind as a duplicate
		// of the command name. Real openers carry both tags.
		candidate = strings.TrimSpace(stripXMLTags(raw))
	}
	candidate = strings.Join(strings.Fields(candidate), " ")
	if candidate == "" {
		return noTitle
	}

	if reTitleNotePath.MatchString(candidate) {
		base := filepath.Base(candidate)
		candidate = strings.TrimSuffix(base, filepath.Ext(base))
	}

	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return noTitle
	}
	return truncateTitle(candidate)
}

// truncateTitle caps s at maxTitleLen bytes, backing the cut up to the last
// space so a word is never split. Falls back to a rune-safe cut when the text
// has no space within the limit (a single very long token).
func truncateTitle(s string) string {
	if len(s) <= maxTitleLen {
		return s
	}
	cut := maxTitleLen
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	if idx := strings.LastIndexByte(s[:cut], ' '); idx > 0 {
		cut = idx
	}
	return strings.TrimRight(s[:cut], " ") + "..."
}
