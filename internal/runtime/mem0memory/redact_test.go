package mem0memory

import (
	"strings"
	"testing"
)

func TestRedactRemovesSecrets(t *testing.T) {
	cases := []struct {
		in       string
		leak     string
		mustKeep string
	}{
		{in: "my key is sk-ant-api03-abcdefghijklmnopqrstuvwxyz0123", leak: "abcdefghijklmnop", mustKeep: "my key is"},
		{in: "token ghp_abcdefghijklmnopqrstuvwxyz0123456789", leak: "ghp_abcdef"},
		{in: "use github_pat_11ABCDEFG0123456789_abcdefghijklmnop", leak: "github_pat_11ABC"},
		{in: "gitlab glpat-abcdefghijklmnopqrst12", leak: "glpat-abc"},
		{in: "mem0 m0sk_live_12345678", leak: "m0sk_live"},
		{in: "aws AKIAABCDEFGHIJKLMNOP", leak: "AKIAABCDEF"},
		{in: "google AIzaSyA1234567890abcdefghijklmnopqrstuv", leak: "AIzaSyA123"},
		{in: "slack xoxb-1234567890-abcdefghij", leak: "xoxb-12345"},
		{in: "bot 123456789:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsawA", leak: "AAHdqTcvCH1"},
		{in: "jwt eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N", leak: "eyJhbGciOiJIUzI1"},
		{in: "Authorization: Bearer abcdefghijklmnop1234567890", leak: "abcdefghijklmnop123", mustKeep: "Bearer [REDACTED]"},
		{in: "set api_key: hunter2hunter2 now", leak: "hunter2", mustKeep: "api_key: [REDACTED] now"},
		{in: `DB_PASSWORD="correct horse battery"`, leak: "correct horse", mustKeep: "DB_PASSWORD=[REDACTED]"},
		{in: "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA\n-----END RSA PRIVATE KEY-----", leak: "MIIEpAIBAAKCAQEA"},
	}
	for _, tc := range cases {
		got := Redact(tc.in)
		if strings.Contains(got, tc.leak) {
			t.Errorf("Redact(%q) = %q, still contains %q", tc.in, got, tc.leak)
		}
		if !strings.Contains(got, redacted) {
			t.Errorf("Redact(%q) = %q, no redaction marker", tc.in, got)
		}
		if tc.mustKeep != "" && !strings.Contains(got, tc.mustKeep) {
			t.Errorf("Redact(%q) = %q, want it to keep %q", tc.in, got, tc.mustKeep)
		}
	}
}

func TestRedactLeavesOrdinaryTextAlone(t *testing.T) {
	for _, in := range []string{
		"we deploy on fridays with pnpm",
		"the token budget is 128k and the password policy requires rotation",
		"call the /api/v1/tokens endpoint",
		"sk-short",
	} {
		if got := Redact(in); got != in {
			t.Errorf("Redact(%q) = %q, want unchanged", in, got)
		}
	}
}
