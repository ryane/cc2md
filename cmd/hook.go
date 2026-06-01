package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

type hookFlags struct {
	Dir        string
	Flavor     string
	Thinking   bool
	Collapse   bool
	MaxLines   int
	Transcript string
	SessionID  string
}

type hookConfig struct {
	Dir      string
	Flavor   string
	Thinking bool
	Collapse bool
	MaxLines int
}

// resolveHookConfig applies precedence: flag (when *Set) > env > default.
// The five booleans report whether the user passed each flag on the command line
// (cmd.Flags().Changed(name)).
func resolveHookConfig(f hookFlags, dirSet, flavorSet, thinkingSet, collapseSet, maxLinesSet bool) hookConfig {
	cfg := hookConfig{
		Dir:      "~/claude-code-logs",
		Flavor:   "obsidian",
		Thinking: true,
		Collapse: true,
		MaxLines: 100,
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
	hookCmd.Flags().IntVar(&hookFlagValues.MaxLines, "max-lines", 100, "Max lines per tool output (env: CC2MD_HOOK_MAX_LINES)")
	hookCmd.Flags().StringVar(&hookFlagValues.Transcript, "transcript", "", "Override transcript_path from stdin")
	hookCmd.Flags().StringVar(&hookFlagValues.SessionID, "session-id", "", "Override session_id from stdin")
	rootCmd.AddCommand(hookCmd)
}

// runHook is implemented in Task 8.
func runHook(cmd *cobra.Command, args []string) error { return nil }
