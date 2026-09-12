package formatter

import (
	"fmt"
	"strings"
	"time"

	"github.com/magarcia/ccsession-viewer/parser"
)

type FormatOptions struct {
	IncludeThinking bool
	Collapse        bool
	MaxLines        int
	Flavor          MarkdownFlavor
	// OmitToolOutput drops tool-call result blocks, keeping only the
	// one-line headers. Zero value keeps output.
	OmitToolOutput bool
	// Frontmatter, when non-nil, prepends a YAML frontmatter block to the
	// rendered session. Only the hook path sets it; stdout and TUI rendering
	// leave it nil so their output is unchanged.
	Frontmatter *FrontmatterFields
}

func FormatSession(meta parser.SessionMetadata, turns []parser.ConversationTurn, opts FormatOptions) string {
	sections := []string{FormatMetadata(meta, opts.Flavor)}

	for _, turn := range turns {
		ts := formatTimestamp(turn.Timestamp)

		switch turn.Type {
		case "user":
			if s := FormatUserTurn(turn.Text, ts, opts.Flavor); s != "" {
				sections = append(sections, s)
			}

		case "local-command":
			sections = append(sections, FormatLocalCommand(turn.Text, ts))

		case "teammate":
			name := turn.TeammateName
			if name == "" {
				name = "agent"
			}
			content := strings.Join(turn.Text, "\n")
			sections = append(sections, FormatTeammateMessage(name, content, ts, opts.Flavor))

		case "assistant":
			var parts []string
			if opts.IncludeThinking && len(turn.Thinking) > 0 {
				parts = append(parts, FormatThinking(turn.Thinking, opts.Collapse, opts.Flavor))
			}
			if len(turn.Text) > 0 {
				parts = append(parts, FormatAssistantText(turn.Text))
			}
			if len(turn.ToolCalls) > 0 {
				parts = append(parts, FormatToolCalls(turn.ToolCalls, ToolFormatOptions{
					Collapse:   opts.Collapse,
					MaxLines:   opts.MaxLines,
					Flavor:     opts.Flavor,
					OmitOutput: opts.OmitToolOutput,
				}))
			}
			if len(parts) > 0 {
				body := strings.Join(parts, "\n\n")
				if ts != "" {
					body = fmt.Sprintf("> **Claude** `%s`\n\n%s", ts, body)
				}
				sections = append(sections, body)
			}
		}
	}

	out := strings.Join(sections, "\n\n") + "\n"
	if opts.Frontmatter != nil {
		return FormatFrontmatter(*opts.Frontmatter) + out
	}
	return out
}

func formatTimestamp(iso string) string {
	if iso == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339Nano, iso)
	if err != nil {
		return ""
	}
	return t.Local().Format("01/02 15:04:05")
}
