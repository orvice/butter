package application

import (
	"testing"

	"connectrpc.com/connect"
	skillmemory "go.orx.me/apps/butter/internal/repo/skill/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
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

	_, err := svc.GetSkill(ctxB, connect.NewRequest(&agentsv1.GetSkillRequest{Name: "pdf-report"}))
	requireConnectCode(t, err, connect.CodeNotFound)
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

	for _, md := range []string{validSkillMD, "---\nname: csv-export\ndescription: Exports CSV.\n---\nbody\n"} {
		fmName := "pdf-report"
		if md != validSkillMD {
			fmName = "csv-export"
		}
		if _, err := svc.CreateSkill(ctx, connect.NewRequest(&agentsv1.CreateSkillRequest{
			Name: fmName, SkillMd: md,
		})); err != nil {
			t.Fatalf("create %s: %v", fmName, err)
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
