package mem0memory

import "regexp"

// redacted replaces a secret found in text sent to mem0.
const redacted = "[REDACTED]"

// secretPatterns match common credential shapes. Memory is shared across the
// whole workspace (ADR-0013 §3), so a key pasted into one conversation must
// not become a recallable "fact" for every agent. The list is deliberately
// conservative: shapes with a distinctive prefix, plus `key = value`
// assignments whose key names a secret.
var secretPatterns = []*regexp.Regexp{
	// PEM private keys, whole block.
	regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`),
	// OpenAI / Anthropic style keys (sk-..., sk-ant-..., sk-proj-...).
	regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{20,}`),
	// GitHub tokens.
	regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{30,}`),
	regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{22,}`),
	// GitLab personal access tokens.
	regexp.MustCompile(`\bglpat-[A-Za-z0-9_-]{20,}`),
	// mem0 API keys.
	regexp.MustCompile(`\bm0sk_[A-Za-z0-9_-]{8,}`),
	// AWS access key IDs.
	regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`),
	// Google API keys.
	regexp.MustCompile(`\bAIza[0-9A-Za-z_-]{35}`),
	// Slack tokens.
	regexp.MustCompile(`\bxox[abprs]-[A-Za-z0-9-]{10,}`),
	// Telegram bot tokens.
	regexp.MustCompile(`\b\d{8,10}:[A-Za-z0-9_-]{35}\b`),
	// JSON Web Tokens.
	regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`),
}

// bearerPattern keeps the scheme and redacts the credential.
var bearerPattern = regexp.MustCompile(`(?i)\b(bearer|basic)\s+[A-Za-z0-9._~+/=-]{16,}`)

// assignmentPattern keeps a secret-named key and redacts its value, e.g.
// `api_key: abc123` or `PASSWORD=hunter2`.
var assignmentPattern = regexp.MustCompile(`(?i)\b([A-Za-z0-9_-]*(?:api[_-]?key|secret|password|passwd|token|access[_-]?key)[A-Za-z0-9_-]*)(\s*[:=]\s*)("[^"]*"|'[^']*'|\S+)`)

// Redact removes common secret shapes from text before it is sent to mem0.
func Redact(text string) string {
	for _, p := range secretPatterns {
		text = p.ReplaceAllString(text, redacted)
	}
	text = bearerPattern.ReplaceAllString(text, "$1 "+redacted)
	return assignmentPattern.ReplaceAllString(text, "$1$2"+redacted)
}
