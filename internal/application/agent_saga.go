package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/agentcontent"
	agentoprepo "go.orx.me/apps/butter/internal/repo/agentop"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// agentContentCoordinator is the narrow seam through which the Agent lifecycle
// Saga reaches Git-owned Agent Content. It takes plain arguments (no
// connect.Request, no transport types) so the domain path stays decoupled.
// Implemented by *RepoBindingServiceServer; faked in tests.
type agentContentCoordinator interface {
	// HasBinding reports whether the workspace has a repository binding.
	HasBinding(ctx context.Context, ws string) (bool, error)
	// IsContentGitOwned reports whether Agent Content is Git-owned — i.e. the
	// workspace is bound AND an Active Revision has been published. A bound but
	// not-yet-onboarded workspace is still database-owned, so its content stays
	// editable through the normal Agent API until onboarding succeeds (#219).
	IsContentGitOwned(ctx context.Context, ws string) (bool, error)
	// CommitContent applies a changeset as a single direct Git commit and
	// publishes it. Returns the commit SHA and any content validation errors
	// (non-empty means no commit was created).
	CommitContent(ctx context.Context, ws string, actions []*agentsv1.ContentFileAction, operation, message string) (commitSHA string, validationErrors []string, err error)
	// SyncAndPublish re-syncs the repository and publishes the Active Revision.
	SyncAndPublish(ctx context.Context, ws string) error
}

// agentOperationCoordinator drives durable, retryable Agent lifecycle Sagas
// (issue #218). It composes the single-step mutateWithRuntime seam (#183) into
// multi-step operations, persisting each step so partial failures are visible
// and operations resume idempotently. It coordinates DB config and Git content
// but never claims a distributed transaction.
type agentOperationCoordinator struct {
	agents                 configrepo.AgentRepository
	ops                    agentoprepo.Repository
	content                agentContentCoordinator
	reload                 func(ctx context.Context) error
	now                    func() time.Time
	leaseDuration          time.Duration
	leaseHeartbeatInterval time.Duration
}

const agentOperationLeaseDuration = 15 * time.Minute

func newAgentOperationCoordinator(agents configrepo.AgentRepository, ops agentoprepo.Repository, content agentContentCoordinator, reload func(context.Context) error) *agentOperationCoordinator {
	return &agentOperationCoordinator{
		agents:                 agents,
		ops:                    ops,
		content:                content,
		reload:                 reload,
		now:                    time.Now,
		leaseDuration:          agentOperationLeaseDuration,
		leaseHeartbeatInterval: agentOperationLeaseDuration / 3,
	}
}

func (c *agentOperationCoordinator) nowTS() *timestamppb.Timestamp {
	return timestamppb.New(c.now())
}

// ── step plumbing ───────────────────────────────────────────────────────────

// stepFunc performs one Saga step. It returns validationErrors (a soft failure
// surfaced to the caller with the operation marked FAILED, no Go error) or a
// hard err. It may mutate op (e.g. record committed_sha).
type stepFunc func(ctx context.Context, op *agentsv1.AgentOperation) (validationErrors []string, err error)

type sagaStep struct {
	kind agentsv1.AgentOperationStepKind
	fn   stepFunc
}

