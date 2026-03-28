package telegram

import "testing"

func TestParseAgentCommand(t *testing.T) {
	tests := []struct {
		input      string
		wantSub    string
		wantArg    string
	}{
		{"/agent list", "list", ""},
		{"/agent summarizer", "switch", "summarizer"},
		{"/agent", "list", ""},
		{"/agent  list", "list", ""},
		{"/agent  foo", "switch", "foo"},
		{"hello", "", ""},
		{"/start", "", ""},
		{" /agent list ", "list", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sub, arg := parseAgentCommand(tt.input)
			if sub != tt.wantSub {
				t.Errorf("parseAgentCommand(%q) sub = %q, want %q", tt.input, sub, tt.wantSub)
			}
			if arg != tt.wantArg {
				t.Errorf("parseAgentCommand(%q) arg = %q, want %q", tt.input, arg, tt.wantArg)
			}
		})
	}
}
