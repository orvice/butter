// AG-UI protocol client for a thread's endpoints,
// /api/agui/:agent_id/threads/:thread_id: its reads, Stop, and attaching to
// its run. Runs themselves (POST /api/agui/:agent_id) are driven by
// @ag-ui/client (ButterAGUIAgent, features/agui-chat/a2ui/agent.ts).
//
// Attaching streams the run's AG-UI events as SSE, read with a hand-rolled
// fetch + ReadableStream parser (sseEvents): native EventSource cannot send
// the auth headers. Failures before a stream opens arrive as non-200 JSON
// {error}. See docs/api.md.
import { ApiError, BASE_URL, authHeaders } from './client'

// AGUIEvent is one decoded SSE frame; `type` discriminates, everything else
// is event-specific and read defensively by the consumer.
export interface AGUIEvent {
  type: string
  [key: string]: unknown
}

// sseEvents reads an SSE body as AG-UI events, one per frame, as they
// arrive. A frame without data, such as a `: heartbeat` comment, carries no
// event, and a malformed frame is dropped rather than ending the stream.
// Leaving the loop early cancels the body.
async function* sseEvents(
  body: ReadableStream<Uint8Array>
): AsyncGenerator<AGUIEvent, void, undefined> {
  const reader = body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  try {
    for (;;) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      // SSE frames are separated by a blank line; the trailing partial frame
      // stays buffered until its terminator arrives.
      for (;;) {
        const sep = buffer.indexOf('\n\n')
        if (sep < 0) break
        const event = parseFrame(buffer.slice(0, sep))
        buffer = buffer.slice(sep + 2)
        if (event) yield event
      }
    }
    const last = parseFrame(buffer)
    if (last) yield last
  } finally {
    reader.cancel().catch(() => {})
  }
}

function parseFrame(frame: string): AGUIEvent | null {
  const dataLines = frame
    .split('\n')
    .filter((line) => line.startsWith('data:'))
    .map((line) => line.slice(5).trimStart())
  if (dataLines.length === 0) return null
  try {
    const parsed = JSON.parse(dataLines.join('\n')) as AGUIEvent
    return parsed && typeof parsed.type === 'string' ? parsed : null
  } catch {
    return null
  }
}

// applyAGUIStateDelta applies RFC 6902 operations to the client's state
// mirror. Butter emits top-level add/replace/remove ops; unsupported shapes
// return null so the caller can fall back to requesting a snapshot (by
// sending its stale mirror on the next run).
export function applyAGUIStateDelta(
  state: Record<string, unknown>,
  ops: Array<{ op: string; path: string; value?: unknown }>
): Record<string, unknown> | null {
  const next = { ...state }
  for (const op of ops) {
    if (!op.path.startsWith('/') || op.path.indexOf('/', 1) >= 0) return null
    const key = op.path.slice(1).split('~1').join('/').split('~0').join('~')
    switch (op.op) {
      case 'add':
      case 'replace':
        next[key] = op.value
        break
      case 'remove':
        delete next[key]
        break
      default:
        return null
    }
  }
  return next
}

// AGUIRunningRun is the run a thread read found in flight (`running`). The
// read then shows the thread as the run found it, plus the turn that started
// it (docs/api.md "Reads during a run").
export interface AGUIRunningRun {
  runId: string
  invocationId: string
}

// fetchAGUIUISnapshot reads a thread's current A2UI surfaces — read-only
// cards and unanswered forms — rebuilt from the persisted session, without
// starting a run. A thread the caller does not own, or one without UI,
// answers with no surfaces. It never waits for a run: during one it answers
// at once, with `running` and the surfaces as the run found them.
export function fetchAGUIUISnapshot<T>(
  agentId: string,
  threadId: string,
  signal?: AbortSignal
): Promise<T> {
  return fetchAGUIThread<T>(agentId, threadId, 'ui', 'UI snapshot', signal)
}