// execute runs the ordered steps, persisting operation state across each
// transition. Steps already recorded SUCCEEDED are skipped (idempotent
// resume). The persist context is detached from the caller so a client
// disconnect cannot corrupt the durable record.
func (c *agentOperationCoordinator) execute(ctx context.Context, op *agentsv1.AgentOperation, steps []sagaStep) ([]string, error) {
	pctx := context.WithoutCancel(ctx)
	logger := log.FromContext(ctx)
	ws := op.GetWorkspaceId()
	claimedAt := c.now()
	leaseToken := uuid.NewString()
	claimed, err := c.ops.Claim(pctx, ws, op.GetId(), leaseToken, claimedAt, claimedAt.Add(c.leaseDuration))
	if errors.Is(err, agentoprepo.ErrInProgress) || errors.Is(err, agentoprepo.ErrCompleted) {
		if claimed == nil {
			claimed, err = c.ops.Get(pctx, ws, op.GetId())
			if err != nil {
				return nil, connectx.InternalWith(err)
			}
		}
		replaceOperation(op, claimed)
		return nil, nil
	}
	if err != nil {
		return nil, connectx.InternalWith(fmt.Errorf("claim operation: %w", err))
	}
	replaceOperation(op, claimed)
	stepCtx, stopHeartbeat, heartbeatErr := c.startLeaseHeartbeat(ctx, pctx, ws, op.GetId(), leaseToken)
	defer stopHeartbeat()
	save := func() error {
		if err := c.ops.SaveClaimed(pctx, ws, leaseToken, op); err != nil {
			if errors.Is(err, agentoprepo.ErrLeaseLost) {
				if latest, getErr := c.ops.Get(pctx, ws, op.GetId()); getErr == nil {
					replaceOperation(op, latest)
				}
				return connect.NewError(connect.CodeAborted,
					fmt.Errorf("%w: reload the operation", agentoprepo.ErrLeaseLost))
			}
			return err
		}
		return nil
	}

	for _, st := range steps {
		if err := heartbeatErr(); err != nil {
			return nil, err
		}
		rec := findOrAddStep(op, st.kind)
		if rec.GetStatus() == agentsv1.AgentOperationStepStatus_AGENT_OPERATION_STEP_STATUS_SUCCEEDED {
			continue
		}
		rec.Status = agentsv1.AgentOperationStepStatus_AGENT_OPERATION_STEP_STATUS_PENDING
		rec.AttemptCount++
		rec.StartedAt = c.nowTS()
		rec.Error = ""
		op.UpdatedAt = c.nowTS()
		if err := save(); err != nil {
			return nil, connectx.InternalWith(fmt.Errorf("persist step start: %w", err))
		}

		vErrs, err := st.fn(stepCtx, op)
		if leaseErr := heartbeatErr(); leaseErr != nil {
			return nil, leaseErr
		}
		rec.FinishedAt = c.nowTS()
		switch {
		case len(vErrs) > 0:
			rec.Status = agentsv1.AgentOperationStepStatus_AGENT_OPERATION_STEP_STATUS_FAILED
			rec.Error = strings.Join(vErrs, "; ")
			op.Status = agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED
			op.Error = fmt.Sprintf("%s: content validation failed", st.kind.String())
			op.UpdatedAt = c.nowTS()
			if sErr := save(); sErr != nil {
				logger.Error("failed to persist step failure", "operation_id", op.GetId(), "step", st.kind.String(), "err", sErr)
				return vErrs, connectx.InternalWith(fmt.Errorf("persist failed step %s: %w", st.kind.String(), sErr))
			}
			logger.Warn("agent lifecycle step validation failed", "operation_id", op.GetId(), "step", st.kind.String(), "errors", vErrs)
			return vErrs, nil
		case err != nil:
			rec.Status = agentsv1.AgentOperationStepStatus_AGENT_OPERATION_STEP_STATUS_FAILED
			rec.Error = err.Error()
			op.Status = agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_FAILED
			op.Error = fmt.Sprintf("%s: %s", st.kind.String(), err.Error())
			op.UpdatedAt = c.nowTS()
			if sErr := save(); sErr != nil {
				logger.Error("failed to persist step failure", "operation_id", op.GetId(), "step", st.kind.String(), "err", sErr)
				return nil, connectx.InternalWith(fmt.Errorf("step %s failed: %v; persist failure: %w", st.kind.String(), err, sErr))
			}
			logger.Error("agent lifecycle step failed", "operation_id", op.GetId(), "step", st.kind.String(), "err", err)
			return nil, err
		default:
			rec.Status = agentsv1.AgentOperationStepStatus_AGENT_OPERATION_STEP_STATUS_SUCCEEDED
			op.UpdatedAt = c.nowTS()
			if sErr := save(); sErr != nil {
				logger.Error("failed to persist step success", "operation_id", op.GetId(), "step", st.kind.String(), "err", sErr)
				return nil, connectx.InternalWith(fmt.Errorf("persist step success: %w", sErr))
			}
		}
	}

	op.Status = agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED
	op.Error = ""
	op.UpdatedAt = c.nowTS()
	if err := save(); err != nil {
		return nil, connectx.InternalWith(fmt.Errorf("persist operation: %w", err))
	}
	return nil, nil
}

