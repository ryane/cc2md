package cmd

import "testing"

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
