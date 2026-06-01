package formatter

import "fmt"

// MarkdownFlavor controls which markdown dialect is used for output.
type MarkdownFlavor string

const (
	FlavorGFM        MarkdownFlavor = "gfm"
	FlavorCommonMark MarkdownFlavor = "commonmark"
	FlavorObsidian   MarkdownFlavor = "obsidian"
)

func ParseFlavor(s string) (MarkdownFlavor, error) {
	switch MarkdownFlavor(s) {
	case FlavorGFM:
		return FlavorGFM, nil
	case FlavorCommonMark:
		return FlavorCommonMark, nil
	case FlavorObsidian:
		return FlavorObsidian, nil
	default:
		return "", fmt.Errorf("invalid markdown flavor %q, must be one of: gfm, commonmark, obsidian", s)
	}
}
