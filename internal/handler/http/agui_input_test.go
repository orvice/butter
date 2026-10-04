package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	"go.orx.me/apps/butter/internal/userinput"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// testPNG stands in for an image: the server checks the declared type and the
// size, never the bytes.
var testPNG = []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 1, 2, 3}

func textContent(text string) map[string]any {
	return map[string]any{"type": "text", "text": text}
}

// imageContent is an image as @assistant-ui/react-ag-ui sends an attachment:
// inline, as a base64 data source.
func imageContent(mimeType string, data []byte) map[string]any {
	return map[string]any{"type": "image", "source": map[string]any{
		"type": "data", "value": base64.StdEncoding.EncodeToString(data), "mimeType": mimeType,
	}}
}

// userContentBody is a run whose trailing user message carries content parts.
func userContentBody(threadID string, content ...map[string]any) map[string]any {
	return withAGUIField(minimalAGUIBody(threadID, ""), "messages", []map[string]any{
		{"id": "m1", "role": "user", "content": content},
	})
}

// Text plus a PNG reaches the runner as two parts, in order: the text, then
// the image decoded into inline data.
func TestAGUIRun_TextAndImageReachRunner(t *testing.T) {
	mock := &mockRunner{runResult: "a cat"}
	router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

	w := postAGUI(t, router, "writer", userContentBody("t-1",
		textContent("what is in this picture?"), imageContent("image/png", testPNG)))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"type":"RUN_FINISHED"`) {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if len(mock.lastParts) != 2 {
		t.Fatalf("parts = %+v, want the text and the image", mock.lastParts)
	}
	if got := mock.lastParts[0].Text; got != "what is in this picture?" {
		t.Errorf("first part text = %q", got)
	}
	img := mock.lastParts[1].InlineData
	if img == nil || img.MIMEType != "image/png" || !bytes.Equal(img.Data, testPNG) {
		t.Fatalf("second part = %+v, want the decoded PNG", mock.lastParts[1])
	}
}

// A message with only an image runs; the image is the whole turn.
func TestAGUIRun_ImageOnlyMessageRuns(t *testing.T) {
	mock := &mockRunner{runResult: "a screenshot"}
	router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

	jpeg := []byte{0xff, 0xd8, 0xff, 0xe0}
	w := postAGUI(t, router, "writer", userContentBody("t-1", imageContent("image/jpeg", jpeg)))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"delta":"a screenshot"`) {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if len(mock.lastParts) != 1 || mock.lastParts[0].InlineData == nil ||
		!bytes.Equal(mock.lastParts[0].InlineData.Data, jpeg) {
		t.Fatalf("parts = %+v, want the one image", mock.lastParts)
	}
}

// The legacy binary part carrying inline data is an image like any other.
func TestAGUIRun_LegacyBinaryImageIsDecoded(t *testing.T) {
	mock := &mockRunner{runResult: "ok"}
	router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

	w := postAGUI(t, router, "writer", userContentBody("t-1", map[string]any{
		"type": "binary", "mimeType": "image/webp", "filename": "photo.webp",
		"data": base64.StdEncoding.EncodeToString(testPNG),
	}))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if len(mock.lastParts) != 1 || mock.lastParts[0].InlineData == nil ||
		mock.lastParts[0].InlineData.MIMEType != "image/webp" || !bytes.Equal(mock.lastParts[0].InlineData.Data, testPNG) {
		t.Fatalf("parts = %+v, want the decoded image", mock.lastParts)
	}
}

