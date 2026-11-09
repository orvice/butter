package plane

import (
	"encoding/json"
	"strings"
)

// Plane's documented Agent payload uses workspace, agent_user and actor on
// resource objects. Keep the older aliases, but never let one override another
// contradictory identity (including across a data envelope).
type webhookRun struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Workspace   string `json:"workspace"`
	AgentUserID string `json:"agent_user"`
}

type webhookActivity struct {
	ID          string `json:"id"`
	Signal      string `json:"signal"`
	UserID      string `json:"user_id"`
	Actor       string `json:"actor"`
	Workspace   string `json:"workspace"`
	WorkspaceID string `json:"workspace_id"`
	AgentRunID  string `json:"agent_run"`
	Content     struct {
		Body string `json:"body"`
	} `json:"content"`
}

type webhookPayload struct {
	Prompt, Type, Action, AppClientID, AppInstallationID, AgentUserID, WorkspaceID string
	Run                                                                            webhookRun
	Activity                                                                       webhookActivity
}

type webhookFields struct {
	Prompt            string          `json:"prompt"`
	Type              string          `json:"type"`
	Action            string          `json:"action"`
	AppClientID       string          `json:"app_client_id"`
	AppInstallationID string          `json:"app_installation_id"`
	AgentUserID       string          `json:"agent_user_id"`
	WorkspaceID       string          `json:"workspace_id"`
	AgentRun          webhookRun      `json:"agent_run"`
	Run               webhookRun      `json:"run"`
	AgentRunActivity  webhookActivity `json:"agent_run_activity"`
	AgentActivity     webhookActivity `json:"agent_activity"`
	Activity          webhookActivity `json:"activity"`
}

func (p *webhookPayload) UnmarshalJSON(data []byte) error {
	var raw struct {
		webhookFields
		Data *webhookFields `json:"data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = webhookPayload{}
	fields := []webhookFields{raw.webhookFields}
	if raw.Data != nil {
		fields = append(fields, *raw.Data)
	}
	for _, f := range fields {
		for _, pair := range []struct {
			field  string
			target *string
			values []string
		}{
			{"type", &p.Type, []string{f.Type}},
			{"action", &p.Action, []string{f.Action}},
			{"app_client_id", &p.AppClientID, []string{f.AppClientID}},
			{"app_installation_id", &p.AppInstallationID, []string{f.AppInstallationID}},
			{"agent_user_id", &p.AgentUserID, []string{f.AgentUserID}},
			{"workspace_id", &p.WorkspaceID, []string{f.WorkspaceID}},
			{"agent_run.id", &p.Run.ID, []string{f.AgentRun.ID, f.Run.ID}},
			{"agent_run.workspace", &p.Run.WorkspaceID, []string{f.AgentRun.Workspace, f.AgentRun.WorkspaceID, f.Run.Workspace, f.Run.WorkspaceID}},
			{"agent_run.agent_user", &p.Run.AgentUserID, []string{f.AgentRun.AgentUserID, f.Run.AgentUserID}},
		} {
			if err := mergeWebhookIdentity(pair.field, pair.target, pair.values...); err != nil {
				return err
			}
		}
		if p.Prompt == "" {
			p.Prompt = f.Prompt
		}
		for _, a := range []webhookActivity{f.AgentRunActivity, f.AgentActivity, f.Activity} {
			for _, pair := range []struct {
				field  string
				target *string
				values []string
			}{
				{"activity.id", &p.Activity.ID, []string{a.ID}},
				{"activity.actor", &p.Activity.UserID, []string{a.Actor, a.UserID}},
				{"activity.signal", &p.Activity.Signal, []string{a.Signal}},
				{"activity.workspace", &p.Activity.WorkspaceID, []string{a.Workspace, a.WorkspaceID}},
				{"activity.agent_run", &p.Activity.AgentRunID, []string{a.AgentRunID}},
			} {
				if err := mergeWebhookIdentity(pair.field, pair.target, pair.values...); err != nil {
					return err
				}
			}
			if p.Prompt == "" {
				p.Prompt = a.Content.Body
			}
		}
	}
	return nil
}

func mergeWebhookIdentity(field string, target *string, values ...string) error {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if *target != "" && *target != value {
			return rejectWebhookIdentity("conflict", field, *target, value)
		}
		*target = value
	}
	return nil
}
