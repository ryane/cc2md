package discovery

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/magarcia/ccsession-viewer/internal"
)

// SessionEntry represents a single Claude Code session log file.
type SessionEntry struct {
	Path       string
	SessionID  string
	Project    string
	ModifiedAt time.Time
	Size       int64
	Name       string
}

// Compiled regexes for stripping XML tags from user messages.
var (
	// Tags where we strip the tag AND its content.
	reStripWithContent = regexp.MustCompile(
		`(?s)<(?:system-reminder|command-name|command-message|local-command-caveat|local-command-stdout|local-command-stderr|task-notification)[^>]*>.*?</(?:system-reminder|command-name|command-message|local-command-caveat|local-command-stdout|local-command-stderr|task-notification)>`,
	)
	reStripTeammate = regexp.MustCompile(`(?s)<teammate-message[^>]*>.*?</teammate-message>`)
	// Tags where we strip only the tags, keeping inner text.
	reStripTagOnly = regexp.MustCompile(`</?command-args[^>]*>`)
)

// ListSessions scans the Claude projects directory for .jsonl session files.
// The directory is $CLAUDE_CONFIG_DIR/projects/ if set, else ~/.claude/projects/.
// If projectFilter is non-empty, only projects whose decoded or raw name
// contains the filter (case-insensitive) are included.
func ListSessions(projectFilter string) []SessionEntry {
	return listSessionsIn(defaultProjectsDir(), defaultSessionNamesDir(), projectFilter)
}

// ProjectsDir returns the path that ListSessions scans for session files.
// Honors $CLAUDE_CONFIG_DIR; falls back to ~/.claude/projects.
func ProjectsDir() string {
	return defaultProjectsDir()
}

func listSessionsIn(projectsDir, sessionNamesDir, projectFilter string) []SessionEntry {
	dirEntries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil
	}

	filterLower := strings.ToLower(strings.TrimSpace(projectFilter))

	var entries []SessionEntry
	for _, d := range dirEntries {
		if !d.IsDir() {
			continue
		}
		name := d.Name()

		if filterLower != "" {
			decoded := DecodeProjectName(name)
			if !strings.Contains(strings.ToLower(decoded), filterLower) &&
				!strings.Contains(strings.ToLower(name), filterLower) {
				continue
			}
		}

		projectPath := filepath.Join(projectsDir, name)
		files, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}
			filePath := filepath.Join(projectPath, f.Name())
			info, err := f.Info()
			if err != nil {
				continue
			}
			sessionID := strings.TrimSuffix(f.Name(), ".jsonl")

			sessionName := readSessionName(sessionNamesDir, sessionID)
			if sessionName == "" {
				sessionName = ExtractFirstUserMessage(filePath, 60)
			}
			if sessionName == "" {
				sessionName = "(no title)"
			}

			entries = append(entries, SessionEntry{
				Path:       filePath,
				SessionID:  sessionID,
				Project:    DecodeProjectName(name),
				ModifiedAt: info.ModTime(),
				Size:       info.Size(),
				Name:       sessionName,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ModifiedAt.After(entries[j].ModifiedAt)
	})

	return entries
}