func (c *agentOperationCoordinator) startLeaseHeartbeat(ctx, persistCtx context.Context, workspaceID, operationID, leaseToken string) (context.Context, context.CancelFunc, func() error) {
	stepCtx, cancelStep := context.WithCancel(ctx)
	heartbeatCtx, cancelHeartbeat := context.WithCancel(persistCtx)
	errCh := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(c.leaseHeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-ticker.C:
				expiresAt := c.now().Add(c.leaseDuration)
				if err := c.ops.RenewLease(heartbeatCtx, workspaceID, operationID, leaseToken, expiresAt); err != nil {
					err = connect.NewError(connect.CodeAborted, fmt.Errorf("renew operation lease: %w", err))
					errCh <- err
					cancelStep()
					return
				}
			}
		}
	}()
	var leaseErr error
	readErr := func() error {
		if leaseErr != nil {
			return leaseErr
		}
		select {
		case leaseErr = <-errCh:
		default:
		}
		return leaseErr
	}
	stop := func() {
		cancelHeartbeat()
		cancelStep()
		<-done
	}
	return stepCtx, stop, readErr
}

func replaceOperation(dst, src *agentsv1.AgentOperation) {
	proto.Reset(dst)
	proto.Merge(dst, src)
}

func findOrAddStep(op *agentsv1.AgentOperation, kind agentsv1.AgentOperationStepKind) *agentsv1.AgentOperationStep {
	for _, s := range op.GetSteps() {
		if s.GetKind() == kind {
			return s
		}
	}
	s := &agentsv1.AgentOperationStep{Kind: kind, Status: agentsv1.AgentOperationStepStatus_AGENT_OPERATION_STEP_STATUS_PENDING}
	op.Steps = append(op.Steps, s)
	return s
}

// getOrCreate loads an existing operation by ID (idempotent resume) or creates
// a new PENDING one. An empty opID always creates a fresh operation.
func (c *agentOperationCoordinator) getOrCreate(ctx context.Context, ws, opID string, typ agentsv1.AgentOperationType, agent *agentsv1.Agent, requestJSON string, expectedVersion int64, baseCommitSHA string) (*agentsv1.AgentOperation, error) {
	if opID != "" {
		existing, err := c.ops.Get(ctx, ws, opID)
		if err == nil {
			if err := validateOperationIdentity(existing, typ, agent.GetAgentId(), requestJSON, expectedVersion, baseCommitSHA); err != nil {
				return nil, err
			}
			return existing, nil
		}
		if !errors.Is(err, agentoprepo.ErrNotFound) {
			return nil, connectx.InternalWith(err)
		}
	} else {
		opID = uuid.NewString()
	}
	op := &agentsv1.AgentOperation{
		Id:                   opID,
		WorkspaceId:          ws,
		AgentId:              agent.GetAgentId(),
		AgentName:            agent.GetName(),
		Type:                 typ,
		Status:               agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_PENDING,
		RequestJson:          requestJSON,
		ExpectedAgentVersion: expectedVersion,
		BaseCommitSha:        baseCommitSHA,
		CreatedAt:            c.nowTS(),
		UpdatedAt:            c.nowTS(),
	}
	if err := c.ops.Create(context.WithoutCancel(ctx), ws, op); err != nil {
		if !errors.Is(err, agentoprepo.ErrAlreadyExists) {
			return nil, connectx.InternalWith(err)
		}
		existing, getErr := c.ops.Get(ctx, ws, opID)
		if getErr != nil {
			return nil, connectx.InternalWith(getErr)
		}
		if identityErr := validateOperationIdentity(existing, typ, agent.GetAgentId(), requestJSON, expectedVersion, baseCommitSHA); identityErr != nil {
			return nil, identityErr
		}
		return existing, nil
	}
	return op, nil
}

func validateOperationIdentity(op *agentsv1.AgentOperation, typ agentsv1.AgentOperationType, agentID, requestJSON string, expectedVersion int64, baseCommitSHA string) error {
	if op.GetType() != typ || op.GetAgentId() != agentID || op.GetRequestJson() != requestJSON ||
		op.GetExpectedAgentVersion() != expectedVersion || op.GetBaseCommitSha() != baseCommitSHA {
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("operation %q is already bound to a different request", op.GetId()))
	}
	return nil
}

