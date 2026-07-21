package application

import (
	"context"
	"errors"
	"fmt"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	adkskill "google.golang.org/adk/v2/tool/skilltoolset/skill"

	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const defaultSkillMDMaxBytes int64 = 256 * 1024

// skillResourceMaxBytes is hard-aligned with ADK's per-resource read cap
// (adkskill maxResourceSize): anything larger could be stored but never
// loaded by load_skill_resource, so the write is rejected instead.
const skillResourceMaxBytes int64 = 10 * 1024 * 1024

const defaultSkillResourceMaxCount = 100

type SkillServiceServer struct {
	repo                  skillrepo.Repository
	skillMDMaxBytes       int64
	skillResourceMaxCount int
}

func NewSkillServiceServer(repo skillrepo.Repository) *SkillServiceServer {
	return &SkillServiceServer{
		repo:                  repo,
		skillMDMaxBytes:       defaultSkillMDMaxBytes,
		skillResourceMaxCount: defaultSkillResourceMaxCount,
	}
}

func (s *SkillServiceServer) SetRepo(repo skillrepo.Repository) {
	s.repo = repo
}

func (s *SkillServiceServer) SetSkillMDMaxBytes(max int64) {
	if max <= 0 {
		max = defaultSkillMDMaxBytes
	}
	s.skillMDMaxBytes = max
}

func (s *SkillServiceServer) SetSkillResourceMaxCount(max int) {
	if max <= 0 {
		max = defaultSkillResourceMaxCount
	}
	s.skillResourceMaxCount = max
}

func (s *SkillServiceServer) requireRepo() (skillrepo.Repository, error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("skill repository not available"))
	}
	return s.repo, nil
}

// parseSkillMD validates a SKILL.md document against the agentskills.io spec
// and checks the frontmatter name equals the addressed skill name.
func (s *SkillServiceServer) parseSkillMD(name, skillMD string) (*agentsv1.Skill, error) {
	if name == "" {
		return nil, connectx.RequiredArgument("name")
	}
	if skillMD == "" {
		return nil, connectx.RequiredArgument("skill_md")
	}
	if max := s.skillMDMaxBytes; max > 0 && int64(len(skillMD)) > max {
		return nil, connectx.InvalidArgument("skill_md", fmt.Sprintf("SKILL.md exceeds max size of %d bytes", max))
	}
	fm, _, err := adkskill.ParseBytes([]byte(skillMD))
	if err != nil {
		return nil, connectx.InvalidArgument("skill_md", err.Error())
	}
	if fm.Name != name {
		return nil, connectx.InvalidArgument("name", fmt.Sprintf("frontmatter name %q does not match skill name %q", fm.Name, name))
	}
	return &agentsv1.Skill{
		Name:          fm.Name,
		Description:   fm.Description,
		License:       fm.License,
		Compatibility: fm.Compatibility,
		Metadata:      fm.Metadata,
		AllowedTools:  fm.AllowedTools,
		SizeBytes:     int64(len(skillMD)),
	}, nil
}

