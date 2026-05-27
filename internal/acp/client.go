package acp

import (
	"context"
	"net/http"

	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type ButterClient struct {
	agentSvc    agentsv1.AgentService
	workspaceID string
	token       string
}

func NewButterClient(serverURL, workspaceID, token string, httpClient *http.Client) *ButterClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ButterClient{
		agentSvc:    agentsv1.NewAgentServiceJSONClient(serverURL, httpClient, twirp.WithClientPathPrefix("/api")),
		workspaceID: workspaceID,
		token:       token,
	}
}

func (c *ButterClient) Invoke(ctx context.Context, req InvokeRequest) (InvokeResponse, error) {
	headers := http.Header{}
	if c.workspaceID != "" {
		headers.Set(workspace.HeaderName, c.workspaceID)
	}
	if c.token != "" {
		headers.Set("Authorization", "Bearer "+c.token)
	}
	if len(headers) > 0 {
		var err error
		ctx, err = twirp.WithHTTPRequestHeaders(ctx, headers)
		if err != nil {
			return InvokeResponse{}, err
		}
	}

	resp, err := c.agentSvc.InvokeAgent(ctx, &agentsv1.InvokeAgentRequest{
		AgentName:     req.AgentName,
		Input:         req.Input,
		AppName:       req.AppName,
		UserId:        req.UserID,
		SessionId:     req.SessionID,
		ModelOverride: req.ModelOverride,
	})
	if err != nil {
		return InvokeResponse{}, err
	}
	return InvokeResponse{SessionID: resp.GetSessionId(), Response: resp.GetResponse()}, nil
}