func (c *agentOperationCoordinator) resumeExisting(ctx context.Context, ws, opID string, typ agentsv1.AgentOperationType, agentID, requestJSON string, expectedVersion int64, baseCommitSHA string) (*agentsv1.Agent, *agentsv1.AgentOperation, []string, bool, error) {
	if opID == "" {
		return nil, nil, nil, false, nil
	}
	op, err := c.ops.Get(ctx, ws, opID)
	if errors.Is(err, agentoprepo.ErrNotFound) {
		return nil, nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, nil, false, connectx.InternalWith(err)
	}
	if err := validateOperationIdentity(op, typ, agentID, requestJSON, expectedVersion, baseCommitSHA); err != nil {
		return nil, op, nil, true, err
	}
	if op.GetStatus() == agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_RUNNING ||
		op.GetStatus() == agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		agent, _ := c.agents.GetAgentByID(ctx, ws, op.GetAgentId())
		return agent, op, nil, true, nil
	}
	agent, op, vErrs, err := c.Retry(ctx, ws, opID)
	return agent, op, vErrs, true, err
}

// setLifecycle flips an agent's lifecycle_status via the mutateWithRuntime seam
// (#183): write then reload, rolling the write back if the reload fails. It is
// idempotent (a no-op when the agent is already at target).
func (c *agentOperationCoordinator) setLifecycle(ctx context.Context, ws, name string, target agentsv1.AgentLifecycleStatus) error {
	for attempt := 0; attempt < 5; attempt++ {
		prev, err := c.agents.GetAgent(ctx, ws, name)
		if err != nil {
			return err
		}
		if prev.GetLifecycleStatus() == target {
			return nil
		}
		updated := proto.Clone(prev).(*agentsv1.Agent)
		updated.LifecycleStatus = target
		updated.UpdatedAt = c.nowTS()
		if target == agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
			updated.DeletedAt = c.nowTS()
		} else {
			updated.DeletedAt = nil
		}
		stored, err := c.agents.UpdateAgentCAS(ctx, ws, updated, prev.GetVersion())
		if errors.Is(err, configrepo.ErrVersionConflict) {
			continue
		}
		if err != nil {
			return err
		}
		if err := c.reload(ctx); err != nil {
			if _, rollbackErr := c.agents.UpdateAgentCAS(ctx, ws, prev, stored.GetVersion()); rollbackErr != nil {
				return connectx.InternalWith(fmt.Errorf("reload runtime: %w; lifecycle rollback failed: %v", err, rollbackErr))
			}
			if rollbackReloadErr := c.reload(ctx); rollbackReloadErr != nil {
				return connectx.InternalWith(fmt.Errorf("reload runtime: %w; rollback reload failed: %v", err, rollbackReloadErr))
			}
			return err
		}
		return nil
	}
	return connect.NewError(connect.CodeAborted,
		errors.New("agent version kept changing while updating lifecycle; retry the operation"))
}

// ── content action builders ─────────────────────────────────────────────────

// contentActionsFromInput builds the PUT changeset for an agent's initial
// content. Empty fields are omitted (absent file == empty value).
func contentActionsFromInput(agentID string, in *agentsv1.AgentContentInput) []*agentsv1.ContentFileAction {
	if in == nil {
		return nil
	}
	var actions []*agentsv1.ContentFileAction
	put := func(path, content string) {
		if strings.TrimSpace(content) == "" {
			return
		}
		actions = append(actions, &agentsv1.ContentFileAction{
			Path:      path,
			Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
			Content:   content,
		})
	}
	put(agentcontent.DescriptionPath(agentID), in.GetDescription())
	put(agentcontent.PromptPath(agentID), in.GetPrompt())
	put(agentcontent.GlobalPromptPath(agentID), in.GetGlobalPrompt())
	return actions
}

// ── Create ───────────────────────────────────────────────────────────────────

