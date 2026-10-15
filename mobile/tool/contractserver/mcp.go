package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"go.orx.me/apps/butter/internal/application"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/mcpoauth"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	mcpoauthmemory "go.orx.me/apps/butter/internal/repo/mcpoauth/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// The MCP and Remote Agent fixture (#454): the production MCPServerService,
// GlobalMCPServerService, RemoteAgentService, MCP OAuth service, token
// resolver and public OAuth callback, over in-memory storage. Only the things
// outside Butter are test doubles served on the fixture's own listener:
//
//	/fake-mcp/<profile>        a real MCP server built on the MCP Go SDK
//	/fake-as/                  an OAuth authorization server (PKCE S256)
//	/fake-a2a/<profile>/       an A2A agent card endpoint
//	/fake-opencode/<profile>/  an opencode server's health endpoint
//
// The "browser" of an OAuth flow is the contract test itself: it follows the
// authorization URL to /fake-as/authorize and then to Butter's public
// callback, exactly the two hops a system browser makes. None of this is a
// production MCP server, identity provider or a device run.
//
// MCP profiles (the URL path chooses the behavior):
//
//	open    no authentication; tools echo and danger
//	bearer  requires the static header "Authorization: Bearer mcpStaticToken"
//	oauth   requires the access token the fake authorization server issued
//	sse     no authentication, served over the SSE transport
//	broken  always answers 500
//
// OAuth scopes steer the authorization server: a scope containing "deny" makes
// the user refuse consent, and one containing "tokenfail" makes the token
// endpoint refuse the code.
const (
	mcpWorkspace       = "mcp"
	mcpDeniedWorkspace = "mcp-denied"
	mcpStaticToken     = "mcp-static-secret"
	mcpAccessToken     = "mcp-oauth-access-token"
	mcpClientID        = "contract-client"
	mcpClientSecret    = "contract-client-secret"
	mcpOpencodeUser    = "opencode-user"
	mcpOpencodePass    = "opencode-secret"
	// mcpEncryptionKey encrypts stored OAuth tokens; public test data.
	mcpEncryptionKey = "0123456789abcdef0123456789abcdef"
)

type mcpFixture struct {
	mcpSvc    *application.MCPServerServiceServer
	globalSvc *application.GlobalMCPServerServiceServer
	remoteSvc *application.RemoteAgentServiceServer
	callback  *httpHandler.MCPOAuthCallbackHandler

	// publicBase is the address a client reaches the fixture at, learned from
	// the first request: the OAuth callback URL and the preset URLs need it,
	// and the listener's port is only known once serving started.
	publicBase atomic.Value
	seed       sync.Once
	agentRepo  *configmemory.Store
	as         *fakeAuthServer
}

func newMCPFixture(agentRepo *configmemory.Store) *mcpFixture {
	f := &mcpFixture{agentRepo: agentRepo, as: newFakeAuthServer()}
	f.publicBase.Store("")
	config := func() mcpoauth.Config {
		return mcpoauth.Config{
			CallbackBaseURL:   f.publicBase.Load().(string),
			EncryptionKey:     mcpEncryptionKey,
			AllowInsecureHTTP: true,
		}
	}
	connections := mcpoauthmemory.New()
	f.mcpSvc = application.NewMCPServerServiceServer(agentRepo)
	f.mcpSvc.SetOAuthService(mcpoauth.NewService(connections, mcpoauth.NewMemoryFlowStore(), config))
	f.mcpSvc.SetMCPHTTPClientFactory(mcpoauth.NewResolver(connections, config))
	f.globalSvc = application.NewGlobalMCPServerServiceServer(agentRepo, f.mcpSvc)
	f.remoteSvc = application.NewRemoteAgentServiceServer(agentRepo)
	f.callback = httpHandler.NewMCPOAuthCallbackHandler(f.mcpSvc, func() string { return "" })
	return f
}

