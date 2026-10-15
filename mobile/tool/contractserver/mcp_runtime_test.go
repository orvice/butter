package main

// The mobile MCP and Remote Agent screens (#454) talk only to the Butter
// server. These tests drive the production MCPServerService,
// GlobalMCPServerService, RemoteAgentService, MCP OAuth service and public
// callback through the fixture's HTTP handlers, against the doubles in mcp.go
// (a real MCP server on the MCP Go SDK, an OAuth authorization server, A2A and
// opencode endpoints). They show what the App's configuration does on real
// protocol calls; they are not a production MCP server or a device run.

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type mcpHarness struct {
	*providerFixture
	mcp    agentsv1connect.MCPServerServiceClient
	global agentsv1connect.GlobalMCPServerServiceClient
	remote agentsv1connect.RemoteAgentServiceClient
}

func newMCPHarness(t *testing.T, workspaceID string) *mcpHarness {
	t.Helper()
	f := newProviderFixture(t)
	interceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+f.token)
			req.Header().Set("X-Workspace-ID", workspaceID)
			return next(ctx, req)
		}
	})
	with := connect.WithInterceptors(interceptor)
	return &mcpHarness{
		providerFixture: f,
		mcp:             agentsv1connect.NewMCPServerServiceClient(http.DefaultClient, f.url+"/api", with),
		global:          agentsv1connect.NewGlobalMCPServerServiceClient(http.DefaultClient, f.url+"/api", with),
		remote:          agentsv1connect.NewRemoteAgentServiceClient(http.DefaultClient, f.url+"/api", with),
	}
}

func (h *mcpHarness) create(t *testing.T, server *agentsv1.MCPServer) *agentsv1.MCPServer {
	t.Helper()
	resp, err := h.mcp.CreateMCPServer(context.Background(), connect.NewRequest(&agentsv1.CreateMCPServerRequest{McpServer: server}))
	if err != nil {
		t.Fatalf("create %s: %v", server.GetId(), err)
	}
	return resp.Msg.GetMcpServer()
}

func (h *mcpHarness) status(t *testing.T, id string) *agentsv1.MCPServerStatus {
	t.Helper()
	resp, err := h.mcp.GetMCPServerStatus(context.Background(), connect.NewRequest(&agentsv1.GetMCPServerStatusRequest{Id: id}))
	if err != nil {
		t.Fatalf("status %s: %v", id, err)
	}
	return resp.Msg.GetStatus()
}

func (h *mcpHarness) tools(t *testing.T, id string) *agentsv1.ListMCPToolsResponse {
	t.Helper()
	resp, err := h.mcp.ListMCPTools(context.Background(), connect.NewRequest(&agentsv1.ListMCPToolsRequest{ServerId: id}))
	if err != nil {
		t.Fatalf("tools %s: %v", id, err)
	}
	return resp.Msg
}

func (h *mcpHarness) oauth(t *testing.T, id string) *agentsv1.MCPOAuthConnectionStatus {
	t.Helper()
	resp, err := h.mcp.GetMCPServerOAuthStatus(context.Background(), connect.NewRequest(&agentsv1.GetMCPServerOAuthStatusRequest{ServerId: id}))
	if err != nil {
		t.Fatalf("oauth status %s: %v", id, err)
	}
	return resp.Msg.GetStatus()
}

func httpServer(id, name, path string) *agentsv1.MCPServer {
	return &agentsv1.MCPServer{
		Id: id, Name: name, Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		Url: "REPLACE" + path,
	}
}

// at points a double's path at this fixture.
func (h *mcpHarness) at(server *agentsv1.MCPServer) *agentsv1.MCPServer {
	server.Url = strings.Replace(server.Url, "REPLACE", h.url, 1)
	return server
}

func oauthServer(h *mcpHarness, id string, scopes ...string) *agentsv1.MCPServer {
	return h.at(&agentsv1.MCPServer{
		Id: id, Name: id, Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		Url: "REPLACE/fake-mcp/oauth",
		Auth: &agentsv1.MCPServerAuth{
			Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2,
			Oauth2: &agentsv1.MCPServerOAuth2Config{
				ClientId: mcpClientID, ClientSecret: mcpClientSecret, Scopes: scopes,
				AuthorizationUrl: h.url + "/fake-as/authorize", TokenUrl: h.url + "/fake-as/token",
			},
		},
	})
}

// browse plays the system browser: it follows the authorization URL to the
// authorization server and then to Butter's public callback, stopping at the
// final redirect, which it returns.
func browse(t *testing.T, authorizationURL string) *url.URL {
	t.Helper()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	next := authorizationURL
	var last *url.URL
	for hop := 0; hop < 2; hop++ {
		resp, err := client.Get(next)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusFound {
			t.Fatalf("hop %d: status %d", hop, resp.StatusCode)
		}
		loc, err := resp.Location()
		if err != nil {
			t.Fatal(err)
		}
		last, next = loc, loc.String()
	}
	return last
}

