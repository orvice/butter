package pipeline

import (
	"strings"

	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/runner"
)

const emptyTurnResponse = "The agent did not produce a text response. Please try again."

// TurnResponseText returns user-facing text for a completed turn. Empty model
// results are translated into a recovery-oriented message instead of exposing
// an implementation placeholder such as "(no response)".
func TurnResponseText(turn *runner.TurnResult) string {
	if TurnHasVisibleText(turn) {
		return turn.Output
	}
	if turn == nil {
		return emptyTurnResponse
	}

	reason := turn.FinishReason
	if reason == "" || reason == genai.FinishReasonUnspecified {
		reason = genai.FinishReason(strings.ToUpper(turn.ErrorCode))
	}
	switch reason {
	case genai.FinishReasonMaxTokens:
		return "The agent reached its response limit before producing a reply. Please try a shorter request."
	case genai.FinishReasonSafety,
		genai.FinishReasonRecitation,
		genai.FinishReasonBlocklist,
		genai.FinishReasonProhibitedContent,
		genai.FinishReasonSPII,
		genai.FinishReasonImageSafety:
		return "The agent could not provide a response because it was blocked by the model's safety checks. Please revise your message and try again."
	default:
		return emptyTurnResponse
	}
}

func TurnHasVisibleText(turn *runner.TurnResult) bool {
	return turn != nil && strings.TrimSpace(turn.Output) != ""
}

func turnEventCount(turn *runner.TurnResult) int {
	if turn == nil {
		return 0
	}
	return turn.EventCount
}

func turnFinishReason(turn *runner.TurnResult) genai.FinishReason {
	if turn == nil {
		return ""
	}
	return turn.FinishReason
}

func turnErrorCode(turn *runner.TurnResult) string {
	if turn == nil {
		return ""
	}
	return turn.ErrorCode
}
