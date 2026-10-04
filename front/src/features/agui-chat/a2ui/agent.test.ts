import { describe, expect, it } from 'vitest'
import {
  ButterAGUIAgent,
  clientResume,
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
})