func TestMCPServerCRUDProbeAndToolFilterAgainstARealMCPServer(t *testing.T) {
	h := newMCPHarness(t, mcpWorkspace)
	ctx := context.Background()

	open := httpServer("open", "Open", "/fake-mcp/open")
	open.ToolFilter = []string{"echo"}
	open.Metadata = map[string]string{"owner": "someone"}
	withUnknown(open, 99, 7)
	created := h.create(t, h.at(open))
	if created.GetWorkspaceId() != mcpWorkspace || created.GetMetadata()["owner"] != "someone" ||
		len(created.ProtoReflect().GetUnknown()) == 0 {
		t.Fatalf("created = %v", created)
	}

	if st := h.status(t, "open"); st.GetState() != agentsv1.MCPServerStatus_STATE_CONNECTED || st.GetToolCount() != 1 {
		t.Fatalf("status with filter = %v", st)
	}
	tools := h.tools(t, "open")
	allowed := map[string]bool{}
	for _, tool := range tools.GetTools() {
		allowed[tool.GetName()] = tool.GetAllowed()
	}
	if len(allowed) != 2 || !allowed["echo"] || allowed["danger"] {
		t.Fatalf("tools = %v errors=%v", allowed, tools.GetErrors())
	}

	// An update replaces the whole object, so what the App does not edit it
	// sends back; this is the proof the sent copy is what is stored.
	edit := proto.Clone(created).(*agentsv1.MCPServer)
	edit.ToolFilter = nil
	edit.Name = "Open renamed"
	updated, err := h.mcp.UpdateMCPServer(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerRequest{McpServer: edit}))
	if err != nil {
		t.Fatal(err)
	}
	stored := updated.Msg.GetMcpServer()
	if stored.GetName() != "Open renamed" || stored.GetMetadata()["owner"] != "someone" || len(stored.ProtoReflect().GetUnknown()) == 0 {
		t.Fatalf("updated = %v", stored)
	}
	if st := h.status(t, "open"); st.GetToolCount() != 2 {
		t.Fatalf("status without filter = %v", st)
	}

	sse := httpServer("sse", "SSE", "/fake-mcp/sse")
	sse.Transport = agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_SSE
	h.create(t, h.at(sse))
	if st := h.status(t, "sse"); st.GetState() != agentsv1.MCPServerStatus_STATE_CONNECTED || st.GetToolCount() != 2 {
		t.Fatalf("sse status = %v", st)
	}

	_, err = h.mcp.CreateMCPServer(ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{McpServer: h.at(httpServer("open", "dup", "/fake-mcp/open"))}))
	wantCode(t, err, connect.CodeAlreadyExists)
	bad := h.at(httpServer("bad", "bad", "/x"))
	bad.Url = "ftp://example.com"
	_, err = h.mcp.CreateMCPServer(ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{McpServer: bad}))
	wantCode(t, err, connect.CodeInvalidArgument)
	bad.Url = h.url
	bad.Transport = agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_UNSPECIFIED
	_, err = h.mcp.CreateMCPServer(ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{McpServer: bad}))
	wantCode(t, err, connect.CodeInvalidArgument)

	if _, err := h.mcp.DeleteMCPServer(ctx, connect.NewRequest(&agentsv1.DeleteMCPServerRequest{Id: "sse"})); err != nil {
		t.Fatal(err)
	}
	_, err = h.mcp.GetMCPServer(ctx, connect.NewRequest(&agentsv1.GetMCPServerRequest{Id: "sse"}))
	wantCode(t, err, connect.CodeNotFound)
	_, err = h.mcp.DeleteMCPServer(ctx, connect.NewRequest(&agentsv1.DeleteMCPServerRequest{Id: "sse"}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestMCPStaticHeaderCredentialsAreTheOnesTheServerSends(t *testing.T) {
	h := newMCPHarness(t, mcpWorkspace)
	ctx := context.Background()

	server := h.at(httpServer("secured", "Secured", "/fake-mcp/bearer"))
	server.Headers = map[string]string{"Authorization": "Bearer wrong"}
	server.Auth = &agentsv1.MCPServerAuth{Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_STATIC_HEADERS}
	h.create(t, server)
	// A wrong credential is an answer, not a transport failure: the probe says
	// disconnected and the per-server tool errors name the server.
	if st := h.status(t, "secured"); st.GetState() != agentsv1.MCPServerStatus_STATE_DISCONNECTED || st.GetDetail() == "" {
		t.Fatalf("wrong header status = %v", st)
	}
	if errs := h.tools(t, "secured").GetErrors(); errs["secured"] == "" {
		t.Fatalf("tool errors = %v", errs)
	}

	fixed := proto.Clone(server).(*agentsv1.MCPServer)
	fixed.Headers = map[string]string{"Authorization": "Bearer " + mcpStaticToken}
	if _, err := h.mcp.UpdateMCPServer(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerRequest{McpServer: fixed})); err != nil {
		t.Fatal(err)
	}
	if st := h.status(t, "secured"); st.GetState() != agentsv1.MCPServerStatus_STATE_CONNECTED {
		t.Fatalf("fixed status = %v", st)
	}
	// A workspace server returns its own headers: the App may keep them.
	got, err := h.mcp.GetMCPServer(ctx, connect.NewRequest(&agentsv1.GetMCPServerRequest{Id: "secured"}))
	if err != nil || got.Msg.GetMcpServer().GetHeaders()["Authorization"] != "Bearer "+mcpStaticToken {
		t.Fatalf("get = %v, %v", got, err)
	}

	h.create(t, h.at(httpServer("broken", "Broken", "/fake-mcp/broken")))
	st := h.status(t, "broken")
	if st.GetState() != agentsv1.MCPServerStatus_STATE_DISCONNECTED || st.GetDetail() == "" {
		t.Fatalf("broken status = %v", st)
	}
	// The probe's own text is whatever the failing server and the transport
	// produced; the App never shows it, so it must classify, not echo.
	t.Logf("probe detail for a broken server (untrusted, never displayed by the App): %q", st.GetDetail())
}

func TestMCPOAuthFlowThroughTheServerCallback(t *testing.T) {
	h := newMCPHarness(t, mcpWorkspace)
	ctx := context.Background()
	h.create(t, oauthServer(h, "oa"))

	before := h.oauth(t, "oa")
	if before.GetState() != agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_DISCONNECTED || before.GetConnectedAt() != nil {
		t.Fatalf("before = %v", before)
	}
	// Not connected yet: the probe is refused by the protected server.
	if st := h.status(t, "oa"); st.GetState() != agentsv1.MCPServerStatus_STATE_DISCONNECTED {
		t.Fatalf("status before oauth = %v", st)
	}

	start, err := h.mcp.StartMCPServerOAuth(ctx, connect.NewRequest(&agentsv1.StartMCPServerOAuthRequest{ServerId: "oa"}))
	if err != nil {
		t.Fatal(err)
	}
	if start.Msg.GetFlowId() == "" || !strings.HasPrefix(start.Msg.GetAuthorizationUrl(), h.url+"/fake-as/authorize?") {
		t.Fatalf("start = %v", start.Msg)
	}
	// Starting changes nothing: the connection only exists after the browser
	// returned through the callback.
	if got := h.oauth(t, "oa"); got.GetState() != agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_DISCONNECTED {
		t.Fatalf("status after start = %v", got)
	}

	final := browse(t, start.Msg.GetAuthorizationUrl())
	if final.Query().Get("mcp_oauth") != "success" || final.Query().Get("server_id") != "oa" || final.Path != "/mcp-servers" {
		t.Fatalf("callback redirected to %s", final)
	}
	after := h.oauth(t, "oa")
	if after.GetState() != agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_CONNECTED || after.GetConnectedAt() == nil {
		t.Fatalf("after = %v", after)
	}
	if st := h.status(t, "oa"); st.GetState() != agentsv1.MCPServerStatus_STATE_CONNECTED || st.GetToolCount() != 2 {
		t.Fatalf("status after oauth = %v", st)
	}

	// A reconnect moves connected_at forward: this is how the App tells a new
	// authorization from the one it started from.
	again, err := h.mcp.StartMCPServerOAuth(ctx, connect.NewRequest(&agentsv1.StartMCPServerOAuthRequest{ServerId: "oa"}))
	if err != nil {
		t.Fatal(err)
	}
	browse(t, again.Msg.GetAuthorizationUrl())
	second := h.oauth(t, "oa")
	if !second.GetConnectedAt().AsTime().After(after.GetConnectedAt().AsTime()) {
		t.Fatalf("connected_at did not advance: %v then %v", after.GetConnectedAt(), second.GetConnectedAt())
	}

	if _, err := h.mcp.DisconnectMCPServerOAuth(ctx, connect.NewRequest(&agentsv1.DisconnectMCPServerOAuthRequest{ServerId: "oa"})); err != nil {
		t.Fatal(err)
	}
	if got := h.oauth(t, "oa"); got.GetState() != agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_DISCONNECTED {
		t.Fatalf("after disconnect = %v", got)
	}
	if st := h.status(t, "oa"); st.GetState() != agentsv1.MCPServerStatus_STATE_DISCONNECTED {
		t.Fatalf("status after disconnect = %v", st)
	}

	// A state is single-use: the authorization server issues a fresh code
	// for the already-consumed state, and the callback refuses it.
	replay := browse(t, again.Msg.GetAuthorizationUrl())
	if replay.Query().Get("mcp_oauth") != "error" {
		t.Fatalf("replayed callback redirected to %s", replay)
	}
}

func TestMCPOAuthRefusedConsentAndFailedExchangeConnectNothing(t *testing.T) {
	h := newMCPHarness(t, mcpWorkspace)
	ctx := context.Background()
	for _, scenario := range []string{"deny", "tokenfail"} {
		h.create(t, oauthServer(h, scenario, scenario))
		start, err := h.mcp.StartMCPServerOAuth(ctx, connect.NewRequest(&agentsv1.StartMCPServerOAuthRequest{ServerId: scenario}))
		if err != nil {
			t.Fatal(err)
		}
		final := browse(t, start.Msg.GetAuthorizationUrl())
		if final.Query().Get("mcp_oauth") != "error" {
			t.Fatalf("%s: callback redirected to %s", scenario, final)
		}
		// The failure is visible only as the absence of a connection; the
		// server keeps no failed-flow record, which the App accounts for.
		if got := h.oauth(t, scenario); got.GetState() != agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_DISCONNECTED {
			t.Fatalf("%s: status = %v", scenario, got)
		}
	}
}

func TestMCPOAuthStartRefusals(t *testing.T) {
	h := newMCPHarness(t, mcpWorkspace)
	ctx := context.Background()

	h.create(t, h.at(httpServer("plain", "Plain", "/fake-mcp/open")))
	_, err := h.mcp.StartMCPServerOAuth(ctx, connect.NewRequest(&agentsv1.StartMCPServerOAuthRequest{ServerId: "plain"}))
	wantCode(t, err, connect.CodeFailedPrecondition)

	_, err = h.mcp.StartMCPServerOAuth(ctx, connect.NewRequest(&agentsv1.StartMCPServerOAuthRequest{ServerId: "missing"}))
	wantCode(t, err, connect.CodeNotFound)

	// No client ID and no advertised registration endpoint: discovery fails.
	dynamic := oauthServer(h, "dyn")
	dynamic.Auth.Oauth2.ClientId, dynamic.Auth.Oauth2.ClientSecret = "", ""
	h.create(t, dynamic)
	_, err = h.mcp.StartMCPServerOAuth(ctx, connect.NewRequest(&agentsv1.StartMCPServerOAuthRequest{ServerId: "dyn"}))
	wantCode(t, err, connect.CodeFailedPrecondition)

	// A client secret without a client ID is refused on save.
	badSecret := oauthServer(h, "badsecret")
	badSecret.Auth.Oauth2.ClientId = ""
	_, err = h.mcp.CreateMCPServer(ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{McpServer: badSecret}))
	wantCode(t, err, connect.CodeInvalidArgument)
}

