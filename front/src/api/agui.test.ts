import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { stopAGUIRun } from './agui'

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
