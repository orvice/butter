package linear

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/linearapi"
)

// defaultProgressInterval is the most often one progress activity posts.
const defaultProgressInterval = 3 * time.Second

// progress turns a turn's events into Linear activities without ever
// blocking the run: observe only records the latest update, and one worker
// posts at most one activity per interval, dropping consecutive repeats.
// Model reasoning and text deltas are never forwarded.
type progress struct {
	send     func(linearapi.Activity)
	interval time.Duration

	mu      sync.Mutex
	pending *linearapi.Activity
	lastKey string

	wake   chan struct{}
	done   chan struct{}
	exited chan struct{}
}

func newProgress(send func(linearapi.Activity), interval time.Duration) *progress {
	if interval <= 0 {
		interval = defaultProgressInterval
	}
	p := &progress{
		send:     send,
		interval: interval,
		wake:     make(chan struct{}, 1),
		done:     make(chan struct{}),
		exited:   make(chan struct{}),
	}
	go p.run()
	return p
}

// observe is a runner.EventCallback.
func (p *progress) observe(evt *session.Event) {
	if evt == nil || evt.Partial || evt.Content == nil {
		return
	}
	for _, part := range evt.Content.Parts {
		if part == nil || part.FunctionCall == nil || part.FunctionCall.Name == "" {
			continue
		}
		verb, parameter := describeTool(part.FunctionCall.Name, part.FunctionCall.Args)
		p.offer(linearapi.Activity{Type: linearapi.ActivityAction, Action: verb, Parameter: parameter})
	}
}

// compaction is a runner.CompactionCallback.
func (p *progress) compaction(string) {
	p.offer(linearapi.Activity{Type: linearapi.ActivityThought, Body: "Compacting the conversation context.", Ephemeral: true})
}

func (p *progress) offer(a linearapi.Activity) {
	key := a.Type + "\x00" + a.Action + "\x00" + a.Parameter + "\x00" + a.Body
	p.mu.Lock()
	if key == p.lastKey {
		p.mu.Unlock()
		return
	}
	p.lastKey = key
	p.pending = &a
	p.mu.Unlock()
	select {
	case p.wake <- struct{}{}:
	default:
	}
}

func (p *progress) run() {
	defer close(p.exited)
	for {
		select {
		case <-p.done:
			return
		case <-p.wake:
		}
		p.mu.Lock()
		a := p.pending
		p.pending = nil
		p.mu.Unlock()
		if a == nil {
			continue
		}
		p.send(*a)
		select {
		case <-p.done:
			return
		case <-time.After(p.interval):
		}
	}
}

// close drops any unsent update and waits for an in-flight post, so nothing
// lands after the turn's final activity.
func (p *progress) close() {
	close(p.done)
	<-p.exited
}

// toolVerbs names what well-known tools are doing, keyed by lowercase name.
var toolVerbs = map[string]string{
	"bash":          "Running",
	"shell":         "Running",
	"run_command":   "Running",
	"exec":          "Running",
	"read":          "Reading",
	"read_file":     "Reading",
	"write":         "Writing",
	"write_file":    "Writing",
	"edit":          "Editing",
	"edit_file":     "Editing",
	"ls":            "Listing",
	"list_files":    "Listing",
	"grep":          "Searching",
	"find":          "Searching",
	"search":        "Searching",
	"web_search":    "Searching the web",
	"fetch":         "Fetching",
	"web_fetch":     "Fetching",
	"http_get":      "Fetching",
	"search_memory": "Searching memory",
	"add_memory":    "Remembering",
}

// summaryKeys are the arguments that best name what a tool call is about.
var summaryKeys = []string{"command", "cmd", "path", "file_path", "query", "pattern", "url", "name", "title"}

// describeTool picks an action verb and a one-line parameter for a call.
func describeTool(name string, args map[string]any) (verb, parameter string) {
	verb, ok := toolVerbs[strings.ToLower(name)]
	if !ok {
		verb = "Using " + name
	}
	for _, key := range summaryKeys {
		if s, ok := args[key].(string); ok && strings.TrimSpace(s) != "" {
			return verb, strings.Join(strings.Fields(s), " ")
		}
	}
	if len(args) == 0 {
		return verb, ""
	}
	raw, err := json.Marshal(args)
	if err != nil {
		return verb, ""
	}
	return verb, string(raw)
}