// RunCreate provisions a Git-bound agent: DB row in PROVISIONING → commit and
// publish initial content → activate. The agent must already be validated by
// the caller.
func (c *agentOperationCoordinator) RunCreate(ctx context.Context, ws string, agent *agentsv1.Agent, initial *agentsv1.AgentContentInput, opID string) (*agentsv1.Agent, *agentsv1.AgentOperation, []string, error) {
	reqJSON, err := createOperationRequestJSON(agent, initial)
	if err != nil {
		return nil, nil, nil, connectx.InternalWith(err)
	}
	op, err := c.getOrCreate(ctx, ws, opID, agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_CREATE, agent, reqJSON, 0, "")
	if err != nil {
		return nil, nil, nil, err
	}
	actions := contentActionsFromInput(agent.GetAgentId(), initial)
	vErrs, err := c.execute(ctx, op, c.createSteps(ws, agent, actions))
	if err != nil || len(vErrs) > 0 {
		if errors.Is(err, agentoprepo.ErrLeaseLost) {
			return nil, op, vErrs, err
		}
		if markErr := c.markProvisioningErrored(ctx, ws, agent.GetName()); markErr != nil {
			if err != nil {
				return nil, op, vErrs, connectx.InternalWith(fmt.Errorf("create failed: %v; mark agent ERROR: %w", err, markErr))
			}
			return nil, op, vErrs, connectx.InternalWith(fmt.Errorf("mark agent ERROR after validation failure: %w", markErr))
		}
		return nil, op, vErrs, err
	}
	final, _ := c.agents.GetAgentByID(ctx, ws, agent.GetAgentId())
	return final, op, nil, nil
}

// markProvisioningErrored transitions a still-PROVISIONING agent to ERROR after
// a create Saga fails, so the agent reflects the documented "Saga left it in a
// partial state, not runnable" contract (PRD §11.1). A later
// RetryAgentOperation moves it to ACTIVE.
func (c *agentOperationCoordinator) markProvisioningErrored(ctx context.Context, ws, name string) error {
	agent, err := c.agents.GetAgent(ctx, ws, name)
	if err != nil {
		if errors.Is(err, configrepo.ErrNotFound) {
			return nil // create rolled the row back, or it never existed
		}
		return err
	}
	if agent.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_PROVISIONING {
		return nil
	}
	return c.setLifecycle(ctx, ws, name, agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ERROR)
}

func (c *agentOperationCoordinator) createSteps(ws string, agent *agentsv1.Agent, actions []*agentsv1.ContentFileAction) []sagaStep {
	return []sagaStep{
		{
			kind: agentsv1.AgentOperationStepKind_AGENT_OPERATION_STEP_KIND_DB_PROVISION,
			fn: func(ctx context.Context, op *agentsv1.AgentOperation) ([]string, error) {
				toCreate := proto.Clone(agent).(*agentsv1.Agent)
				toCreate.LifecycleStatus = agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_PROVISIONING
				toCreate.Version = 1
				toCreate.CreatedAt = c.nowTS()
				toCreate.UpdatedAt = c.nowTS()
				toCreate.DeletedAt = nil
				_, err := mutateWithRuntime(
					func() (*agentsv1.Agent, error) { return c.agents.CreateAgent(ctx, ws, toCreate) },
					func() error { return c.reload(ctx) },
					func() error {
						if e := c.agents.DeleteAgent(ctx, ws, toCreate.GetName()); e != nil {
							return e
						}
						return c.reload(ctx)
					},
				)
				if errors.Is(err, configrepo.ErrAlreadyExists) {
					return nil, nil // idempotent resume: row already provisioned
				}
				return nil, err
			},
		},
		{
			kind: agentsv1.AgentOperationStepKind_AGENT_OPERATION_STEP_KIND_CONTENT_COMMIT,
			fn: func(ctx context.Context, op *agentsv1.AgentOperation) ([]string, error) {
				if op.GetCommittedSha() != "" || len(actions) == 0 {
					return nil, nil
				}
				sha, vErrs, err := c.content.CommitContent(ctx, ws, actions, "create", "Create agent "+agent.GetAgentId())
				if err != nil {
					return nil, err
				}
				if len(vErrs) > 0 {
					return vErrs, nil
				}
				op.CommittedSha = sha
				return nil, nil
			},
		},
		{
			kind: agentsv1.AgentOperationStepKind_AGENT_OPERATION_STEP_KIND_SYNC_PUBLISH,
			fn: func(ctx context.Context, op *agentsv1.AgentOperation) ([]string, error) {
				return nil, c.content.SyncAndPublish(ctx, ws)
			},
		},
		{
			kind: agentsv1.AgentOperationStepKind_AGENT_OPERATION_STEP_KIND_ACTIVATE,
			fn: func(ctx context.Context, op *agentsv1.AgentOperation) ([]string, error) {
				return nil, c.setLifecycle(ctx, ws, agent.GetName(), agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE)
			},
		},
	}
}

