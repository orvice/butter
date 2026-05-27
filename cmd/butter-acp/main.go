package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.orx.me/apps/butter/internal/acp"
)

func main() {
	serverURL := flag.String("server", env("BUTTER_SERVER_URL", "http://127.0.0.1:8080"), "Butter server base URL")
	workspaceID := flag.String("workspace", env("BUTTER_WORKSPACE_ID", ""), "Butter workspace ID")
	agentName := flag.String("agent", env("BUTTER_AGENT_NAME", ""), "Butter agent name to expose over ACP")
	token := flag.String("token", env("BUTTER_API_TOKEN", ""), "Butter API token")
	appName := flag.String("app", env("BUTTER_ACP_APP_NAME", "acp"), "Butter invocation app/channel name")
	userID := flag.String("user", env("BUTTER_ACP_USER_ID", "acp"), "Butter invocation user ID")
	modelOverride := flag.String("model", env("BUTTER_MODEL_OVERRIDE", ""), "optional Butter model override")
	flag.Parse()

	if strings.TrimSpace(*agentName) == "" {
		slog.Error("agent is required; set -agent or BUTTER_AGENT_NAME")
		os.Exit(2)
	}

	client := acp.NewButterClient(*serverURL, *workspaceID, *token, nil)
	server, err := acp.NewServer(acp.Config{
		AgentName:     *agentName,
		AppName:       *appName,
		UserID:        *userID,
		ModelOverride: *modelOverride,
	}, client)
	if err != nil {
		slog.Error("failed to create ACP server", "err", err)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := server.Serve(ctx, os.Stdin, os.Stdout); err != nil && ctx.Err() == nil {
		slog.Error("ACP server exited with error", "err", err)
		os.Exit(1)
	}
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
