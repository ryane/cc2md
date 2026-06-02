package formatter

import (
	"fmt"
	"strings"

	"github.com/magarcia/ccsession-viewer/internal"
	"github.com/magarcia/ccsession-viewer/parser"
)

type ToolFormatOptions struct {
	Collapse bool
	MaxLines int
	Flavor   MarkdownFlavor
	// OmitOutput drops each tool call's result block, keeping only the
	// one-line header (e.g. "- **Read** `file`"). Zero value keeps output.
	OmitOutput bool
}

func FormatToolCalls(calls []parser.LinkedToolCall, opts ToolFormatOptions) string {
	parts := make([]string, len(calls))
	for i, c := range calls {
		parts[i] = formatSingleTool(c, opts)
	}
	body := strings.Join(parts, "\n\n")

	// Obsidian renders neither foldable cleanly: it won't render markdown
	// inside <details> (shows it raw), and fenced code inside a `>` callout
	// breaks at scale. So obsidian — like commonmark — always renders tool
	// output as plain markdown. Only GFM gets the <details> foldable.
	if !opts.Collapse || opts.Flavor == FlavorCommonMark || opts.Flavor == FlavorObsidian {
		return body
	}

	label := fmt.Sprintf("Tool calls (%d)", len(calls))
	if len(calls) == 1 {
		label = "Tool call (1)"
	}

	return fmt.Sprintf("<details>\n<summary>%s</summary>\n\n%s\n\n</details>", label, body)
}

var summaryKeys = map[string]string{
	"Bash":       "command",
	"Read":       "file_path",
	"Write":      "file_path",
	"Edit":       "file_path",
	"Glob":       "pattern",
	"Grep":       "pattern",
	"WebSearch":  "query",
	"WebFetch":   "url",
	"Task":       "description",
	"TaskCreate": "subject",
	"TaskUpdate": "taskId",
}

func formatSingleTool(call parser.LinkedToolCall, opts ToolFormatOptions) string {
	summary := inlineSummary(call)

	header := fmt.Sprintf("- **%s**", call.Name)
	if summary != "" {
		header += fmt.Sprintf(" `%s`", summary)
	}

	if opts.OmitOutput || call.Result == nil || *call.Result == "" {
		return header
	}

	result := *call.Result
	if opts.MaxLines > 0 {
		result = internal.TruncateLines(result, opts.MaxLines).Text
	}

	// Use a fence longer than any backtick run in the output so embedded
	// ``` (e.g. a Read of a markdown file) cannot close the wrapper early.
	fence := fenceFor(result)
	return fmt.Sprintf("%s\n\n  %s\n%s\n  %s", header, fence, indentBlock(result, 2), fence)
}

// fenceFor returns a backtick fence at least three long, and always longer
// than the longest run of backticks appearing in s, so s can be embedded
// verbatim without prematurely closing the fence.
func fenceFor(s string) string {
	longest, run := 0, 0
	for _, r := range s {
		if r == '`' {
			run++
			if run > longest {
				longest = run
			}
		} else {
			run = 0
		}
	}
	n := longest + 1
	if n < 3 {
		n = 3
	}
	return strings.Repeat("`", n)
}

func inlineSummary(call parser.LinkedToolCall) string {
	if key, ok := summaryKeys[call.Name]; ok {
		if val, ok := call.Input[key]; ok {
			if s, ok := val.(string); ok {
				return internal.TruncateString(s, 80)
			}
		}
		return ""
	}

	if strings.HasPrefix(call.Name, "mcp__") {
		parts := strings.Split(call.Name, "__")
		return parts[len(parts)-1]
	}

	return ""
}

func indentBlock(text string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
