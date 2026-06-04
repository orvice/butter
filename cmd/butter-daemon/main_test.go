package main

import (
	"path/filepath"
	"testing"

	"go.orx.me/apps/butter/cmd/butter-daemon/executor"
)

func TestACPProfilesFromConfig(t *testing.T) {
	workDir := t.TempDir()
	cfg := &Config{
		Executors: ExecutorConfig{
			ACP: []executor.ACPConfig{{
				Capability: "gemini",
				Command:    "gemini",
				Args:       []string{"--experimental-acp"},
				WorkDir:    workDir,
			}},
		},
	}

	profiles, err := acpProfilesFromConfig(cfg)
	if err != nil {
		t.Fatalf("acpProfilesFromConfig: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Capability != "gemini" || profiles[0].Command != "gemini" {
		t.Fatalf("unexpected profile: %+v", profiles[0])
	}
	if profiles[0].WorkDir != filepath.Clean(workDir) {
		t.Fatalf("expected normalized work dir %q, got %q", filepath.Clean(workDir), profiles[0].WorkDir)
	}
}

func TestACPProfilesFromConfigRejectsIncompleteProfile(t *testing.T) {
	cfg := &Config{
		Executors: ExecutorConfig{
			ACP: []executor.ACPConfig{{Command: "opencode"}},
		},
	}
	if _, err := acpProfilesFromConfig(cfg); err == nil {
		t.Fatalf("expected missing capability error")
	}

	cfg = &Config{
		Executors: ExecutorConfig{
			ACP: []executor.ACPConfig{{Capability: "opencode"}},
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
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	got := profiles[0]
	if got.Capability != "opencode" {
		t.Fatalf("expected opencode capability, got %q", got.Capability)
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
	if len(executors) != 1 {
		t.Fatalf("expected 1 executor, got %d", len(executors))
	}
	if executors[0].Capability() != "shell" {
		t.Fatalf("expected shell capability, got %q", executors[0].Capability())
	}
}