// ── Composite save ────────────────────────────────────────────────────────────

// RunUpdateConfiguration coordinates a DB operational patch with Git content
// changes: commit and publish content, then apply the patch with optimistic
// concurrency. The two sides retain separate ownership.
func (c *agentOperationCoordinator) RunUpdateConfiguration(ctx context.Context, ws string, prev, patch *agentsv1.Agent, contentChanges []*agentsv1.ContentFileAction, expectedVersion int64, baseCommitSHA, opID string) (*agentsv1.Agent, *agentsv1.AgentOperation, []string, error) {
	reqJSON, err := updateOperationRequestJSON(patch, contentChanges, expectedVersion, baseCommitSHA)
	if err != nil {
		return nil, nil, nil, connectx.InternalWith(err)
	}
	op, err := c.getOrCreate(ctx, ws, opID, agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_UPDATE_CONFIGURATION, prev, reqJSON, expectedVersion, baseCommitSHA)
	if err != nil {
		return nil, nil, nil, err
	}
	steps := c.updateSteps(ws, prev, patch, contentChanges, expectedVersion)
	vErrs, err := c.execute(ctx, op, steps)
	if err != nil || len(vErrs) > 0 {
		return nil, op, vErrs, err
	}
	final, _ := c.agents.GetAgent(ctx, ws, prev.GetName())
	return final, op, nil, nil
}

func (c *agentOperationCoordinator) updateSteps(ws string, prev, patch *agentsv1.Agent, contentChanges []*agentsv1.ContentFileAction, expectedVersion int64) []sagaStep {
	return []sagaStep{
		{
			kind: agentsv1.AgentOperationStepKind_AGENT_OPERATION_STEP_KIND_CONTENT_COMMIT,
			fn: func(ctx context.Context, op *agentsv1.AgentOperation) ([]string, error) {
				if op.GetCommittedSha() != "" || len(contentChanges) == 0 {
					return nil, nil
				}
				sha, vErrs, err := c.content.CommitContent(ctx, ws, contentChanges, "commit", "Update agent "+prev.GetAgentId())
				if err != nil {
					return nil, err
				}
				if len(vErrs) > 0 {
					return vErrs, nil
				}
				op.CommittedSha = sha
				return nil, nil
			},
		},
		{
			kind: agentsv1.AgentOperationStepKind_AGENT_OPERATION_STEP_KIND_SYNC_PUBLISH,
			fn: func(ctx context.Context, op *agentsv1.AgentOperation) ([]string, error) {
				if len(contentChanges) == 0 {
					return nil, nil
				}
				return nil, c.content.SyncAndPublish(ctx, ws)
			},
		},
		{
			kind: agentsv1.AgentOperationStepKind_AGENT_OPERATION_STEP_KIND_DB_PATCH,
			fn: func(ctx context.Context, op *agentsv1.AgentOperation) ([]string, error) {
				current, err := c.agents.GetAgentByID(ctx, ws, prev.GetAgentId())
				if err != nil {
					return nil, err
				}
				if current.GetVersion() != expectedVersion {
					if configPatchAlreadyApplied(current, patch) {
						return nil, c.reload(ctx)
					}
					return nil, connect.NewError(connect.CodeAborted,
						errors.New("agent version changed since it was read; re-read and retry"))
				}
				patched := applyConfigPatch(current, patch)
				_, err = mutateWithRuntime(
					func() (*agentsv1.Agent, error) { return c.agents.UpdateAgentCAS(ctx, ws, patched, expectedVersion) },
					func() error { return c.reload(ctx) },
					nil,
				)
				if errors.Is(err, configrepo.ErrVersionConflict) {
					return nil, connect.NewError(connect.CodeAborted,
						errors.New("agent version changed since it was read; re-read and retry"))
				}
				return nil, err
			},
		},
	}
}

func configPatchAlreadyApplied(current, patch *agentsv1.Agent) bool {
	desired := applyConfigPatch(current, patch)
	desired.Version = current.GetVersion()
	desired.UpdatedAt = current.GetUpdatedAt()
	return proto.Equal(current, desired)
}

