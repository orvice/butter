import type { ChatModelRunResult, ThreadMessage } from '@assistant-ui/react'
import { describe, expect, it } from 'vitest'
import type { AGUIEvent } from '@/api/agui'
// The aggregator the AG-UI runtime folds a run it streams with, and the
// parser its events go through first, as the pinned release ships them. The
// package exports neither, so they are read from its files: RunFold must
// fold Butter's events into the snapshots they fold them into.
import { RunAggregator } from '../../../node_modules/@assistant-ui/react-ag-ui/dist/runtime/adapter/run-aggregator.js'
import { parseAgUiEvent } from '../../../node_modules/@assistant-ui/react-ag-ui/dist/runtime/event-parser.js'
import { RunFold, toolCallIdsOf, type RunEndEvent } from './run-fold'

// Runs as the server streams them (internal/handler/http/agui_sink.go): a
// detached run opens with RUN_STARTED and a STATE_SNAPSHOT, gives its reply
// one message ID, and sends a tool call's arguments whole.

const THREAD = 't-trip'
const RUN = 'run-2'
const MSG = 'msg-2'
const V = 'v0.9.1'

const started = (snapshot: Record<string, unknown> = {}): AGUIEvent[] => [
  { type: 'RUN_STARTED', threadId: THREAD, runId: RUN },
  { type: 'STATE_SNAPSHOT', snapshot },
]

const text = (...deltas: string[]): AGUIEvent[] => [
  { type: 'TEXT_MESSAGE_START', messageId: MSG, role: 'assistant' },
  ...deltas.map((delta) => ({
    type: 'TEXT_MESSAGE_CONTENT',
    messageId: MSG,
    delta,
  })),
  { type: 'TEXT_MESSAGE_END', messageId: MSG },
]

const toolCall = (id: string, name: string, args: unknown): AGUIEvent[] => [
  { type: 'TOOL_CALL_START', toolCallId: id, toolCallName: name },
  { type: 'TOOL_CALL_ARGS', toolCallId: id, delta: JSON.stringify(args) },
  { type: 'TOOL_CALL_END', toolCallId: id },
]

const toolResult = (id: string, result: unknown): AGUIEvent => ({
  type: 'TOOL_CALL_RESULT',
  messageId: MSG,
  toolCallId: id,
  content: JSON.stringify(result),
  role: 'tool',
})

const finished = (outcome: Record<string, unknown> = { type: 'success' }) => ({
  type: 'RUN_FINISHED',
  threadId: THREAD,
  runId: RUN,
  outcome,
})

function a2ui(
  surfaceId: string,
  kind: 'card' | 'form',
  seq: number,
  envelope: Record<string, unknown>,
  form?: Record<string, unknown>
): AGUIEvent {
  return {
    type: 'CUSTOM',
    name: 'butter.a2ui',
    value: {
      version: V,
      surfaceId,
      kind,
      revision: 1,
      seq,
      threadId: THREAD,
      runId: RUN,
      messageId: MSG,
      envelope: { version: V, ...envelope },
      ...(form ? { form } : {}),
    },
  }
}

// cardCreated is a Result Card's creation, as render_ui's write streams it.
const cardCreated = (surfaceId: string): AGUIEvent[] => [
  a2ui(surfaceId, 'card', 0, {
    createSurface: { surfaceId, catalogId: 'butter-basic-v1' },
  }),
  a2ui(surfaceId, 'card', 1, {
    updateComponents: {
      surfaceId,
      components: [{ id: 'root', component: 'Text', text: 'Lisbon' }],
    },
  }),
]

const QUESTION = { id: 'int-1', reason: 'human_input', message: 'Which hotel?' }

