package main

import (
	"path/filepath"
	"testing"

	"go.orx.me/apps/butter/cmd/butter-daemon/executor"
)

func TestApplyRuntimeConfigUsesEnvFallbacks(t *testing.T) {
	t.Setenv(envDaemonURL, "https://env-host/api")
	t.Setenv(envDaemonToken, "env-token")

	cfg := &Config{}
	applyRuntimeConfig(cfg, "", "")

	if cfg.Server != "https://env-host/api" {
		t.Fatalf("expected env url, got %q", cfg.Server)
	}
	if cfg.Credential != "env-token" {
		t.Fatalf("expected env token, got %q", cfg.Credential)
	}
}

func TestApplyRuntimeConfigKeepsConfigOverEnv(t *testing.T) {
	t.Setenv(envDaemonURL, "https://env-host/api")
	t.Setenv(envDaemonToken, "env-token")

	cfg := &Config{
		URL:   "https://config-host/api",
		Token: "config-token",
	}
	applyRuntimeConfig(cfg, "", "")

	if cfg.Server != "https://config-host/api" {
		t.Fatalf("expected config url, got %q", cfg.Server)
	}
	if cfg.Credential != "config-token" {
		t.Fatalf("expected config token, got %q", cfg.Credential)
	}
}

func TestApplyRuntimeConfigFlagsOverrideConfigAndEnv(t *testing.T) {
	t.Setenv(envDaemonURL, "https://env-host/api")
	t.Setenv(envDaemonToken, "env-token")

	cfg := &Config{
		Server:     "https://config-host/api",
		Credential: "config-token",
	}
	applyRuntimeConfig(cfg, "https://flag-host/api", "flag-token")

	if cfg.Server != "https://flag-host/api" {
		t.Fatalf("expected flag url, got %q", cfg.Server)
	}
	if cfg.Credential != "flag-token" {
		t.Fatalf("expected flag token, got %q", cfg.Credential)
	}
}

func TestACPProfilesFromConfig(t *testing.T) {
	workDir := t.TempDir()
	cfg := &Config{
		Executors: ExecutorConfig{
			ACP: []executor.ACPConfig{{
				Runtime: "gemini",
				Command: "gemini",
				Args:    []string{"--experimental-acp"},
				WorkDir: workDir,
			}},
		},
	}

	profiles, err := acpProfilesFromConfig(cfg)
	if err != nil {
		t.Fatalf("acpProfilesFromConfig: %v", err)
	}
	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(profiles))
	}
	got := profiles[2]
	if got.Runtime != "gemini" || got.Command != "gemini" {
		t.Fatalf("unexpected profile: %+v", got)
	}
	if got.WorkDir != filepath.Clean(workDir) {
		t.Fatalf("expected normalized work dir %q, got %q", filepath.Clean(workDir), got.WorkDir)
	}
}

func TestACPProfilesFromConfigRejectsIncompleteProfile(t *testing.T) {
	cfg := &Config{
		Executors: ExecutorConfig{
			ACP: []executor.ACPConfig{{Command: "opencode"}},
		},
	}
	if _, err := acpProfilesFromConfig(cfg); err == nil {
		t.Fatalf("expected missing runtime error")
	}

	cfg = &Config{
		Executors: ExecutorConfig{
			ACP: []executor.ACPConfig{{Runtime: "opencode"}},
		},
	}
	if _, err := acpProfilesFromConfig(cfg); err == nil {
		t.Fatalf("expected missing command error")
	}
}

func TestLegacyOpenCodeConfigBecomesACPProfile(t *testing.T) {
	workDir := t.TempDir()
	cfg := &Config{
		Executors: ExecutorConfig{
			OpenCode: &OpenCodeConfig{
				Binary:  "custom-opencode",
				WorkDir: workDir,
			},
		},
	}

	profiles, err := acpProfilesFromConfig(cfg)
	if err != nil {
		t.Fatalf("acpProfilesFromConfig: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	got := profiles[0]
	if got.Runtime != "opencode" {
		t.Fatalf("expected opencode runtime, got %q", got.Runtime)
	}
	if got.Command != "custom-opencode" {
		t.Fatalf("expected custom binary, got %q", got.Command)
	}
	if len(got.Args) != 1 || got.Args[0] != "acp" {
		t.Fatalf("expected args [acp], got %#v", got.Args)
	}
	if got.PermissionPolicy != executor.PermissionPolicyDeny {
		t.Fatalf("expected deny policy, got %q", got.PermissionPolicy)
	}
	if !got.FS.Read || !got.FS.Write || !got.Terminal {
		t.Fatalf("expected legacy opencode to enable fs and terminal, got %+v", got)
	}
}

func TestBuildExecutorsPreservesShell(t *testing.T) {
	cfg := &Config{
		Executors: ExecutorConfig{
			Shell: &ShellConfig{WorkDir: t.TempDir()},
		},
	}
	executors, err := buildExecutors(cfg)
	if err != nil {
		t.Fatalf("buildExecutors: %v", err)
	}
	if len(executors) != 3 {
		t.Fatalf("expected 3 executors, got %d", len(executors))
	}
	foundShell := false
	for _, exec := range executors {
		if exec.Runtime() == "shell" {
			foundShell = true
		}
	}
	if !foundShell {
		t.Fatalf("expected shell runtime in executors")
	}
}
