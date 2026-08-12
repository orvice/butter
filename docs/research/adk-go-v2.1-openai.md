# ADK Go v2.1.0 OpenAI support

Research date: 2026-08-12

## Conclusion

ADK Go v2.1.0 adds a native OpenAI `model.LLM`, but Butter should **not replace
its current adapter with v2.1.0 as-is**. The official adapter has an open,
deterministic multi-turn bug: it serializes prior assistant messages as
`input_text`, while the Responses API requires `output_text`, so the second user
turn receives HTTP 400. This is directly relevant to Telegram's persistent
sessions. The bug is documented in [issue #1197], is visible in the v2.1.0
request converter, and remains unfixed in v2.2.0; [PR #1205] and [PR #1291] are
still open as of the research date.

Even after that bug is fixed, the migration is not protocol-neutral: Google's
adapter uses the OpenAI **Responses API**, whereas Butter's current adapter uses
Chat Completions. Every custom `base_url` endpoint must first be verified to
implement `/responses`.

The official package is also explicitly marked **experimental**, and ADK v2.1.0
raises the required Go version to 1.26.5. These constraints should be treated as
release/deployment requirements, not just code changes.

Sources: [v2.1.0 release], [OpenAI PR #1178], [issue #1197], [v2.1.0 request
conversion], [v2.2.0 request conversion], [package documentation], [v2.1.0
go.mod].

## Exact API

The actual import path and constructor are:

```go
import openaimodel "google.golang.org/adk/v2/model/openaimodel"

llm, err := openaimodel.NewModel(ctx, modelName, &openaimodel.ClientConfig{
	APIKey:  apiKey,
	BaseURL: baseURL,
})
```

`NewModel` returns `(model.LLM, error)`. It rejects an empty model name, accepts
a nil config, and does not perform a network request during construction. Its
context argument is currently unused and exists for parity with other model
constructors.

`ClientConfig` contains:

- `APIKey string`
- `BaseURL string`
- `HTTPClient *http.Client`
- `Options []option.RequestOption`, appended after options derived from the
  other fields

When `APIKey` or `BaseURL` is empty, the underlying OpenAI Go SDK uses its
defaults, including `OPENAI_API_KEY` and `OPENAI_BASE_URL`. An explicit base URL
can point at an OpenAI-compatible service.

There is a documentation typo worth avoiding: the v2.1.0 example README says
`google.golang.org/adk/v2/model/openai`, but both the shipped package and the
compilable example use `google.golang.org/adk/v2/model/openaimodel`.

Sources: [OpenAI constructor], [package documentation], [official example],
[official example README].

## Model and endpoint semantics

Any non-empty model string is accepted; no `openai/` prefix is required. The
constructor model name is sent as the Responses API model. If
`model.LLMRequest.Model` is non-empty, it overrides that model for the individual
request, while `LLM.Name()` continues to return the constructor name.

For a custom endpoint, `BaseURL` normally includes `/v1`; Google's Ollama
example uses `http://localhost:11434/v1`. The endpoint must implement the OpenAI
**Responses API**. The official documentation explicitly says endpoints that
only implement the older Chat Completions API do not work; it names recent
Ollama, LM Studio, and vLLM releases as potential compatible endpoints.

Sources: [request conversion], [official example README].

## Butter replacement

Butter's current OpenAI construction:

```go
return adkopenai.New(adkopenai.Config{
	APIKey:    p.GetApiKey(),
	BaseURL:   p.GetBaseUrl(),
	ModelName: modelName,
}), nil
```

becomes:

```go
return openaimodel.NewModel(ctx, modelName, &openaimodel.ClientConfig{
	APIKey:  p.GetApiKey(),
	BaseURL: p.GetBaseUrl(),
})
```

ADK v2.1.0 also adds `model.Register` / `model.NewLLM`, but registration is
opt-in and provider packages do not self-register. Butter already resolves
workspace provider aliases and credentials before model construction, so direct
construction remains the appropriate integration point.

Replacing the model adapter will not remove the entire
`github.com/achetronic/adk-utils-go` dependency: Butter still uses its Langfuse
and context-guard plugins.

Sources: [OpenAI constructor], [model registry], Butter
`internal/agent/model.go`, `internal/app/runtime.go`, and
`internal/runtime/runner/runner.go`.

## Behavior differences and risks

### Blocking multi-turn defect

The v2.1.0 request converter sends every text-bearing role through `newMessage`.
That helper always creates `ResponseInputTextParam` with type `input_text`, even
when the ADK role is `model` and is normalized to the OpenAI `assistant` role.
On turn two, session history contains the previous assistant message, and the
OpenAI Responses API rejects it because assistant content must be
`output_text`.

This is not an intermittent provider behavior: it follows deterministically
from the shipped request construction. Google's open issue reports failure from
the second message for both reasoning and non-reasoning OpenAI models. The same
code is still present in v2.2.0. Two upstream fixes exist but neither has been
merged.

For Butter, this blocks direct adoption in Telegram and every other multi-turn
channel. Viable choices are to retain the current adapter until a fixed release,
or carry a reviewed upstream patch/fork plus a two-turn wire-level regression
test. Retaining the current adapter is the lower-risk choice for this change.

Sources: [issue #1197], [v2.1.0 request conversion], [v2.2.0 request
conversion], [PR #1205], [PR #1291].

### API compatibility

The current adapter calls Chat Completions. The Google adapter calls
`client.Responses.New` / `NewStreaming`. A configured provider can therefore
work today and fail after migration even when it advertises itself as
"OpenAI-compatible". Provider acceptance should include a real Responses API
text, streaming, and tool-call request before switching it.

Sources: [OpenAI constructor], [official example README], [current adapter
source].

### Input modalities

The Google request converter accepts text, `FunctionCall`, and
`FunctionResponse` parts. Other content parts return an `unsupported content
part` error. In particular, v2.1.0 does not convert inline image, audio, or file
data. Butter's current adapter does convert those media types, so routing an
attachment to an OpenAI-backed agent would regress unless Butter keeps a
compatibility path or blocks unsupported inputs clearly.

Source: [request conversion]; comparison: [current adapter source].

### Tools and generation settings

The native adapter supports ADK function tools and tool-choice modes. It rejects
non-function Gemini tools such as Google Search, retrieval, Maps, computer use,
and code execution.

It supports temperature, top-p, max output tokens, system instructions,
log-probabilities, JSON mode, and strict JSON Schema output. It returns explicit
errors for top-k, stop sequences, more than one candidate, frequency/presence
penalties, request labels, safety settings, and unsupported MIME types.
`ThinkingConfig` is not mapped in v2.1.0, so Butter should not assume it controls
OpenAI reasoning effort through this adapter.

Sources: [request conversion], [tool conversion], [official example README].

### Empty responses and diagnostics

For non-streaming responses, the native adapter returns explicit errors when the
response is nil, contains no output items, or contains neither text nor a tool
call. Refusals are converted to visible text. Reasoning text and summaries are
converted to `genai.Part` values with `Thought: true`.

Finish reasons map as follows:

| Responses API state | ADK finish reason |
| --- | --- |
| `max_output_tokens` | `MAX_TOKENS` |
| `content_filter` | `SAFETY` |
| no incomplete reason | `STOP` |
| another incomplete reason | `OTHER` |

Content filtering also produces prompt feedback. Responses carry token usage,
and the adapter attaches `openai_response_id` and `openai_model` to
`LLMResponse.CustomMetadata`. This is better diagnostic material for Butter's
empty-response logging, though the runner still needs to preserve and report it.

Streaming translates text deltas, reasoning text/summary deltas, and buffered
function-call arguments. Unknown stream event types are ignored.

Sources: [response conversion], [stream conversion], [OpenAI errors], [OpenAI
constructor].

## v2.0.0 to v2.1.0 upgrade notes

- OpenAI support is new in v2.1.0 and is marked experimental.
- The ADK module's `go` directive changes from `1.25.0` to `1.26.5`. Butter
  now declares Go `1.26.5`; CI and the builder image already use Go 1.26, so
  production toolchains must meet the same minimum rather than relying on
  automatic toolchain download.
- `google.golang.org/genai` moves from 1.57.0 to 1.63.0, and ADK adds a direct
  dependency on `github.com/openai/openai-go/v3` 3.8.1.
- The release also adds the name-based model registry, `TaskRunner`, public
  `PackTool`, `runner.NewInMemory`, MCP per-request auth, and workflow/parallel
  fixes. There is no dedicated OpenAI migration guide; the release, PR, example,
  and source are the authoritative upgrade material.

Sources: [v2.1.0 release], [v2.0.0...v2.1.0 comparison], [v2.0.0 go.mod],
[v2.1.0 go.mod].

## Recommended migration gate

1. Keep the current OpenAI adapter; Butter's Telegram empty-response fixes are
   implemented independently of the provider migration.
2. Wait for an ADK release containing the fix for issue #1197, or deliberately
   carry one of the upstream patches with a local multi-turn regression test.
3. Upgrade the repository and build environments to Go 1.26.5 before adopting
   ADK v2.1.0 or later.
4. Switch Butter's OpenAI constructor to `openaimodel.NewModel`.
5. Run contract tests against api.openai.com and every configured custom base
   URL for non-streaming text, streaming text, and function calling.
6. Include a mandatory two-turn session test that inspects the real API outcome
   or request wire format, so assistant history cannot regress to `input_text`.
7. Decide explicitly how OpenAI-backed agents handle inline media before making
   the native adapter the only path.
8. Keep Butter's runner-side empty-output fallback and diagnostics work; the
   native adapter improves failure signaling but does not render workflow
   `Event.Output` or serialize Telegram turns for Butter.

[v2.1.0 release]: https://github.com/google/adk-go/releases/tag/v2.1.0
[OpenAI PR #1178]: https://github.com/google/adk-go/pull/1178
[v2.0.0...v2.1.0 comparison]: https://github.com/google/adk-go/compare/v2.0.0...v2.1.0
[package documentation]: https://github.com/google/adk-go/blob/v2.1.0/model/openaimodel/doc.go
[OpenAI constructor]: https://github.com/google/adk-go/blob/v2.1.0/model/openaimodel/openai.go
[request conversion]: https://github.com/google/adk-go/blob/v2.1.0/model/openaimodel/request.go
[v2.1.0 request conversion]: https://github.com/google/adk-go/blob/v2.1.0/model/openaimodel/request.go
[v2.2.0 request conversion]: https://github.com/google/adk-go/blob/v2.2.0/model/openaimodel/request.go
[response conversion]: https://github.com/google/adk-go/blob/v2.1.0/model/openaimodel/response.go
[stream conversion]: https://github.com/google/adk-go/blob/v2.1.0/model/openaimodel/stream.go
[tool conversion]: https://github.com/google/adk-go/blob/v2.1.0/model/openaimodel/tools.go
[OpenAI errors]: https://github.com/google/adk-go/blob/v2.1.0/model/openaimodel/errors.go
[model registry]: https://github.com/google/adk-go/blob/v2.1.0/model/registry.go
[official example]: https://github.com/google/adk-go/blob/v2.1.0/examples/openai/main.go
[official example README]: https://github.com/google/adk-go/blob/v2.1.0/examples/openai/README.md
[v2.0.0 go.mod]: https://github.com/google/adk-go/blob/v2.0.0/go.mod
[v2.1.0 go.mod]: https://github.com/google/adk-go/blob/v2.1.0/go.mod
[issue #1197]: https://github.com/google/adk-go/issues/1197
[PR #1205]: https://github.com/google/adk-go/pull/1205
[PR #1291]: https://github.com/google/adk-go/pull/1291
[current adapter source]: https://github.com/achetronic/adk-utils-go/blob/v0.22.0/genai/openai/openai.go
