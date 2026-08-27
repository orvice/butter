package cursorbox

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	"go.orx.me/apps/butter/internal/secretbox"
)

// ClientFactory resolves a workspace's ButterBox into a CursorClient.
// Implementations resolve per call — no client cache — so a base-URL change
// or token rotation takes effect on the next turn.
type ClientFactory interface {
	ClientFor(ctx context.Context, workspaceID, butterboxID string) (CursorClient, error)
}

// Factory is the production ClientFactory: it reads the box from the
// repository and decrypts its access token through the secretbox keyring,
// then builds a Connect-based CursorClient targeting the box's CursorService.
type Factory struct {
	repo       butterboxrepo.Repository
	keyring    *secretbox.Keyring
	httpClient connect.HTTPClient
}

func NewFactory(repo butterboxrepo.Repository, keyring *secretbox.Keyring) *Factory {
	return &Factory{repo: repo, keyring: keyring}
}

func (f *Factory) ClientFor(ctx context.Context, workspaceID, butterboxID string) (CursorClient, error) {
	if f == nil || f.repo == nil {
		return nil, fmt.Errorf("cursorbox: butterbox repository is not configured")
	}
	box, err := f.repo.Get(ctx, workspaceID, butterboxID)
	if err != nil {
		if errors.Is(err, butterboxrepo.ErrNotFound) {
			return nil, fmt.Errorf("cursorbox: butterbox %q no longer exists in this workspace; point the agent at a registered ButterBox", butterboxID)
		}
		return nil, fmt.Errorf("cursorbox: resolve butterbox %q: %w", butterboxID, err)
	}

	token := ""
	if box.GetCredentialSet() {
		if f.keyring == nil {
			return nil, fmt.Errorf("cursorbox: butterbox %q has a stored token but credential encryption is not configured", box.GetName())
		}
		cred, err := f.repo.GetCredential(ctx, workspaceID, butterboxID)
		if err != nil {
			return nil, fmt.Errorf("cursorbox: read butterbox %q credential: %w", box.GetName(), err)
		}
		plaintext, err := f.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
		if err != nil {
			return nil, fmt.Errorf("cursorbox: decrypt butterbox %q token: %w", box.GetName(), err)
		}
		token = string(plaintext)
	}

	httpClient := f.httpClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return newConnectCursorClient(httpClient, box.GetBaseUrl(), token), nil
}

// connectCursorClient wraps HTTP calls to the box's CursorService endpoint
// using plain Connect unary RPCs. When butter-box publishes the cursor.v1
// proto (issue #315), this implementation should be replaced by the generated
// Connect client.
type connectCursorClient struct {
	baseURL    string
	httpClient connect.HTTPClient
	opts       []connect.ClientOption
}

func newConnectCursorClient(httpClient connect.HTTPClient, baseURL, token string) *connectCursorClient {
	var opts []connect.ClientOption
	if token != "" {
		opts = append(opts, connect.WithInterceptors(bearerInterceptor(token)))
	}
	return &connectCursorClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		opts:       opts,
	}
}

func (c *connectCursorClient) CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	type wireReq struct {
		Name       string `json:"name,omitempty"`
		WorkingDir string `json:"working_dir,omitempty"`
		Model      string `json:"model,omitempty"`
		Mode       string `json:"mode,omitempty"`
	}
	type wireSession struct {
		ID string `json:"id"`
	}
	type wireResp struct {
		Session wireSession `json:"session"`
	}
	client := connect.NewClient[wireReq, wireResp](
		c.httpClient,
		c.baseURL+"/butterbox.cursor.v1.CursorService/CreateSession",
		c.opts...,
	)
	resp, err := client.CallUnary(ctx, connect.NewRequest(&wireReq{
		Name:       req.Name,
		WorkingDir: req.WorkingDir,
		Model:      req.Model,
		Mode:       req.Mode,
	}))
	if err != nil {
		return nil, err
	}
	return &CreateSessionResponse{SessionID: resp.Msg.Session.ID}, nil
}

func (c *connectCursorClient) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	type wireImage struct {
		MIMEType string `json:"mime_type,omitempty"`
		Data     []byte `json:"data,omitempty"`
	}
	type wireReq struct {
		SessionID string      `json:"session_id"`
		Message   string      `json:"message"`
		Images    []wireImage `json:"images,omitempty"`
	}
	type wireResp struct {
		Text string `json:"text"`
	}
	var wireImages []wireImage
	for _, img := range req.Images {
		wireImages = append(wireImages, wireImage{MIMEType: img.MIMEType, Data: img.Data})
	}
	client := connect.NewClient[wireReq, wireResp](
		c.httpClient,
		c.baseURL+"/butterbox.cursor.v1.CursorService/SendMessage",
		c.opts...,
	)
	resp, err := client.CallUnary(ctx, connect.NewRequest(&wireReq{
		SessionID: req.SessionID,
		Message:   req.Message,
		Images:    wireImages,
	}))
	if err != nil {
		return nil, err
	}
	return &SendMessageResponse{Text: resp.Msg.Text}, nil
}

func (c *connectCursorClient) AbortSession(ctx context.Context, req *AbortSessionRequest) error {
	type wireReq struct {
		SessionID string `json:"session_id"`
	}
	type wireResp struct{}
	client := connect.NewClient[wireReq, wireResp](
		c.httpClient,
		c.baseURL+"/butterbox.cursor.v1.CursorService/AbortSession",
		c.opts...,
	)
	_, err := client.CallUnary(ctx, connect.NewRequest(&wireReq{SessionID: req.SessionID}))
	return err
}

func bearerInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}