// fold folds a run's events as a page that follows the run does, from a
// shared state of state, and records every snapshot, every butter.a2ui
// value and the state as the run left it.
function fold(
  events: AGUIEvent[],
  { state = undefined as unknown, earlier = [] as ThreadMessage[] } = {}
) {
  const a2uiValues: unknown[] = []
  let shared = state
  const run = new RunFold(
    {
      a2ui: (value) => a2uiValues.push(value),
      state: (update) => (shared = update(shared)),
    },
    toolCallIdsOf(earlier)
  )
  const snapshots: ChatModelRunResult[] = []
  for (const event of events) {
    const snapshot = run.handle(event)
    if (snapshot) snapshots.push(snapshot)
  }
  return { snapshots, a2ui: a2uiValues, state: shared, end: run.end }
}

// aggregate folds the same events as the AG-UI runtime folds a run it
// streams: each parsed, then handed to the pinned aggregator, which the
// runtime keeps the shared state from (STATE_*). Its snapshots carry the
// stream's timing, which a replay leaves out.
function aggregate(events: AGUIEvent[]): ChatModelRunResult[] {
  const snapshots: ChatModelRunResult[] = []
  const aggregator = new RunAggregator({
    showThinking: true,
    logger: { debug: () => {}, error: () => {} },
    emit: (snapshot) => snapshots.push(withoutTiming(snapshot)),
  })
  for (const raw of events) {
    const event = parseAgUiEvent(raw)
    if (!event || event.type === 'STATE_SNAPSHOT') continue
    if (event.type === 'STATE_DELTA') continue
    aggregator.handle(event)
  }
  return snapshots
}

function withoutTiming(snapshot: ChatModelRunResult): ChatModelRunResult {
  if (!snapshot.metadata) return snapshot
  const { timing: _timing, ...metadata } = snapshot.metadata
  const { metadata: _metadata, ...rest } = snapshot
  return Object.keys(metadata).length > 0 ? { ...rest, metadata } : rest
}

const last = (snapshots: ChatModelRunResult[]) =>
  snapshots[snapshots.length - 1]

const statuses = (snapshots: ChatModelRunResult[]) =>
  snapshots.map((s) => s.status?.type)

