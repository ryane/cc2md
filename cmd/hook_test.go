package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveHookFlags_Defaults(t *testing.T) {
	t.Setenv("CC2MD_HOOK_DIR", "")
	t.Setenv("CC2MD_HOOK_FLAVOR", "")
	t.Setenv("CC2MD_HOOK_THINKING", "")
	t.Setenv("CC2MD_HOOK_COLLAPSE", "")
	t.Setenv("CC2MD_HOOK_MAX_LINES", "")

	cfg := resolveHookConfig(hookFlags{}, false, false, false, false, false)
	if cfg.Dir != "~/claude-code-logs" {
		t.Errorf("Dir default: got %q", cfg.Dir)
	}
	if cfg.Flavor != "obsidian" {
		t.Errorf("Flavor default: got %q", cfg.Flavor)
	}
	if cfg.Thinking != true {
		t.Errorf("Thinking default: got %v", cfg.Thinking)
	}
	if cfg.Collapse != true {
		t.Errorf("Collapse default: got %v", cfg.Collapse)
	}
	if cfg.MaxLines != 100 {
		t.Errorf("MaxLines default: got %d", cfg.MaxLines)
	}
}

func TestResolveHookFlags_EnvOverridesDefault(t *testing.T) {
	t.Setenv("CC2MD_HOOK_DIR", "/tmp/from-env")
	t.Setenv("CC2MD_HOOK_THINKING", "false")
	t.Setenv("CC2MD_HOOK_MAX_LINES", "42")

	cfg := resolveHookConfig(hookFlags{}, false, false, false, false, false)
	if cfg.Dir != "/tmp/from-env" {
		t.Errorf("Dir from env: got %q", cfg.Dir)
	}
	if cfg.Thinking != false {
		t.Errorf("Thinking from env: got %v", cfg.Thinking)
	}
	if cfg.MaxLines != 42 {
		t.Errorf("MaxLines from env: got %d", cfg.MaxLines)
	}
}

func TestResolveHookFlags_FlagOverridesEnv(t *testing.T) {
	t.Setenv("CC2MD_HOOK_DIR", "/tmp/from-env")
	t.Setenv("CC2MD_HOOK_THINKING", "false")

	f := hookFlags{Dir: "/tmp/from-flag", Thinking: true}
	cfg := resolveHookConfig(f, true /*dirSet*/, false, true /*thinkingSet*/, false, false)
	if cfg.Dir != "/tmp/from-flag" {
		t.Errorf("Dir from flag: got %q", cfg.Dir)
	}
	if cfg.Thinking != true {
		t.Errorf("Thinking from flag: got %v", cfg.Thinking)
	}
}

func TestResolveHookFlags_BadEnvBoolFallsBack(t *testing.T) {
	t.Setenv("CC2MD_HOOK_THINKING", "notabool")
	cfg := resolveHookConfig(hookFlags{}, false, false, false, false, false)
	if cfg.Thinking != true {
		t.Errorf("expected default true on bad bool, got %v", cfg.Thinking)
	}
}

func TestHookCmd_WritesObsidianFile(t *testing.T) {
	wd, _ := os.Getwd()
	fixture := filepath.Join(wd, "testdata", "sample.jsonl")
	tmp := t.TempDir()

	stdinPayload, _ := json.Marshal(map[string]string{
		"session_id":      "0b9c1f3a-7e4d-4f2a-b8c1-3d4e5f6a7b8c",
		"transcript_path": fixture,
		"hook_event_name": "Stop",
	})

	resetHookFlagsForTest(t)
	rootCmd.SetIn(bytes.NewReader(stdinPayload))
	rootCmd.SetArgs([]string{"hook", "--dir", tmp})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("hook returned error: %v", err)
	}

	// Project slug is derived from filepath.Base(filepath.Dir(fixture)).
	// fixture is .../cmd/testdata/sample.jsonl, so the parent dir name is
	// "testdata". DecodeProjectName leaves it unchanged, and filepath.Base
	// of "testdata" is "testdata". So files land at <tmp>/testdata/...
	projectDir := filepath.Join(tmp, "testdata")
	matches, _ := filepath.Glob(filepath.Join(projectDir, "*.md"))
	if len(matches) != 1 {
		t.Fatalf("expected exactly 1 file under %s; got %v", projectDir, matches)
	}
	got, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	body := string(got)
	if !strings.Contains(body, "Hello, world.") {
		t.Errorf("output missing user text: %s", body)
	}
	if !strings.Contains(body, "Hi there!") {
		t.Errorf("output missing assistant text: %s", body)
	}

	// no temp leftovers
	leftovers, _ := filepath.Glob(filepath.Join(projectDir, ".cc2md-*.tmp"))
	if len(leftovers) != 0 {
		t.Errorf("expected no temp files; got %v", leftovers)
	}
}

