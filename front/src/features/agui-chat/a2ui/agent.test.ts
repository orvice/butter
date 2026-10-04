import type { Message } from '@ag-ui/client'
import { describe, expect, it } from 'vitest'
import {
  ButterAGUIAgent,
  clientResume,
  runMessages,
  wireResume,
  type NextRun,
  type ResumeEntry,
} from './agent'

const answer = (interruptId: string, payload: unknown): NextRun => ({
  kind: 'answer',
  entry: { interruptId, status: 'resolved', payload },
})
const message: NextRun = { kind: 'message' }
const open = (...ids: string[]) =>
  ids.map((id) => ({ id, reason: 'human_input' }))

describe('clientResume', () => {
  it('covers every other open Interrupt with a placeholder', () => {
    expect(clientResume(answer('b', 'yes'), open('a', 'b', 'c'))).toEqual([
      { interruptId: 'b', status: 'resolved', payload: 'yes' },
      { interruptId: 'a', status: 'cancelled' },
      { interruptId: 'c', status: 'cancelled' },
    ])
  })

  it('covers every open Interrupt for a plain message', () => {
    expect(clientResume(message, open('a', 'b'))).toEqual([
      { interruptId: 'a', status: 'cancelled' },
      { interruptId: 'b', status: 'cancelled' },
    ])
  })

  it('keeps an answer to an Interrupt the client does not track', () => {
    // After a reload the client tracks nothing, yet the thread's question
    // can still be answered.
    expect(clientResume(answer('a', 'yes'), [])).toEqual([
      { interruptId: 'a', status: 'resolved', payload: 'yes' },
    ])
    expect(clientResume(message, [])).toEqual([])
  })
})

describe('wireResume', () => {
  it('sends exactly the one answer', () => {
    expect(wireResume(answer('b', { butterForm: {} }))).toEqual([
      { interruptId: 'b', status: 'resolved', payload: { butterForm: {} } },
    ])
  })

  it('sends no resume for a plain message', () => {
    expect(wireResume(message)).toBeUndefined()
  })
})

// A transcript as the assistant-ui runtime hands it to the client: a turn
// that carried an image, a reply that called a frontend tool, its result.
const photo: Message = {
  id: 'u1',
  role: 'user',
  content: [
    { type: 'text', text: 'What is in this picture?' },
    {
      type: 'image',
      source: { type: 'data', value: 'iVBORw0KGgo=', mimeType: 'image/png' },
    },
  ],
}
const user = (id: string, content: string): Message => ({
  id,
  role: 'user',
  content,
})
const reply = (id: string, ...toolCallIds: string[]): Message => ({
  id,
  role: 'assistant',
  content: '',
  ...(toolCallIds.length > 0 && {
    toolCalls: toolCallIds.map((callId) => ({
      id: callId,
      type: 'function' as const,
      function: { name: 'pickDate', arguments: '{}' },
    })),
  }),
})
const result = (toolCallId: string): Message => ({
  id: `${toolCallId}:tool`,
  role: 'tool',
  content: '{"date":"2026-10-05"}',
  toolCallId,
})

describe('runMessages', () => {
  it('sends only the trailing user message, and no earlier image', () => {
    const sent = runMessages([
      photo,
      reply('a1'),
      user('u2', 'And the colours?'),
    ])
    expect(sent).toEqual([user('u2', 'And the colours?')])
  })

  it('sends a trailing user message with its images', () => {
    expect(runMessages([user('u1', 'hi'), reply('a1'), photo])).toEqual([photo])
  })

  it('sends the trailing tool results of a frontend tool continuation', () => {
    const sent = runMessages([
      photo,
      reply('a1', 'call-1', 'call-2'),
      result('call-1'),
      result('call-2'),
    ])
    expect(sent).toEqual([result('call-1'), result('call-2')])
  })

  it('sends a message written after tool results on its own', () => {
    const sent = runMessages([
      user('u1', 'book it'),
      reply('a1', 'call-1'),
      result('call-1'),
      user('u2', 'actually, wait'),
    ])
    expect(sent).toEqual([user('u2', 'actually, wait')])
  })

  it('sends the last user message when the transcript ends on a reply', () => {
    // A run that only answers an Interrupt by its resume: the server reads
    // the resume, and would read this message without one.
    expect(runMessages([photo, reply('a1')])).toEqual([photo])
  })

  it('sends nothing when there is nothing the server reads', () => {
    expect(runMessages([])).toEqual([])
    expect(runMessages([reply('a1')])).toEqual([])
  })
})

