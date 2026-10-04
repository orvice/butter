import type {
  ChatModelRunResult,
  MessageStatus,
  ThreadAssistantMessagePart,
  ThreadMessage,
  ToolCallMessagePart,
} from '@assistant-ui/react'
import type { AgUiInterrupt } from '@assistant-ui/react-ag-ui'
import { applyAGUIStateDelta, type AGUIEvent } from '@/api/agui'
import { EVENT_NAME } from './a2ui/protocol'
import type { RunErrorEvent } from './errors'
import { AGUI_METADATA_NAMESPACE, parseJSON } from './history'

// A run the page follows from its log (RunFollower) reaches the AG-UI runtime
// through the history adapter's resume(), and resume() hands the runtime
// snapshots of the run's reply, not AG-UI events (ADR-0016 decision 8). On
// that path the runtime parses no event, and the aggregator it folds a run's
// events with is not exported. RunFold folds them itself, as that aggregator
// does in the release the dashboard pins (@assistant-ui/react-ag-ui 0.0.63,
// src/runtime/adapter/run-aggregator.ts), and applies what the runtime would
// apply besides the reply: butter.a2ui to the A2UI store, STATE_SNAPSHOT and
// STATE_DELTA to the shared state (RunEffects).
//
// It folds what Butter's server sends (docs/api.md "AG-UI Protocol"):
// RUN_STARTED, TEXT_MESSAGE_*, TOOL_CALL_*, TOOL_CALL_RESULT, STATE_*, CUSTOM,
// RUN_FINISHED and RUN_ERROR. Like the aggregator it ignores what it does not
// know, and the aggregator's reasoning, activity, MCP and subagent events
// never come. Three things the aggregator does are left out:
//   - one reply: the server gives a run one message ID, so the aggregator's
//     split of a run into several replies never happens;
//   - no timing: a replay's would be the replay's, not the run's;
//   - no tool approvals: Butter's Interrupts are human_input, never the
//     tool_call gates the aggregator projects onto tool calls.

// RunEffects is what a followed run's events change besides its reply, which
// the runtime applies only for the runs it streams itself.
export interface RunEffects {
  // a2ui takes the value of each butter.a2ui CUSTOM event, in order, as the
  // A2UI store takes them from a run the page streams (A2UIStore.apply).
  a2ui(value: unknown): void
  // state moves the thread's shared state: update takes the state as it
  // stands and returns the next one.
  state(update: (prev: unknown) => unknown): void
}

// RunEndEvent is the event that ended a followed run: RUN_FINISHED, or
// RUN_ERROR, which carries the stop code when a person stopped the run.
export type RunEndEvent =
  { type: 'RUN_FINISHED' } | ({ type: 'RUN_ERROR' } & RunErrorEvent)

// Part is one part of the reply, in the order the run produced it.
type Part =
  | { kind: 'text'; key: string }
  | { kind: 'tool-call'; toolCallId: string }
  | { kind: 'data'; name: string; value: unknown }

interface ToolCall {
  toolCallId: string
  toolName: string
  argsText: string
  args: Record<string, unknown> | undefined
  result: unknown
  isError: boolean | undefined
  parentMessageId?: string
  toolMessageId?: string
}

// RunFold folds one run's events into snapshots of its reply. earlier holds
// the tool calls of the replies before it (toolCallIdsOf).
export class RunFold {
  private readonly effects: RunEffects
  private readonly earlier: ReadonlySet<string>
  private status: MessageStatus | undefined
  private interrupts: AgUiInterrupt[] | undefined
  private readonly parts: Part[] = []
  private readonly texts = new Map<string, string>()
  // activeText is the text part an event without a message ID adds to.
  private activeText: string | undefined
  private readonly toolCalls = new Map<string, ToolCall>()
  // textPartCount numbers the text parts that arrive without a message ID.
  private textPartCount = 0
  private endEvent: RunEndEvent | undefined

