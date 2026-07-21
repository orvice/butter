package application

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	skillmemory "go.orx.me/apps/butter/internal/repo/skill/memory"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const validSkillMD = `---
name: pdf-report
description: Generates PDF reports from structured data.
license: MIT
metadata:
  author: butter
allowed-tools:
  - agent_files_read_file
---
# PDF Report

Follow these steps to build a report.
`

func newSkillTestService() *SkillServiceServer {
	return NewSkillServiceServer(skillmemory.New())
}

func TestSkillServiceCreateThenGetRoundTrip(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()

	created, err := svc.CreateSkill(ctx, connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name:    "pdf-report",
		SkillMd: validSkillMD,
	}))
	if err != nil {
		t.Fatalf("CreateSkill: %v", err)
	}
	skill := created.Msg.GetSkill()
	if skill.GetName() != "pdf-report" {
		t.Fatalf("expected name pdf-report, got %q", skill.GetName())
	}
	if skill.GetDescription() != "Generates PDF reports from structured data." {
		t.Fatalf("unexpected description: %q", skill.GetDescription())
	}
	if skill.GetLicense() != "MIT" {
		t.Fatalf("unexpected license: %q", skill.GetLicense())
	}
	if skill.GetMetadata()["author"] != "butter" {
		t.Fatalf("unexpected metadata: %v", skill.GetMetadata())
	}
	if len(skill.GetAllowedTools()) != 1 || skill.GetAllowedTools()[0] != "agent_files_read_file" {
		t.Fatalf("unexpected allowed tools: %v", skill.GetAllowedTools())
	}
	if skill.GetSizeBytes() != int64(len(validSkillMD)) {
		t.Fatalf("expected size %d, got %d", len(validSkillMD), skill.GetSizeBytes())
	}

	got, err := svc.GetSkill(ctx, connect.NewRequest(&agentsv1.GetSkillRequest{Name: "pdf-report"}))
	if err != nil {
		t.Fatalf("GetSkill: %v", err)
	}
	if got.Msg.GetSkillMd() != validSkillMD {
		t.Fatalf("SKILL.md did not round-trip:\n%s", got.Msg.GetSkillMd())
	}
	if got.Msg.GetSkill().GetDescription() != skill.GetDescription() {
		t.Fatalf("metadata mismatch between create and get")
	}
}

func requireConnectCode(t *testing.T, err error, want connect.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s error, got nil", want)
	}
	cerr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}
	if cerr.Code() != want {
		t.Fatalf("expected %s, got %s (%s)", want, cerr.Code(), cerr.Message())
	}
}

func TestSkillServiceCreateRejectsInvalidSkillMD(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()

	cases := []struct {
		label   string
		name    string
		skillMD string
	}{
		{"missing frontmatter separator", "pdf-report", "# Just markdown\n"},
		{"frontmatter name mismatch", "other-name", validSkillMD},
		{"invalid name characters", "Bad_Name", "---\nname: Bad_Name\ndescription: x\n---\nbody\n"},
		{"missing description", "pdf-report", "---\nname: pdf-report\n---\nbody\n"},
		{"over-long description", "pdf-report", "---\nname: pdf-report\ndescription: " + strings.Repeat("x", 1025) + "\n---\nbody\n"},
		{"unclosed frontmatter", "pdf-report", "---\nname: pdf-report\ndescription: x\n"},
	}
	for _, tc := range cases {
		_, err := svc.CreateSkill(ctx, connect.NewRequest(&agentsv1.CreateSkillRequest{
			Name:    tc.name,
			SkillMd: tc.skillMD,
		}))
		if err == nil {
			t.Fatalf("%s: expected error", tc.label)
		}
		requireConnectCode(t, err, connect.CodeInvalidArgument)
	}
}

func TestSkillServiceCreateDuplicateAndCrossWorkspace(t *testing.T) {
	svc := newSkillTestService()
	ctxA := workspace.WithID(t.Context(), "ws-a")
	ctxB := workspace.WithID(t.Context(), "ws-b")

	if _, err := svc.CreateSkill(ctxA, connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name: "pdf-report", SkillMd: validSkillMD,
	})); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := svc.CreateSkill(ctxA, connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name: "pdf-report", SkillMd: validSkillMD,
	}))
	requireConnectCode(t, err, connect.CodeAlreadyExists)

	if _, err := svc.CreateSkill(ctxB, connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name: "pdf-report", SkillMd: validSkillMD,
	})); err != nil {
		t.Fatalf("same name in another workspace should succeed: %v", err)
	}
}