// Content the endpoint does not take is refused with 400 before the stream
// opens, and nothing runs: a URL the server would have to fetch, media other
// than images, and anything past the limits the RPCs share.
func TestAGUIRun_UnsupportedContentIsRejected(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString(testPNG)
	elevenImages := make([]map[string]any, 0, userinput.MaxImages+1)
	for range userinput.MaxImages + 1 {
		elevenImages = append(elevenImages, imageContent("image/png", testPNG))
	}
	sevenMiB := make([]byte, 7<<20)

	cases := []struct {
		name      string
		content   []map[string]any
		wantError string
	}{
		{
			name: "url image source",
			content: []map[string]any{{"type": "image", "source": map[string]any{
				"type": "url", "value": "https://example.com/cat.png", "mimeType": "image/png",
			}}},
			wantError: "image url sources are not supported",
		},
		{
			name: "audio part",
			content: []map[string]any{{"type": "audio", "source": map[string]any{
				"type": "data", "value": encoded, "mimeType": "audio/wav",
			}}},
			wantError: "audio content is not supported",
		},
		{
			name: "video part",
			content: []map[string]any{{"type": "video", "source": map[string]any{
				"type": "data", "value": encoded, "mimeType": "video/mp4",
			}}},
			wantError: "video content is not supported",
		},
		{
			name: "document part",
			content: []map[string]any{{"type": "document", "source": map[string]any{
				"type": "data", "value": encoded, "mimeType": "application/pdf",
			}}},
			wantError: "document content is not supported",
		},
		{name: "an eleventh image", content: elevenImages, wantError: "too many images"},
		{
			name:      "oversized image",
			content:   []map[string]any{imageContent("image/png", make([]byte, userinput.MaxImageBytes+1))},
			wantError: "image exceeds maximum allowed size",
		},
		{
			name: "message over the total",
			content: []map[string]any{
				imageContent("image/png", sevenMiB), imageContent("image/png", sevenMiB), imageContent("image/png", sevenMiB),
			},
			wantError: "total parts payload exceeds",
		},
		{
			name:      "unsupported image type",
			content:   []map[string]any{imageContent("image/svg+xml", []byte("<svg/>"))},
			wantError: `unsupported mime_type "image/svg+xml"`,
		},
		{
			name:      "binary part that is not an image",
			content:   []map[string]any{{"type": "binary", "mimeType": "application/pdf", "data": encoded}},
			wantError: `unsupported mime_type "application/pdf"`,
		},
		{
			name:      "binary url reference",
			content:   []map[string]any{{"type": "binary", "mimeType": "image/png", "url": "https://example.com/cat.png"}},
			wantError: "binary content must carry inline data",
		},
		{
			name:      "malformed binary part",
			content:   []map[string]any{{"type": "binary", "data": encoded}},
			wantError: "invalid user message content",
		},
		{
			name:      "image without a source",
			content:   []map[string]any{{"type": "image"}},
			wantError: "image content requires a source",
		},
		{
			name: "image data that is not base64",
			content: []map[string]any{{"type": "image", "source": map[string]any{
				"type": "data", "value": "data:image/png;base64," + encoded, "mimeType": "image/png",
			}}},
			wantError: "image data is not valid base64",
		},
		{
			name:      "unknown part type",
			content:   []map[string]any{{"type": "sticker"}},
			wantError: `unsupported content type "sticker"`,
		},
		{
			name:      "text part over the cap",
			content:   []map[string]any{textContent(strings.Repeat("a", userinput.MaxTextBytes+1))},
			wantError: "text part exceeds maximum allowed size",
		},
		{
			name:      "nothing but blank text",
			content:   []map[string]any{textContent("  ")},
			wantError: "non-empty user message",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockRunner{runResult: "should not run"}
			router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

			w := postAGUI(t, router, "writer", userContentBody("t-1", tc.content...))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (body %.300s)", w.Code, w.Body.String())
			}
			var resp aguiErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("body %.300s is not an error response: %v", w.Body.String(), err)
			}
			if !strings.Contains(resp.Error, tc.wantError) {
				t.Errorf("error %q, want it to mention %q", resp.Error, tc.wantError)
			}
			if mock.lastAgentName != "" {
				t.Error("the runner was invoked for a rejected message")
			}
		})
	}
}

// repeatByte is an endless reader of one byte, so a test can send a body
// larger than the cap without holding it in memory.
type repeatByte byte

func (r repeatByte) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(r)
	}
	return len(p), nil
}

// overCapBody is a run request whose transcript re-sends an earlier message
// that alone outgrows the body cap.
func overCapBody() io.Reader {
	return io.MultiReader(
		strings.NewReader(`{"threadId":"t-1","messages":[{"id":"m1","role":"user","content":"`),
		io.LimitReader(repeatByte('a'), aguiMaxRequestBytes),
		strings.NewReader(`"},{"id":"m2","role":"user","content":"hi"}]}`),
	)
}

