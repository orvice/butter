## Scope

This note records the smallest Plane Activity client contract that is supported by the repository's research and SDK evidence. It intentionally does not choose an endpoint or request shape beyond the documented contract.

## Confirmed provider contract

The research record documents the Cloud endpoint as `POST /api/v1/workspaces/{slug}/runs/{run_id}/activities/`. Authentication is a Bearer bot token minted with `client_credentials`; the token source must be installation scoped and must not expose credentials to management responses.

The create request has `content`, optional `content_metadata`, optional `signal`, optional `signal_metadata`, `type`, and optional `project`. Valid agent-created types are `thought`, `action`, `response`, `elicitation`, and `error`; `prompt` is provider-created. `response` is the user-visible reply. `elicitation` is a user-visible waiting request. `thought` and `action` are progress activities. `error` has contradictory provider documentation and must never be the only user-visible failure explanation: send a safe `response` first, then record the error activity when the provider supports it.

The provider documents no idempotency key for activity creation. A retry can duplicate a comment. Butter therefore needs a durable delivery record keyed by Installation + Run + Activity + `reply`, a serialized claim, and explicit owner-confirmed resend after an uncertain result. A 2xx is delivered; request validation/auth failures are permanent until configuration changes; rate limits, timeouts, and transport failures are retryable only while the outcome is known to be unsent. Deadline or connection loss after request transmission is UNKNOWN and requires reconciliation or owner confirmation. The actual status/body mapping must be confirmed against a named Plane Cloud instance before implementation.

## Minimal internal interface

```go
type ActivityClient interface {
    CreateActivity(context.Context, ActivityRequest) (ActivityResult, error)
    ListActivities(context.Context, ActivityQuery) ([]Activity, error)
}

type ActivityRequest struct {
    WorkspaceSlug string
    RunID         string
    InstallationID string
    Type          ActivityType
    Body          string
    Signal        string
    Metadata      map[string]any
}
```

`ActivityClient` is deliberately narrower than `planeapi.Client`: it accepts provider-neutral typed data, returns a provider activity identifier/status, and owns token acquisition and HTTP details. `CreateActivity` must classify errors as delivered, retryable-unsent, permanent, or unknown; callers must never infer that classification from a generic error string. `ListActivities` is required for UNKNOWN reconciliation because provider creation is not idempotent.

The interface must not include `ephemeral` on create, because the published SDK does not accept it. It must not promise `external_link` update because the provider has no run update operation. The first run event also has no prompt body; context retrieval remains a separate client contract and is an external prerequisite.

## Application ownership boundary

`BootstrapResult` owns long-lived repositories and provider-independent services. It should eventually expose Plane DeliveryStore, ProcessingRepo, ProgressSink/Store, DiagnosticStore, ActivityClient factory, and a Plane Worker handle. `Handlers.Wire` owns HTTP/Connect handler injection only. A dedicated Plane runtime owner created during bootstrap should construct one Processor from `result.RunnerSvc`, ActivityClient-backed ReplyPublisher, and the durable stores; construct the Worker from that Processor, queue, and AppResolver; and retain the Worker for shutdown.

The application shutdown owner (`Handlers`/runtime shutdown coordinator, alongside existing Telegram and Linear Stop methods) must call Plane Worker Stop and wait before closing Redis/Mongo. `StartChannels` must fail if the configured production mode cannot construct the ActivityClient or durable stores. Memory implementations are test-only and must be selected explicitly by test bootstrap, never silently by production wiring.

Receiver and worker must receive the same ProcessingRepo and DeliveryStore instances (or the same Mongo collection boundary), and management RPCs must use that same DeliveryStore. Queue acceptance is complete only after Redis append; ProcessingRecord state transitions and delivery state transitions are durable writes around execution and publication.

## Unresolved provider evidence

The following require a named Plane Cloud or supported commercial instance: exact status codes and response body for each error class, rate-limit headers and retry-after semantics, maximum activity body size, whether `error` is ephemeral in the deployed version, activity-list consistency after a timeout, token scope enforcement, and whether the provider offers any hidden idempotency behavior. Community Edition cannot satisfy this contract because the documented Agent Run endpoints are absent.
