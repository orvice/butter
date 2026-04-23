package system

import (
	"context"
	"fmt"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/cron"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// --- Agent query tools ---

type listAgentsArgs struct{}

type agentInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type listAgentsResult struct {
	Agents []agentInfo `json:"agents"`
}

func newListAgentsTool(agentRepo configrepo.AgentRepository) (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "list_agents",
		Description: "List all registered agents with their names, types, and descriptions.",
	}, func(_ tool.Context, _ listAgentsArgs) (listAgentsResult, error) {
		agents, err := agentRepo.ListAgents(context.Background())
		if err != nil {
			return listAgentsResult{}, err
		}
		infos := make([]agentInfo, 0, len(agents))
		for _, a := range agents {
			infos = append(infos, agentInfo{
				Name:        a.GetName(),
				Description: a.GetDescription(),
				Type:        a.GetType().String(),
			})
		}
		return listAgentsResult{Agents: infos}, nil
	})
}

type getAgentArgs struct {
	Name string `json:"name" jsonschema:"the name of the agent to retrieve"`
}

type getAgentResult struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Model       string   `json:"model,omitempty"`
	Instruction string   `json:"instruction,omitempty"`
	SubAgents   []string `json:"sub_agents,omitempty"`
}

func newGetAgentTool(agentRepo configrepo.AgentRepository) (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "get_agent",
		Description: "Get detailed configuration of a specific agent by name.",
	}, func(_ tool.Context, args getAgentArgs) (getAgentResult, error) {
		a, err := agentRepo.GetAgent(context.Background(), args.Name)
		if err != nil {
			return getAgentResult{}, fmt.Errorf("agent %q not found", args.Name)
		}
		var subAgents []string
		for _, sub := range a.GetSubAgents() {
			subAgents = append(subAgents, sub.GetName())
		}
		return getAgentResult{
			Name:        a.GetName(),
			Description: a.GetDescription(),
			Type:        a.GetType().String(),
			Model:       a.GetConfig().GetModel(),
			Instruction: a.GetConfig().GetInstruction(),
			SubAgents:   subAgents,
		}, nil
	})
}

// --- Cron management tools ---

type listCronJobsArgs struct{}

type cronJobInfo struct {
	Name      string `json:"name"`
	Schedule  string `json:"schedule"`
	AgentName string `json:"agent_name"`
	Enabled   bool   `json:"enabled"`
	Timezone  string `json:"timezone,omitempty"`
	Input     string `json:"input,omitempty"`
}

type listCronJobsResult struct {
	Jobs []cronJobInfo `json:"jobs"`
}

func newListCronJobsTool(scheduler *cron.Scheduler) (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "list_cron_jobs",
		Description: "List all configured cron jobs with their schedule, agent, and enabled status.",
	}, func(tc tool.Context, _ listCronJobsArgs) (listCronJobsResult, error) {
		jobs, err := scheduler.ListJobs(context.Background())
		if err != nil {
			return listCronJobsResult{}, err
		}
		infos := make([]cronJobInfo, 0, len(jobs))
		for _, j := range jobs {
			infos = append(infos, cronJobInfo{
				Name:      j.GetName(),
				Schedule:  j.GetSchedule(),
				AgentName: j.GetAgentName(),
				Enabled:   j.GetEnabled(),
				Timezone:  j.GetTimezone(),
				Input:     j.GetInput(),
			})
		}
		return listCronJobsResult{Jobs: infos}, nil
	})
}

type createCronJobArgs struct {
	Name      string `json:"name" jsonschema:"unique name for the cron job"`
	Schedule  string `json:"schedule" jsonschema:"cron expression in 5-field format, e.g. 0 9 * * *"`
	AgentName string `json:"agent_name" jsonschema:"name of the agent to execute"`
	Input     string `json:"input,omitempty" jsonschema:"input text to send to the agent"`
	Timezone  string `json:"timezone,omitempty" jsonschema:"IANA timezone e.g. Asia/Shanghai, defaults to UTC"`
	Enabled   bool   `json:"enabled" jsonschema:"whether the job is enabled"`
}

type createCronJobResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func newCreateCronJobTool(scheduler *cron.Scheduler) (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "create_cron_job",
		Description: "Create a new cron job that runs an agent on a schedule.",
	}, func(tc tool.Context, args createCronJobArgs) (createCronJobResult, error) {
		job := &agentsv1.CronJob{
			Name:      args.Name,
			Schedule:  args.Schedule,
			AgentName: args.AgentName,
			Input:     args.Input,
			Timezone:  args.Timezone,
			Enabled:   args.Enabled,
		}
		if err := scheduler.AddJob(context.Background(), job); err != nil {
			return createCronJobResult{Success: false, Message: err.Error()}, nil
		}
		return createCronJobResult{Success: true, Message: fmt.Sprintf("Cron job %q created successfully", args.Name)}, nil
	})
}