  constructor(effects: RunEffects, earlier: ReadonlySet<string> = new Set()) {
    this.effects = effects
    this.earlier = earlier
  }

  // end is the event that ended the run, once one was folded.
  get end(): RunEndEvent | undefined {
    return this.endEvent
  }

  // handle folds one event, and returns the reply as it now stands, or
  // undefined when the event leaves the reply as it was.
  handle(event: AGUIEvent): ChatModelRunResult | undefined {
    switch (event.type) {
      case 'RUN_STARTED':
        if (!stringOf(event.runId)) return undefined
        this.parts.length = 0
        this.texts.clear()
        this.toolCalls.clear()
        this.activeText = undefined
        this.textPartCount = 0
        this.interrupts = undefined
        this.endEvent = undefined
        this.status = { type: 'running' }
        return this.reply()
      case 'RUN_FINISHED': {
        if (!stringOf(event.runId)) return undefined
        this.endEvent = { type: 'RUN_FINISHED' }
        const interrupts = outcomeInterrupts(event.outcome)
        if (interrupts) {
          this.interrupts = interrupts
          this.status = { type: 'requires-action', reason: 'interrupt' }
          return this.reply()
        }
        this.interrupts = undefined
        // A call the run left without a result is a client tool's, which
        // the client answers (docs/api.md "Frontend tools").
        const pending = [...this.toolCalls.values()].some(
          (call) => call.result === undefined
        )
        this.status = pending
          ? { type: 'requires-action', reason: 'tool-calls' }
          : { type: 'complete', reason: 'unknown' }
        return this.reply()
      }
      case 'RUN_ERROR': {
        const message = stringOf(event.message)
        const code = stringOf(event.code)
        this.endEvent = {
          type: 'RUN_ERROR',
          ...(message !== undefined ? { message } : {}),
          ...(code !== undefined ? { code } : {}),
        }
        this.status = {
          type: 'incomplete',
          reason: 'error',
          ...(message !== undefined ? { error: message } : {}),
        }
        return this.reply()
      }
      case 'TEXT_MESSAGE_START':
        this.activeText = this.textPart(
          stringOf(event.messageId) ?? this.newTextKey()
        )
        return this.reply()
      case 'TEXT_MESSAGE_CONTENT': {
        const delta = stringOf(event.delta)
        if (!delta) return undefined
        const messageId = stringOf(event.messageId)
        const key = messageId
          ? this.textPart(messageId)
          : (this.activeText ?? this.textPart(this.newTextKey()))
        this.activeText = key
        this.texts.set(key, (this.texts.get(key) ?? '') + delta)
        return this.reply()
      }
      case 'TEXT_MESSAGE_END': {
        const messageId = stringOf(event.messageId)
        if (messageId && this.activeText === messageId) {
          this.activeText = undefined
        }
        return this.reply()
      }
      case 'CUSTOM': {
        const name = stringOf(event.name)
        if (!name) return undefined
        if (name === EVENT_NAME) this.effects.a2ui(event.value)
        this.activeText = undefined
        this.parts.push({ kind: 'data', name, value: event.value })
        return this.reply()
      }
      case 'TOOL_CALL_START': {
        const id = stringOf(event.toolCallId)
        if (!id) return undefined
        // A tool call ends the text before it: text after it is a new part.
        this.activeText = undefined
        const parentMessageId = stringOf(event.parentMessageId)
        if (!this.hasToolCallPart(id)) this.placeToolCall(id, parentMessageId)
        this.toolCalls.set(
          id,
          newToolCall(id, stringOf(event.toolCallName), parentMessageId)
        )
        return this.reply()
      }
      case 'TOOL_CALL_ARGS': {
        const id = stringOf(event.toolCallId)
        const delta = stringOf(event.delta)
        if (!id || !delta) return undefined
        const call = this.toolCalls.get(id)
        if (call) {
          call.argsText += delta
          call.args = argsOf(call.argsText)
        }
        return this.reply()
      }
      case 'TOOL_CALL_END':
        return stringOf(event.toolCallId) ? this.reply() : undefined
      case 'TOOL_CALL_RESULT': {
        const id = stringOf(event.toolCallId)
        if (!id) return undefined
        // A result for a call of an earlier reply: the runtime applies it to
        // that reply, which the thread's read already shows with its result.
        if (!this.toolCalls.has(id) && this.earlier.has(id)) return undefined
        let call = this.toolCalls.get(id)
        if (!call) {
          call = newToolCall(id)
          this.toolCalls.set(id, call)
        }
        if (!this.hasToolCallPart(id)) {
          this.parts.push({ kind: 'tool-call', toolCallId: id })
        }
        call.result = parseJSON(stringOf(event.content) ?? '')
        call.isError = event.role === 'tool' ? false : undefined
        const toolMessageId = stringOf(event.messageId)
        if (toolMessageId) call.toolMessageId = toolMessageId
        return this.reply()
      }
      case 'STATE_SNAPSHOT': {
        if (event.snapshot === undefined) return undefined
        const snapshot = event.snapshot
        this.effects.state(() => snapshot)
        return undefined
      }
      case 'STATE_DELTA': {
        const ops = Array.isArray(event.delta) ? event.delta : []
        if (ops.length === 0) return undefined
        // The server sends top-level add, replace and remove. A delta past
        // that leaves the state as it was, as the runtime leaves it on a
        // patch it cannot apply; the next run's STATE_SNAPSHOT corrects the
        // mirror that run sends.
        this.effects.state((prev) => {
          try {
            return (
              applyAGUIStateDelta(isPlainObject(prev) ? prev : {}, ops) ?? prev
            )
          } catch {
            return prev
          }
        })
        return undefined
      }
      default:
        return undefined
    }
  }