// seedPresets stores the global presets once the fixture's address is known.
// Global presets have no public write API here (administration is out of the
// App's scope), so they are put straight into the store.
func (f *mcpFixture) seedPresets(base string) {
	f.seed.Do(func() {
		ctx := context.Background()
		for _, preset := range []*agentsv1.MCPServer{
			{
				Id: "preset-open", Name: "Open preset",
				Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
				Url:       base + "/fake-mcp/open",
			},
			{
				Id: "preset-secret", Name: "Secret preset",
				Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
				Url:       base + "/fake-mcp/bearer",
				Headers:   map[string]string{"Authorization": "Bearer " + mcpStaticToken},
				Auth:      &agentsv1.MCPServerAuth{Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_STATIC_HEADERS},
			},
			{
				Id: "preset-oauth", Name: "OAuth preset",
				Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
				Url:       base + "/fake-mcp/oauth",
				Auth: &agentsv1.MCPServerAuth{
					Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2,
					Oauth2: &agentsv1.MCPServerOAuth2Config{
						ClientId: mcpClientID, ClientSecret: mcpClientSecret,
						AuthorizationUrl: base + "/fake-as/authorize",
						TokenUrl:         base + "/fake-as/token",
					},
				},
			},
		} {
			if _, err := f.agentRepo.CreateGlobalMCPServer(ctx, preset); err != nil {
				panic(err)
			}
		}
	})
}

// register mounts the production services and the doubles. Only the listed
// procedures are reachable; global preset writes are deliberately absent.
func (f *mcpFixture) register(r *gin.Engine, mux *http.ServeMux, opts []connect.HandlerOption) {
	_, mcpHandler := agentsv1connect.NewMCPServerServiceHandler(f.mcpSvc, opts...)
	for _, procedure := range []string{
		agentsv1connect.MCPServerServiceListMCPServersProcedure,
		agentsv1connect.MCPServerServiceGetMCPServerProcedure,
		agentsv1connect.MCPServerServiceCreateMCPServerProcedure,
		agentsv1connect.MCPServerServiceUpdateMCPServerProcedure,
		agentsv1connect.MCPServerServiceDeleteMCPServerProcedure,
		agentsv1connect.MCPServerServiceGetMCPServerStatusProcedure,
		agentsv1connect.MCPServerServiceListMCPToolsProcedure,
		agentsv1connect.MCPServerServiceStartMCPServerOAuthProcedure,
		agentsv1connect.MCPServerServiceGetMCPServerOAuthStatusProcedure,
		agentsv1connect.MCPServerServiceDisconnectMCPServerOAuthProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", mcpHandler)))
	}
	_, globalHandler := agentsv1connect.NewGlobalMCPServerServiceHandler(f.globalSvc, opts...)
	for _, procedure := range []string{
		agentsv1connect.GlobalMCPServerServiceListGlobalMCPServersProcedure,
		agentsv1connect.GlobalMCPServerServiceInstallGlobalMCPServerProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", globalHandler)))
	}
	_, remoteHandler := agentsv1connect.NewRemoteAgentServiceHandler(f.remoteSvc, opts...)
	for _, procedure := range []string{
		agentsv1connect.RemoteAgentServiceListRemoteAgentsProcedure,
		agentsv1connect.RemoteAgentServiceGetRemoteAgentProcedure,
		agentsv1connect.RemoteAgentServiceCreateRemoteAgentProcedure,
		agentsv1connect.RemoteAgentServiceUpdateRemoteAgentProcedure,
		agentsv1connect.RemoteAgentServiceDeleteRemoteAgentProcedure,
		agentsv1connect.RemoteAgentServiceGetRemoteAgentStatusProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", remoteHandler)))
	}
	f.callback.Register(r)

	mux.Handle("/fake-mcp/", mcpDouble())
	mux.Handle("/fake-as/", f.as.handler())
	mux.Handle("/fake-a2a/", a2aDouble())
	mux.Handle("/fake-opencode/", opencodeDouble())
}

// wrap records the address clients use and seeds the presets before the first
// request is served.
func (f *mcpFixture) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := "http://" + r.Host
		f.publicBase.Store(base)
		f.seedPresets(base)
		next.ServeHTTP(w, r)
	})
}

func newMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "contract-mcp", Version: "1.0.0"}, nil)
	type args struct {
		Text string `json:"text"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "echo", Description: "Echo the given text."},
		func(_ context.Context, _ *mcp.CallToolRequest, in args) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: in.Text}}}, nil, nil
		})
	mcp.AddTool(server, &mcp.Tool{Name: "danger", Description: "A tool a filter should hide."},
		func(_ context.Context, _ *mcp.CallToolRequest, _ args) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "danger"}}}, nil, nil
		})
	return server
}

func mcpDouble() http.Handler {
	mux := http.NewServeMux()
	streamable := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return newMCPServer() },
		&mcp.StreamableHTTPOptions{Stateless: true})
	sse := mcp.NewSSEHandler(func(*http.Request) *mcp.Server { return newMCPServer() }, nil)
	guard := func(want string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != want {
				w.Header().Set("WWW-Authenticate", `Bearer realm="fake-mcp"`)
				http.Error(w, "credentials missing or wrong", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	mux.Handle("/fake-mcp/open", streamable)
	mux.Handle("/fake-mcp/bearer", guard("Bearer "+mcpStaticToken, streamable))
	mux.Handle("/fake-mcp/oauth", guard("Bearer "+mcpAccessToken, streamable))
	mux.Handle("/fake-mcp/sse", sse)
	mux.HandleFunc("/fake-mcp/broken", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error with secret detail sk-should-never-be-shown", http.StatusInternalServerError)
	})
	return mux
}

func a2aDouble() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/fake-a2a/ok/.well-known/agent.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"contract-a2a"}`))
	})
	mux.HandleFunc("/fake-a2a/gone/.well-known/agent.json", func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	})
	return mux
}

func opencodeDouble() http.Handler {
	mux := http.NewServeMux()
	health := func(requireAuth bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if requireAuth {
				user, pass, ok := r.BasicAuth()
				if !ok || user != mcpOpencodeUser || pass != mcpOpencodePass {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"healthy":true,"version":"1.0.0"}`))
		}
	}
	mux.HandleFunc("/fake-opencode/ok/global/health", health(true))
	mux.HandleFunc("/fake-opencode/open/global/health", health(false))
	mux.HandleFunc("/fake-opencode/broken/global/health", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	return mux
}

// fakeAuthServer is the OAuth authorization server of the "oauth" MCP
// profile. It enforces PKCE S256 and single-use codes, the parts of the
// protocol Butter's own code is responsible for sending correctly.
type fakeAuthServer struct {
	mu    sync.Mutex
	codes map[string]authCode
}

type authCode struct {
	challenge string
	redirect  string
	scope     string
}

func newFakeAuthServer() *fakeAuthServer { return &fakeAuthServer{codes: map[string]authCode{}} }

func (a *fakeAuthServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/fake-as/authorize", a.authorize)
	mux.HandleFunc("/fake-as/token", a.token)
	return mux
}

func (a *fakeAuthServer) authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	redirect, err := url.Parse(q.Get("redirect_uri"))
	if err != nil || redirect.Host == "" || q.Get("client_id") != mcpClientID ||
		q.Get("response_type") != "code" || q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" {
		http.Error(w, "invalid authorization request", http.StatusBadRequest)
		return
	}
	out := redirect.Query()
	out.Set("state", q.Get("state"))
	if strings.Contains(q.Get("scope"), "deny") {
		out.Set("error", "access_denied")
	} else {
		raw := make([]byte, 12)
		_, _ = rand.Read(raw)
		code := hex.EncodeToString(raw)
		a.mu.Lock()
		a.codes[code] = authCode{challenge: q.Get("code_challenge"), redirect: q.Get("redirect_uri"), scope: q.Get("scope")}
		a.mu.Unlock()
		out.Set("code", code)
	}
	redirect.RawQuery = out.Encode()
	http.Redirect(w, r, redirect.String(), http.StatusFound)
}

func (a *fakeAuthServer) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	id, secret, basic := r.BasicAuth()
	if !basic {
		id, secret = r.PostForm.Get("client_id"), r.PostForm.Get("client_secret")
	}
	refuse := func(status int, reason string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": reason})
	}
	if id != mcpClientID || secret != mcpClientSecret {
		refuse(http.StatusUnauthorized, "invalid_client")
		return
	}
	a.mu.Lock()
	issued, ok := a.codes[r.PostForm.Get("code")]
	delete(a.codes, r.PostForm.Get("code"))
	a.mu.Unlock()
	sum := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
	if r.PostForm.Get("grant_type") != "authorization_code" || !ok ||
		issued.challenge != base64.RawURLEncoding.EncodeToString(sum[:]) ||
		issued.redirect != r.PostForm.Get("redirect_uri") || strings.Contains(issued.scope, "tokenfail") {
		refuse(http.StatusBadRequest, "invalid_grant")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": mcpAccessToken, "token_type": "Bearer", "expires_in": 3600, "refresh_token": "mcp-refresh-token",
	})
}
