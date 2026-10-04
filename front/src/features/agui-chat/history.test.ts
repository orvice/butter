import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fetchAGUIThreadHistory, fetchAGUIUISnapshot } from '@/api/agui'
import { ApiError } from '@/api/client'
import type { UISnapshot } from './a2ui/protocol'
import { A2UIStore } from './a2ui/store'
import {
  loadThread,
  restoreThread,
  runningOf,
  threadRepository,
  type ThreadHistory,
} from './history'

vi.mock('@/api/agui', () => ({
  fetchAGUIThreadHistory: vi.fn(),
  fetchAGUIUISnapshot: vi.fn(),
}))

const V = 'v0.9.1'

function snapshot(...surfaceIds: string[]): UISnapshot {
  return {
    version: V,
    catalogId: 'butter-basic-v1',
    threadId: 't',
    surfaces: surfaceIds.map((surfaceId) => ({
      surfaceId,
      kind: surfaceId.startsWith('form') ? 'form' : 'card',
      revision: 1,
      fallback: `${surfaceId} (text)`,
      envelopes: [
        {
          version: V,
          createSurface: { surfaceId, catalogId: 'butter-basic-v1' },
        },
        { version: V, updateComponents: { surfaceId, components: [] } },
      ],
    })),
  }
}

function history(partial: Partial<ThreadHistory>): ThreadHistory {
  return {
    threadId: 't',
    messages: [],
    interrupts: [],
    surfaces: [],
    ...partial,
  }
}

const conversation = history({
  messages: [
    { id: 'u1', role: 'user', content: 'Which tickets expire soon?' },
    {
      id: 'a1',
      role: 'assistant',
      content: 'Two tickets expire this month.',
      toolCalls: [
        {
          id: 'c1',
          type: 'function',
          function: { name: 'queryRecords', arguments: '{"status":"open"}' },
        },
      ],
    },
    { id: 'result:c1', role: 'tool', toolCallId: 'c1', content: '{"rows":2}' },
    { id: 'u2', role: 'user', content: 'Thanks' },
    { id: 'a2', role: 'assistant', content: 'You are welcome.' },
  ],
  surfaces: [{ surfaceId: 'card-1', messageId: 'a1' }],
})

