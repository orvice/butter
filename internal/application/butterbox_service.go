package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"

	"go.orx.me/apps/butter/pkg/proto/butterbox/cursor/v1"
	"go.orx.me/apps/butter/pkg/proto/butterbox/cursor/v1/cursorv1connect"

	"butterfly.orx.me/core/log"
	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// butterBoxProbeTimeout bounds the status/catalog calls to a box. Session
// turns are long; these read-only probes are not.
const butterBoxProbeTimeout = 15 * time.Second

// ButterBoxServiceServer implements agentsv1connect.ButterBoxServiceHandler.
// ButterBoxes are workspace-scoped agent VMs (ADR-0011): the resource holds
// name/URL/enabled, the access token lives behind the secretbox credential
// seam, and status/catalog reads are proxied to the box's PiService.
type ButterBoxServiceServer struct {
	repo    butterboxrepo.Repository
	keyring *secretbox.Keyring
	// agentRepo backs the delete reference guard: a box a PI agent still
	// points at cannot be removed.
	agentRepo configrepo.AgentRepository
}

func NewButterBoxServiceServer(repo butterboxrepo.Repository) *ButterBoxServiceServer {
	return &ButterBoxServiceServer{repo: repo}
}

// SetRepo wires the repository after bootstrap.
func (s *ButterBoxServiceServer) SetRepo(repo butterboxrepo.Repository) { s.repo = repo }

// SetKeyring wires credential encryption after bootstrap. Without it, token
// writes are refused (a nil keyring fails closed).
func (s *ButterBoxServiceServer) SetKeyring(k *secretbox.Keyring) { s.keyring = k }

// SetAgentRepo wires the agent repository used by the delete reference guard.
func (s *ButterBoxServiceServer) SetAgentRepo(repo configrepo.AgentRepository) { s.agentRepo = repo }

func (s *ButterBoxServiceServer) requireRepo() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("butterbox repository not configured"))
	}
	return nil
}

func (s *ButterBoxServiceServer) requireKeyring() error {
	if s.keyring == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("credential encryption is not configured"))
	}
	return nil
}

func mapButterBoxErr(err error) *connect.Error {
	if errors.Is(err, butterboxrepo.ErrNotFound) {
		return connectx.NotFound(err.Error())
	}
	if errors.Is(err, butterboxrepo.ErrAlreadyExists) {
		return connect.NewError(connect.CodeAlreadyExists, err)
	}
	return connectx.InternalWith(err)
}

// validateButterBoxFields checks operator-supplied fields. The base URL must
// be an absolute http(s) URL; anything else is rejected before storage.
func validateButterBoxFields(name, baseURL string) error {
	if strings.TrimSpace(name) == "" {
		return connectx.RequiredArgument("name")
	}
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return connectx.InvalidArgument("base_url", "must be an absolute http(s) URL")
	}
	return nil
}

func normalizeBaseURL(baseURL string) string {
	return strings.TrimRight(strings.TrimSpace(baseURL), "/")
}

// encryptToken turns a write-only token into a storable credential. An empty
// token yields an unset credential (used to clear).
func (s *ButterBoxServiceServer) encryptToken(ctx context.Context, token string) (butterboxrepo.Credential, error) {
	if token == "" {
		return butterboxrepo.Credential{}, nil
	}
	if err := s.requireKeyring(); err != nil {
		return butterboxrepo.Credential{}, err
	}
	ciphertext, keyID, err := s.keyring.Encrypt(ctx, []byte(token))
	if err != nil {
		return butterboxrepo.Credential{}, connectx.InternalWith(fmt.Errorf("encrypt butterbox token: %w", err))
	}
	return butterboxrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}, nil
}