func TestGlobalMCPPresetsAreReadOnlyAndRedactedForMembers(t *testing.T) {
	h := newMCPHarness(t, mcpWorkspace)
	ctx := context.Background()

	list, err := h.global.ListGlobalMCPServers(ctx, connect.NewRequest(&agentsv1.ListGlobalMCPServersRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]*agentsv1.MCPServer{}
	for _, p := range list.Msg.GetMcpServers() {
		byID[p.GetId()] = p
	}
	if len(byID) != 3 {
		t.Fatalf("presets = %v", byID)
	}
	secret := byID["preset-secret"]
	if secret.GetHeaders()["Authorization"] != "***" || strings.Contains(secret.String(), mcpStaticToken) {
		t.Fatalf("a member's preset listing leaks headers: %v", secret)
	}
	oauth := byID["preset-oauth"].GetAuth().GetOauth2()
	if oauth.GetClientSecret() != "" || strings.Contains(byID["preset-oauth"].String(), mcpClientSecret) {
		t.Fatalf("a member's preset listing leaks the client secret: %v", byID["preset-oauth"])
	}

	installed, err := h.global.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "preset-secret"}))
	if err != nil {
		t.Fatal(err)
	}
	if installed.Msg.GetMcpServer().GetHeaders()["Authorization"] != "***" ||
		installed.Msg.GetMcpServer().GetMetadata()["butter.global_mcp_preset_id"] != "preset-secret" ||
		installed.Msg.GetMcpServer().GetWorkspaceId() != mcpWorkspace {
		t.Fatalf("installed = %v", installed.Msg.GetMcpServer())
	}
	// The installed copy works: the server holds the real header.
	if st := h.status(t, "preset-secret"); st.GetState() != agentsv1.MCPServerStatus_STATE_CONNECTED {
		t.Fatalf("installed preset status = %v", st)
	}
	// Reinstalling is create-or-overwrite, so it succeeds again.
	if _, err := h.global.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "preset-secret"})); err != nil {
		t.Fatal(err)
	}

	// What an old client that round-trips the redacted copy still gets, since
	// the whole-object update is unchanged: the mask overwrites the stored
	// credential. The App never does this; it names what it keeps instead (see
	// TestEditingAnInstalledPresetKeepsItsHiddenCredentials).
	got, err := h.mcp.GetMCPServer(ctx, connect.NewRequest(&agentsv1.GetMCPServerRequest{Id: "preset-secret"}))
	if err != nil {
		t.Fatal(err)
	}
	corrupt := proto.Clone(got.Msg.GetMcpServer()).(*agentsv1.MCPServer)
	corrupt.Name = "renamed"
	if _, err := h.mcp.UpdateMCPServer(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerRequest{McpServer: corrupt})); err != nil {
		t.Fatal(err)
	}
	if st := h.status(t, "preset-secret"); st.GetState() != agentsv1.MCPServerStatus_STATE_DISCONNECTED {
		t.Fatalf("a redacted copy written back should have broken the credential, status = %v", st)
	}

	_, err = h.global.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "no-such-preset"}))
	wantCode(t, err, connect.CodeNotFound)

	h.create(t, h.at(httpServer("preset-open", "Mine", "/fake-mcp/open")))
	_, err = h.global.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "preset-open"}))
	wantCode(t, err, connect.CodeAlreadyExists)

	// Preset administration is not reachable from the App at all.
	resp, err := http.Post(h.url+"/api/agents.v1.GlobalMCPServerService/CreateGlobalMCPServer", "application/proto", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatal("global preset writes must not be served")
	}
}

