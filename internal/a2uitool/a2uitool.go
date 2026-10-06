// Package a2uitool provides render_ui, the tool an LLM agent uses to show a
// read-only result card in an A2UI-capable AG-UI client.
//
// The tool exists for one run at a time: its Toolset is attached to every
// LLM agent but offers render_ui only when the run's context carries an
// a2ui.Run — an AG-UI run whose client negotiated A2UI and whose session is
// bound to the caller. Every other entry point (Telegram, the classic chat,
// cron, OpenAI-compatible, Pi/Cursor leaves) sees no tool and pays nothing.
// Each Toolset is also built with the Card Policy of its agent's place in the
// tree, and offers nothing where that policy turns cards off.
//
// A batch is validated completely before anything is written; the card is
// then persisted through the tool's state delta and reaches the client only
// after the runner has stored that event.
package a2uitool

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"

	"go.orx.me/apps/butter/internal/a2ui"
)

// ToolName is the model-facing name of the tool.
const ToolName = "render_ui"

// Toolset offers render_ui for A2UI runs only, where its Card Policy allows
// cards.
type Toolset struct {
	policy a2ui.CardPolicy
	tool   tool.Tool
}

// NewToolset builds the render_ui toolset of an agent whose place in the
// agent tree has policy. It is inert outside A2UI runs and wherever policy
// turns cards off. Offering the tool and running it read the same policy, so
// a run sees one policy from start to end.
func NewToolset(policy a2ui.CardPolicy) Toolset {
	handler := func(ctx agent.Context, args renderArgs) (renderResult, error) {
		return render(ctx, policy, args)
	}
	t, err := functiontool.New(functiontool.Config{Name: ToolName, Description: description()}, handler)
	if err != nil {
		// The handler signature is fixed at compile time; a failure here is
		// a programming error, not a runtime condition.
		panic(fmt.Sprintf("a2uitool: build %s: %v", ToolName, err))
	}
	return Toolset{policy: policy, tool: t}
}

var _ tool.Toolset = Toolset{}

func (Toolset) Name() string { return "a2ui" }

func (t Toolset) Tools(ctx agent.ReadonlyContext) ([]tool.Tool, error) {
	if !t.policy.Allowed() {
		return nil, nil
	}
	if _, ok := a2ui.RunFrom(ctx); !ok {
		return nil, nil
	}
	return []tool.Tool{t.tool}, nil
}

type renderArgs struct {
	SurfaceID string           `json:"surface_id,omitempty" jsonschema:"Handle of a card you created earlier in this conversation, to update or delete it. Omit to create a new card; the server assigns the handle and returns it."`
	Messages  []map[string]any `json:"messages" jsonschema:"A2UI v0.9.1 messages applied in order: {\"updateComponents\": {\"components\": [...]}}, {\"updateDataModel\": {\"path\": \"/\", \"value\": {...}}}, or {\"deleteSurface\": {}}."`
	Fallback  string           `json:"fallback,omitempty" jsonschema:"Plain-text version of the card for clients that cannot render it. Required when creating."`
}

type renderResult struct {
	SurfaceID string `json:"surface_id"`
	Revision  int    `json:"revision"`
	Status    string `json:"status"`
}

// maxConsecutiveFailures is how many render_ui calls in a row may fail in
// one run before the tool refuses the rest. Every failure hands the model an
// error it can try to fix, and nothing else bounds how often it tries: ADK
// caps model calls only in live runs.
const maxConsecutiveFailures = 3

// errRenderStopped tells the model why render_ui refuses the rest of a run.
var errRenderStopped = fmt.Errorf("render_ui has failed %d times in a row, so cards are off for the rest of this turn; answer in text", maxConsecutiveFailures)

