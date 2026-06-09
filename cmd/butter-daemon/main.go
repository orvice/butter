package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"gopkg.in/yaml.v3"

	"go.orx.me/apps/butter/cmd/butter-daemon/executor"
)

const (
	envDaemonURL   = "BUTTER_DAEMON_URL"
	envDaemonToken = "BUTTER_DAEMON_TOKEN"
)

// Config holds daemon configuration.
type Config struct {
	Server     string            `yaml:"server"`
	URL        string            `yaml:"url"`
	Credential string            `yaml:"credential"`
	Token      string            `yaml:"token"`
	Name       string            `yaml:"name"`
	Labels     map[string]string `yaml:"labels"`

	Executors ExecutorConfig `yaml:"executors"`
}

// ExecutorConfig holds per-executor settings.
type ExecutorConfig struct {
	ACP      []executor.ACPConfig `yaml:"acp"`
	OpenCode *OpenCodeConfig      `yaml:"opencode"`
	Shell    *ShellConfig         `yaml:"shell"`
}

// OpenCodeConfig is the legacy opencode executor shape. It is translated to an
// ACP profile that starts "opencode acp".
type OpenCodeConfig struct {
	WorkDir string `yaml:"work_dir"`
	Binary  string `yaml:"binary"`
}

// ShellConfig holds shell executor settings.
type ShellConfig struct {
	WorkDir string `yaml:"work_dir"`
}

func main() {
	configPath := flag.String("config", "", "optional path to config file")
	url := flag.String("url", "", "Butter API base URL for daemon connector")
	token := flag.String("token", "", "daemon runtime token")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil && *configPath != "" {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}
	if cfg == nil {
		cfg = &Config{}
	}
	applyRuntimeConfig(cfg, *url, *token)

	if cfg.Server == "" {
		slog.Error("url is required")
		os.Exit(1)
	}
	if cfg.Credential == "" {
		slog.Error("token is required")
		os.Exit(1)
	}

	executors, err := buildExecutors(cfg)
	if err != nil {
		slog.Error("failed to build executors", "err", err)
		os.Exit(1)
	}
	if len(executors) == 0 {
		slog.Error("no executors configured")
		os.Exit(1)
	}

	connector := NewConnector(cfg, executors)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("starting butter-daemon",
		"server", cfg.Server,
		"name", cfg.Name,
	)

	if err := connector.Run(ctx); err != nil && ctx.Err() == nil {
		slog.Error("daemon exited with error", "err", err)
		os.Exit(1)
	}

	slog.Info("daemon stopped")
}

func applyRuntimeConfig(cfg *Config, urlFlag, tokenFlag string) {
	if cfg.Server == "" && cfg.URL != "" {
		cfg.Server = cfg.URL
	}
	if cfg.Credential == "" && cfg.Token != "" {
		cfg.Credential = cfg.Token
	}
	if cfg.Server == "" {
		cfg.Server = os.Getenv(envDaemonURL)
	}
	if cfg.Credential == "" {
		cfg.Credential = os.Getenv(envDaemonToken)
	}
	if urlFlag != "" {
		cfg.Server = urlFlag
	}
	if tokenFlag != "" {
		cfg.Credential = tokenFlag
	}
}

func loadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func buildExecutors(cfg *Config) ([]executor.Executor, error) {
	var out []executor.Executor

	acpProfiles, err := acpProfilesFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	for _, profile := range acpProfiles {
		exec, err := executor.NewACPExecutor(profile)
		if err != nil {
			return nil, err
		}
		out = append(out, exec)
	}

	if cfg.Executors.Shell != nil {
		workDir, err := normalizeWorkDir(cfg.Executors.Shell.WorkDir)
		if err != nil {
			return nil, fmt.Errorf("shell work_dir: %w", err)
		}
		out = append(out, &executor.ShellExecutor{WorkDir: workDir})
	}

	return out, nil
}

func acpProfilesFromConfig(cfg *Config) ([]executor.ACPConfig, error) {
	raw := []executor.ACPConfig{
		defaultACPProfile("opencode", "opencode", []string{"acp"}),
		defaultACPProfile("codex", "codex-acp", nil),
	}
	raw = append(raw, cfg.Executors.ACP...)

	if cfg.Executors.OpenCode != nil {
		binary := cfg.Executors.OpenCode.Binary
		if binary == "" {
			binary = "opencode"
		}
		slog.Warn("executors.opencode is deprecated; use executors.acp with command 'opencode' and args ['acp']")
		raw = append(raw, executor.ACPConfig{
			Runtime:          "opencode",
			Command:          binary,
			Args:             []string{"acp"},
			WorkDir:          cfg.Executors.OpenCode.WorkDir,
			PermissionPolicy: executor.PermissionPolicyDeny,
			FS: executor.ACPFSConfig{
				Read:  true,
				Write: true,
			},
			Terminal: true,
		})
	}

	var profiles []executor.ACPConfig
	indexByRuntime := make(map[string]int)
	for i := range raw {
		profile := raw[i]
		if profile.Runtime == "" {
			profile.Runtime = profile.Capability
		}
		if profile.Runtime == "" {
			return nil, fmt.Errorf("acp executor %d: runtime is required", i)
		}
		if profile.Command == "" {
			return nil, fmt.Errorf("acp executor %q: command is required", profile.Runtime)
		}
		workDir, err := normalizeWorkDir(profile.WorkDir)
		if err != nil {
			return nil, fmt.Errorf("acp executor %q work_dir: %w", profile.Runtime, err)
		}
		profile.WorkDir = workDir
		if idx, ok := indexByRuntime[profile.Runtime]; ok {
			profiles[idx] = profile
			continue
		}
		indexByRuntime[profile.Runtime] = len(profiles)
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func defaultACPProfile(runtimeName, command string, args []string) executor.ACPConfig {
	return executor.ACPConfig{
		Runtime:          runtimeName,
		Command:          command,
		Args:             args,
		PermissionPolicy: executor.PermissionPolicyDeny,
		FS: executor.ACPFSConfig{
			Read:  true,
			Write: true,
		},
		Terminal: true,
	}
}

func normalizeWorkDir(path string) (string, error) {
	if path == "" {
		return os.Getwd()
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}
