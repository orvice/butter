// AG-UI protocol client for POST /api/agui/:agent_id.
//
// The endpoint streams AG-UI events as SSE over a POST body, which rules out
// native EventSource — this is a hand-rolled fetch + ReadableStream parser.
// Pre-stream failures arrive as non-200 JSON {error}; once the stream opens,
// failures arrive in-band as RUN_ERROR events. See docs/api.md.
import { ApiError, BASE_URL, authHeaders } from './client'

export interface AGUIToolCallRef {
  id: string
  type: 'function'
  function: { name: string; arguments: string }
}

export interface AGUIMessage {
  id: string
  role: 'user' | 'assistant' | 'tool' | 'system' | 'developer'
  content?: string
  toolCalls?: AGUIToolCallRef[]
  toolCallId?: string
  error?: string
}

export interface AGUIResumeEntry {
  interruptId: string
  status: 'resolved'
  payload?: unknown
}

export interface AGUIRunInput {
  threadId: string
  runId: string
  messages: AGUIMessage[]
  tools?: unknown[]
  state?: Record<string, unknown> | null
  resume?: AGUIResumeEntry[]
}

export interface AGUIInterrupt {
  id: string
  reason?: string
  message?: string
}

// AGUIEvent is one decoded SSE frame; `type` discriminates, everything else
// is event-specific and read defensively by the consumer.
export interface AGUIEvent {
  type: string
  [key: string]: unknown
}

// runAGUIAgent POSTs one run and invokes onEvent for every streamed AG-UI
// event, resolving when the stream ends. Abort via the signal detaches the
// client; the server cancels the run and releases its session lease.
export async function runAGUIAgent(
  agentId: string,
  input: AGUIRunInput,
  opts: { signal?: AbortSignal; onEvent: (event: AGUIEvent) => void }
): Promise<void> {
  const res = await fetch(
    `${BASE_URL}/api/agui/${encodeURIComponent(agentId)}`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'text/event-stream',
        ...authHeaders(),
      },
      body: JSON.stringify(input),
      signal: opts.signal,
    }
  )

  if (!res.ok) {
    let message = `AG-UI request failed (${res.status})`
    try {
      const data = (await res.json()) as { error?: string }
      if (data?.error) message = data.error
    } catch {
      // Non-JSON error body; keep the status message.
    }
    throw new ApiError(String(res.status), message)
  }
  if (!res.body) {
    throw new ApiError('stream', 'response has no body')
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    // SSE frames are separated by a blank line; the trailing partial frame
    // stays buffered until its terminator arrives.
    for (;;) {
      const sep = buffer.indexOf('\n\n')
      if (sep < 0) break
      const frame = buffer.slice(0, sep)
      buffer = buffer.slice(sep + 2)
      dispatchFrame(frame, opts.onEvent)
    }
  }
  if (buffer.trim() !== '') {
    dispatchFrame(buffer, opts.onEvent)
  }
}

function dispatchFrame(frame: string, onEvent: (event: AGUIEvent) => void) {
  const dataLines = frame
    .split('\n')
    .filter((line) => line.startsWith('data:'))
    .map((line) => line.slice(5).trimStart())
  if (dataLines.length === 0) return
  try {
    const parsed = JSON.parse(dataLines.join('\n')) as AGUIEvent
    if (parsed && typeof parsed.type === 'string') {
      onEvent(parsed)
    }
  } catch {
    // A malformed frame is dropped rather than killing the stream.
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