func postAGUIBody(router *gin.Engine, agentID string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/agui/"+agentID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// A body past the cap is refused with 413 before the stream opens, and
// nothing runs.
func TestAGUIRun_OversizedBodyIsRejected(t *testing.T) {
	mock := &mockRunner{runResult: "should not run"}
	router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

	w := postAGUIBody(router, "writer", overCapBody())
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413 (body %.300s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "send only the trailing message") {
		t.Errorf("error body %s, want it to say what to send", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "RUN_STARTED") || mock.lastAgentName != "" {
		t.Fatalf("an oversized body ran: %.300s", w.Body.String())
	}
}

// One message at the limits fits under the body cap with its images
// base64-encoded: the cap stops only a client that sends more than the
// trailing message.
func TestAGUIRun_MessageAtTheLimitsFitsTheBodyCap(t *testing.T) {
	mock := &mockRunner{runResult: "ok"}
	router := setupAGUIRouter(aguiEnabledRepo(), mock, true)

	image := make([]byte, userinput.MaxTotalBytes/2)
	w := postAGUI(t, router, "writer", userContentBody("t-1",
		imageContent("image/png", image), imageContent("image/png", image)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %.300s", w.Code, w.Body.String())
	}
	if len(mock.lastParts) != 2 {
		t.Fatalf("parts = %d, want both images", len(mock.lastParts))
	}
}

// modelImageURLs returns the image URLs of the last user message the model
// received.
func modelImageURLs(t *testing.T, req openaifake.ChatCompletionRequest) []string {
	t.Helper()
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role != "user" {
			continue
		}
		var parts []struct {
			Type     string `json:"type"`
			ImageURL struct {
				URL string `json:"url"`
			} `json:"image_url"`
		}
		if err := json.Unmarshal(req.Messages[i].Content, &parts); err != nil {
			t.Fatalf("user content %s is not a list of parts: %v", req.Messages[i].Content, err)
		}
		var urls []string
		for _, p := range parts {
			if p.Type == "image_url" {
				urls = append(urls, p.ImageURL.URL)
			}
		}
		return urls
	}
	t.Fatal("no user message reached the model")
	return nil
}

// Through a real runner and session store: the image reaches the model as an
// image, and the thread's history gives it back with its text. A refused
// message on the same thread runs nothing and appends nothing.
func TestAGUIRun_ImageReachesTheModelAndTheHistory(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	h.backend.ScriptRequest("card-model", func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
		openaifake.WriteReply(w, req, "A cat.")
	})

	if w := h.post("carder", userContentBody("t-img",
		textContent("What is this?"), imageContent("image/png", testPNG))); w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	req, ok := h.backend.LastRequest("card-model")
	if !ok {
		t.Fatal("the model was never called")
	}
	encoded := base64.StdEncoding.EncodeToString(testPNG)
	if urls := modelImageURLs(t, req); len(urls) != 1 || urls[0] != "data:image/png;base64,"+encoded {
		t.Fatalf("image URLs the model received = %v", urls)
	}

	code, history := h.history("carder", "t-img")
	if code != http.StatusOK || len(history.Messages) != 2 {
		t.Fatalf("history status = %d, messages = %+v", code, history.Messages)
	}
	requireJSON(t, "user message", history.Messages[0].Content, []map[string]any{
		{"type": "text", "text": "What is this?"},
		{"type": "image", "source": map[string]any{"type": "data", "value": encoded, "mimeType": "image/png"}},
	})
	if history.Messages[1].Content != "A cat." {
		t.Fatalf("assistant message = %+v", history.Messages[1])
	}

	events := func() int {
		resp, err := h.sessions.Get(t.Context(), &adksession.GetRequest{AppName: aguiAppName, UserID: "u1", SessionID: "agui-t-img"})
		if err != nil {
			t.Fatal(err)
		}
		return resp.Session.Events().Len()
	}
	elevenImages := make([]map[string]any, 0, userinput.MaxImages+1)
	for range userinput.MaxImages + 1 {
		elevenImages = append(elevenImages, imageContent("image/png", testPNG))
	}
	before, calls := events(), h.backend.CallCount("card-model")
	for name, content := range map[string][]map[string]any{
		"url image":      {{"type": "image", "source": map[string]any{"type": "url", "value": "https://example.com/cat.png"}}},
		"audio":          {{"type": "audio", "source": map[string]any{"type": "data", "value": encoded, "mimeType": "audio/wav"}}},
		"eleventh image": elevenImages,
		"oversized":      {imageContent("image/png", make([]byte, userinput.MaxImageBytes+1))},
	} {
		if w := h.post("carder", userContentBody("t-img", content...)); w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, w.Code)
		}
	}
	if w := postAGUIBody(h.router, "carder", overCapBody()); w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized body: status = %d, want 413", w.Code)
	}
	if got := events(); got != before {
		t.Errorf("refused messages appended %d events", got-before)
	}
	if got := h.backend.CallCount("card-model"); got != calls {
		t.Errorf("refused messages called the model %d times", got-calls)
	}
}