// An installed Workspace preset is edited through UpdateMCPServerKeepingCredentials: the hidden
// header and client secret survive, every explicit edit applies, and nothing
// the caller could not read comes back. The proof that a credential survived is
// that the real MCP server still accepts it, or the real token endpoint does.
func TestEditingAnInstalledPresetKeepsItsHiddenCredentials(t *testing.T) {
	h := newMCPHarness(t, mcpWorkspace)
	ctx := context.Background()
	install := func(id string) *agentsv1.MCPServer {
		resp, err := h.global.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: id}))
		if err != nil {
			t.Fatal(err)
		}
		return resp.Msg.GetMcpServer()
	}
	update := func(req *agentsv1.UpdateMCPServerRequest) (*agentsv1.MCPServer, error) {
		resp, err := h.mcp.UpdateMCPServer(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err
		}
		return resp.Msg.GetMcpServer(), nil
	}
	keep := func(req *agentsv1.UpdateMCPServerKeepingCredentialsRequest) (*agentsv1.MCPServer, error) {
		resp, err := h.mcp.UpdateMCPServerKeepingCredentials(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err
		}
		return resp.Msg.GetMcpServer(), nil
	}

	t.Run("a hidden header survives a rename, a filter and a timeout", func(t *testing.T) {
		edit := install("preset-secret")
		edit.Name, edit.TimeoutSeconds, edit.ToolFilter = "Renamed", 30, []string{"echo"}
		delete(edit.Headers, "Authorization")
		answer, err := keep(&agentsv1.UpdateMCPServerKeepingCredentialsRequest{McpServer: edit, KeepHeaderKeys: []string{"Authorization"}})
		if err != nil {
			t.Fatal(err)
		}
		if answer.GetHeaders()["Authorization"] != "***" || strings.Contains(answer.String(), mcpStaticToken) {
			t.Fatalf("the answer leaks the kept header: %v", answer)
		}
		if st := h.status(t, "preset-secret"); st.GetState() != agentsv1.MCPServerStatus_STATE_CONNECTED {
			t.Fatalf("the real server no longer accepts the kept header: %v", st)
		}
		read, err := h.mcp.GetMCPServer(ctx, connect.NewRequest(&agentsv1.GetMCPServerRequest{Id: "preset-secret"}))
		if err != nil {
			t.Fatal(err)
		}
		got := read.Msg.GetMcpServer()
		if got.GetName() != "Renamed" || got.GetTimeoutSeconds() != 30 || len(got.GetToolFilter()) != 1 ||
			got.GetHeaders()["Authorization"] != "***" || got.GetMetadata()["butter.global_mcp_preset_id"] != "preset-secret" {
			t.Fatalf("read back = %v", got)
		}
	})

	t.Run("an explicit replacement applies, and so does a removal", func(t *testing.T) {
		edit := install("preset-secret")
		edit.Headers = map[string]string{"Authorization": "Bearer wrong"}
		if _, err := update(&agentsv1.UpdateMCPServerRequest{McpServer: edit}); err != nil {
			t.Fatal(err)
		}
		if st := h.status(t, "preset-secret"); st.GetState() == agentsv1.MCPServerStatus_STATE_CONNECTED {
			t.Fatalf("a replaced header must be what is sent now: %v", st)
		}
		edit = install("preset-secret")
		edit.Headers = nil
		if _, err := update(&agentsv1.UpdateMCPServerRequest{McpServer: edit}); err != nil {
			t.Fatal(err)
		}
		read, _ := h.mcp.GetMCPServer(ctx, connect.NewRequest(&agentsv1.GetMCPServerRequest{Id: "preset-secret"}))
		if len(read.Msg.GetMcpServer().GetHeaders()) != 0 {
			t.Fatalf("an omitted header must be removed: %v", read.Msg.GetMcpServer())
		}
	})

	t.Run("a hidden client secret survives, and a replaced or cleared one is what is used", func(t *testing.T) {
		complete := func() string {
			start, err := h.mcp.StartMCPServerOAuth(ctx, connect.NewRequest(&agentsv1.StartMCPServerOAuthRequest{ServerId: "preset-oauth"}))
			if err != nil {
				t.Fatal(err)
			}
			return browse(t, start.Msg.GetAuthorizationUrl()).Query().Get("mcp_oauth")
		}
		edit := install("preset-oauth")
		if edit.GetAuth().GetOauth2().GetClientSecret() != "" {
			t.Fatalf("the install answer leaks the client secret: %v", edit)
		}
		edit.Name = "Renamed OAuth"
		answer, err := keep(&agentsv1.UpdateMCPServerKeepingCredentialsRequest{McpServer: edit, KeepOauthClientSecret: true})
		if err != nil {
			t.Fatal(err)
		}
		if answer.GetAuth().GetOauth2().GetClientSecret() != "" || strings.Contains(answer.String(), mcpClientSecret) {
			t.Fatalf("the answer leaks the client secret: %v", answer)
		}
		if got := complete(); got != "success" {
			t.Fatalf("the kept secret must still authorize, callback = %s", got)
		}

		edit = install("preset-oauth")
		edit.Auth.Oauth2.ClientSecret = "wrong-secret"
		if _, err := update(&agentsv1.UpdateMCPServerRequest{McpServer: edit}); err != nil {
			t.Fatal(err)
		}
		if got := complete(); got != "error" {
			t.Fatalf("a replaced secret is what the token endpoint sees, callback = %s", got)
		}

		edit = install("preset-oauth")
		edit.Auth.Oauth2.ClientSecret = ""
		if _, err := update(&agentsv1.UpdateMCPServerRequest{McpServer: edit}); err != nil {
			t.Fatal(err)
		}
		if got := complete(); got != "error" {
			t.Fatalf("a cleared secret is not sent, callback = %s", got)
		}
	})

	t.Run("a refused keep changes nothing", func(t *testing.T) {
		edit := install("preset-oauth")
		edit.Name = "Changed"
		edit.Auth.Oauth2.ClientId = "another-client"
		_, err := keep(&agentsv1.UpdateMCPServerKeepingCredentialsRequest{McpServer: edit, KeepOauthClientSecret: true})
		wantCode(t, err, connect.CodeInvalidArgument)
		_, err = keep(&agentsv1.UpdateMCPServerKeepingCredentialsRequest{McpServer: install("preset-secret"), KeepHeaderKeys: []string{"X-Missing"}})
		wantCode(t, err, connect.CodeInvalidArgument)
		read, _ := h.mcp.GetMCPServer(ctx, connect.NewRequest(&agentsv1.GetMCPServerRequest{Id: "preset-oauth"}))
		if read.Msg.GetMcpServer().GetName() == "Changed" {
			t.Fatal("a refused update was applied")
		}
	})

	t.Run("the global preset itself is untouched by edits to its copy", func(t *testing.T) {
		list, err := h.global.ListGlobalMCPServers(ctx, connect.NewRequest(&agentsv1.ListGlobalMCPServersRequest{}))
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range list.Msg.GetMcpServers() {
			if p.GetId() == "preset-secret" && p.GetName() != "Secret preset" {
				t.Fatalf("the preset was changed: %v", p)
			}
		}
	})
}

