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

// Config holds daemon configuration.
type Config struct {
	Server     string            `yaml:"server"`
	Credential string            `yaml:"credential"`
	DaemonID   string            `yaml:"daemon_id"`
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
	configPath := flag.String("config", "daemon.yaml", "path to config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	if cfg.Server == "" {
		slog.Error("server address is required")
		os.Exit(1)
	}
	if cfg.Credential == "" {
		slog.Error("daemon credential is required")
		os.Exit(1)
	}
	if cfg.DaemonID == "" {
		hostname, _ := os.Hostname()
		cfg.DaemonID = hostname
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
		"daemon_id", cfg.DaemonID,
		"name", cfg.Name,
	)

	if err := connector.Run(ctx); err != nil && ctx.Err() == nil {
		slog.Error("daemon exited with error", "err", err)
		os.Exit(1)
	}

	slog.Info("daemon stopped")
}

func loadConfig(path string) (*Config, error) {
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
	profiles := append([]executor.ACPConfig(nil), cfg.Executors.ACP...)

	if cfg.Executors.OpenCode != nil {
		binary := cfg.Executors.OpenCode.Binary
		if binary == "" {
			binary = "opencode"
		}
		slog.Warn("executors.opencode is deprecated; use executors.acp with command 'opencode' and args ['acp']")
		profiles = append(profiles, executor.ACPConfig{
			Capability:       "opencode",
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

	for i := range profiles {
		if profiles[i].Capability == "" {
			return nil, fmt.Errorf("acp executor %d: capability is required", i)
		}
		if profiles[i].Command == "" {
			return nil, fmt.Errorf("acp executor %q: command is required", profiles[i].Capability)
		}
		workDir, err := normalizeWorkDir(profiles[i].WorkDir)
		if err != nil {
			return nil, fmt.Errorf("acp executor %q work_dir: %w", profiles[i].Capability, err)
		}
		profiles[i].WorkDir = workDir
	}

	return profiles, nil
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