// The agent against the real AG-UI client: its pre-run check and the
// request body it hands to fetch.
describe('ButterAGUIAgent', () => {
  function agent(finish: Record<string, unknown> = {}) {
    const bodies: Array<Record<string, unknown>> = []
    const a = new ButterAGUIAgent({
      url: 'http://butter.test/api/agui/a1',
      threadId: 't',
      fetch: async (_url, init) => {
        bodies.push(JSON.parse(String(init.body)))
        const events = [
          { type: 'RUN_STARTED', threadId: 't', runId: 'r' },
          { type: 'RUN_FINISHED', threadId: 't', runId: 'r', ...finish },
        ]
        return new Response(
          events.map((e) => `data: ${JSON.stringify(e)}\n\n`).join(''),
          { headers: { 'Content-Type': 'text/event-stream' } }
        )
      },
    })
    return { agent: a, bodies }
  }

  // What the assistant-ui runtime passes when it steers away from open
  // Interrupts: every one it was not given an answer for is cancelled.
  const steerAway = (...ids: string[]): ResumeEntry[] =>
    ids.map((interruptId) => ({ interruptId, status: 'cancelled' }))

  it('sends a plain message while Interrupts are open', async () => {
    const { agent: a, bodies } = agent()
    a.pendingInterrupts = open('a', 'b')
    a.sendNextRunAsMessage()
    await a.runAgent({ resume: steerAway('a', 'b') })
    expect(bodies).toHaveLength(1)
    expect(bodies[0]).not.toHaveProperty('resume')
    expect(bodies[0].forwardedProps).toMatchObject({
      butterA2UI: { catalogs: ['butter-basic-v1'] },
    })
  })

  it('sends a plain message even when the runtime no longer lists the open Interrupts', async () => {
    const { agent: a, bodies } = agent()
    a.pendingInterrupts = open('a')
    a.sendNextRunAsMessage()
    await a.runAgent({})
    expect(bodies[0]).not.toHaveProperty('resume')
  })

  it('sends one answer and nothing for the other open Interrupts', async () => {
    const { agent: a, bodies } = agent()
    a.pendingInterrupts = open('a', 'b')
    a.resumeNextRunWith('b', '2.4.1')
    await a.runAgent({ resume: steerAway('a', 'b') })
    expect(bodies[0].resume).toEqual([
      { interruptId: 'b', status: 'resolved', payload: '2.4.1' },
    ])
  })

  it('arms only the next run', async () => {
    const { agent: a, bodies } = agent({
      outcome: {
        type: 'interrupt',
        interrupts: [{ id: 'a', reason: 'human_input' }],
      },
    })
    a.sendNextRunAsMessage()
    await a.runAgent({})
    expect(a.pendingInterrupts.map((i) => i.id)).toEqual(['a'])
    // An unarmed run must address the open Interrupt itself.
    await expect(a.runAgent({})).rejects.toThrow(/not addressed by resume/)
    expect(bodies).toHaveLength(1)
  })

  it('leaves an unarmed run alone', async () => {
    const { agent: a, bodies } = agent()
    const resume: ResumeEntry[] = [
      { interruptId: 'a', status: 'resolved', payload: 'yes' },
    ]
    await a.runAgent({ resume })
    expect(bodies[0].resume).toEqual(resume)
  })

  it('forgets an armed answer that was cleared', async () => {
    const { agent: a, bodies } = agent()
    a.resumeNextRunWith('a', 'yes')
    a.clearNextRun()
    await a.runAgent({})
    expect(bodies[0]).not.toHaveProperty('resume')
  })

  it('sends the trailing message alone and keeps the transcript', async () => {
    const { agent: a, bodies } = agent()
    // The assistant-ui runtime hands the client the whole transcript.
    a.messages = [photo, reply('a1'), user('u2', 'And the colours?')]
    await a.runAgent({})
    expect(bodies[0].messages).toEqual([user('u2', 'And the colours?')])
    expect(a.messages.map((m) => m.id)).toEqual(['u1', 'a1', 'u2'])
  })

  it('sends the trailing tool results of a continuation', async () => {
    const { agent: a, bodies } = agent()
    a.messages = [photo, reply('a1', 'call-1'), result('call-1')]
    await a.runAgent({})
    expect(bodies[0].messages).toEqual([result('call-1')])
  })

  it('asks every run to detach, next to A2UI and what the runtime forwards', async () => {
    const { agent: a, bodies } = agent()
    a.messages = [user('u1', 'hi')]
    await a.runAgent({ forwardedProps: { fromRuntime: true } })
    expect(bodies[0].forwardedProps).toEqual({
      fromRuntime: true,
      butterA2UI: { version: 'v0.9.1', catalogs: ['butter-basic-v1'] },
      butterRun: { detach: true },
    })
  })

  it('asks to detach whatever the run carries', async () => {
    const { agent: a, bodies } = agent()
    // An answer to one open Interrupt.
    a.pendingInterrupts = open('a', 'b')
    a.resumeNextRunWith('b', '2.4.1')
    await a.runAgent({ resume: steerAway('a', 'b') })
    // A message the server takes as the answer to its oldest Interrupt.
    a.pendingInterrupts = open('a')
    a.sendNextRunAsMessage()
    await a.runAgent({ resume: steerAway('a') })
    // A continuation after frontend tool calls.
    a.pendingInterrupts = []
    a.messages = [
      user('u1', 'book it'),
      reply('a1', 'call-1'),
      result('call-1'),
    ]
    await a.runAgent({})
    expect(bodies).toHaveLength(3)
    for (const body of bodies) {
      expect(body.forwardedProps).toMatchObject({
        butterRun: { detach: true },
      })
    }
  })
})
