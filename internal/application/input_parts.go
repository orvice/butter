package application

import (
	"strconv"

	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/userinput"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// resolveUserParts resolves a request's user input to the genai parts the
// runner executes. Non-empty `parts` is validated and converted, and
// `message` is ignored; empty `parts` falls back to `message` as a single
// text part so pre-multimodal clients keep working unchanged. The message
// fallback carries the same 1 MiB cap as a text part, so no input path is
// unbounded. Shared by StreamAgent and ReplySession.
func resolveUserParts(inputs []*agentsv1.InputPart, message string) ([]*genai.Part, error) {
	if len(inputs) > 0 {
		return convertInputParts(inputs)
	}
	if len(message) > userinput.MaxTextBytes {
		return nil, connectx.InvalidArgument("message",
			"exceeds maximum allowed size of "+strconv.Itoa(userinput.MaxTextBytes)+" bytes")
	}
	return []*genai.Part{genai.NewPartFromText(message)}, nil
}

// convertInputParts validates a request's multimodal input parts against the
// limits every entry point shares (internal/userinput) and converts them to
// the genai parts the runner executes. Violations are returned as
// connect.CodeInvalidArgument errors. Shared by StreamAgent and ReplySession.
func convertInputParts(inputs []*agentsv1.InputPart) ([]*genai.Part, error) {
	var turn userinput.Turn
	for _, input := range inputs {
		var err error
		switch p := input.GetPart().(type) {
		case *agentsv1.InputPart_Text:
			err = turn.AddText(p.Text)
		case *agentsv1.InputPart_InlineData:
			err = turn.AddImage(p.InlineData.GetMimeType(), p.InlineData.GetData())
		default:
			return nil, connectx.InvalidArgument("parts", "part must set text or inline_data")
		}
		if err != nil {
			return nil, connectx.InvalidArgument("parts", err.Error())
		}
	}
	return turn.Parts(), nil
}
