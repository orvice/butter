package http

import (
	"context"
	"encoding/json"
	"io"

	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	aguitypes "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/types"
	aguisse "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/encoding/sse"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"

	"go.orx.me/apps/butter/internal/runtime/streamorch"
)

// aguiInterruptReason is the AG-UI Interrupt.Reason for a Workflow Agent
// pausing on a Human Input node. AG-UI leaves the vocabulary to the server;
// "human_input" names what the pause actually is.
const aguiInterruptReason = "human_input"

// aguiEmitter delivers one encoded AG-UI event. Tests substitute a recorder so
// event ordering can be asserted without going through SSE.
type aguiEmitter func(aguievents.Event) error

// newAGUISSEEmitter returns an aguiEmitter writing events to w as SSE frames,
// flushing each one so the client sees it immediately.
//
// ctx must be non-nil: the SDK's JSON encoder calls ctx.Err() unguarded and
// panics on a nil context.
func newAGUISSEEmitter(ctx context.Context, w io.Writer, flush func()) aguiEmitter {
	writer := aguisse.NewSSEWriter()
	return func(ev aguievents.Event) error {
		if err := writer.WriteEvent(ctx, w, ev); err != nil {
			return err
		}
		if flush != nil {
			flush()
		}
		return nil
	}
}

// aguiSink implements streamorch.Sink by translating a run's frames into AG-UI
// protocol events.
//
// It lives beside the handler rather than in streamorch because streamorch is
// the protocol-neutral orchestration seam; every Sink implementation belongs
// with its own transport (compare streamAgentSink in internal/application and
// asyncrun's hubSink).
//
// Not concurrency-safe, which is what streamorch.Run guarantees: it calls the
// Sink serially from inside the ADK event loop.
type aguiSink struct {
	threadID  string
	runID     string
	messageID string
	emit      aguiEmitter

	// msgOpen tracks whether a TEXT_MESSAGE_START has been emitted without a
	// matching END, so text can stream lazily and still be closed exactly once.
	msgOpen bool
	// interrupts are the pauses observed in-stream via
	// session.Event.RequestedInput. Collecting them here is what lets Final
	// pick the run's outcome without reaching for runner.TurnResult — see
	// docs/research/ag-ui-integration.md.
	interrupts []aguitypes.Interrupt
}

func newAGUISink(threadID, runID, messageID string, emit aguiEmitter) *aguiSink {
	return &aguiSink{threadID: threadID, runID: runID, messageID: messageID, emit: emit}
}

func (s *aguiSink) Started(streamorch.RunIdentity) error {
	return s.emit(aguievents.NewRunStartedEvent(s.threadID, s.runID))
}

func (s *aguiSink) TextDelta(_ streamorch.RunIdentity, text string) error {
	if err := s.openMessage(); err != nil {
		return err
	}
	return s.emit(aguievents.NewTextMessageContentEvent(s.messageID, text))
}

func (s *aguiSink) RunEvent(_ streamorch.RunIdentity, evt *session.Event) error {
	if evt == nil {
		return nil
	}

	// A Workflow Agent pausing for human input is signalled in-stream: the
	// runner forwards request-input events to the callback even though they
	// count as final responses (internal/runtime/runner/runner.go).
	if req := evt.RequestedInput; req != nil {
		s.interrupts = append(s.interrupts, aguitypes.Interrupt{
			ID:      req.InterruptID,
			Reason:  aguiInterruptReason,
			Message: req.Message,
		})
	}

	if evt.Content == nil {
		return nil
	}
	for _, part := range evt.Content.Parts {
		if part == nil {
			continue
		}
		// The request-input FunctionCall/Response pair *is* the interrupt
		// handshake, already reported above and below as an Interrupt. Emitting
		// TOOL_CALL_* for it too would make clients render a tool the user
		// never called.
		if fc := part.FunctionCall; fc != nil && fc.Name != workflow.WorkflowInputFunctionCallName {
			if err := s.emitToolCall(fc.ID, fc.Name, fc.Args); err != nil {
				return err
			}
		}
		if fr := part.FunctionResponse; fr != nil && fr.Name != workflow.WorkflowInputFunctionCallName {
			out, err := json.Marshal(fr.Response)
			if err != nil {
				out = []byte("{}")
			}
			if err := s.emit(aguievents.NewToolCallResultEvent(s.messageID, fr.ID, string(out))); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *aguiSink) Final(_ streamorch.RunIdentity, response string) error {
	// streamorch streams TextDelta only for *partial* events, so a
	// non-streaming turn carries its whole answer in response. Emit it so the
	// client is not left with an empty message.
	//
	// The exception is a paused workflow: runner.run appends each pending
	// question to the turn output, so response is the question that is already
	// travelling as Interrupt.Message. Emitting it again would show the client
	// the same prompt twice.
	if !s.msgOpen && response != "" && len(s.interrupts) == 0 {
		if err := s.openMessage(); err != nil {
			return err
		}
		if err := s.emit(aguievents.NewTextMessageContentEvent(s.messageID, response)); err != nil {
			return err
		}
	}
	if err := s.closeMessage(); err != nil {
		return err
	}
	if len(s.interrupts) > 0 {
		return s.emit(aguievents.NewRunFinishedEventWithOptions(
			s.threadID, s.runID, aguievents.WithInterruptOutcome(s.interrupts)))
	}
	return s.emit(aguievents.NewRunFinishedEventWithOptions(
		s.threadID, s.runID, aguievents.WithSuccessOutcome()))
}

// Error emits RUN_ERROR. streamorch.Sink has no error frame — streamorch.Run
// returns the run error to its caller — so the handler calls this after a
// failed run. Any open message is closed first so the client is not left
// waiting on a TEXT_MESSAGE_END that never arrives.
func (s *aguiSink) Error(runErr error) error {
	if err := s.closeMessage(); err != nil {
		return err
	}
	return s.emit(aguievents.NewRunErrorEvent(runErr.Error(), aguievents.WithRunID(s.runID)))
}

func (s *aguiSink) openMessage() error {
	if s.msgOpen {
		return nil
	}
	if err := s.emit(aguievents.NewTextMessageStartEvent(
		s.messageID, aguievents.WithRole("assistant"))); err != nil {
		return err
	}
	s.msgOpen = true
	return nil
}

func (s *aguiSink) closeMessage() error {
	if !s.msgOpen {
		return nil
	}
	if err := s.emit(aguievents.NewTextMessageEndEvent(s.messageID)); err != nil {
		return err
	}
	s.msgOpen = false
	return nil
}

// emitToolCall renders one ADK FunctionCall as the AG-UI three-event sequence.
// ADK hands over complete arguments rather than a token stream, so the whole
// JSON object goes out as a single TOOL_CALL_ARGS delta.
func (s *aguiSink) emitToolCall(id, name string, args map[string]any) error {
	if err := s.emit(aguievents.NewToolCallStartEvent(id, name)); err != nil {
		return err
	}
	encoded, err := json.Marshal(args)
	if err != nil {
		encoded = []byte("{}")
	}
	if err := s.emit(aguievents.NewToolCallArgsEvent(id, string(encoded))); err != nil {
		return err
	}
	return s.emit(aguievents.NewToolCallEndEvent(id))
}

var _ streamorch.Sink = (*aguiSink)(nil)