// piClient builds a PiService client for one box, carrying the box's
// decrypted token as a bearer credential. An unset credential yields an
// unauthenticated client (a box without MCP_AUTH_TOKEN accepts that).
func (s *ButterBoxServiceServer) piClient(ctx context.Context, workspaceID string, box *agentsv1.ButterBox) (piv1connect.PiServiceClient, error) {
	token := ""
	if box.GetCredentialSet() {
		if err := s.requireKeyring(); err != nil {
			return nil, err
		}
		cred, err := s.repo.GetCredential(ctx, workspaceID, box.GetId())
		if err != nil {
			return nil, mapButterBoxErr(err)
		}
		plaintext, err := s.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
		if err != nil {
			return nil, connectx.InternalWith(fmt.Errorf("decrypt butterbox token: %w", err))
		}
		token = string(plaintext)
	}

	httpClient := &http.Client{Timeout: butterBoxProbeTimeout}
	opts := []connect.ClientOption{}
	if token != "" {
		opts = append(opts, connect.WithInterceptors(bearerInterceptor(token)))
	}
	return piv1connect.NewPiServiceClient(httpClient, box.GetBaseUrl(), opts...), nil
}

// cursorClient builds a CursorService client for one box, carrying the box's
// decrypted token as a bearer credential — same shape as piClient.
func (s *ButterBoxServiceServer) cursorClient(ctx context.Context, workspaceID string, box *agentsv1.ButterBox) (cursorv1connect.CursorServiceClient, error) {
	token := ""
	if box.GetCredentialSet() {
		if err := s.requireKeyring(); err != nil {
			return nil, err
		}
		cred, err := s.repo.GetCredential(ctx, workspaceID, box.GetId())
		if err != nil {
			return nil, mapButterBoxErr(err)
		}
		plaintext, err := s.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
		if err != nil {
			return nil, connectx.InternalWith(fmt.Errorf("decrypt butterbox token: %w", err))
		}
		token = string(plaintext)
	}

	httpClient := &http.Client{Timeout: butterBoxProbeTimeout}
	opts := []connect.ClientOption{}
	if token != "" {
		opts = append(opts, connect.WithInterceptors(bearerInterceptor(token)))
	}
	return cursorv1connect.NewCursorServiceClient(httpClient, box.GetBaseUrl(), opts...), nil
}

func bearerInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

