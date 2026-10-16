package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/repo/agentfile"
	agentfilememory "go.orx.me/apps/butter/internal/repo/agentfile/memory"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// The file space fixture (#456): the production AgentFileService over the
// in-memory repository the real one is backed by (the same Repository
// contract the Mongo store implements), plus the file/skill/MCP backends the
// production runner mounts into an Agent's tools. Only persistence is in
// memory and the model is a scripted double (see fakeModel); nothing here is a
// production database, object store or language model.
//
// The Files Workspace is the one the account is a plain member of, and whose
// Agents the chat runner serves, so a contract test can save an Agent with
// associations through the real AgentService and run it. files-peer holds a
// same-named space to prove spaces never cross Workspaces.
const (
	filesWorkspace     = "files"
	filesPeerWorkspace = "files-peer"
	// filesSkill is a SKILL.md bundle seeded in the Files Workspace, so the
	// Agent's skill association has something to list and load.
	filesSkill = "files-skill"
)

// fileBackends are the stores the chat runner is built with, so the Agents the
// App saves get real file, skill and MCP tools.
type fileBackends struct {
	files *agentfilememory.Store
}

func newFileBackends() *fileBackends {
	return &fileBackends{files: agentfilememory.New()}
}

// register mounts the production AgentFileService's ten procedures.
func (b *fileBackends) register(r *gin.Engine, opts []connect.HandlerOption) {
	svc := application.NewAgentFileServiceServer(b.files)
	_, handler := agentsv1connect.NewAgentFileServiceHandler(svc, opts...)
	for _, procedure := range []string{
		agentsv1connect.AgentFileServiceListAgentFileSpacesProcedure,
		agentsv1connect.AgentFileServiceGetAgentFileSpaceProcedure,
		agentsv1connect.AgentFileServiceCreateAgentFileSpaceProcedure,
		agentsv1connect.AgentFileServiceUpdateAgentFileSpaceProcedure,
		agentsv1connect.AgentFileServiceDeleteAgentFileSpaceProcedure,
		agentsv1connect.AgentFileServiceListAgentFilesProcedure,
		agentsv1connect.AgentFileServiceGetAgentFileProcedure,
		agentsv1connect.AgentFileServiceWriteAgentFileProcedure,
		agentsv1connect.AgentFileServiceDeleteAgentFileProcedure,
		agentsv1connect.AgentFileServiceSearchAgentFilesProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", handler)))
	}
}

var _ agentfile.Repository = (*agentfilememory.Store)(nil)

// toolCommand turns the scripted model's input convention into a reply:
//
//	tools?                        answers with the tool names the run offered
//	tool:<name> <json arguments>  calls that tool once, then reports its result
//
// Anything else is the plain echo. The model decides nothing: what the
// Agent can do is exactly what its saved configuration mounted.
func toolCommand(w http.ResponseWriter, req openaifake.ChatCompletionRequest, text string) bool {
	if req.LastRole() == "tool" {
		last := req.Messages[len(req.Messages)-1]
		var content string
		if json.Unmarshal(last.Content, &content) != nil {
			content = string(last.Content)
		}
		openaifake.WriteReply(w, req, "fixture-tool-result("+content+")")
		return true
	}
	if strings.HasPrefix(text, "tools?") {
		openaifake.WriteReply(w, req, "fixture-tools("+strings.Join(offeredTools(req), ",")+")")
		return true
	}
	if rest, ok := strings.CutPrefix(text, "tool:"); ok {
		name, args, _ := strings.Cut(strings.TrimSpace(rest), " ")
		if strings.TrimSpace(args) == "" {
			args = "{}"
		}
		openaifake.WriteReply(w, req, "", openaifake.ToolCall{ID: "call-1", Name: name, Arguments: args})
		return true
	}
	return false
}

func offeredTools(req openaifake.ChatCompletionRequest) []string {
	var names []string
	tools, _ := req.Decoded["tools"].([]any)
	for _, t := range tools {
		fn, _ := t.(map[string]any)["function"].(map[string]any)
		if name, _ := fn["name"].(string); name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func init() {
	registerFixtureWorkspaces(
		fixtureWorkspace{ID: filesWorkspace, NoSeedAgent: true},
		fixtureWorkspace{ID: filesPeerWorkspace, NoSeedAgent: true},
	)
}
