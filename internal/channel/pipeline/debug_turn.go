package pipeline

import (
	"context"
	"sync"
	"time"

	"google.golang.org/adk/v2/session"
)

const debugEditInterval = time.Second

// debugTurn aggregates debug activity for one runner turn and serializes
// throttled processing-message edits through a single worker.
type debugTurn struct {
	mu      sync.Mutex
	summary DebugSummary
	notify  chan struct{}
	stop    chan chan DebugSummary
	closed  bool
}

func newDebugTurn(
	ctx context.Context,
	interval time.Duration,
	edit func(DebugSummary),
) *debugTurn {
	d := &debugTurn{
		summary: DebugSummary{ToolCounts: make(map[string]int)},
		notify:  make(chan struct{}, 1),
		stop:    make(chan chan DebugSummary),
	}
	go d.run(ctx, interval, edit)
	return d
}

func (d *debugTurn) observeEvent(evt *session.Event) {
	if evt == nil {
		return
	}

	relevant := false
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return
	}
	if evt.Actions.TransferToAgent != "" {
		d.summary.Transfers++
		relevant = true
	}
	if evt.Content != nil {
		for _, part := range evt.Content.Parts {
			if part.FunctionCall == nil {
				continue
			}
			d.summary.ToolCalls++
			d.summary.ToolCounts[part.FunctionCall.Name]++
			relevant = true
		}
	}
	if relevant {
		d.summary.LatestEvent = evt
		d.summary.LatestCompaction = ""
	}
	d.mu.Unlock()

	if relevant {
		d.signal()
	}
}

func (d *debugTurn) observeCompaction(agentName string) {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return
	}
	d.summary.Compactions++
	d.summary.LatestEvent = nil
	d.summary.LatestCompaction = agentName
	d.mu.Unlock()
	d.signal()
}

func (d *debugTurn) finish() DebugSummary {
	d.mu.Lock()
	d.closed = true
	d.mu.Unlock()

	result := make(chan DebugSummary)
	d.stop <- result
	return <-result
}

func (d *debugTurn) signal() {
	select {
	case d.notify <- struct{}{}:
	default:
	}
}

func (d *debugTurn) snapshot() DebugSummary {
	d.mu.Lock()
	defer d.mu.Unlock()
	return cloneDebugSummary(d.summary)
}

func (d *debugTurn) run(ctx context.Context, interval time.Duration, edit func(DebugSummary)) {
	var (
		lastEdit time.Time
		timer    *time.Timer
		timerC   <-chan time.Time
	)

	stopTimer := func() {
		if timer == nil {
			return
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timerC = nil
	}

	flush := func() {
		stopTimer()
		edit(d.snapshot())
		lastEdit = time.Now()
	}

	for {
		select {
		case <-d.notify:
			if lastEdit.IsZero() || time.Since(lastEdit) >= interval {
				flush()
				continue
			}
			remaining := time.Until(lastEdit.Add(interval))
			if timer == nil {
				timer = time.NewTimer(remaining)
			} else {
				stopTimer()
				timer.Reset(remaining)
			}
			timerC = timer.C
		case <-timerC:
			timerC = nil
			flush()
		case result := <-d.stop:
			stopTimer()
			result <- d.snapshot()
			return
		case <-ctx.Done():
			// The runner may finish with a canceled request context. Keep waiting
			// for finish so the final edit can safely follow any edit in flight.
			ctx = context.WithoutCancel(ctx)
		}
	}
}

func cloneDebugSummary(in DebugSummary) DebugSummary {
	out := in
	out.ToolCounts = make(map[string]int, len(in.ToolCounts))
	for name, count := range in.ToolCounts {
		out.ToolCounts[name] = count
	}
	return out
}