func (s *ButterBoxServiceServer) ListButterBoxes(ctx context.Context, _ *connect.Request[agentsv1.ListButterBoxesRequest]) (*connect.Response[agentsv1.ListButterBoxesResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	boxes, err := s.repo.List(ctx, workspaceID)
	if err != nil {
		return nil, mapButterBoxErr(err)
	}
	return connect.NewResponse(&agentsv1.ListButterBoxesResponse{ButterBoxes: boxes}), nil
}

func (s *ButterBoxServiceServer) GetButterBox(ctx context.Context, req *connect.Request[agentsv1.GetButterBoxRequest]) (*connect.Response[agentsv1.GetButterBoxResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	box, err := s.repo.Get(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		return nil, mapButterBoxErr(err)
	}
	return connect.NewResponse(&agentsv1.GetButterBoxResponse{ButterBox: box}), nil
}

func (s *ButterBoxServiceServer) CreateButterBox(ctx context.Context, req *connect.Request[agentsv1.CreateButterBoxRequest]) (*connect.Response[agentsv1.CreateButterBoxResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateButterBoxFields(req.Msg.GetName(), req.Msg.GetBaseUrl()); err != nil {
		return nil, err
	}

	cred, err := s.encryptToken(ctx, req.Msg.GetToken())
	if err != nil {
		return nil, err
	}

	box := &agentsv1.ButterBox{
		Id:      uuid.NewString(),
		Name:    strings.TrimSpace(req.Msg.GetName()),
		BaseUrl: normalizeBaseURL(req.Msg.GetBaseUrl()),
		Enabled: req.Msg.GetEnabled(),
	}
	created, err := s.repo.Create(ctx, workspaceID, box, cred)
	if err != nil {
		return nil, mapButterBoxErr(err)
	}
	log.FromContext(ctx).Info("butterbox created",
		"id", created.GetId(), "name", created.GetName(), "workspace", workspaceID)
	return connect.NewResponse(&agentsv1.CreateButterBoxResponse{ButterBox: created}), nil
}

func (s *ButterBoxServiceServer) UpdateButterBox(ctx context.Context, req *connect.Request[agentsv1.UpdateButterBoxRequest]) (*connect.Response[agentsv1.UpdateButterBoxResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateButterBoxFields(req.Msg.GetName(), req.Msg.GetBaseUrl()); err != nil {
		return nil, err
	}
	box := &agentsv1.ButterBox{
		Id:      req.Msg.GetId(),
		Name:    strings.TrimSpace(req.Msg.GetName()),
		BaseUrl: normalizeBaseURL(req.Msg.GetBaseUrl()),
		Enabled: req.Msg.GetEnabled(),
	}
	updated, err := s.repo.Update(ctx, workspaceID, box)
	if err != nil {
		return nil, mapButterBoxErr(err)
	}
	return connect.NewResponse(&agentsv1.UpdateButterBoxResponse{ButterBox: updated}), nil
}

func (s *ButterBoxServiceServer) DeleteButterBox(ctx context.Context, req *connect.Request[agentsv1.DeleteButterBoxRequest]) (*connect.Response[agentsv1.DeleteButterBoxResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	// A PI agent whose box vanished is an agent that silently stops working;
	// refuse the delete and name the agents instead (same contract as the
	// Telegram reference guard). Tombstoned agents count too — restoring one
	// must not resurrect a dangling box reference.
	if err := s.checkBoxRemovable(ctx, workspaceID, req.Msg.GetId()); err != nil {
		return nil, err
	}
	if err := s.repo.Delete(ctx, workspaceID, req.Msg.GetId()); err != nil {
		return nil, mapButterBoxErr(err)
	}
	log.FromContext(ctx).Info("butterbox deleted", "id", req.Msg.GetId(), "workspace", workspaceID)
	return connect.NewResponse(&agentsv1.DeleteButterBoxResponse{}), nil
}

// checkBoxRemovable returns a FailedPrecondition error naming every PI or
// Cursor agent (any lifecycle status) whose config binds the box.
func (s *ButterBoxServiceServer) checkBoxRemovable(ctx context.Context, workspaceID, boxID string) error {
	if s.agentRepo == nil {
		return nil
	}
	agents, err := s.agentRepo.ListAgents(ctx, workspaceID)
	if err != nil {
		return connectx.InternalWith(err)
	}
	var refs []string
	for _, a := range agents {
		switch a.GetType() {
		case agentsv1.AgentType_AGENT_TYPE_PI:
			if strings.TrimSpace(a.GetConfig().GetPi().GetButterboxId()) == strings.TrimSpace(boxID) {
				refs = append(refs, a.GetAgentId())
			}
		case agentsv1.AgentType_AGENT_TYPE_CURSOR:
			if strings.TrimSpace(a.GetConfig().GetCursor().GetButterboxId()) == strings.TrimSpace(boxID) {
				refs = append(refs, a.GetAgentId())
			}
		}
	}
	if len(refs) == 0 {
		return nil
	}
	sort.Strings(refs)
	return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf(
		"butterbox %q is referenced by agents: %s; repoint or delete them first", boxID, strings.Join(refs, ", ")))
}

func (s *ButterBoxServiceServer) SetButterBoxToken(ctx context.Context, req *connect.Request[agentsv1.SetButterBoxTokenRequest]) (*connect.Response[agentsv1.SetButterBoxTokenResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}

	cred, err := s.encryptToken(ctx, req.Msg.GetToken())
	if err != nil {
		return nil, err
	}
	updated, err := s.repo.SetCredential(ctx, workspaceID, req.Msg.GetId(), cred)
	if err != nil {
		return nil, mapButterBoxErr(err)
	}
	log.FromContext(ctx).Info("butterbox token updated",
		"id", req.Msg.GetId(), "workspace", workspaceID, "set", cred.Set())
	return connect.NewResponse(&agentsv1.SetButterBoxTokenResponse{ButterBox: updated}), nil
}

func (s *ButterBoxServiceServer) GetButterBoxStatus(ctx context.Context, req *connect.Request[agentsv1.GetButterBoxStatusRequest]) (*connect.Response[agentsv1.GetButterBoxStatusResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	box, err := s.repo.Get(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		return nil, mapButterBoxErr(err)
	}
	client, err := s.piClient(ctx, workspaceID, box)
	if err != nil {
		// A broken credential (missing keyring, undecryptable token) is a
		// probe outcome too: the status view must render it, not break.
		return connect.NewResponse(&agentsv1.GetButterBoxStatusResponse{
			Reachable: false,
			Error:     err.Error(),
		}), nil
	}

	probeCtx, cancel := context.WithTimeout(ctx, butterBoxProbeTimeout)
	defer cancel()
	sessions, err := client.ListSessions(probeCtx, connect.NewRequest(&piv1.ListSessionsRequest{}))
	if err != nil {
		// Unreachability is the status, not a failure to report it.
		return connect.NewResponse(&agentsv1.GetButterBoxStatusResponse{
			Reachable: false,
			Error:     err.Error(),
		}), nil
	}
	return connect.NewResponse(&agentsv1.GetButterBoxStatusResponse{
		Reachable:      true,
		ActiveSessions: int32(len(sessions.Msg.GetSessions())),
	}), nil
}

func (s *ButterBoxServiceServer) ListButterBoxModels(ctx context.Context, req *connect.Request[agentsv1.ListButterBoxModelsRequest]) (*connect.Response[agentsv1.ListButterBoxModelsResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	box, err := s.repo.Get(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		return nil, mapButterBoxErr(err)
	}
	client, err := s.piClient(ctx, workspaceID, box)
	if err != nil {
		return nil, err
	}

	probeCtx, cancel := context.WithTimeout(ctx, butterBoxProbeTimeout)
	defer cancel()
	// Session-less: the box answers from a short-lived pi process and caches
	// the catalog (butter-box#10).
	catalog, err := client.GetAvailableModels(probeCtx, connect.NewRequest(&piv1.GetAvailableModelsRequest{}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeResourceExhausted {
			return nil, connect.NewError(connect.CodeResourceExhausted,
				fmt.Errorf("butterbox %q is at its session capacity; retry later or raise PI_API_MAX_SESSIONS on the box: %w", box.GetName(), err))
		}
		return nil, connect.NewError(connect.CodeUnavailable,
			fmt.Errorf("butterbox %q did not answer the model catalog; check the box is running and the token is valid: %w", box.GetName(), err))
	}
	models := make([]*agentsv1.ButterBoxModel, 0, len(catalog.Msg.GetModels()))
	for _, m := range catalog.Msg.GetModels() {
		models = append(models, &agentsv1.ButterBoxModel{
			Id:        m.GetId(),
			Provider:  m.GetProvider(),
			Name:      m.GetName(),
			Reasoning: m.GetReasoning(),
		})
	}
	return connect.NewResponse(&agentsv1.ListButterBoxModelsResponse{Models: models}), nil
}

func (s *ButterBoxServiceServer) ListCursorModels(ctx context.Context, req *connect.Request[agentsv1.ListCursorModelsRequest]) (*connect.Response[agentsv1.ListCursorModelsResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	box, err := s.repo.Get(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		return nil, mapButterBoxErr(err)
	}
	client, err := s.cursorClient(ctx, workspaceID, box)
	if err != nil {
		return nil, err
	}

	probeCtx, cancel := context.WithTimeout(ctx, butterBoxProbeTimeout)
	defer cancel()
	// Session-less: the box answers from a short-lived bridge and shuts it
	// down, so this never consumes a session slot (butter-box #315).
	catalog, err := client.ListModels(probeCtx, connect.NewRequest(&cursorv1.ListModelsRequest{}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeResourceExhausted {
			return nil, connect.NewError(connect.CodeResourceExhausted,
				fmt.Errorf("butterbox %q is at its session capacity; retry later or raise CURSOR_MAX_SESSIONS on the box: %w", box.GetName(), err))
		}
		return nil, connect.NewError(connect.CodeUnavailable,
			fmt.Errorf("butterbox %q did not answer the cursor model catalog; check the box is running, the token is valid, and CURSOR_API_KEY is configured on the box: %w", box.GetName(), err))
	}
	models := make([]*agentsv1.CursorBoxModel, 0, len(catalog.Msg.GetModels()))
	for _, m := range catalog.Msg.GetModels() {
		models = append(models, &agentsv1.CursorBoxModel{
			Id:   m.GetId(),
			Name: m.GetName(),
		})
	}
	return connect.NewResponse(&agentsv1.ListCursorModelsResponse{Models: models}), nil
}
