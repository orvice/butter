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

type SkillServiceServer struct {
	repo            skillrepo.Repository
	skillMDMaxBytes int64
}

func NewSkillServiceServer(repo skillrepo.Repository) *SkillServiceServer {
	return &SkillServiceServer{repo: repo, skillMDMaxBytes: defaultSkillMDMaxBytes}
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
