package automation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Step templating: `{{ selector }}` placeholders in step inputs are resolved
// against the same roots the conditions read — `payload` (the trigger
// payload), `context` (automation name, workspace, trigger type), and
// `steps.<name>` (parsed output of every earlier step, e.g.
// `steps.ask.response` for an invoke_agent step). This is what lets a later
// step consume an earlier step's output instead of only static text.
//
// A missing selector fails the step explicitly — a placeholder that silently
// rendered empty would produce an agent prompt or webhook body that looks
// intentional but isn't.
//
// The optional `| json` filter renders the value as its JSON encoding
// (strings become quoted/escaped), which is what a placeholder inside
// call_webhook.payload_json needs to keep the body valid JSON.
var templateExpr = regexp.MustCompile(`\{\{([^{}]*)\}\}`)

func renderTemplate(s string, roots map[string]any) (string, error) {
	if !strings.Contains(s, "{{") {
		return s, nil
	}
	var firstErr error
	out := templateExpr.ReplaceAllStringFunc(s, func(m string) string {
		if firstErr != nil {
			return m
		}
		expr, filter, hasFilter := strings.Cut(m[2:len(m)-2], "|")
		selector := strings.TrimSpace(expr)
		if selector == "" {
			firstErr = fmt.Errorf("empty template selector in %q", m)
			return m
		}
		value, ok := selectValue(roots, selector)
		if !ok {
			firstErr = fmt.Errorf("template selector %q not found", selector)
			return m
		}
		if !hasFilter {
			rendered, err := templateValueToString(value)
			if err != nil {
				firstErr = fmt.Errorf("render %q: %w", selector, err)
				return m
			}
			return rendered
		}
		switch strings.TrimSpace(filter) {
		case "json":
			b, err := json.Marshal(value)
			if err != nil {
				firstErr = fmt.Errorf("render %q as JSON: %w", selector, err)
				return m
			}
			return string(b)
		default:
			firstErr = fmt.Errorf("unsupported template filter %q (only \"json\")", strings.TrimSpace(filter))
			return m
		}
	})
	if firstErr != nil {
		return "", firstErr
	}
	return out, nil
}

// templateValueToString renders a selected value as substitution text:
// scalars as their natural text, structured values as compact JSON.
func templateValueToString(v any) (string, error) {
	switch typed := v.(type) {
	case nil:
		return "", nil
	case string:
		return typed, nil
	case json.Number:
		return typed.String(), nil
	case bool:
		return fmt.Sprintf("%t", typed), nil
	default:
		b, err := json.Marshal(typed)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

// renderStepTemplates resolves placeholders in a step's inputs against the
// execution state. The step is cloned so retries and the persisted definition
// keep the original template text; only agent/address references (agent_id,
// thread_id, notify_group_name, webhook headers' keys) stay literal — they
// are validated resources, not content.
func renderStepTemplates(step *agentsv1.AutomationStep, state *executionState) (*agentsv1.AutomationStep, error) {
	if step == nil || state == nil {
		return step, nil
	}
	out := proto.Clone(step).(*agentsv1.AutomationStep)
	render := func(field string, s *string) error {
		rendered, err := renderTemplate(*s, state.roots)
		if err != nil {
			return fmt.Errorf("%s: %w", field, err)
		}
		*s = rendered
		return nil
	}
	if invoke := out.GetInvokeAgent(); invoke != nil {
		if err := render("invoke_agent.input", &invoke.Input); err != nil {
			return nil, err
		}
	}
	if webhook := out.GetCallWebhook(); webhook != nil {
		if err := render("call_webhook.url", &webhook.Url); err != nil {
			return nil, err
		}
		if err := render("call_webhook.payload_json", &webhook.PayloadJson); err != nil {
			return nil, err
		}
		for k := range webhook.GetHeaders() {
			v := webhook.Headers[k]
			if err := render(fmt.Sprintf("call_webhook.headers[%q]", k), &v); err != nil {
				return nil, err
			}
			webhook.Headers[k] = v
		}
	}
	if notify := out.GetSendNotifyGroup(); notify != nil {
		if err := render("send_notify_group.title", &notify.Title); err != nil {
			return nil, err
		}
		if err := render("send_notify_group.message", &notify.Message); err != nil {
			return nil, err
		}
	}
	if post := out.GetCreateForumPost(); post != nil {
		if err := render("create_forum_post.body", &post.Body); err != nil {
			return nil, err
		}
	}
	return out, nil
}
