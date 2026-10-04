import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { attachAGUIRun, stopAGUIRun, type AGUIEvent } from './agui'

// The Stop endpoint as docs/api.md "Stopping a run" describes it: 202 with
// the run it reached, 204 when no detached run is in flight, and an
// {error} body otherwise.
describe('stopAGUIRun', () => {
  const sent: Array<{ url: string; init: RequestInit }> = []

  function answer(response: Response) {
    vi.stubGlobal('fetch', async (url: string, init: RequestInit) => {
      sent.push({ url, init })
      return response
    })
  }

  beforeEach(() => {
    sent.length = 0
    const items: Record<string, string> = {
      butter_token: 'tok-1',
      butter_workspace_id: 'ws-1',
    }
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => items[key] ?? null,
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('posts to the thread’s stop endpoint as the caller, in the workspace', async () => {
    answer(new Response(null, { status: 204 }))
    await stopAGUIRun('agent/1', 't 1')
    expect(sent).toHaveLength(1)
    expect(sent[0].url).toBe('/api/agui/agent%2F1/threads/t%201/stop')
    expect(sent[0].init.method).toBe('POST')
    expect(sent[0].init.headers).toEqual({
      Authorization: 'Bearer tok-1',
      'X-Workspace-ID': 'ws-1',
    })
    expect(sent[0].init.body).toBeUndefined()
  })

  it('returns the run a Stop reached', async () => {
    const run = { threadId: 't-1', runId: 'run-1', invocationId: 'inv-1' }
    answer(Response.json(run, { status: 202 }))
    await expect(stopAGUIRun('a1', 't-1')).resolves.toEqual(run)
  })

  it('returns null when no run was in flight', async () => {
    answer(new Response(null, { status: 204 }))
    await expect(stopAGUIRun('a1', 't-1')).resolves.toBeNull()
  })

  it('rejects with the server’s error', async () => {
    answer(
      Response.json({ error: 'stop unavailable, retry later' }, { status: 503 })
    )
    await expect(stopAGUIRun('a1', 't-1')).rejects.toMatchObject({
      name: 'ApiError',
      code: '503',
      message: 'stop unavailable, retry later',
    })
  })

  it('names the status when the error has no body', async () => {
    answer(new Response('upstream down', { status: 502 }))
    await expect(stopAGUIRun('a1', 't-1')).rejects.toMatchObject({
      code: '502',
      message: 'Stop failed (502)',
    })
  })
})

// Attaching as docs/api.md "Attaching to a run" describes it: 204 when there
// is no log to follow, else the run's AG-UI events as SSE, with heartbeat
// comments while the run is quiet, and an {error} body on failure.
describe('attachAGUIRun', () => {
  const sent: Array<{ url: string; init: RequestInit }> = []
  const encoder = new TextEncoder()

  function answer(response: Response) {
    vi.stubGlobal('fetch', async (url: string, init: RequestInit) => {
      sent.push({ url, init })
      return response
    })
  }

  // sse is an SSE body the test hands out in chunks, as they arrive. It
  // records whether the reader let go of it.
  function sse() {
    let controller!: ReadableStreamDefaultController<Uint8Array>
    const record = { cancelled: false }
    const body = new ReadableStream<Uint8Array>({
      start: (c) => {
        controller = c
      },
      cancel: () => {
        record.cancelled = true
      },
    })
    return {
      body,
      record,
      chunk: (text: string) => controller.enqueue(encoder.encode(text)),
      close: () => controller.close(),
    }
  }

  const frame = (event: Record<string, unknown>) =>
    `data: ${JSON.stringify(event)}\n\n`

  beforeEach(() => {
    sent.length = 0
    const items: Record<string, string> = {
      butter_token: 'tok-1',
      butter_workspace_id: 'ws-1',
    }
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => items[key] ?? null,
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('gets the thread’s run as the caller, in the workspace, as SSE', async () => {
    const aborted = new AbortController()
    answer(new Response(null, { status: 204 }))
    await attachAGUIRun('agent/1', 't 1', aborted.signal)
    expect(sent).toHaveLength(1)
    expect(sent[0].url).toBe('/api/agui/agent%2F1/threads/t%201/run')
    expect(sent[0].init.method).toBeUndefined()
    expect(sent[0].init.headers).toEqual({
      Accept: 'text/event-stream',
      Authorization: 'Bearer tok-1',
      'X-Workspace-ID': 'ws-1',
    })
    expect(sent[0].init.signal).toBe(aborted.signal)
  })

  it('returns null when there is no log to follow', async () => {
    answer(new Response(null, { status: 204 }))
    await expect(attachAGUIRun('a1', 't-1')).resolves.toBeNull()
  })

  it('streams each event as its frame arrives, and skips heartbeats', async () => {
    const body = sse()
    answer(
      new Response(body.body, {
        status: 200,
        headers: { 'Content-Type': 'text/event-stream' },
      })
    )
    const events = await attachAGUIRun('a1', 't-1')
    expect(events).not.toBeNull()
    const started = { type: 'RUN_STARTED', threadId: 't-1', runId: 'run-1' }
    const text = {
      type: 'TEXT_MESSAGE_CONTENT',
      messageId: 'm1',
      delta: 'Hi',
    }
    const finished = { type: 'RUN_FINISHED', threadId: 't-1', runId: 'run-1' }
    body.chunk(frame(started) + ': heartbeat\n\n')
    expect((await events!.next()).value).toEqual(started)
    // A frame split across chunks arrives whole, and a malformed one is
    // dropped.
    const half = Math.floor(frame(text).length / 2)
    body.chunk(frame(text).slice(0, half))
    const second = events!.next()
    body.chunk(frame(text).slice(half) + 'data: {not json}\n\n')
    expect((await second).value).toEqual(text)
    body.chunk(frame(finished))
    body.close()
    const rest: AGUIEvent[] = []
    for await (const event of events!) rest.push(event)
    expect(rest).toEqual([finished])
  })

  it('lets go of the stream when its reader stops early', async () => {
    const body = sse()
    answer(new Response(body.body, { status: 200 }))
    const events = await attachAGUIRun('a1', 't-1')
    body.chunk(frame({ type: 'RUN_STARTED', threadId: 't-1', runId: 'r' }))
    for await (const event of events!) {
      expect(event.type).toBe('RUN_STARTED')
      break
    }
    expect(body.record.cancelled).toBe(true)
  })

  it('rejects with the server’s error', async () => {
    answer(
      Response.json(
        { error: 'run log unavailable, retry later' },
        { status: 503 }
      )
    )
    await expect(attachAGUIRun('a1', 't-1')).rejects.toMatchObject({
      name: 'ApiError',
      code: '503',
      message: 'run log unavailable, retry later',
    })
  })
})
