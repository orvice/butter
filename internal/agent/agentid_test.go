package agent

import "testing"

func TestValidateAgentID(t *testing.T) {
	valid := []string{
		"a",
		"my-agent",
		"agent-1",
		"a1b2c3",
		"x",
		"a-b-c-d",
	}
	for _, id := range valid {
		if err := ValidateAgentID(id); err != nil {
			t.Errorf("ValidateAgentID(%q) = %v; want nil", id, err)
		}
	}

	invalid := []struct {
		id   string
		desc string
	}{
		{"", "empty"},
		{"-abc", "leading hyphen"},
		{"abc-", "trailing hyphen"},
		{"ABC", "uppercase"},
		{"my agent", "space"},
		{"my_agent", "underscore"},
		{"a/b", "slash"},
		{"user", "reserved"},
		{"system", "reserved"},
		{"admin", "reserved"},
		{"start", "reserved"},
		{"default", "reserved"},
		{"api", "reserved"},
		{"new", "reserved"},
		{string(make([]byte, 65)), "too long"},
	}
	for _, tc := range invalid {
		if err := ValidateAgentID(tc.id); err == nil {
			t.Errorf("ValidateAgentID(%q) [%s] = nil; want error", tc.id, tc.desc)
		}
	}
}
