package memoryhook

import (
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/plugin"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/mem0memory"
)

// MaxBlockChars caps the injected block. It rides on every model call of
// the turn, so it is kept small.
const MaxBlockChars = 4000

const blockHeader = `<memories>
These memories were recalled from earlier conversations. They are historical and may be outdated or superseded; when they conflict with the current conversation, the current conversation wins. Use them only when relevant, and do not mention this block itself.`

const blockFooter = "</memories>"

// InjectionPlugin returns the ADK plugin that appends the turn's recall
// block to the system instruction of every model call. Register it before
// plugins that inspect the final request (Langfuse, ContextGuard) so they
// see — and count — the injected text.
func InjectionPlugin() *plugin.Plugin {
	p, _ := plugin.New(plugin.Config{
		Name:                "memory_recall",
		BeforeModelCallback: llmagent.BeforeModelCallback(inject),
	})
	return p
}

func inject(ctx agent.Context, req *model.LLMRequest) (*model.LLMResponse, error) {
	block := TurnFromContext(ctx).Block()
	if block == "" || req == nil {
		return nil, nil
	}
	appendInstruction(req, block)
	return nil, nil
}

// appendInstruction mirrors ADK's internal utils.AppendInstructions. It is
// idempotent per request so a re-run callback cannot stack the block.
func appendInstruction(req *model.LLMRequest, text string) {
	if req.Config == nil {
		req.Config = &genai.GenerateContentConfig{}
	}
	si := req.Config.SystemInstruction
	if si == nil {
		req.Config.SystemInstruction = genai.NewContentFromText(text, genai.RoleUser)
		return
	}
	for _, p := range si.Parts {
		if p != nil && strings.Contains(p.Text, text) {
			return
		}
	}
	if n := len(si.Parts); n > 0 && si.Parts[n-1] != nil && si.Parts[n-1].Text != "" {
		si.Parts[n-1].Text += "\n\n" + text
		return
	}
	si.Parts = append(si.Parts, genai.NewPartFromText(text))
}

// FormatBlock renders recalled memories as the injected block: Workspace
// Memory and Agent Memory in separate sections, each entry dated, within
// maxChars runes. Entries that do not fit are dropped whole. It returns ""
// when nothing fits.
func FormatBlock(recalled []mem0memory.Recalled, maxChars int) string {
	sections := []struct {
		target mem0memory.Target
		title  string
	}{
		{mem0memory.TargetWorkspace, "Workspace memories (shared across this workspace):"},
		{mem0memory.TargetAgent, "Agent memories (specific to you):"},
	}
	budget := maxChars - utf8.RuneCountInString(blockHeader) - utf8.RuneCountInString(blockFooter) - 4
	var body strings.Builder
	entries := 0
	for _, sec := range sections {
		var lines []string
		for _, m := range recalled {
			if m.Target != sec.target || strings.TrimSpace(m.Memory.Memory) == "" {
				continue
			}
			lines = append(lines, "- "+datePrefix(m.CreatedAt)+strings.TrimSpace(m.Memory.Memory))
		}
		if len(lines) == 0 {
			continue
		}
		title := "\n" + sec.title
		if utf8.RuneCountInString(title)+utf8.RuneCountInString(lines[0])+1 > budget {
			break
		}
		body.WriteString(title)
		budget -= utf8.RuneCountInString(title)
		for _, line := range lines {
			cost := utf8.RuneCountInString(line) + 1
			if cost > budget {
				break
			}
			body.WriteString("\n" + line)
			budget -= cost
			entries++
		}
	}
	if entries == 0 {
		return ""
	}
	return blockHeader + "\n" + body.String() + "\n" + blockFooter
}

// datePrefix renders mem0's created_at as "[YYYY-MM-DD] ", or nothing when
// it is missing or unparseable.
func datePrefix(createdAt string) string {
	if createdAt == "" {
		return ""
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.999999", "2006-01-02"} {
		if ts, err := time.Parse(layout, createdAt); err == nil {
			return "[" + ts.Format("2006-01-02") + "] "
		}
	}
	return ""
}