// applyConfigPatch produces the agent to store: operational fields come from the
// patch; identity, lifecycle, and Git-owned content fields are preserved from
// prev. Version is handled by the CAS write.
func applyConfigPatch(prev, patch *agentsv1.Agent) *agentsv1.Agent {
	out := proto.Clone(patch).(*agentsv1.Agent)
	out.Name = prev.GetName()
	out.AgentId = prev.GetAgentId()
	out.LegacyName = prev.GetLegacyName()
	out.LifecycleStatus = prev.GetLifecycleStatus()
	out.WorkspaceId = prev.GetWorkspaceId()
	out.CreatedAt = prev.GetCreatedAt()
	out.DeletedAt = prev.GetDeletedAt()
	out.SubAgents = nil
	// Git-owned content fields: these are the sole source of truth for bound
	// workspaces and must never be written to the DB config store.
	out.Description = prev.GetDescription()
	if out.Config == nil && prev.GetConfig() != nil {
		out.Config = &agentsv1.AgentConfig{}
	}
	if out.Config != nil {
		out.Config.Instruction = prev.GetConfig().GetInstruction()
		out.Config.GlobalInstruction = prev.GetConfig().GetGlobalInstruction()
	}
	return out
}

// ── Delete (tombstone) ─────────────────────────────────────────────────────────

// RunDelete tombstones an agent: a single step flips lifecycle to DELETED and
// reloads the runner. No Git content is touched; the Agent ID stays reserved.
func (c *agentOperationCoordinator) RunDelete(ctx context.Context, ws string, agent *agentsv1.Agent, opID string) (*agentsv1.AgentOperation, error) {
	reqJSON, err := deleteOperationRequestJSON(agent)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	op, err := c.getOrCreate(ctx, ws, opID, agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_DELETE, agent, reqJSON, 0, "")
	if err != nil {
		return nil, err
	}
	_, err = c.execute(ctx, op, c.deleteSteps(ws, agent.GetName()))
	return op, err
}

func (c *agentOperationCoordinator) deleteSteps(ws, name string) []sagaStep {
	return []sagaStep{
		{
			kind: agentsv1.AgentOperationStepKind_AGENT_OPERATION_STEP_KIND_TOMBSTONE,
			fn: func(ctx context.Context, _ *agentsv1.AgentOperation) ([]string, error) {
				if err := c.setLifecycle(ctx, ws, name, agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETING); err != nil {
					return nil, err
				}
				return nil, c.setLifecycle(ctx, ws, name, agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED)
			},
		},
	}
}

// ── Restore ────────────────────────────────────────────────────────────────────

// RunRestore reactivates a tombstoned agent: publish the retained Git content,
// then flip lifecycle DELETED→ACTIVE.
func (c *agentOperationCoordinator) RunRestore(ctx context.Context, ws string, agent *agentsv1.Agent, opID string) (*agentsv1.Agent, *agentsv1.AgentOperation, []string, error) {
	reqJSON, err := restoreOperationRequestJSON(agent.GetAgentId())
	if err != nil {
		return nil, nil, nil, connectx.InternalWith(err)
	}
	op, err := c.getOrCreate(ctx, ws, opID, agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_RESTORE, agent, reqJSON, 0, "")
	if err != nil {
		return nil, nil, nil, err
	}
	vErrs, err := c.execute(ctx, op, c.restoreSteps(ws, agent.GetName()))
	if err != nil || len(vErrs) > 0 {
		return nil, op, vErrs, err
	}
	final, _ := c.agents.GetAgentByID(ctx, ws, agent.GetAgentId())
	return final, op, nil, nil
}

func (c *agentOperationCoordinator) restoreSteps(ws, name string) []sagaStep {
	return []sagaStep{
		{
			kind: agentsv1.AgentOperationStepKind_AGENT_OPERATION_STEP_KIND_SYNC_PUBLISH,
			fn: func(ctx context.Context, _ *agentsv1.AgentOperation) ([]string, error) {
				hasBinding, err := c.content.HasBinding(ctx, ws)
				if err != nil || !hasBinding {
					return nil, err
				}
				return nil, c.content.SyncAndPublish(ctx, ws)
			},
		},
		{
			kind: agentsv1.AgentOperationStepKind_AGENT_OPERATION_STEP_KIND_RESTORE_DB,
			fn: func(ctx context.Context, _ *agentsv1.AgentOperation) ([]string, error) {
				return nil, c.setLifecycle(ctx, ws, name, agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE)
			},
		},
	}
}

