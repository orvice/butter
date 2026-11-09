package plane

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Identity diagnostics are safe even if the signed payload is misconfigured
// or attacker-controlled: no original value is formatted into the error.
func rejectWebhookIdentity(reason, field, expected, actual string) error {
	return fmt.Errorf("%w: reason=%s field=%s expected_present=%t expected_sha256=%s actual_present=%t actual_sha256=%s",
		ErrIdentity, reason, field, expected != "", identityHash(expected), actual != "", identityHash(actual))
}

func identityHash(value string) string {
	if value == "" {
		return "-"
	}
	digest := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", digest[:6])
}

func resolveWebhookInstallation(p *webhookPayload, installations []*agentsv1.PlaneInstallation) (*agentsv1.PlaneInstallation, error) {
	candidates := installations
	if p.AppInstallationID != "" {
		candidates = nil
		for _, inst := range installations {
			if inst.GetAppInstallationId() == p.AppInstallationID {
				candidates = append(candidates, inst)
			}
		}
	}
	if len(candidates) == 0 {
		return nil, rejectWebhookIdentity("installation_not_found", "app_installation_id", "", p.AppInstallationID)
	}
	workspaceMatches := []*agentsv1.PlaneInstallation{}
	for _, inst := range candidates {
		if inst.GetPlaneWorkspaceId() == p.WorkspaceID {
			workspaceMatches = append(workspaceMatches, inst)
		}
	}
	if len(workspaceMatches) == 0 {
		return nil, rejectWebhookIdentity("workspace_mismatch", "workspace_id", candidates[0].GetPlaneWorkspaceId(), p.WorkspaceID)
	}
	matches := []*agentsv1.PlaneInstallation{}
	for _, inst := range workspaceMatches {
		if inst.GetAppBotUserId() == p.AgentUserID {
			matches = append(matches, inst)
		}
	}
	if len(matches) == 0 {
		return nil, rejectWebhookIdentity("bot_user_mismatch", "agent_user_id", workspaceMatches[0].GetAppBotUserId(), p.AgentUserID)
	}
	// Official Plane Agent webhooks omit app_installation_id. A verified App
	// secret plus an unambiguous stored workspace/bot binding locates it; never
	// take the first installation or trust a payload-supplied replacement ID.
	if len(matches) != 1 {
		return nil, rejectWebhookIdentity("installation_ambiguous", "app_installation_id", "", p.AppInstallationID)
	}
	inst := matches[0]
	if inst.GetAppInstallationId() == "" {
		return nil, rejectWebhookIdentity("missing", "installation.app_installation_id", "", "")
	}
	if inst.GetStatus() != "installed" {
		return nil, rejectWebhookIdentity("installation_inactive", "installation.status", "installed", inst.GetStatus())
	}
	var missing []string
	for _, required := range requiredAgentScopes {
		if !planeScopeGranted(inst.GetGrantedScopes(), required) {
			missing = append(missing, required)
		}
	}
	if len(missing) != 0 {
		// Only public names from the fixed requirement list, never arbitrary
		// stored grant values. Report every missing permission in one pass.
		err := rejectWebhookIdentity("scope_incomplete", missing[0], missing[0], "")
		return nil, fmt.Errorf("%w missing_scopes=%s", err, strings.Join(missing, ","))
	}
	return inst, nil
}

// Plane's All Workspace Resources grant uses literal read/write scopes. A
// global verb covers only that verb; unrelated granular grants and lookalikes
// must never satisfy the requirement. Evaluate stored provider grants only.
func planeScopeGranted(granted []string, required string) bool {
	if slices.Contains(granted, required) {
		return true
	}
	_, verb, ok := strings.Cut(required, ":")
	return ok && (verb == "read" || verb == "write") && slices.Contains(granted, verb)
}

func planeEventLogFields(e *Event) []any {
	return []any{"app_hash", identityHash(e.AppID), "installation_hash", identityHash(e.PlaneInstallationID),
		"run_hash", identityHash(e.AgentRunID), "activity_hash", identityHash(e.ActivityID), "invocation_hash", identityHash(e.InvocationID)}
}