func TestHookCmd_Idempotent(t *testing.T) {
	wd, _ := os.Getwd()
	fixture := filepath.Join(wd, "testdata", "sample.jsonl")
	tmp := t.TempDir()
	stdinPayload, _ := json.Marshal(map[string]string{
		"session_id":      "0b9c1f3a-7e4d-4f2a-b8c1-3d4e5f6a7b8c",
		"transcript_path": fixture,
	})

	for i := 0; i < 2; i++ {
		resetHookFlagsForTest(t)
		rootCmd.SetIn(bytes.NewReader(stdinPayload))
		rootCmd.SetArgs([]string{"hook", "--dir", tmp})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
	matches, _ := filepath.Glob(filepath.Join(tmp, "testdata", "*.md"))
	if len(matches) != 1 {
		t.Fatalf("idempotency: expected 1 file, got %d: %v", len(matches), matches)
	}
}

func TestHookCmd_GfmFlavor(t *testing.T) {
	wd, _ := os.Getwd()
	fixture := filepath.Join(wd, "testdata", "sample.jsonl")
	tmp := t.TempDir()
	stdinPayload, _ := json.Marshal(map[string]string{
		"session_id":      "0b9c1f3a-7e4d-4f2a-b8c1-3d4e5f6a7b8c",
		"transcript_path": fixture,
	})

	resetHookFlagsForTest(t)
	rootCmd.SetIn(bytes.NewReader(stdinPayload))
	rootCmd.SetArgs([]string{"hook", "--dir", tmp, "--flavor", "gfm"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("hook returned error: %v", err)
	}
	matches, _ := filepath.Glob(filepath.Join(tmp, "testdata", "*.md"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 file, got %v", matches)
	}
	body, _ := os.ReadFile(matches[0])
	if !strings.Contains(string(body), "Hi there!") {
		t.Errorf("flavor=gfm: missing assistant content")
	}
}

// resetHookFlagsForTest reinitializes the hook subcommand's flag state between
// rootCmd.Execute() runs. cobra retains parsed flag values across invocations
// of the same Command instance; without this, --dir from a previous test
// leaks into the next.
func resetHookFlagsForTest(t *testing.T) {
	t.Helper()
	hookFlagValues = hookFlags{}
	hookCmd.ResetFlags()
	hookCmd.Flags().StringVar(&hookFlagValues.Dir, "dir", "", "")
	hookCmd.Flags().StringVar(&hookFlagValues.Flavor, "flavor", "", "")
	hookCmd.Flags().BoolVar(&hookFlagValues.Thinking, "thinking", true, "")
	hookCmd.Flags().BoolVar(&hookFlagValues.Collapse, "collapse", true, "")
	hookCmd.Flags().IntVar(&hookFlagValues.MaxLines, "max-lines", 100, "")
	hookCmd.Flags().StringVar(&hookFlagValues.Transcript, "transcript", "", "")
	hookCmd.Flags().StringVar(&hookFlagValues.SessionID, "session-id", "", "")

	t.Cleanup(func() {
		rootCmd.SetIn(os.Stdin)
	})
}