describe('threadRepository', () => {
  it('rebuilds the conversation in order as one chain', () => {
    const { repository } = threadRepository(conversation, snapshot('card-1'))
    expect(
      repository.messages.map((m) => [m.message.id, m.message.role, m.parentId])
    ).toEqual([
      ['u1', 'user', null],
      ['a1', 'assistant', 'u1'],
      ['u2', 'user', 'a1'],
      ['a2', 'assistant', 'u2'],
    ])
    expect(repository.headId).toBe('a2')
    expect(repository.messages[0].message.content).toEqual([
      { type: 'text', text: 'Which tickets expire soon?' },
    ])
  })

  it('shows a reply as the live run left it: tool calls with results, its card, then its text', () => {
    const { repository, placed } = threadRepository(
      conversation,
      snapshot('card-1')
    )
    const reply = repository.messages[1].message
    expect(reply.content.map((p) => p.type)).toEqual([
      'tool-call',
      'data',
      'text',
    ])
    expect(reply.content[0]).toMatchObject({
      toolCallId: 'c1',
      toolName: 'queryRecords',
      args: { status: 'open' },
      result: { rows: 2 },
    })
    expect(reply.content[1]).toMatchObject({
      name: 'butter.a2ui',
      data: {
        surfaceId: 'card-1',
        revision: 1,
        seq: 0,
        envelope: { createSurface: { surfaceId: 'card-1' } },
      },
    })
    expect(reply.status).toEqual({ type: 'complete', reason: 'stop' })
    expect([...placed]).toEqual(['card-1'])
  })

  it('leaves a surface the snapshot no longer holds out of the reply', () => {
    const { repository, placed } = threadRepository(conversation, snapshot())
    expect(repository.messages[1].message.content.map((p) => p.type)).toEqual([
      'tool-call',
      'text',
    ])
    expect(placed.size).toBe(0)
  })

  it('hands the open Interrupts to the last reply, as a live run ending on them does', () => {
    const interrupts = [
      { id: 'ask-1', reason: 'human_input', message: 'Approve?' },
    ]
    const { repository } = threadRepository(
      history({
        ...conversation,
        interrupts,
        surfaces: [{ surfaceId: 'form-1', messageId: 'a2' }],
      }),
      snapshot('form-1')
    )
    const last = repository.messages[3].message
    expect(last.status).toEqual({
      type: 'requires-action',
      reason: 'interrupt',
    })
    expect(last.metadata.custom).toEqual({ agui: { interrupts } })
    expect(last.content.map((p) => p.type)).toEqual(['data', 'text'])
    expect(repository.messages[1].message.status).toEqual({
      type: 'complete',
      reason: 'stop',
    })
  })

  it('shows a turn that carried images as the composer sent it: its text, and its images as attachments', () => {
    const { repository } = threadRepository(
      history({
        messages: [
          {
            id: 'u1',
            role: 'user',
            content: [
              { type: 'text', text: 'What is in these?' },
              {
                type: 'image',
                source: {
                  type: 'data',
                  value: 'iVBORw0K',
                  mimeType: 'image/png',
                },
              },
              {
                type: 'image',
                source: {
                  type: 'data',
                  value: '/9j/4AAQ',
                  mimeType: 'image/jpeg',
                },
              },
            ],
          },
          { id: 'a1', role: 'assistant', content: 'A cat and a dog.' },
        ],
      }),
      snapshot()
    )
    const turn = repository.messages[0].message
    expect(turn.role).toBe('user')
    expect(turn.content).toEqual([{ type: 'text', text: 'What is in these?' }])
    expect(turn.role === 'user' && turn.attachments).toEqual([
      {
        id: 'u1:image-1',
        type: 'image',
        name: 'Image 1',
        contentType: 'image/png',
        status: { type: 'complete' },
        content: [{ type: 'image', image: 'data:image/png;base64,iVBORw0K' }],
      },
      {
        id: 'u1:image-2',
        type: 'image',
        name: 'Image 2',
        contentType: 'image/jpeg',
        status: { type: 'complete' },
        content: [{ type: 'image', image: 'data:image/jpeg;base64,/9j/4AAQ' }],
      },
    ])
    expect(repository.messages[1].message.content).toEqual([
      { type: 'text', text: 'A cat and a dog.' },
    ])
  })

  it('shows a turn of images alone without text, and leaves out what is not an inline image', () => {
    const { repository } = threadRepository(
      history({
        messages: [
          {
            id: 'u1',
            role: 'user',
            content: [
              {
                type: 'image',
                source: { type: 'url', value: 'https://example.com/a.png' },
              },
              { type: 'document', source: { type: 'data', value: 'JVBE' } },
              {
                type: 'image',
                source: {
                  type: 'data',
                  value: 'R0lGOD',
                  mimeType: 'image/gif',
                },
              },
            ],
          },
        ],
      }),
      snapshot()
    )
    const turn = repository.messages[0].message
    expect(turn.content).toEqual([])
    expect(
      turn.role === 'user' && turn.attachments.map((a) => a.content)
    ).toEqual([[{ type: 'image', image: 'data:image/gif;base64,R0lGOD' }]])
  })

  it('reads several text parts as paragraphs, as the server writes a turn of text', () => {
    const { repository } = threadRepository(
      history({
        messages: [
          {
            id: 'u1',
            role: 'user',
            content: [
              { type: 'text', text: 'First.' },
              {
                type: 'image',
                source: {
                  type: 'data',
                  value: 'UklGR',
                  mimeType: 'image/webp',
                },
              },
              { type: 'text', text: 'Second.' },
            ],
          },
        ],
      }),
      snapshot()
    )
    expect(repository.messages[0].message.content).toEqual([
      { type: 'text', text: 'First.\n\nSecond.' },
    ])
  })

  it('gives open Interrupts a reply of their own when the history has none', () => {
    const interrupts = [
      { id: 'ask-1', reason: 'human_input', message: 'Approve?' },
    ]
    const { repository } = threadRepository(
      history({
        messages: [{ id: 'u1', role: 'user', content: 'Go' }],
        interrupts,
      }),
      snapshot()
    )
    expect(repository.messages.map((m) => m.message.role)).toEqual([
      'user',
      'assistant',
    ])
    expect(repository.messages[1].message.status).toEqual({
      type: 'requires-action',
      reason: 'interrupt',
    })
  })
})