func render(ctx agent.Context, policy a2ui.CardPolicy, args renderArgs) (renderResult, error) {
	run, ok := a2ui.RunFrom(ctx)
	if !ok || !policy.Allowed() {
		return renderResult{}, errors.New("cards cannot be shown in this conversation; answer in text instead")
	}
	if run.RenderFailures() >= maxConsecutiveFailures {
		return renderResult{}, errRenderStopped
	}
	res, err := apply(ctx, run, args)
	if err != nil {
		if run.RenderFailed() >= maxConsecutiveFailures {
			return renderResult{}, fmt.Errorf("%w; %w", err, errRenderStopped)
		}
		return renderResult{}, err
	}
	run.RenderSucceeded()
	return res, nil
}

// apply validates one batch against the session's cards and persists the
// card it produces.
func apply(ctx agent.Context, run *a2ui.Run, args renderArgs) (renderResult, error) {
	cards := a2ui.Cards(ctx.State())
	res, err := a2ui.Apply(cards, a2ui.Batch{
		SurfaceID: strings.TrimSpace(args.SurfaceID),
		Messages:  args.Messages,
		Fallback:  args.Fallback,
	}, a2ui.Anchor{
		ThreadID:     run.ThreadID,
		RunID:        run.RunID,
		MessageID:    run.MessageID,
		InvocationID: ctx.InvocationID(),
		At:           time.Now().UTC(),
	}, a2ui.DefaultLimits)
	if err != nil {
		return renderResult{}, fmt.Errorf("card not rendered, nothing was changed: %w", err)
	}
	card := res.Card
	if err := ctx.State().Set(a2ui.CardKey(card.ID), card.StateValue()); err != nil {
		return renderResult{}, fmt.Errorf("card not rendered: %w", err)
	}
	status := "updated"
	switch {
	case res.Created:
		status = "created"
	case card.Deleted:
		status = "deleted"
	}
	return renderResult{SurfaceID: card.ID, Revision: card.Revision, Status: status}, nil
}

func description() string {
	limits := a2ui.DefaultLimits
	return fmt.Sprintf(`Show a read-only result card to the user, next to your text answer. Use it when a structured view (a summary with key facts, a status, a short list of results) is easier to read than prose. The card cannot collect input or contain buttons; keep answering in text as well.

Pass A2UI v0.9.1 messages. To create a card, omit surface_id and send an updateComponents message whose components form a tree under a component with id "root", optionally followed by updateDataModel. The result returns the card's surface_id; pass it later to update the same card (send only the components that change, and/or new data) or to remove it with {"deleteSurface": {}}.

Components (catalog butter-basic-v1). Every component is {"id": "...", "component": "<Name>", ...properties}; children are referenced by id:
- Card: {"child": "<id>"} — a bordered container for one child (usually a Column).
- Column / Row: {"children": ["<id>", ...], "justify"?: start|center|end|spaceBetween|spaceAround|spaceEvenly|stretch, "align"?: start|center|end|stretch}.
- Text: {"text": <text>, "variant"?: h1|h2|h3|h4|h5|caption|body}.
- KeyValue: {"label": <text>, "value": <text>} — one labelled result.
- Status: {"text": <text>, "tone"?: neutral|info|success|warning|error}.
- Divider: {"axis"?: horizontal|vertical}.
<text> is a plain string, or {"path": "/key"} to read a string from the card's data model (set with {"updateDataModel": {"path": "/", "value": {"key": "..."}}}).

Rules: plain text only (no HTML, markdown links, images or URLs); at most %d components per card, %d KiB per call, and %d cards per conversation. An invalid call changes nothing and returns an error explaining why; fix it or answer in text. After %d failed calls in a row, render_ui refuses the rest of this turn.

Example: {"messages": [{"updateComponents": {"components": [{"id": "root", "component": "Card", "child": "col"}, {"id": "col", "component": "Column", "children": ["title", "env", "state"]}, {"id": "title", "component": "Text", "text": "Deploy summary", "variant": "h3"}, {"id": "env", "component": "KeyValue", "label": "Environment", "value": "production"}, {"id": "state", "component": "Status", "text": "Healthy", "tone": "success"}]}}], "fallback": "Deploy summary: production, healthy."}`,
		limits.MaxComponents, limits.MaxBatchBytes>>10, limits.MaxSurfaces, maxConsecutiveFailures)
}