  private newTextKey(): string {
    this.textPartCount += 1
    return `text-${this.textPartCount}`
  }

  // textPart is the text part key names, added at the end of the reply when
  // it is new.
  private textPart(key: string): string {
    if (!this.texts.has(key)) {
      this.texts.set(key, '')
      this.parts.push({ kind: 'text', key })
    }
    return key
  }

  private hasToolCallPart(id: string): boolean {
    return this.parts.some(
      (part) => part.kind === 'tool-call' && part.toolCallId === id
    )
  }

  // placeToolCall adds a tool call's part: after its parent message's text,
  // and the calls already placed there, when the run names one, else at the
  // end of the reply.
  private placeToolCall(id: string, parentMessageId: string | undefined) {
    const part: Part = { kind: 'tool-call', toolCallId: id }
    const parent = parentMessageId
      ? this.parts.findIndex(
          (p) => p.kind === 'text' && p.key === parentMessageId
        )
      : -1
    if (parent === -1) {
      this.parts.push(part)
      return
    }
    let at = parent + 1
    while (at < this.parts.length) {
      const next = this.parts[at]
      if (
        next.kind !== 'tool-call' ||
        this.toolCalls.get(next.toolCallId)?.parentMessageId !== parentMessageId
      ) {
        break
      }
      at++
    }
    this.parts.splice(at, 0, part)
  }

  // reply is the reply as the run left it so far: its parts in order, text
  // that is not blank, its status, and the Interrupts it ended on, where the
  // runtime reads them (metadata.custom.agui.interrupts).
  private reply(): ChatModelRunResult {
    const content: ThreadAssistantMessagePart[] = []
    for (const part of this.parts) {
      switch (part.kind) {
        case 'text': {
          const text = this.texts.get(part.key) ?? ''
          if (text.trim().length > 0) content.push({ type: 'text', text })
          break
        }
        case 'tool-call': {
          const call = this.toolCalls.get(part.toolCallId)
          if (call) content.push(toolCallPart(call))
          break
        }
        case 'data':
          content.push({ type: 'data', name: part.name, data: part.value })
          break
      }
    }
    return {
      content,
      ...(this.status ? { status: this.status } : {}),
      ...(this.interrupts
        ? {
            metadata: {
              custom: {
                [AGUI_METADATA_NAMESPACE]: { interrupts: this.interrupts },
              },
            },
          }
        : {}),
    }
  }
}