describe('loadThread', () => {
  const read = { history: history({}), snapshot: snapshot() }

  beforeEach(() => {
    vi.mocked(fetchAGUIThreadHistory).mockReset()
    vi.mocked(fetchAGUIUISnapshot).mockReset()
    vi.mocked(fetchAGUIThreadHistory).mockResolvedValue(read.history)
    vi.mocked(fetchAGUIUISnapshot).mockResolvedValue(read.snapshot)
  })

  it('reads the history and the UI snapshot of the thread, once each', async () => {
    await expect(loadThread('a', 't')).resolves.toEqual(read)
    expect(fetchAGUIThreadHistory).toHaveBeenCalledTimes(1)
    expect(fetchAGUIThreadHistory).toHaveBeenCalledWith('a', 't', undefined)
    expect(fetchAGUIUISnapshot).toHaveBeenCalledTimes(1)
    expect(fetchAGUIUISnapshot).toHaveBeenCalledWith('a', 't', undefined)
  })

  it('shows a run in flight as the reads report it, without waiting for it', async () => {
    const running = { runId: 'run-2', invocationId: 'inv-2' }
    const cut = history({
      messages: [{ id: 'u1', role: 'user', content: 'plan the trip' }],
      running,
    })
    vi.mocked(fetchAGUIThreadHistory).mockResolvedValue(cut)
    await expect(loadThread('a', 't')).resolves.toEqual({
      history: cut,
      snapshot: read.snapshot,
    })
    expect(fetchAGUIThreadHistory).toHaveBeenCalledTimes(1)
  })

  for (const [code, message] of [
    ['409', 'a run is in progress on this thread'],
    ['503', 'run state unavailable, retry later'],
    ['500', 'session store down'],
  ]) {
    it(`reads nothing again after a ${code}: the page offers Retry`, async () => {
      vi.mocked(fetchAGUIUISnapshot).mockRejectedValue(
        new ApiError(code, message)
      )
      await expect(loadThread('a', 't')).rejects.toThrow(message)
      expect(fetchAGUIUISnapshot).toHaveBeenCalledTimes(1)
      expect(fetchAGUIThreadHistory).toHaveBeenCalledTimes(1)
    })
  }

  it('restores the surfaces alone from a server without the history endpoint', async () => {
    vi.mocked(fetchAGUIThreadHistory).mockRejectedValue(
      new ApiError('404', 'not found')
    )
    vi.mocked(fetchAGUIUISnapshot).mockResolvedValue(snapshot('card-1'))
    await expect(loadThread('a', 't')).resolves.toEqual({
      history: history({}),
      snapshot: snapshot('card-1'),
    })
  })
})

describe('runningOf', () => {
  const running = { runId: 'run-2', invocationId: 'inv-2' }

  it('names the run either read found in flight', () => {
    expect(
      runningOf({ history: history({ running }), snapshot: snapshot() })
    ).toEqual(running)
    // A run that started after the history was read.
    expect(
      runningOf({ history: history({}), snapshot: { ...snapshot(), running } })
    ).toEqual(running)
    expect(runningOf({ history: history({}), snapshot: snapshot() })).toBeNull()
  })
})

describe('restoreThread', () => {
  it('puts the surfaces in the store, and returns the conversation with each in its reply', () => {
    const store = new A2UIStore()
    const repository = restoreThread(
      { history: conversation, snapshot: snapshot('card-1', 'card-2') },
      store
    )
    expect(repository.headId).toBe('a2')
    expect(repository.messages[1].message.content.map((p) => p.type)).toEqual([
      'tool-call',
      'data',
      'text',
    ])
    expect(store.list().map((e) => e.id)).toEqual(['card-1', 'card-2'])
    // card-2 is in no reply, so it shows on its own.
    expect(store.isPlaced('card-1')).toBe(true)
    expect(store.isPlaced('card-2')).toBe(false)
  })

  it('drops the surfaces a later read of the thread no longer has', () => {
    const store = new A2UIStore()
    restoreThread(
      { history: conversation, snapshot: snapshot('card-1', 'card-2') },
      store
    )
    restoreThread(
      { history: conversation, snapshot: snapshot('card-1') },
      store
    )
    expect(store.list().map((e) => [e.id, e.deleted])).toEqual([
      ['card-1', false],
      ['card-2', true],
    ])
  })
})
