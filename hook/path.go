package hook

import (
	"strings"
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