func TestSkillServiceWorkspaceIsolation(t *testing.T) {
	svc := newSkillTestService()
	ctxA := workspace.WithID(t.Context(), "ws-a")
	ctxB := workspace.WithID(t.Context(), "ws-b")

	if _, err := svc.CreateSkill(ctxA, connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name: "pdf-report", SkillMd: validSkillMD,
	})); err != nil {
		t.Fatalf("create: %v", err)
	}

	_, getErr := svc.GetSkill(ctxB, connect.NewRequest(&agentsv1.GetSkillRequest{Name: "pdf-report"}))
	requireConnectCode(t, getErr, connect.CodeNotFound)

	_, updateErr := svc.UpdateSkill(ctxB, connect.NewRequest(&agentsv1.UpdateSkillRequest{
		Name: "pdf-report", SkillMd: updatedSkillMD,
	}))
	requireConnectCode(t, updateErr, connect.CodeNotFound)

	_, deleteErr := svc.DeleteSkill(ctxB, connect.NewRequest(&agentsv1.DeleteSkillRequest{Name: "pdf-report"}))
	requireConnectCode(t, deleteErr, connect.CodeNotFound)

	// The skill remains intact in its own workspace.
	if _, err := svc.GetSkill(ctxA, connect.NewRequest(&agentsv1.GetSkillRequest{Name: "pdf-report"})); err != nil {
		t.Fatalf("skill should survive cross-workspace mutation attempts: %v", err)
	}
}

const updatedSkillMD = `---
name: pdf-report
description: Updated description for the report skill.
---
# PDF Report v2

New instructions.
`

func TestSkillServiceListReturnsMetadataOnly(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()

	cases := []struct {
		name    string
		skillMD string
	}{
		{"pdf-report", validSkillMD},
		{"csv-export", "---\nname: csv-export\ndescription: Exports CSV.\n---\nbody\n"},
	}
	for _, tc := range cases {
		if _, err := svc.CreateSkill(ctx, connect.NewRequest(&agentsv1.CreateSkillRequest{
			Name: tc.name, SkillMd: tc.skillMD,
		})); err != nil {
			t.Fatalf("create %s: %v", tc.name, err)
		}
	}

	res, err := svc.ListSkills(ctx, connect.NewRequest(&agentsv1.ListSkillsRequest{}))
	if err != nil {
		t.Fatalf("ListSkills: %v", err)
	}
	skills := res.Msg.GetSkills()
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
	if skills[0].GetName() != "csv-export" || skills[1].GetName() != "pdf-report" {
		t.Fatalf("expected sorted names, got %q, %q", skills[0].GetName(), skills[1].GetName())
	}
	if skills[1].GetDescription() != "Generates PDF reports from structured data." {
		t.Fatalf("unexpected description: %q", skills[1].GetDescription())
	}
}

func TestSkillServiceUpdateReplacesSkillMD(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()

	if _, err := svc.CreateSkill(ctx, connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name: "pdf-report", SkillMd: validSkillMD,
	})); err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := svc.UpdateSkill(ctx, connect.NewRequest(&agentsv1.UpdateSkillRequest{
		Name: "pdf-report", SkillMd: updatedSkillMD,
	}))
	if err != nil {
		t.Fatalf("UpdateSkill: %v", err)
	}
	if updated.Msg.GetSkill().GetDescription() != "Updated description for the report skill." {
		t.Fatalf("description not updated: %q", updated.Msg.GetSkill().GetDescription())
	}

	got, err := svc.GetSkill(ctx, connect.NewRequest(&agentsv1.GetSkillRequest{Name: "pdf-report"}))
	if err != nil {
		t.Fatalf("GetSkill after update: %v", err)
	}
	if got.Msg.GetSkillMd() != updatedSkillMD {
		t.Fatalf("SKILL.md not replaced")
	}
}

func TestSkillServiceUpdateRejectsNameChangeAndMissing(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()

	if _, err := svc.CreateSkill(ctx, connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name: "pdf-report", SkillMd: validSkillMD,
	})); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Frontmatter declares a different name than the addressed skill.
	_, err := svc.UpdateSkill(ctx, connect.NewRequest(&agentsv1.UpdateSkillRequest{
		Name: "pdf-report", SkillMd: "---\nname: renamed\ndescription: x\n---\nbody\n",
	}))
	requireConnectCode(t, err, connect.CodeInvalidArgument)

	// Updating a skill that does not exist.
	_, err = svc.UpdateSkill(ctx, connect.NewRequest(&agentsv1.UpdateSkillRequest{
		Name: "ghost", SkillMd: "---\nname: ghost\ndescription: x\n---\nbody\n",
	}))
	requireConnectCode(t, err, connect.CodeNotFound)
}

