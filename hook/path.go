package hook

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magarcia/ccsession-viewer/discovery"
)

const maxTitleSlugLen = 50

// SlugifyTitle converts s to a filename-safe slug: only ASCII [a-z0-9] are
// kept; runs of any other rune (including non-ASCII) become a single '-';
// leading/trailing '-' are trimmed; the result is capped at maxTitleSlugLen
// (cutting at the last '-' boundary within the cap when possible). Returns ""
// if the slug is empty after sanitization. Note: non-ASCII runes (e.g. 'é',
// '世', U+212A KELVIN SIGN) are always treated as separators, not letters —
// this matches the spec's literal "non-[a-z0-9]" rule. The function
// intentionally avoids strings.ToLower because its Unicode-aware mapping
// (e.g. U+212A → 'k') would allow non-ASCII letters to sneak through the
// ASCII-only gate.
func SlugifyTitle(s string) string {
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		// Reject non-ASCII before any lowercasing so that runes like U+212A
		// (KELVIN SIGN, which Unicode-aware ToLower maps to ASCII 'k') cannot
		// sneak through the ASCII-only gate.
		if r > 127 {
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
			continue
		}
		// ASCII range — apply ASCII-only lowercasing manually.
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
			continue
		}
		if !prevHyphen && b.Len() > 0 {
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) <= maxTitleSlugLen {
		return out
	}
	clipped := out[:maxTitleSlugLen]
	if idx := strings.LastIndex(clipped, "-"); idx > 0 {
		clipped = clipped[:idx]
	}
	return strings.Trim(clipped, "-")
}

const idShortLen = 8

// IDShort returns the first idShortLen characters of sessionID, or the full
// string if it is shorter than that. Returns "unknown" if sessionID is empty.
func IDShort(sessionID string) string {
	if sessionID == "" {
		return "unknown"
	}
	if len(sessionID) < idShortLen {
		return sessionID
	}
	return sessionID[:idShortLen]
}

// BuildFilename assembles a session export filename from the session's date,
// a slugified title, and the short session id. If titleSlug is empty, the
// title segment is omitted: <YYYY-MM-DD>-<idShort>.md. Otherwise the
// filename is <YYYY-MM-DD>-<titleSlug>-<idShort>.md.
func BuildFilename(date time.Time, titleSlug, idShort string) string {
	d := date.Format("2006-01-02")
	if titleSlug == "" {
		return d + "-" + idShort + ".md"
	}
	return d + "-" + titleSlug + "-" + idShort + ".md"
}

// ResolveDir expands a leading "~/" or bare "~" against the user's home directory.
// Any other path is returned verbatim. If the home dir cannot be determined,
// the input is returned unchanged.
//
// Note: POSIX "~username" syntax is not supported and is returned unchanged.
func ResolveDir(dir string) string {
	if dir == "~" {
		if h, err := os.UserHomeDir(); err == nil {
			return h
		}
		return dir
	}
	if strings.HasPrefix(dir, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			return filepath.Join(h, dir[2:])
		}
	}
	return dir
}

// OutputPath assembles the full archive path for a session export.
// hookDir may use ~ for the home directory. cwd is the session's real working
// directory (from the hook payload or transcript); pass "" if unknown.
func OutputPath(hookDir, cwd, transcriptPath, sessionID, titleSlug string, date time.Time) string {
	return filepath.Join(
		ResolveDir(hookDir),
		ProjectSlug(cwd, transcriptPath),
		BuildFilename(date, titleSlug, IDShort(sessionID)),
	)
}

// ProjectSlug derives a per-project folder name for a session.
//
// When cwd is a usable working directory, its basename is used directly. This
// is the preferred source because it is unambiguous: it preserves both path
// separators and literal hyphens (e.g. cwd "/Users/ryan/org-tools" yields
// "org-tools").
//
// When cwd is empty or degenerate ("/", "."), it falls back to decoding the
// transcript path's parent directory name (e.g. "-Users-ryan-Projects-cc2md")
// via discovery.DecodeProjectName and taking the basename of the result.
//
// Caveat (fallback only): DecodeProjectName replaces every '-' in the encoded
// name with '/', which is lossy for project directories that contain hyphens —
// "-Users-ryan-Projects-my-go-app" decodes to "/Users/ryan/Projects/my/go/app"
// and yields "app", not "my-go-app". The cwd-based path above avoids this; the
// fallback only runs when no real cwd is available.
func ProjectSlug(cwd, transcriptPath string) string {
	if slug := filepath.Base(filepath.Clean(cwd)); cwd != "" && slug != "." && slug != string(filepath.Separator) {
		return slug
	}

	encoded := filepath.Base(filepath.Dir(transcriptPath))
	if encoded == "" || encoded == "." || encoded == "/" {
		return "unknown"
	}
	decoded := discovery.DecodeProjectName(encoded)
	slug := filepath.Base(decoded)
	if slug == "" || slug == "." || slug == "/" {
		// Defensive: unreachable for Claude-emitted paths after the encoded
		// sentinel check above, but kept in case DecodeProjectName behavior changes.
		return strings.TrimPrefix(strings.ReplaceAll(encoded, "/", "-"), "-")
	}
	return slug
}