type updateCronJobArgs struct {
	Name      string `json:"name" jsonschema:"name of the cron job to update"`
	Schedule  string `json:"schedule,omitempty" jsonschema:"new cron expression"`
	AgentName string `json:"agent_name,omitempty" jsonschema:"new agent name"`
	Input     string `json:"input,omitempty" jsonschema:"new input text"`
	Timezone  string `json:"timezone,omitempty" jsonschema:"new IANA timezone"`
	Enabled   bool   `json:"enabled" jsonschema:"whether the job is enabled"`
}

type updateCronJobResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func newUpdateCronJobTool(scheduler *cron.Scheduler) (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "update_cron_job",
		Description: "Update an existing cron job's schedule, agent, input, or enabled status.",
	}, func(tc tool.Context, args updateCronJobArgs) (updateCronJobResult, error) {
		existing, err := scheduler.GetJob(context.Background(), args.Name)
		if err != nil {
			return updateCronJobResult{Success: false, Message: fmt.Sprintf("cron job %q not found", args.Name)}, nil
		}
		if args.Schedule != "" {
			existing.Schedule = args.Schedule
		}
		if args.AgentName != "" {
			existing.AgentName = args.AgentName
		}
		if args.Input != "" {
			existing.Input = args.Input
		}
		if args.Timezone != "" {
			existing.Timezone = args.Timezone
		}
		existing.Enabled = args.Enabled

		if err := scheduler.UpdateJob(context.Background(), existing); err != nil {
			return updateCronJobResult{Success: false, Message: err.Error()}, nil
		}
		return updateCronJobResult{Success: true, Message: fmt.Sprintf("Cron job %q updated successfully", args.Name)}, nil
	})
}

type deleteCronJobArgs struct {
	Name string `json:"name" jsonschema:"name of the cron job to delete"`
}

type deleteCronJobResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func newDeleteCronJobTool(scheduler *cron.Scheduler) (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "delete_cron_job",
		Description: "Delete an existing cron job by name.",
	}, func(tc tool.Context, args deleteCronJobArgs) (deleteCronJobResult, error) {
		if err := scheduler.RemoveJob(context.Background(), args.Name); err != nil {
			return deleteCronJobResult{Success: false, Message: fmt.Sprintf("cron job %q not found", args.Name)}, nil
		}
		return deleteCronJobResult{Success: true, Message: fmt.Sprintf("Cron job %q deleted successfully", args.Name)}, nil
	})
}

type listCronExecutionsArgs struct {
	JobName  string `json:"job_name,omitempty" jsonschema:"filter by job name"`
	PageSize int32  `json:"page_size,omitempty" jsonschema:"number of results to return, default 10"`
}

type cronExecutionInfo struct {
	ID        string `json:"id"`
	JobName   string `json:"job_name"`
	AgentName string `json:"agent_name"`
	Status    string `json:"status"`
	Input     string `json:"input,omitempty"`
	Output    string `json:"output,omitempty"`
	StartedAt string `json:"started_at"`
}

type listCronExecutionsResult struct {
	Executions []cronExecutionInfo `json:"executions"`
}

func newListCronExecutionsTool(execRepo cron.ExecutionRepo) (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "list_cron_executions",
		Description: "List recent cron job execution history, optionally filtered by job name.",
	}, func(tc tool.Context, args listCronExecutionsArgs) (listCronExecutionsResult, error) {
		pageSize := args.PageSize
		if pageSize <= 0 {
			pageSize = 10
		}
		execs, _, err := execRepo.List(context.Background(), args.JobName, pageSize, "")
		if err != nil {
			return listCronExecutionsResult{}, err
		}
		infos := make([]cronExecutionInfo, 0, len(execs))
		for _, e := range execs {
			infos = append(infos, cronExecutionInfo{
				ID:        e.GetId(),
				JobName:   e.GetJobName(),
				AgentName: e.GetAgentName(),
				Status:    e.GetStatus().String(),
				Input:     e.GetInput(),
				Output:    e.GetOutput(),
				StartedAt: e.GetStartedAt().AsTime().Format("2006-01-02T15:04:05Z07:00"),
			})
		}
		return listCronExecutionsResult{Executions: infos}, nil
	})
}