// ── Retry ──────────────────────────────────────────────────────────────────────

// Retry resumes a FAILED operation from its first unfinished step, rebuilding
// the Saga plan from the persisted request. Steps already SUCCEEDED are skipped.
func (c *agentOperationCoordinator) Retry(ctx context.Context, ws, opID string) (*agentsv1.Agent, *agentsv1.AgentOperation, []string, error) {
	op, err := c.ops.Get(ctx, ws, opID)
	if err != nil {
		if errors.Is(err, agentoprepo.ErrNotFound) {
			return nil, nil, nil, connectx.NotFound("operation not found")
		}
		return nil, nil, nil, connectx.InternalWith(err)
	}
	if op.GetStatus() == agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		agent, _ := c.agents.GetAgentByID(ctx, ws, op.GetAgentId())
		return agent, op, nil, nil
	}

	steps, err := c.rebuildSteps(ctx, ws, op)
	if err != nil {
		return nil, op, nil, err
	}
	vErrs, err := c.execute(ctx, op, steps)
	if err != nil || len(vErrs) > 0 {
		return nil, op, vErrs, err
	}
	agent, _ := c.agents.GetAgentByID(ctx, ws, op.GetAgentId())
	return agent, op, nil, nil
}

func (c *agentOperationCoordinator) rebuildSteps(ctx context.Context, ws string, op *agentsv1.AgentOperation) ([]sagaStep, error) {
	switch op.GetType() {
	case agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_CREATE:
		var req agentsv1.CreateAgentRequest
		if err := protojson.Unmarshal([]byte(op.GetRequestJson()), &req); err != nil {
			return nil, connectx.InternalWith(fmt.Errorf("decode create request: %w", err))
		}
		actions := contentActionsFromInput(req.GetAgent().GetAgentId(), req.GetInitialContent())
		return c.createSteps(ws, req.GetAgent(), actions), nil
	case agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_UPDATE_CONFIGURATION:
		var req agentsv1.UpdateAgentConfigurationRequest
		if err := protojson.Unmarshal([]byte(op.GetRequestJson()), &req); err != nil {
			return nil, connectx.InternalWith(fmt.Errorf("decode update request: %w", err))
		}
		prev, err := c.agents.GetAgentByID(ctx, ws, op.GetAgentId())
		if err != nil {
			return nil, toConnectError(err)
		}
		return c.updateSteps(ws, prev, req.GetAgentPatch(), req.GetContentChanges(), op.GetExpectedAgentVersion()), nil
	case agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_DELETE:
		return c.deleteSteps(ws, op.GetAgentName()), nil
	case agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_RESTORE:
		return c.restoreSteps(ws, op.GetAgentName()), nil
	default:
		return nil, connectx.InvalidArgument("operation", "is not retryable")
	}
}

func marshalJSON(m proto.Message) (string, error) {
	b, err := protojson.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func createOperationRequestJSON(agent *agentsv1.Agent, initial *agentsv1.AgentContentInput) (string, error) {
	return marshalJSON(&agentsv1.CreateAgentRequest{Agent: agent, InitialContent: initial})
}

func updateOperationRequestJSON(patch *agentsv1.Agent, contentChanges []*agentsv1.ContentFileAction, expectedVersion int64, baseCommitSHA string) (string, error) {
	return marshalJSON(&agentsv1.UpdateAgentConfigurationRequest{
		AgentPatch: patch, ContentChanges: contentChanges,
		ExpectedAgentVersion: expectedVersion, BaseCommitSha: baseCommitSHA,
	})
}

func deleteOperationRequestJSON(agent *agentsv1.Agent) (string, error) {
	return marshalJSON(&agentsv1.DeleteAgentRequest{Name: agent.GetName(), AgentId: agent.GetAgentId()})
}

func restoreOperationRequestJSON(agentID string) (string, error) {
	return marshalJSON(&agentsv1.RestoreAgentRequest{AgentId: agentID})
}