// A server without UpdateMCPServerKeepingCredentials answers it as an unknown
// path. That answer is Unimplemented, it changes nothing, and the ordinary
// create / update / delete procedures keep working there, so a client can tell
// the two apart before it stores anything.
func TestAServerWithoutTheKeepMethodAnswersUnimplementedAndChangesNothing(t *testing.T) {
	h := newMCPHarness(t, mcpWorkspace)
	ctx := context.Background()
	interceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+h.token)
			req.Header().Set("X-Workspace-ID", mcpWorkspace)
			return next(ctx, req)
		}
	})
	legacy := agentsv1connect.NewMCPServerServiceClient(http.DefaultClient, h.url+legacyPrefix+"/api", connect.WithInterceptors(interceptor))

	installed, err := h.global.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "preset-secret"}))
	if err != nil {
		t.Fatal(err)
	}
	edit := installed.Msg.GetMcpServer()
	edit.Name = "Never applied"
	delete(edit.Headers, "Authorization")
	_, err = legacy.UpdateMCPServerKeepingCredentials(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerKeepingCredentialsRequest{
		McpServer: edit, KeepHeaderKeys: []string{"Authorization"},
	}))
	wantCode(t, err, connect.CodeUnimplemented)
	if st := h.status(t, "preset-secret"); st.GetState() != agentsv1.MCPServerStatus_STATE_CONNECTED {
		t.Fatalf("the stored credential must be untouched: %v", st)
	}
	read, err := h.mcp.GetMCPServer(ctx, connect.NewRequest(&agentsv1.GetMCPServerRequest{Id: "preset-secret"}))
	if err != nil {
		t.Fatal(err)
	}
	if read.Msg.GetMcpServer().GetName() == "Never applied" {
		t.Fatal("an unimplemented method changed the server")
	}

	// The ordinary procedures are all there.
	plain := h.at(httpServer("legacy-plain", "Legacy plain", "/fake-mcp/open"))
	if _, err := legacy.CreateMCPServer(ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{McpServer: plain})); err != nil {
		t.Fatal(err)
	}
	plain.Name = "Renamed"
	if _, err := legacy.UpdateMCPServer(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerRequest{McpServer: plain})); err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.DeleteMCPServer(ctx, connect.NewRequest(&agentsv1.DeleteMCPServerRequest{Id: "legacy-plain"})); err != nil {
		t.Fatal(err)
	}
	_, err = h.mcp.DeleteMCPServer(ctx, connect.NewRequest(&agentsv1.DeleteMCPServerRequest{Id: "preset-secret"}))
	if err != nil {
		t.Fatal(err)
	}
}