// readSessionName reads a session name from ~/.claude/session-names/<id>.name.
func readSessionName(sessionNamesDir, sessionID string) string {
	if sessionNamesDir == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(sessionNamesDir, sessionID+".name"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// ExtractFirstUserMessage reads the first user message from a JSONL file
// and returns a cleaned, truncated version suitable as a session name.
//
// Unlike ExtractFirstUserMessageRaw, this skips messages that are empty *after*
// tag stripping (e.g. an opener consisting solely of
// <local-command-caveat>…</local-command-caveat>) and continues to the next
// candidate, preserving the original single-pass behavior.
func ExtractFirstUserMessage(filePath string, maxLen int) string {
	for _, raw := range extractUserMessages(filePath) {
		text := stripXMLTags(raw)
		text = strings.Join(strings.Fields(text), " ")
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		return internal.TruncateString(text, maxLen)
	}
	return ""
}

// ExtractFirstUserMessageRaw returns the first user message that carries usable
// title text, verbatim — XML tags intact, untruncated. Title derivation needs
// this because <command-args> carries the topic for slash-command sessions, and
// stripXMLTags discards it.
//
// Messages that contribute nothing are skipped, mirroring ExtractFirstUserMessage:
// sessions frequently open with a <local-command-caveat> block that strips to
// nothing, and returning it verbatim would shadow the real opener on the next
// line. A message counts as usable if it has non-empty <command-args> or any
// text surviving stripXMLTags. When nothing qualifies, the first message is
// returned so a caveat-only session still derives its sentinel.
func ExtractFirstUserMessageRaw(filePath string) string {
	msgs := extractUserMessages(filePath)
	if len(msgs) == 0 {
		return ""
	}
	for _, raw := range msgs {
		if m := reTitleCommandArgs.FindStringSubmatch(raw); m != nil && strings.TrimSpace(m[1]) != "" {
			return raw
		}
		if strings.TrimSpace(stripXMLTags(raw)) != "" {
			return raw
		}
	}
	return msgs[0]
}

// extractUserMessages returns the raw text of each user message in the first 20
// lines of a JSONL file, in order, skipping entries that are blank before any
// cleaning. Callers apply their own filtering.
func extractUserMessages(filePath string) []string {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	var out []string
	// Only scan the first 20 lines to keep it fast.
	for i := 0; i < 20 && scanner.Scan(); i++ {
		line := scanner.Bytes()

		// Quick check before full parse.
		if !strings.Contains(string(line), `"user"`) {
			continue
		}

		var jl struct {
			Type    string          `json:"type"`
			Message json.RawMessage `json:"message"`
		}
		if err := json.Unmarshal(line, &jl); err != nil {
			continue
		}
		if jl.Type != "user" {
			continue
		}

		var msg struct {
			Content json.RawMessage `json:"content"`
		}
		if err := json.Unmarshal(jl.Message, &msg); err != nil {
			continue
		}

		// Only handle string content, skip arrays.
		if len(msg.Content) == 0 || msg.Content[0] != '"' {
			continue
		}
		var text string
		if err := json.Unmarshal(msg.Content, &text); err != nil {
			continue
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		out = append(out, text)
	}
	return out
}

// stripXMLTags removes known XML tags from user message content.
func stripXMLTags(s string) string {
	s = reStripWithContent.ReplaceAllString(s, "")
	s = reStripTeammate.ReplaceAllString(s, "")
	s = reStripTagOnly.ReplaceAllString(s, "")
	return s
}

// FormatSessionList returns a human-readable listing of sessions.
func FormatSessionList(sessions []SessionEntry) string {
	if len(sessions) == 0 {
		return "No sessions found."
	}

	// Compute max name width, capped at 40.
	maxNameLen := 0
	for _, s := range sessions {
		if l := len(s.Name); l > maxNameLen {
			maxNameLen = l
		}
	}
	if maxNameLen > 40 {
		maxNameLen = 40
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d session(s):\n\n", len(sessions))

	for i, s := range sessions {
		date := s.ModifiedAt.Format("2006-01-02 15:04")
		sizeKB := (s.Size + 512) / 1024
		name := s.Name
		if len(name) > 40 {
			name = internal.TruncateString(name, 37)
		}
		shortID := s.SessionID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		fmt.Fprintf(&b, "  %s  %-*s  %5dKB  %s  %s",
			date, maxNameLen, name, sizeKB, s.Project, shortID)
		if i < len(sessions)-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

// DecodeProjectName converts a Claude project directory name (e.g.
// "-Users-ryan-Projects-cc2md") to a human-readable path (e.g.
// "/Users/ryan/Projects/cc2md"). Names that do not start with "-" are
// returned unchanged.
func DecodeProjectName(encoded string) string {
	if strings.HasPrefix(encoded, "-") {
		return strings.ReplaceAll(encoded, "-", "/")
	}
	return encoded
}

// claudeConfigDir returns the Claude config directory, honoring the
// CLAUDE_CONFIG_DIR environment variable when set and falling back to
// ~/.claude otherwise. Returns "" if no directory can be determined.
func claudeConfigDir() string {
	if dir := strings.TrimSpace(os.Getenv("CLAUDE_CONFIG_DIR")); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

func defaultProjectsDir() string {
	base := claudeConfigDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "projects")
}

func defaultSessionNamesDir() string {
	base := claudeConfigDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "session-names")
}
