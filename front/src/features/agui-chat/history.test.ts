import { describe, expect, it } from 'vitest'
import type { UISnapshot } from './a2ui/protocol'
import { threadRepository, type ThreadHistory } from './history'

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