// fetchAGUIThreadHistory reads a thread's conversation as AG-UI messages,
// with its open Interrupts and the reply each restorable surface belongs
// to, rebuilt from the persisted session without starting a run. A thread
// the caller does not own answers with an empty history. It never waits for
// a run: during one it answers at once, with `running` and the conversation
// up to the turn that started the run.
export function fetchAGUIThreadHistory<T>(
  agentId: string,
  threadId: string,
  signal?: AbortSignal
): Promise<T> {
  return fetchAGUIThread<T>(
    agentId,
    threadId,
    'messages',
    'Thread history',
    signal
  )
}

// AGUIStoppedRun is the run a Stop reached. It ends shortly after, and its
// stream ends with a RUN_ERROR whose code is "stopped".
export interface AGUIStoppedRun {
  threadId: string
  runId: string
  invocationId: string
}

// stopAGUIRun stops the thread's detached run, on whichever Pod runs it
// (POST /api/agui/:agent_id/threads/:thread_id/stop, ADR-0016 decision 4).
// The server answers at once and never waits for the run to end. The answer
// is the run that will end, or null when no detached run is in flight: the
// thread is idle, or its run already ended.
export async function stopAGUIRun(
  agentId: string,
  threadId: string,
  signal?: AbortSignal
): Promise<AGUIStoppedRun | null> {
  const res = await fetch(aguiThreadURL(agentId, threadId, 'stop'), {
    method: 'POST',
    headers: authHeaders(),
    signal,
  })
  if (res.status === 204) return null
  if (!res.ok) throw await aguiFailure(res, 'Stop')
  return (await res.json()) as AGUIStoppedRun
}

// AGUI_FALLBACK_EVENT names the CUSTOM event that ends a stream of a
// detached run's log when it cannot follow the run to its end (docs/api.md
// "The fallback marker"): the log was truncated, expired or lost. The run is
// not affected, and the client reads the thread instead.
export const AGUI_FALLBACK_EVENT = 'butter.fallback'

// attachAGUIRun follows the thread's detached run, on whichever Pod runs it
// (GET /api/agui/:agent_id/threads/:thread_id/run, ADR-0016 decision 5). It
// resolves with the run's AG-UI events as they arrive: replayed from
// RUN_STARTED under the run's own runId, then each new one, up to the run's
// RUN_FINISHED or RUN_ERROR, or the fallback marker in their place. It
// resolves null when there is no log to follow (204): the thread is idle, its
// run ended over 5 minutes ago, or it did not detach. Aborting signal ends
// this observer only; the run goes on.
export async function attachAGUIRun(
  agentId: string,
  threadId: string,
  signal?: AbortSignal
): Promise<AsyncGenerator<AGUIEvent, void, undefined> | null> {
  const res = await fetch(aguiThreadURL(agentId, threadId, 'run'), {
    headers: { Accept: 'text/event-stream', ...authHeaders() },
    signal,
  })
  if (res.status === 204) return null
  if (!res.ok) throw await aguiFailure(res, 'Attach')
  if (!res.body) throw new ApiError('stream', 'response has no body')
  return sseEvents(res.body)
}

async function fetchAGUIThread<T>(
  agentId: string,
  threadId: string,
  resource: 'ui' | 'messages',
  label: string,
  signal?: AbortSignal
): Promise<T> {
  const res = await fetch(aguiThreadURL(agentId, threadId, resource), {
    headers: authHeaders(),
    signal,
  })
  if (!res.ok) throw await aguiFailure(res, label)
  return (await res.json()) as T
}

function aguiThreadURL(
  agentId: string,
  threadId: string,
  resource: 'ui' | 'messages' | 'stop' | 'run'
): string {
  return `${BASE_URL}/api/agui/${encodeURIComponent(agentId)}/threads/${encodeURIComponent(threadId)}/${resource}`
}

// aguiFailure is the error a failed request answers with: the server's
// {error} message, else the status.
async function aguiFailure(res: Response, label: string): Promise<ApiError> {
  let message = `${label} failed (${res.status})`
  try {
    const data = (await res.json()) as { error?: string }
    if (data?.error) message = data.error
  } catch {
    // Non-JSON error body; keep the status message.
  }
  return new ApiError(String(res.status), message)
}