// toolCallIdsOf lists the tool calls of a thread's replies.
export function toolCallIdsOf(messages: readonly ThreadMessage[]): Set<string> {
  const ids = new Set<string>()
  for (const message of messages) {
    if (message.role !== 'assistant') continue
    for (const part of message.content) {
      if (part.type === 'tool-call') ids.add(part.toolCallId)
    }
  }
  return ids
}

function toolCallPart(call: ToolCall): ThreadAssistantMessagePart {
  return {
    type: 'tool-call',
    toolCallId: call.toolCallId,
    toolName: call.toolName,
    args: (call.args ?? {}) as ToolCallMessagePart['args'],
    argsText: call.argsText,
    ...(call.result !== undefined ? { result: call.result } : {}),
    ...(call.isError !== undefined ? { isError: call.isError } : {}),
    ...(call.parentMessageId ? { parentId: call.parentMessageId } : {}),
    ...(call.toolMessageId
      ? { unstable_toolMessageId: call.toolMessageId }
      : {}),
  } as ToolCallMessagePart & { unstable_toolMessageId?: string }
}

// outcomeInterrupts is the Interrupts of a RUN_FINISHED interrupt outcome.
// An outcome without a well-formed Interrupt counts as none, as the AG-UI
// runtime reads it.
function outcomeInterrupts(outcome: unknown): AgUiInterrupt[] | undefined {
  if (
    !isPlainObject(outcome) ||
    outcome.type !== 'interrupt' ||
    !Array.isArray(outcome.interrupts)
  ) {
    return undefined
  }
  const interrupts = outcome.interrupts.flatMap(interruptOf)
  return interrupts.length > 0 ? interrupts : undefined
}

// interruptOf keeps what the AG-UI runtime keeps of an Interrupt.
function interruptOf(raw: unknown): AgUiInterrupt[] {
  if (
    !isPlainObject(raw) ||
    typeof raw.id !== 'string' ||
    typeof raw.reason !== 'string'
  ) {
    return []
  }
  const interrupt: AgUiInterrupt = { id: raw.id, reason: raw.reason }
  if (typeof raw.message === 'string') interrupt.message = raw.message
  if (typeof raw.toolCallId === 'string') interrupt.toolCallId = raw.toolCallId
  if (typeof raw.expiresAt === 'string') interrupt.expiresAt = raw.expiresAt
  if (isPlainObject(raw.responseSchema)) {
    interrupt.responseSchema = raw.responseSchema
  }
  if (isPlainObject(raw.metadata)) interrupt.metadata = raw.metadata
  return [interrupt]
}

// argsOf is a tool call's arguments once their text is a whole JSON object
// or array, as far as it arrived. The server sends a call's arguments whole,
// in one delta.
function argsOf(text: string): Record<string, unknown> | undefined {
  try {
    const parsed: unknown = JSON.parse(text)
    return parsed && typeof parsed === 'object'
      ? (parsed as Record<string, unknown>)
      : undefined
  } catch {
    return undefined
  }
}

// newToolCall is a tool call as it starts, with no arguments or result yet.
// A call with no name is a "tool", as the aggregator names it.
function newToolCall(
  id: string,
  name = 'tool',
  parentMessageId?: string
): ToolCall {
  return {
    toolCallId: id,
    toolName: name,
    argsText: '',
    args: undefined,
    result: undefined,
    isError: undefined,
    ...(parentMessageId ? { parentMessageId } : {}),
  }
}

function stringOf(value: unknown): string | undefined {
  return typeof value === 'string' ? value : undefined
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === 'object' && !Array.isArray(value)
}
