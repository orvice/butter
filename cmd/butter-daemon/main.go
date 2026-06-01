package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gopkg.in/yaml.v3"

	"go.orx.me/apps/butter/cmd/butter-daemon/executor"
)

// Config holds daemon configuration.
type Config struct {
	Server   string            `yaml:"server"`
	Host     string            `yaml:"host"`
	Token    string            `yaml:"token"`
	DaemonID string            `yaml:"daemon_id"`
	Name     string            `yaml:"name"`
	Labels   map[string]string `yaml:"labels"`

	Executors ExecutorConfig `yaml:"executors"`
}

// ExecutorConfig holds per-executor settings.
type ExecutorConfig struct {
	OpenCode struct {
		WorkDir string `yaml:"work_dir"`
		Binary  string `yaml:"binary"`
	} `yaml:"opencode"`
	Shell struct {
		WorkDir string `yaml:"work_dir"`
	} `yaml:"shell"`
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
	cfg.Host = resolveHost(cfg.Host, cfg.Server)
	if cfg.DaemonID == "" {
		hostname, _ := os.Hostname()
		cfg.DaemonID = hostname
	}

	executors := buildExecutors(cfg)
	if len(executors) == 0 {
		slog.Error("no executors configured")
		os.Exit(1)
	}

	connector := NewConnector(cfg, executors)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("starting butter-daemon",
		"server", cfg.Server,
		"host", cfg.Host,
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

func resolveHost(host, server string) string {
	if trimmed := strings.TrimSpace(host); trimmed != "" {
		return trimmed
	}

	server = strings.TrimSpace(server)
	if server == "" {
		return ""
	}
	if u, err := url.Parse(server); err == nil && u.Hostname() != "" {
		return u.Hostname()
	}
	if h, _, err := net.SplitHostPort(server); err == nil {
		return strings.Trim(h, "[]")
	}
	if idx := strings.LastIndex(server, ":"); idx > -1 && strings.Count(server, ":") == 1 {
		return server[:idx]
	}
	return strings.Trim(server, "[]")
}

func buildExecutors(cfg *Config) []executor.Executor {
	var executors []executor.Executor

	executors = append(executors, &executor.OpenCodeExecutor{
		WorkDir: cfg.Executors.OpenCode.WorkDir,
		Binary:  cfg.Executors.OpenCode.Binary,
	})

	executors = append(executors, &executor.ShellExecutor{
		WorkDir: cfg.Executors.Shell.WorkDir,
	})

	return executors
}