func TestMCPPermissionAndWorkspaceIsolation(t *testing.T) {
	h := newMCPHarness(t, mcpWorkspace)
	ctx := context.Background()
	h.create(t, h.at(httpServer("mine", "Mine", "/fake-mcp/open")))

	denied := newMCPHarness(t, mcpDeniedWorkspace)
	_, err := denied.mcp.ListMCPServers(ctx, connect.NewRequest(&agentsv1.ListMCPServersRequest{}))
	wantCode(t, err, connect.CodePermissionDenied)
	_, err = denied.mcp.CreateMCPServer(ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{McpServer: denied.at(httpServer("x", "x", "/fake-mcp/open"))}))
	wantCode(t, err, connect.CodePermissionDenied)
	_, err = denied.remote.ListRemoteAgents(ctx, connect.NewRequest(&agentsv1.ListRemoteAgentsRequest{}))
	wantCode(t, err, connect.CodePermissionDenied)

	alpha := newMCPHarness(t, "alpha")
	list, err := alpha.mcp.ListMCPServers(ctx, connect.NewRequest(&agentsv1.ListMCPServersRequest{}))
	if err != nil || len(list.Msg.GetMcpServers()) != 0 {
		t.Fatalf("another Workspace sees %v, %v", list, err)
	}
}

func TestRemoteAgentConfigurationCredentialsAndProbes(t *testing.T) {
	h := newMCPHarness(t, mcpWorkspace)
	ctx := context.Background()
	probe := func(id string) *agentsv1.RemoteAgentStatus {
		resp, err := h.remote.GetRemoteAgentStatus(ctx, connect.NewRequest(&agentsv1.GetRemoteAgentStatusRequest{Id: id}))
		if err != nil {
			t.Fatalf("probe %s: %v", id, err)
		}
		return resp.Msg.GetStatus()
	}
	create := func(agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
		resp, err := h.remote.CreateRemoteAgent(ctx, connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{RemoteAgent: agent}))
		if err != nil {
			return nil, err
		}
		return resp.Msg.GetRemoteAgent(), nil
	}

	a2a := &agentsv1.RemoteAgent{Id: "a2a", Name: "A2A", Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A, Url: h.url + "/fake-a2a/ok"}
	withUnknown(a2a, 99, 5)
	created, err := create(a2a)
	if err != nil || created.GetWorkspaceId() != mcpWorkspace || len(created.ProtoReflect().GetUnknown()) == 0 {
		t.Fatalf("created = %v, %v", created, err)
	}
	if st := probe("a2a"); st.GetState() != agentsv1.RemoteAgentStatus_STATE_ACTIVE {
		t.Fatalf("a2a probe = %v", st)
	}
	if _, err := create(&agentsv1.RemoteAgent{Id: "a2a-gone", Name: "Gone", Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A, Url: h.url + "/fake-a2a/gone"}); err != nil {
		t.Fatal(err)
	}
	if st := probe("a2a-gone"); st.GetState() != agentsv1.RemoteAgentStatus_STATE_UNREACHABLE || st.GetDetail() == "" {
		t.Fatalf("gone probe = %v", st)
	}

	oc := &agentsv1.RemoteAgent{
		Id: "oc", Name: "OpenCode", Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP,
		Url: h.url + "/fake-opencode/ok", Username: mcpOpencodeUser, Password: mcpOpencodePass,
		OpencodeAgent: "review", OpencodeModel: "anthropic/claude",
	}
	createdOC, err := create(oc)
	if err != nil || createdOC.GetPassword() != mcpOpencodePass || createdOC.GetOpencodeAgent() != "review" {
		t.Fatalf("opencode created = %v, %v", createdOC, err)
	}
	if st := probe("oc"); st.GetState() != agentsv1.RemoteAgentStatus_STATE_ACTIVE {
		t.Fatalf("opencode probe = %v", st)
	}

	// Credential semantics are whole-object: sending the password back keeps
	// it, a new value replaces it, an empty one clears it.
	wrong := proto.Clone(createdOC).(*agentsv1.RemoteAgent)
	wrong.Password = "nope"
	if _, err := h.remote.UpdateRemoteAgent(ctx, connect.NewRequest(&agentsv1.UpdateRemoteAgentRequest{RemoteAgent: wrong})); err != nil {
		t.Fatal(err)
	}
	if st := probe("oc"); st.GetState() != agentsv1.RemoteAgentStatus_STATE_ERROR {
		t.Fatalf("wrong password probe = %v", st)
	}
	cleared := proto.Clone(createdOC).(*agentsv1.RemoteAgent)
	cleared.Username, cleared.Password = "", ""
	if _, err := h.remote.UpdateRemoteAgent(ctx, connect.NewRequest(&agentsv1.UpdateRemoteAgentRequest{RemoteAgent: cleared})); err != nil {
		t.Fatal(err)
	}
	got, err := h.remote.GetRemoteAgent(ctx, connect.NewRequest(&agentsv1.GetRemoteAgentRequest{Id: "oc"}))
	if err != nil || got.Msg.GetRemoteAgent().GetPassword() != "" || got.Msg.GetRemoteAgent().GetOpencodeAgent() != "review" {
		t.Fatalf("after clear = %v, %v", got, err)
	}

	// Validation, in the server's own words.
	_, err = create(&agentsv1.RemoteAgent{Id: "u", Name: "u", Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP, Url: h.url, Username: "someone"})
	wantCode(t, err, connect.CodeInvalidArgument)
	_, err = create(&agentsv1.RemoteAgent{Id: "n", Name: "n", Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A, Url: "not a url"})
	wantCode(t, err, connect.CodeInvalidArgument)
	_, err = create(&agentsv1.RemoteAgent{Id: "d", Name: "d", Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON})
	wantCode(t, err, connect.CodeInvalidArgument)
	_, err = create(&agentsv1.RemoteAgent{Id: "a2a", Name: "dup", Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A, Url: h.url})
	wantCode(t, err, connect.CodeAlreadyExists)

	if _, err := h.remote.DeleteRemoteAgent(ctx, connect.NewRequest(&agentsv1.DeleteRemoteAgentRequest{Id: "a2a"})); err != nil {
		t.Fatal(err)
	}
	_, err = h.remote.GetRemoteAgent(ctx, connect.NewRequest(&agentsv1.GetRemoteAgentRequest{Id: "a2a"}))
	wantCode(t, err, connect.CodeNotFound)

	// A Daemon agent stored by the Web is listed with its fields intact and is
	// the App's to show, never to edit (Daemon is a separate ticket).
	list, err := h.remote.ListRemoteAgents(ctx, connect.NewRequest(&agentsv1.ListRemoteAgentsRequest{}))
	if err != nil || len(list.Msg.GetRemoteAgents()) != 2 {
		t.Fatalf("list = %v, %v", list, err)
	}
}
