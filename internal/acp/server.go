package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/google/uuid"
)

const protocolVersion = 1

type AgentClient interface {
	Invoke(ctx context.Context, req InvokeRequest) (InvokeResponse, error)
}

type InvokeRequest struct {
	AgentName     string
	Input         string
	AppName       string
	UserID        string
	SessionID     string
	ModelOverride string
}

type InvokeResponse struct {
	SessionID string
	Response  string
}

type Config struct {
	AgentName     string
	AppName       string
	UserID        string
	ModelOverride string
}

type Server struct {
	cfg    Config
	client AgentClient

	mu       sync.Mutex
	sessions map[string]*sessionState
}

type sessionState struct {
	butterSessionID string
	cancel          context.CancelFunc
}

func NewServer(cfg Config, client AgentClient) (*Server, error) {
	cfg.AgentName = strings.TrimSpace(cfg.AgentName)
	if cfg.AgentName == "" {
		return nil, errors.New("agent name is required")
	}
	if cfg.AppName == "" {
		cfg.AppName = "acp"
	}
	if cfg.UserID == "" {
		cfg.UserID = "acp"
	}
	if client == nil {
		return nil, errors.New("agent client is required")
	}
	return &Server{
		cfg:      cfg,
		client:   client,
		sessions: make(map[string]*sessionState),
	}, nil
}

func (s *Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	enc := json.NewEncoder(out)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var req request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			if err := enc.Encode(errorResponse(nil, -32700, "parse error")); err != nil {
				return err
			}
			continue
		}
		if req.JSONRPC != "2.0" || req.Method == "" {
			if err := enc.Encode(errorResponse(req.ID, -32600, "invalid request")); err != nil {
				return err
			}
			continue
		}

		resp, notifications := s.handle(ctx, req)
		for _, n := range notifications {
			if err := enc.Encode(n); err != nil {
				return err
			}
		}
		if req.ID != nil && resp != nil {
			if err := enc.Encode(resp); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func (s *Server) handle(ctx context.Context, req request) (*response, []notification) {
	switch req.Method {
	case "initialize":
		return resultResponse(req.ID, map[string]any{
			"protocolVersion": protocolVersion,
			"agent": map[string]any{
				"name":         "butter-acp",
				"description":  "Butter ACP adapter",
				"capabilities": []string{"text"},
			},
			"capabilities": map[string]any{
				"sessions": true,
			},
		}), nil
	case "session/new":
		id := uuid.NewString()
		s.mu.Lock()
		s.sessions[id] = &sessionState{}
		s.mu.Unlock()
		return resultResponse(req.ID, map[string]any{"sessionId": id}), nil
	case "session/prompt":
		var params promptParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return errorResponse(req.ID, -32602, "invalid session/prompt params"), nil
		}
		if params.SessionID == "" {
			return errorResponse(req.ID, -32602, "sessionId is required"), nil
		}
		input := extractPromptText(params)
		if strings.TrimSpace(input) == "" {
			return errorResponse(req.ID, -32602, "prompt text is required"), nil
		}

		promptCtx, cancel := context.WithCancel(ctx)
		state := s.ensureSession(params.SessionID)
		state.cancel = cancel
		defer func() {
			s.mu.Lock()
			if cur := s.sessions[params.SessionID]; cur != nil {
				cur.cancel = nil
			}
			s.mu.Unlock()
		}()

		invokeResp, err := s.client.Invoke(promptCtx, InvokeRequest{
			AgentName:     s.cfg.AgentName,
			Input:         input,
			AppName:       s.cfg.AppName,
			UserID:        s.cfg.UserID,
			SessionID:     firstNonEmpty(params.ButterSessionID, state.butterSessionID, params.SessionID),
			ModelOverride: firstNonEmpty(params.ModelOverride, s.cfg.ModelOverride),
		})
		if err != nil {
			return errorResponse(req.ID, -32000, err.Error()), nil
		}
		if invokeResp.SessionID != "" {
			state.butterSessionID = invokeResp.SessionID
		}
		update := notification{
			JSONRPC: "2.0",
			Method:  "session/update",
			Params: map[string]any{
				"sessionId": params.SessionID,
				"content": []map[string]string{{
					"type": "text",
					"text": invokeResp.Response,
				}},
			},
		}
		return resultResponse(req.ID, map[string]any{
			"stopReason": "end_turn",
			"content": []map[string]string{{
				"type": "text",
				"text": invokeResp.Response,
			}},
		}), []notification{update}
	case "session/cancel":
		var params struct {
			SessionID string `json:"sessionId"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil || params.SessionID == "" {
			return errorResponse(req.ID, -32602, "sessionId is required"), nil
		}
		s.mu.Lock()
		if state := s.sessions[params.SessionID]; state != nil && state.cancel != nil {
			state.cancel()
			state.cancel = nil
		}
		s.mu.Unlock()
		return resultResponse(req.ID, map[string]any{}), nil
	default:
		return errorResponse(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method)), nil
	}
}

func (s *Server) ensureSession(id string) *sessionState {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.sessions[id]
	if state == nil {
		state = &sessionState{}
		s.sessions[id] = state
	}
	return state
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *responseError  `json:"error,omitempty"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type promptParams struct {
	SessionID       string          `json:"sessionId"`
	Prompt          string          `json:"prompt"`
	Input           string          `json:"input"`
	Text            string          `json:"text"`
	Content         json.RawMessage `json:"content"`
	Messages        []message       `json:"messages"`
	ButterSessionID string          `json:"butterSessionId"`
	ModelOverride   string          `json:"modelOverride"`
}

type message struct {
	Content json.RawMessage `json:"content"`
}

func resultResponse(id json.RawMessage, result any) *response {
	return &response{JSONRPC: "2.0", ID: id, Result: result}
}

func errorResponse(id json.RawMessage, code int, message string) *response {
	return &response{JSONRPC: "2.0", ID: id, Error: &responseError{Code: code, Message: message}}
}

func extractPromptText(params promptParams) string {
	for _, value := range []string{params.Prompt, params.Input, params.Text} {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	if text := textFromRawContent(params.Content); text != "" {
		return text
	}
	var parts []string
	for _, msg := range params.Messages {
		if text := textFromRawContent(msg.Content); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}

func textFromRawContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}
	var block struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &block); err == nil && strings.TrimSpace(block.Text) != "" {
		return block.Text
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var parts []string
		for _, block := range blocks {
			if strings.TrimSpace(block.Text) != "" {
				parts = append(parts, block.Text)
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