func (s *SkillServiceServer) CreateSkill(ctx context.Context, req *connect.Request[agentsv1.CreateSkillRequest]) (*connect.Response[agentsv1.CreateSkillResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	sk, err := s.parseSkillMD(req.Msg.GetName(), req.Msg.GetSkillMd())
	if err != nil {
		return nil, err
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("creating skill", "workspace_id", wsID, "name", sk.GetName())
	created, err := repo.Create(ctx, wsID, sk, req.Msg.GetSkillMd())
	if err != nil {
		return nil, toConnectSkillError(err)
	}
	return connect.NewResponse(&agentsv1.CreateSkillResponse{Skill: created}), nil
}

func (s *SkillServiceServer) GetSkill(ctx context.Context, req *connect.Request[agentsv1.GetSkillRequest]) (*connect.Response[agentsv1.GetSkillResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetName() == "" {
		return nil, connectx.RequiredArgument("name")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	sk, err := repo.Get(ctx, wsID, req.Msg.GetName())
	if err != nil {
		return nil, toConnectSkillError(err)
	}
	md, err := repo.GetSkillMD(ctx, wsID, req.Msg.GetName())
	if err != nil {
		return nil, toConnectSkillError(err)
	}
	return connect.NewResponse(&agentsv1.GetSkillResponse{Skill: sk, SkillMd: md}), nil
}

func (s *SkillServiceServer) ListSkills(ctx context.Context, _ *connect.Request[agentsv1.ListSkillsRequest]) (*connect.Response[agentsv1.ListSkillsResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	skills, err := repo.List(ctx, wsID)
	if err != nil {
		return nil, toConnectSkillError(err)
	}
	return connect.NewResponse(&agentsv1.ListSkillsResponse{Skills: skills}), nil
}

func (s *SkillServiceServer) UpdateSkill(ctx context.Context, req *connect.Request[agentsv1.UpdateSkillRequest]) (*connect.Response[agentsv1.UpdateSkillResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	sk, err := s.parseSkillMD(req.Msg.GetName(), req.Msg.GetSkillMd())
	if err != nil {
		return nil, err
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := repo.Update(ctx, wsID, sk, req.Msg.GetSkillMd())
	if err != nil {
		return nil, toConnectSkillError(err)
	}
	return connect.NewResponse(&agentsv1.UpdateSkillResponse{Skill: updated}), nil
}

func (s *SkillServiceServer) DeleteSkill(ctx context.Context, req *connect.Request[agentsv1.DeleteSkillRequest]) (*connect.Response[agentsv1.DeleteSkillResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetName() == "" {
		return nil, connectx.RequiredArgument("name")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("deleting skill", "workspace_id", wsID, "name", req.Msg.GetName())
	if err := repo.Delete(ctx, wsID, req.Msg.GetName()); err != nil {
		return nil, toConnectSkillError(err)
	}
	return connect.NewResponse(&agentsv1.DeleteSkillResponse{}), nil
}

// resourceRequestContext validates the shared (skill_name, path) addressing
// of the resource RPCs: repo present, workspace on the context, and the path
// cleaned by CleanResourcePath (traversal or out-of-spec → InvalidArgument).
func (s *SkillServiceServer) resourceRequestContext(ctx context.Context, skillName, resourcePath string) (skillrepo.Repository, string, string, error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, "", "", err
	}
	if skillName == "" {
		return nil, "", "", connectx.RequiredArgument("skill_name")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, "", "", err
	}
	if resourcePath == "" {
		return repo, wsID, "", nil
	}
	cleaned, err := skillrepo.CleanResourcePath(resourcePath)
	if err != nil {
		return nil, "", "", connectx.InvalidArgument("path", err.Error())
	}
	return repo, wsID, cleaned, nil
}

func (s *SkillServiceServer) ListSkillResources(ctx context.Context, req *connect.Request[agentsv1.ListSkillResourcesRequest]) (*connect.Response[agentsv1.ListSkillResourcesResponse], error) {
	repo, wsID, _, err := s.resourceRequestContext(ctx, req.Msg.GetSkillName(), "")
	if err != nil {
		return nil, err
	}
	resources, err := repo.ListResources(ctx, wsID, req.Msg.GetSkillName())
	if err != nil {
		return nil, toConnectSkillError(err)
	}
	return connect.NewResponse(&agentsv1.ListSkillResourcesResponse{Resources: resources}), nil
}

func (s *SkillServiceServer) GetSkillResource(ctx context.Context, req *connect.Request[agentsv1.GetSkillResourceRequest]) (*connect.Response[agentsv1.GetSkillResourceResponse], error) {
	if req.Msg.GetPath() == "" {
		return nil, connectx.RequiredArgument("path")
	}
	repo, wsID, cleaned, err := s.resourceRequestContext(ctx, req.Msg.GetSkillName(), req.Msg.GetPath())
	if err != nil {
		return nil, err
	}
	resource, content, err := repo.GetResource(ctx, wsID, req.Msg.GetSkillName(), cleaned)
	if err != nil {
		return nil, toConnectSkillError(err)
	}
	return connect.NewResponse(&agentsv1.GetSkillResourceResponse{Resource: resource, Content: content}), nil
}

func (s *SkillServiceServer) PutSkillResource(ctx context.Context, req *connect.Request[agentsv1.PutSkillResourceRequest]) (*connect.Response[agentsv1.PutSkillResourceResponse], error) {
	if req.Msg.GetPath() == "" {
		return nil, connectx.RequiredArgument("path")
	}
	repo, wsID, cleaned, err := s.resourceRequestContext(ctx, req.Msg.GetSkillName(), req.Msg.GetPath())
	if err != nil {
		return nil, err
	}
	if int64(len(req.Msg.GetContent())) > skillResourceMaxBytes {
		return nil, connectx.InvalidArgument("content", fmt.Sprintf("resource exceeds max size of %d bytes", skillResourceMaxBytes))
	}
	// The count cap only guards new paths; overwriting an existing resource
	// never changes the count. ListResources doubles as the skill-existence
	// check, so a missing skill surfaces as NotFound before the cap.
	existing, err := repo.ListResources(ctx, wsID, req.Msg.GetSkillName())
	if err != nil {
		return nil, toConnectSkillError(err)
	}
	isOverwrite := false
	for _, res := range existing {
		if res.GetPath() == cleaned {
			isOverwrite = true
			break
		}
	}
	if !isOverwrite && len(existing) >= s.skillResourceMaxCount {
		return nil, connect.NewError(connect.CodeResourceExhausted,
			fmt.Errorf("skill %q already has %d resources (max %d)", req.Msg.GetSkillName(), len(existing), s.skillResourceMaxCount))
	}
	logger := log.FromContext(ctx)
	logger.Info("putting skill resource", "workspace_id", wsID, "skill", req.Msg.GetSkillName(), "path", cleaned, "size", len(req.Msg.GetContent()))
	stored, err := repo.PutResource(ctx, wsID, req.Msg.GetSkillName(), &agentsv1.SkillResource{
		Path:        cleaned,
		ContentType: req.Msg.GetContentType(),
	}, req.Msg.GetContent())
	if err != nil {
		return nil, toConnectSkillError(err)
	}
	return connect.NewResponse(&agentsv1.PutSkillResourceResponse{Resource: stored}), nil
}

func (s *SkillServiceServer) DeleteSkillResource(ctx context.Context, req *connect.Request[agentsv1.DeleteSkillResourceRequest]) (*connect.Response[agentsv1.DeleteSkillResourceResponse], error) {
	if req.Msg.GetPath() == "" {
		return nil, connectx.RequiredArgument("path")
	}
	repo, wsID, cleaned, err := s.resourceRequestContext(ctx, req.Msg.GetSkillName(), req.Msg.GetPath())
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("deleting skill resource", "workspace_id", wsID, "skill", req.Msg.GetSkillName(), "path", cleaned)
	if err := repo.DeleteResource(ctx, wsID, req.Msg.GetSkillName(), cleaned); err != nil {
		return nil, toConnectSkillError(err)
	}
	return connect.NewResponse(&agentsv1.DeleteSkillResourceResponse{}), nil
}

func toConnectSkillError(err error) *connect.Error {
	var cerr *connect.Error
	if errors.As(err, &cerr) {
		return cerr
	}
	if errors.Is(err, skillrepo.ErrNotFound) {
		return connectx.NotFound(err.Error())
	}
	if errors.Is(err, skillrepo.ErrAlreadyExists) {
		return connect.NewError(connect.CodeAlreadyExists, errors.New(err.Error()))
	}
	return connectx.InternalWith(err)
}