func TestSkillServiceDeleteRemovesSkill(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()

	if _, err := svc.CreateSkill(ctx, connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name: "pdf-report", SkillMd: validSkillMD,
	})); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.DeleteSkill(ctx, connect.NewRequest(&agentsv1.DeleteSkillRequest{Name: "pdf-report"})); err != nil {
		t.Fatalf("DeleteSkill: %v", err)
	}
	_, err := svc.GetSkill(ctx, connect.NewRequest(&agentsv1.GetSkillRequest{Name: "pdf-report"}))
	requireConnectCode(t, err, connect.CodeNotFound)

	_, err = svc.DeleteSkill(ctx, connect.NewRequest(&agentsv1.DeleteSkillRequest{Name: "pdf-report"}))
	requireConnectCode(t, err, connect.CodeNotFound)
}

func TestSkillServiceCreateEnforcesMaxSkillMDBytes(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()
	svc.SetSkillMDMaxBytes(int64(len(validSkillMD)) - 1)

	_, err := svc.CreateSkill(ctx, connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name: "pdf-report", SkillMd: validSkillMD,
	}))
	requireConnectCode(t, err, connect.CodeInvalidArgument)
}

func createTestSkill(t *testing.T, ctx context.Context, svc *SkillServiceServer) {
	t.Helper()
	_, err := svc.CreateSkill(ctx, connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name:    "pdf-report",
		SkillMd: validSkillMD,
	}))
	if err != nil {
		t.Fatalf("CreateSkill: %v", err)
	}
}

func TestSkillResourcePutGetListDeleteRoundTrip(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()
	createTestSkill(t, ctx, svc)
	binary := []byte{0x89, 'P', 'N', 'G', 0x00, 0x1f, 0xff, 0x00}

	put, err := svc.PutSkillResource(ctx, connect.NewRequest(&agentsv1.PutSkillResourceRequest{
		SkillName:   "pdf-report",
		Path:        "assets/logo.png",
		Content:     binary,
		ContentType: "image/png",
	}))
	if err != nil {
		t.Fatalf("PutSkillResource: %v", err)
	}
	res := put.Msg.GetResource()
	if res.GetPath() != "assets/logo.png" || res.GetSizeBytes() != int64(len(binary)) || res.GetContentType() != "image/png" {
		t.Fatalf("unexpected resource metadata: %v", res)
	}

	got, err := svc.GetSkillResource(ctx, connect.NewRequest(&agentsv1.GetSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "assets/logo.png",
	}))
	if err != nil {
		t.Fatalf("GetSkillResource: %v", err)
	}
	if !bytes.Equal(got.Msg.GetContent(), binary) {
		t.Fatalf("content did not round-trip: %v", got.Msg.GetContent())
	}

	list, err := svc.ListSkillResources(ctx, connect.NewRequest(&agentsv1.ListSkillResourcesRequest{
		SkillName: "pdf-report",
	}))
	if err != nil {
		t.Fatalf("ListSkillResources: %v", err)
	}
	if n := len(list.Msg.GetResources()); n != 1 {
		t.Fatalf("expected 1 resource, got %d", n)
	}

	if _, err := svc.DeleteSkillResource(ctx, connect.NewRequest(&agentsv1.DeleteSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "assets/logo.png",
	})); err != nil {
		t.Fatalf("DeleteSkillResource: %v", err)
	}
	_, err = svc.GetSkillResource(ctx, connect.NewRequest(&agentsv1.GetSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "assets/logo.png",
	}))
	requireConnectCode(t, err, connect.CodeNotFound)
}

func TestSkillResourceRejectsUnsafePaths(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()
	createTestSkill(t, ctx, svc)

	for _, p := range []string{"docs/readme.md", "../secrets.txt", "references/../../etc/passwd", "references/dir/../notes.md", "/references/abs.md", `references\..\x`} {
		_, err := svc.PutSkillResource(ctx, connect.NewRequest(&agentsv1.PutSkillResourceRequest{
			SkillName: "pdf-report",
			Path:      p,
			Content:   []byte("x"),
		}))
		requireConnectCode(t, err, connect.CodeInvalidArgument)
	}
	// Read-side paths are cleaned with the same rule.
	_, err := svc.GetSkillResource(ctx, connect.NewRequest(&agentsv1.GetSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "../SKILL.md",
	}))
	requireConnectCode(t, err, connect.CodeInvalidArgument)
	_, err = svc.DeleteSkillResource(ctx, connect.NewRequest(&agentsv1.DeleteSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "../SKILL.md",
	}))
	requireConnectCode(t, err, connect.CodeInvalidArgument)
}

