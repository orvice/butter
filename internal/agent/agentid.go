package agent

import (
	"fmt"
	"regexp"
)

// agentIDPattern defines valid Agent ID slugs: 1–64 lowercase alphanumeric
// characters and hyphens, must start and end with an alphanumeric character.
var agentIDPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$`)

// reservedAgentIDs are slugs that may not be assigned as Agent IDs because
// they collide with system routes, sentinel values, or planned identifiers.
var reservedAgentIDs = map[string]bool{
	"user":    true,
	"system":  true,
	"admin":   true,
	"start":   true,
	"default": true,
	"api":     true,
	"new":     true,
}

// ValidateAgentID checks that id is a valid Agent ID slug. It returns nil
// when the value is acceptable for assignment.
func ValidateAgentID(id string) error {
	if id == "" {
		return fmt.Errorf("agent_id must not be empty")
	}
	if !agentIDPattern.MatchString(id) {
		return fmt.Errorf("agent_id %q is invalid: must be 1-64 lowercase alphanumeric characters or hyphens, starting and ending with alphanumeric", id)
	}
	if reservedAgentIDs[id] {
		return fmt.Errorf("agent_id %q is reserved", id)
	}
	return nil
}
