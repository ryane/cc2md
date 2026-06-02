package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/magarcia/ccsession-viewer/discovery"
	"github.com/magarcia/ccsession-viewer/formatter"
	"github.com/magarcia/ccsession-viewer/hook"
	"github.com/magarcia/ccsession-viewer/parser"
)

type hookFlags struct {
	Dir        string
	Flavor     string
	Thinking   bool
	Collapse   bool
	ToolOutput bool
	MaxLines   int
	Transcript string
	SessionID  string
}

type hookConfig struct {
	Dir        string
	Flavor     string
	Thinking   bool
	Collapse   bool
	ToolOutput bool
	MaxLines   int
}

// resolveHookConfig applies precedence: flag (when *Set) > env > default.
// The five booleans report whether the user passed each flag on the command line
// (cmd.Flags().Changed(name)).
func resolveHookConfig(f hookFlags, dirSet, flavorSet, thinkingSet, collapseSet, toolOutputSet, maxLinesSet bool) hookConfig {
	cfg := hookConfig{
		Dir:        "~/claude-code-logs",
		Flavor:     "obsidian",
		Thinking:   true,
		Collapse:   true,
		ToolOutput: true,
		MaxLines:   100,
	}

	if v := os.Getenv("CC2MD_HOOK_DIR"); v != "" {
		cfg.Dir = v
	}
	if v := os.Getenv("CC2MD_HOOK_FLAVOR"); v != "" {
		cfg.Flavor = v
	}
	if v := os.Getenv("CC2MD_HOOK_THINKING"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Thinking = b
		} else {
			fmt.Fprintf(os.Stderr, "cc2md hook: invalid CC2MD_HOOK_THINKING=%q, using default\n", v)
		}
	}
	if v := os.Getenv("CC2MD_HOOK_COLLAPSE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Collapse = b
		} else {
			fmt.Fprintf(os.Stderr, "cc2md hook: invalid CC2MD_HOOK_COLLAPSE=%q, using default\n", v)
		}
	}
	if v := os.Getenv("CC2MD_HOOK_TOOL_OUTPUT"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.ToolOutput = b
		} else {
			fmt.Fprintf(os.Stderr, "cc2md hook: invalid CC2MD_HOOK_TOOL_OUTPUT=%q, using default\n", v)
		}
	}
	if v := os.Getenv("CC2MD_HOOK_MAX_LINES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxLines = n
		} else {
			fmt.Fprintf(os.Stderr, "cc2md hook: invalid CC2MD_HOOK_MAX_LINES=%q, using default\n", v)
		}
	}

	if dirSet {
		cfg.Dir = f.Dir
	}
	if flavorSet {
		cfg.Flavor = f.Flavor
	}
	if thinkingSet {
		cfg.Thinking = f.Thinking
	}
	if collapseSet {
		cfg.Collapse = f.Collapse
	}
	if toolOutputSet {
		cfg.ToolOutput = f.ToolOutput
	}
	if maxLinesSet {
		cfg.MaxLines = f.MaxLines
	}

	return cfg
}

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Run as a Claude Code Stop hook: export the current session to markdown",
	Long:  "Reads a Claude Code Stop-hook JSON payload from stdin and writes the session as markdown.",
	RunE:  runHook, // implemented in Task 8
}

var hookFlagValues hookFlags

func init() {
	hookCmd.Flags().StringVar(&hookFlagValues.Dir, "dir", "", "Archive root directory (env: CC2MD_HOOK_DIR)")
	hookCmd.Flags().StringVar(&hookFlagValues.Flavor, "flavor", "", "Markdown flavor: gfm, commonmark, obsidian (env: CC2MD_HOOK_FLAVOR)")
	hookCmd.Flags().BoolVar(&hookFlagValues.Thinking, "thinking", true, "Include thinking blocks (env: CC2MD_HOOK_THINKING)")
	hookCmd.Flags().BoolVar(&hookFlagValues.Collapse, "collapse", true, "Collapse tool calls (env: CC2MD_HOOK_COLLAPSE)")
	hookCmd.Flags().BoolVar(&hookFlagValues.ToolOutput, "tool-output", true, "Include tool call output; --tool-output=false keeps only the headers (env: CC2MD_HOOK_TOOL_OUTPUT)")
	hookCmd.Flags().IntVar(&hookFlagValues.MaxLines, "max-lines", 100, "Max lines per tool output (env: CC2MD_HOOK_MAX_LINES)")
	hookCmd.Flags().StringVar(&hookFlagValues.Transcript, "transcript", "", "Override transcript_path from stdin")
	hookCmd.Flags().StringVar(&hookFlagValues.SessionID, "session-id", "", "Override session_id from stdin")
	rootCmd.AddCommand(hookCmd)
}

func runHook(cmd *cobra.Command, args []string) error {
	cfg := resolveHookConfig(
		hookFlagValues,
		cmd.Flags().Changed("dir"),
		cmd.Flags().Changed("flavor"),
		cmd.Flags().Changed("thinking"),
		cmd.Flags().Changed("collapse"),
		cmd.Flags().Changed("tool-output"),
		cmd.Flags().Changed("max-lines"),
	)
	flavor, err := formatter.ParseFlavor(cfg.Flavor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cc2md hook: %v, using obsidian\n", err)
		flavor = formatter.FlavorObsidian
	}

	in, parseErr := hook.ParseInput(cmd.InOrStdin())
	transcript := hookFlagValues.Transcript
	sessionID := hookFlagValues.SessionID
	if transcript == "" {
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "cc2md hook: parse stdin: %v\n", parseErr)
			return nil // exit 0
		}
		transcript = in.TranscriptPath
	}
	if sessionID == "" {
		sessionID = in.SessionID
	}
	if transcript == "" {
		fmt.Fprintln(os.Stderr, "cc2md hook: missing transcript_path")
		return nil
	}

	// Date from transcript modtime, else now.
	date := time.Now()
	if st, statErr := os.Stat(transcript); statErr == nil {
		date = st.ModTime()
	} else {
		fmt.Fprintf(os.Stderr, "cc2md hook: stat transcript: %v (falling back to now)\n", statErr)
	}

	lines, err := parser.ReadSessionFile(transcript)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cc2md hook: read transcript: %v\n", err)
		return nil
	}
	meta := parser.ExtractMetadata(lines)
	turns := parser.BuildTurns(lines)
	md := formatter.FormatSession(meta, turns, formatter.FormatOptions{
		IncludeThinking: cfg.Thinking,
		Collapse:        cfg.Collapse,
		MaxLines:        cfg.MaxLines,
		Flavor:          flavor,
		OmitToolOutput:  !cfg.ToolOutput,
	})

	titleSlug := hook.SlugifyTitle(discovery.ExtractFirstUserMessage(transcript, 60))
	target := hook.OutputPath(cfg.Dir, transcript, sessionID, titleSlug, date)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "cc2md hook: mkdir %s: %v\n", filepath.Dir(target), err)
		return nil
	}
	if err := atomicWrite(target, []byte(md)); err != nil {
		fmt.Fprintf(os.Stderr, "cc2md hook: write %s: %v\n", target, err)
		return nil
	}
	fmt.Fprintf(os.Stderr, "cc2md hook: wrote %s\n", target)
	return nil
}

func atomicWrite(target string, data []byte) error {
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".cc2md-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmpName, target); err != nil {
		cleanup()
		return err
	}
	return nil
}