func TestSkillResourceSizeCap(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()
	createTestSkill(t, ctx, svc)

	tooBig := make([]byte, 10*1024*1024+1)
	_, err := svc.PutSkillResource(ctx, connect.NewRequest(&agentsv1.PutSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "assets/huge.bin",
		Content:   tooBig,
	}))
	requireConnectCode(t, err, connect.CodeInvalidArgument)

	// Exactly at the cap is allowed (aligned with ADK's read limit).
	_, err = svc.PutSkillResource(ctx, connect.NewRequest(&agentsv1.PutSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "assets/max.bin",
		Content:   tooBig[:10*1024*1024],
	}))
	if err != nil {
		t.Fatalf("PutSkillResource at cap: %v", err)
	}
}

func TestSkillResourceCountCap(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()
	svc.SetSkillResourceMaxCount(2)
	createTestSkill(t, ctx, svc)

	for _, p := range []string{"assets/a.txt", "assets/b.txt"} {
		if _, err := svc.PutSkillResource(ctx, connect.NewRequest(&agentsv1.PutSkillResourceRequest{
			SkillName: "pdf-report",
			Path:      p,
			Content:   []byte("x"),
		})); err != nil {
			t.Fatalf("PutSkillResource %s: %v", p, err)
		}
	}
	// A third distinct path exceeds the cap.
	_, err := svc.PutSkillResource(ctx, connect.NewRequest(&agentsv1.PutSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "assets/c.txt",
		Content:   []byte("x"),
	}))
	requireConnectCode(t, err, connect.CodeResourceExhausted)

	// Overwriting an existing path does not count against the cap.
	if _, err := svc.PutSkillResource(ctx, connect.NewRequest(&agentsv1.PutSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "assets/a.txt",
		Content:   []byte("longer replacement"),
	})); err != nil {
		t.Fatalf("overwrite at cap: %v", err)
	}
}

func TestSkillResourceMissingSkillIsNotFound(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-skills")
	svc := newSkillTestService()

	_, err := svc.PutSkillResource(ctx, connect.NewRequest(&agentsv1.PutSkillResourceRequest{
		SkillName: "absent",
		Path:      "assets/a.txt",
		Content:   []byte("x"),
	}))
	requireConnectCode(t, err, connect.CodeNotFound)
	_, err = svc.ListSkillResources(ctx, connect.NewRequest(&agentsv1.ListSkillResourcesRequest{SkillName: "absent"}))
	requireConnectCode(t, err, connect.CodeNotFound)
}

// TestSkillResourceBinaryRoundTripOverSnakeCaseJSON exercises the full
// connect handler with the snake_case JSON codec (issue #154 acceptance):
// bytes fields must survive the wire encoding both directions.
func TestSkillResourceBinaryRoundTripOverSnakeCaseJSON(t *testing.T) {
	svc := newSkillTestService()
	mux := http.NewServeMux()
	path, handler := agentsv1connect.NewSkillServiceHandler(svc, connectx.HandlerOptions()...)
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(workspace.WithID(r.Context(), "ws-skills"))
		handler.ServeHTTP(w, r)
	}))
	server := httptest.NewServer(mux)
	defer server.Close()

	client := agentsv1connect.NewSkillServiceClient(server.Client(), server.URL, connect.WithProtoJSON())

	if _, err := client.CreateSkill(t.Context(), connect.NewRequest(&agentsv1.CreateSkillRequest{
		Name:    "pdf-report",
		SkillMd: validSkillMD,
	})); err != nil {
		t.Fatalf("CreateSkill over JSON: %v", err)
	}

	binary := []byte{0x00, 0x01, 0xfe, 0xff, '{', '"', 0x7f}
	if _, err := client.PutSkillResource(t.Context(), connect.NewRequest(&agentsv1.PutSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "assets/blob.bin",
		Content:   binary,
	})); err != nil {
		t.Fatalf("PutSkillResource over JSON: %v", err)
	}

	got, err := client.GetSkillResource(t.Context(), connect.NewRequest(&agentsv1.GetSkillResourceRequest{
		SkillName: "pdf-report",
		Path:      "assets/blob.bin",
	}))
	if err != nil {
		t.Fatalf("GetSkillResource over JSON: %v", err)
	}
	if !bytes.Equal(got.Msg.GetContent(), binary) {
		t.Fatalf("binary content corrupted over JSON wire: %v", got.Msg.GetContent())
	}
}
