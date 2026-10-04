package http

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	aguitypes "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/types"
	"github.com/gin-gonic/gin"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/userinput"
)

// aguiMaxRequestBytes caps the body of POST /api/agui/:agent_id. A run reads
// one message, so the cap is one message at the limits every entry point
// shares (userinput.MaxTotalBytes, about 26.7 MiB once AG-UI base64-encodes
// the images), plus room for the JSON around it: tools, state, context and
// the text of the transcript. A client that re-sends earlier messages' images
// with every run outgrows it, but only the trailing message is ever read, so
// that is all a client needs to send.
const aguiMaxRequestBytes = 32 << 20 // 32 MiB

var errNoUserMessage = errors.New("messages must end with a non-empty user message")

// bindAGUIInput decodes the request body into input, reading at most
// aguiMaxRequestBytes. On failure it returns the status to answer: 413 past
// the cap, 400 for a malformed body.
func bindAGUIInput(c *gin.Context, input *aguitypes.RunAgentInput) (int, error) {
	if c.Request.Body != nil {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, aguiMaxRequestBytes)
	}
	if err := c.ShouldBindJSON(input); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return http.StatusRequestEntityTooLarge, fmt.Errorf(
				"request body exceeds %d bytes; send only the trailing message", aguiMaxRequestBytes)
		}
		return http.StatusBadRequest, errors.New("invalid request body")
	}
	return 0, nil
}

// aguiUserParts builds a run's parts from the last user message, the only
// one a run reads. A string content is one text part. AG-UI multimodal
// content becomes one part per content part, in order:
//   - text stays text, and blank text is dropped;
//   - an image with a data source is base64-decoded into inline data, and so
//     is a legacy binary part that carries data.
//
// The server never fetches a URL, so url sources and binary url or id
// references are refused, as are audio, video and document parts. The parts
// are held to the limits every entry point shares (internal/userinput).
func aguiUserParts(messages []aguitypes.Message) ([]*genai.Part, error) {
	contents, err := lastAGUIUserContent(messages)
	if err != nil {
		return nil, err
	}
	var turn userinput.Turn
	for _, content := range contents {
		if err := addAGUIContent(&turn, content); err != nil {
			return nil, err
		}
	}
	if len(turn.Parts()) == 0 {
		return nil, errNoUserMessage
	}
	return turn.Parts(), nil
}

// lastAGUIUserContent returns the content of the last user message as content
// parts; a string content is a single text part.
func lastAGUIUserContent(messages []aguitypes.Message) ([]aguitypes.InputContent, error) {
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role != aguitypes.RoleUser {
			continue
		}
		if text, ok := msg.ContentString(); ok {
			return []aguitypes.InputContent{{Type: aguitypes.InputContentTypeText, Text: text}}, nil
		}
		raw, ok := msg.Content.([]any)
		if !ok {
			return nil, errNoUserMessage
		}
		// Decoded here rather than by Message.ContentInputContents, which
		// does not say why a part is malformed.
		encoded, err := json.Marshal(raw)
		if err != nil {
			return nil, errors.New("invalid user message content")
		}
		var contents []aguitypes.InputContent
		if err := json.Unmarshal(encoded, &contents); err != nil {
			return nil, fmt.Errorf("invalid user message content: %v", err)
		}
		return contents, nil
	}
	return nil, errNoUserMessage
}

// addAGUIContent adds one content part of a user message to the turn.
func addAGUIContent(turn *userinput.Turn, content aguitypes.InputContent) error {
	switch content.Type {
	case aguitypes.InputContentTypeText:
		text := strings.TrimSpace(content.Text)
		if text == "" {
			return nil
		}
		return turn.AddText(text)
	case aguitypes.InputContentTypeImage:
		source := content.Source
		switch {
		case source == nil:
			return errors.New("image content requires a source")
		case source.Type == aguitypes.InputContentSourceTypeURL:
			return errors.New("image url sources are not supported; send the image inline as a data source")
		case source.Type != aguitypes.InputContentSourceTypeData:
			return fmt.Errorf("unsupported image source type %q", source.Type)
		}
		return addAGUIImage(turn, source.MimeType, source.Value)
	case aguitypes.InputContentTypeBinary:
		if content.Data == "" {
			return errors.New("binary content must carry inline data; url and id references are not supported")
		}
		return addAGUIImage(turn, content.MimeType, content.Data)
	case aguitypes.InputContentTypeAudio, aguitypes.InputContentTypeVideo, aguitypes.InputContentTypeDocument:
		return fmt.Errorf("%s content is not supported; send text and images", content.Type)
	default:
		return fmt.Errorf("unsupported content type %q", content.Type)
	}
}

// addAGUIImage decodes an image's base64 payload and adds it to the turn.
func addAGUIImage(turn *userinput.Turn, mimeType, encoded string) error {
	if encoded == "" {
		return errors.New("image data is empty")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return errors.New("image data is not valid base64")
	}
	return turn.AddImage(mimeType, data)
}