describe('RunFold', () => {
  it('folds streamed text into one text part that grows with each delta', () => {
    const events = [
      ...started(),
      ...text('Three days', ' in Lisbon:', ' Alfama, Belém and Sintra.'),
      finished(),
    ]
    const { snapshots, end } = fold(events)
    expect(snapshots.map((s) => s.content)).toEqual([
      [],
      [],
      [{ type: 'text', text: 'Three days' }],
      [{ type: 'text', text: 'Three days in Lisbon:' }],
      [
        {
          type: 'text',
          text: 'Three days in Lisbon: Alfama, Belém and Sintra.',
        },
      ],
      [
        {
          type: 'text',
          text: 'Three days in Lisbon: Alfama, Belém and Sintra.',
        },
      ],
      [
        {
          type: 'text',
          text: 'Three days in Lisbon: Alfama, Belém and Sintra.',
        },
      ],
    ])
    expect(statuses(snapshots)).toEqual([
      'running',
      'running',
      'running',
      'running',
      'running',
      'running',
      'complete',
    ])
    expect(last(snapshots).status).toEqual({
      type: 'complete',
      reason: 'unknown',
    })
    expect(end).toEqual({ type: 'RUN_FINISHED' })
    expect(snapshots).toEqual(aggregate(events))
  })

  it('folds a tool call with its result ahead of the text that follows it', () => {
    const events = [
      ...started(),
      ...toolCall('call-1', 'search_hotels', { city: 'Lisbon' }),
      toolResult('call-1', { hotels: 3 }),
      ...text('Three hotels have rooms.'),
      finished(),
    ]
    const { snapshots } = fold(events)
    expect(last(snapshots)).toEqual({
      content: [
        {
          type: 'tool-call',
          toolCallId: 'call-1',
          toolName: 'search_hotels',
          args: { city: 'Lisbon' },
          argsText: '{"city":"Lisbon"}',
          result: { hotels: 3 },
          isError: false,
          unstable_toolMessageId: MSG,
        },
        { type: 'text', text: 'Three hotels have rooms.' },
      ],
      status: { type: 'complete', reason: 'unknown' },
    })
    // The call shows while it runs, without a result.
    expect(snapshots[3].content).toEqual([
      {
        type: 'tool-call',
        toolCallId: 'call-1',
        toolName: 'search_hotels',
        args: { city: 'Lisbon' },
        argsText: '{"city":"Lisbon"}',
      },
    ])
    expect(snapshots).toEqual(aggregate(events))
  })

  it('places a render_ui card where its envelopes arrived, and hands each to the A2UI store in order', () => {
    const events = [
      ...started(),
      ...toolCall('call-ui', 'render_ui', { messages: [] }),
      toolResult('call-ui', {
        surface_id: 'card-plan',
        revision: 1,
        status: 'created',
      }),
      ...cardCreated('card-plan'),
      ...text('Here is the plan.'),
      finished(),
    ]
    const { snapshots, a2ui: applied } = fold(events)
    const reply = last(snapshots).content ?? []
    expect(reply.map((part) => part.type)).toEqual([
      'tool-call',
      'data',
      'data',
      'text',
    ])
    expect(reply.slice(1, 3)).toEqual(
      cardCreated('card-plan').map((event) => ({
        type: 'data',
        name: 'butter.a2ui',
        data: event.value,
      }))
    )
    expect(applied).toEqual(cardCreated('card-plan').map((e) => e.value))
    expect(snapshots).toEqual(aggregate(events))
  })

  it('ends on an interrupt as requires-action, with the Interrupts where the runtime reads them', () => {
    const form = {
      interruptId: 'int-1',
      token: 'tok-1',
      revision: 1,
      title: 'Hotel',
      question: 'Which hotel?',
      fields: [],
    }
    const events = [
      ...started(),
      ...text('Two hotels fit.'),
      a2ui(
        'form-1',
        'form',
        0,
        {
          createSurface: { surfaceId: 'form-1', catalogId: 'butter-basic-v1' },
        },
        form
      ),
      finished({ type: 'interrupt', interrupts: [QUESTION] }),
    ]
    const { snapshots, a2ui: applied, end } = fold(events)
    const reply = last(snapshots)
    expect(reply.status).toEqual({
      type: 'requires-action',
      reason: 'interrupt',
    })
    expect(reply.metadata).toEqual({
      custom: { agui: { interrupts: [QUESTION] } },
    })
    expect(reply.content?.map((part) => part.type)).toEqual(['text', 'data'])
    expect(applied).toHaveLength(1)
    expect(end).toEqual({ type: 'RUN_FINISHED' })
    expect(snapshots).toEqual(aggregate(events))
  })

  it('ends with a pending client tool call as requires-action on tool-calls', () => {
    const events = [
      ...started(),
      ...toolCall('call-2', 'confirm_booking', { hotel: 'Lumiares' }),
      finished(),
    ]
    const { snapshots } = fold(events)
    expect(last(snapshots)).toEqual({
      content: [
        {
          type: 'tool-call',
          toolCallId: 'call-2',
          toolName: 'confirm_booking',
          args: { hotel: 'Lumiares' },
          argsText: '{"hotel":"Lumiares"}',
        },
      ],
      status: { type: 'requires-action', reason: 'tool-calls' },
    })
    expect(snapshots).toEqual(aggregate(events))
  })

  it('ends a failed run incomplete, with its error, and keeps what it streamed', () => {
    const events = [
      ...started(),
      ...text('Day one: Alfama.'),
      { type: 'RUN_ERROR', message: 'model exploded', runId: RUN },
    ]
    const { snapshots, end } = fold(events)
    expect(last(snapshots)).toEqual({
      content: [{ type: 'text', text: 'Day one: Alfama.' }],
      status: { type: 'incomplete', reason: 'error', error: 'model exploded' },
    })
    expect(end).toEqual({ type: 'RUN_ERROR', message: 'model exploded' })
    expect(snapshots).toEqual(aggregate(events))
  })

  it('ends a stopped run with the stop code', () => {
    const events: AGUIEvent[] = [
      ...started(),
      ...text('Day one: Alfama.'),
      {
        type: 'RUN_ERROR',
        code: 'stopped',
        message: 'stopped by user',
        runId: RUN,
      },
    ]
    const { snapshots, end } = fold(events)
    expect(last(snapshots).status).toEqual({
      type: 'incomplete',
      reason: 'error',
      error: 'stopped by user',
    })
    expect(end).toEqual<RunEndEvent>({
      type: 'RUN_ERROR',
      code: 'stopped',
      message: 'stopped by user',
    })
    expect(snapshots).toEqual(aggregate(events))
  })

  it('moves the shared state as the run’s snapshot and deltas do, and the reply not at all', () => {
    const events: AGUIEvent[] = [
      ...started({ plan: 'draft', days: 2 }),
      {
        type: 'STATE_DELTA',
        delta: [
          { op: 'replace', path: '/plan', value: 'final' },
          { op: 'add', path: '/hotel', value: 'Lumiares' },
          { op: 'remove', path: '/days' },
        ],
      },
      { type: 'STATE_DELTA', delta: [] },
    ]
    const { snapshots, state } = fold(events, { state: { stale: true } })
    expect(state).toEqual({ plan: 'final', hotel: 'Lumiares' })
    // Only RUN_STARTED changed the reply.
    expect(snapshots).toEqual([{ content: [], status: { type: 'running' } }])
  })

  it('leaves the shared state as it was on a delta it cannot apply', () => {
    const { state } = fold(
      [
        ...started({ plan: { days: 2 } }),
        {
          type: 'STATE_DELTA',
          delta: [{ op: 'replace', path: '/plan/days', value: 3 }],
        },
        { type: 'STATE_DELTA', delta: [{ op: 'replace' }] },
      ],
      {}
    )
    expect(state).toEqual({ plan: { days: 2 } })
  })

  it('leaves out a result for a call of an earlier reply, which the thread already shows', () => {
    const earlier = [
      {
        id: 'a1',
        role: 'assistant',
        content: [
          {
            type: 'tool-call',
            toolCallId: 'call-0',
            toolName: 'confirm_booking',
            args: {},
            argsText: '{}',
            result: { approved: true },
          },
        ],
      },
    ] as unknown as ThreadMessage[]
    const events = [
      ...started(),
      toolResult('call-0', { approved: true }),
      ...text('Booked.'),
      finished(),
    ]
    const { snapshots } = fold(events, { earlier })
    expect(last(snapshots)).toEqual({
      content: [{ type: 'text', text: 'Booked.' }],
      status: { type: 'complete', reason: 'unknown' },
    })
    // A result for a call no reply holds becomes a call of this one, as the
    // aggregator makes it.
    expect(fold(events).snapshots).toEqual(aggregate(events))
  })

  it('folds what the AG-UI runtime would drop as it would: nothing', () => {
    const events: AGUIEvent[] = [
      { type: 'RUN_STARTED', threadId: THREAD },
      ...started(),
      { type: 'TEXT_MESSAGE_CONTENT', messageId: MSG, delta: '' },
      { type: 'TOOL_CALL_ARGS', delta: '{}' },
      { type: 'CUSTOM', value: 1 },
      { type: 'STEP_STARTED', stepName: 'plan' },
      // An interrupt outcome without a well-formed Interrupt is none.
      finished({ type: 'interrupt', interrupts: [{ message: 'no id' }] }),
    ]
    const { snapshots } = fold(events)
    expect(snapshots).toEqual([
      { content: [], status: { type: 'running' } },
      { content: [], status: { type: 'complete', reason: 'unknown' } },
    ])
    expect(snapshots).toEqual(aggregate(events))
  })
})
