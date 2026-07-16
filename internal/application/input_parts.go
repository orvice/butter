package application

import (
	"strconv"

	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// maxInlineImageBytes caps a single inline image; the limit stays under
// typical model-provider inline-data ceilings while protecting server
// memory under concurrent load.
const maxInlineImageBytes = 10 << 20 // 10 MiB

// maxInputPartsBytes caps the combined payload (text plus inline data) of
// all parts in a single request.
const maxInputPartsBytes = 20 << 20 // 20 MiB

// maxInlineImageCount caps how many inline images a single request may
// carry.
const maxInlineImageCount = 10

// allowedInlineMimeTypes is the whitelist of inline image formats accepted
// by multimodal input parts.
var allowedInlineMimeTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/gif":  {},
	"image/webp": {},
}

// convertInputParts validates a request's multimodal input parts and
// converts them to the genai parts the runner executes. Violations are
// returned as connect.CodeInvalidArgument errors. Shared by StreamAgent
// and (in a follow-up) ReplySession.
func convertInputParts(inputs []*agentsv1.InputPart) ([]*genai.Part, error) {
	out := make([]*genai.Part, 0, len(inputs))
	totalBytes := 0
	imageCount := 0
	for _, input := range inputs {
		switch p := input.GetPart().(type) {
		case *agentsv1.InputPart_Text:
			// Same cap as the legacy message field so `parts` cannot
			// bypass the text input limit.
			if len(p.Text) > maxInvokeAgentInputBytes {
				return nil, connectx.InvalidArgument("parts",
					"text part exceeds maximum allowed size of "+strconv.Itoa(maxInvokeAgentInputBytes)+" bytes")
			}
			totalBytes += len(p.Text)
			out = append(out, genai.NewPartFromText(p.Text))
		case *agentsv1.InputPart_InlineData:
			imageCount++
			if imageCount > maxInlineImageCount {
				return nil, connectx.InvalidArgument("parts",
					"too many images; maximum is "+strconv.Itoa(maxInlineImageCount)+" per request")
			}
			totalBytes += len(p.InlineData.GetData())
			mimeType := p.InlineData.GetMimeType()
			if _, ok := allowedInlineMimeTypes[mimeType]; !ok {
				return nil, connectx.InvalidArgument("parts",
					"unsupported mime_type "+strconv.Quote(mimeType)+"; accepted: image/jpeg, image/png, image/gif, image/webp")
			}
			if len(p.InlineData.GetData()) > maxInlineImageBytes {
				return nil, connectx.InvalidArgument("parts",
					"image exceeds maximum allowed size of "+strconv.Itoa(maxInlineImageBytes)+" bytes")
			}
			out = append(out, genai.NewPartFromBytes(p.InlineData.GetData(), mimeType))
		default:
			return nil, connectx.InvalidArgument("parts", "part must set text or inline_data")
		}
		if totalBytes > maxInputPartsBytes {
			return nil, connectx.InvalidArgument("parts",
				"total parts payload exceeds maximum allowed size of "+strconv.Itoa(maxInputPartsBytes)+" bytes")
		}
	}
	return out, nil
}
