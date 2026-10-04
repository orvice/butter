import {
  fromThreadMessageLike,
  type CompleteAttachment,
  type ExportedMessageRepository,
  type ThreadMessageLike,
} from '@assistant-ui/react'
import { fetchAGUIThreadHistory, fetchAGUIUISnapshot } from '@/api/agui'
import { ApiError } from '@/api/client'
import {
  A2UI_VERSION,
  EVENT_NAME,
  envelopeOp,
  type A2UIEventValue,
  type SnapshotSurface,
  type UISnapshot,
} from './a2ui/protocol'

// ThreadHistory is the body of GET /api/agui/:agent_id/threads/:thread_id/messages:
// a thread's conversation as AG-UI messages, its open Interrupts, and the
// reply each restorable surface belongs to.
export interface ThreadHistory {
  threadId: string
  messages: HistoryMessage[]
  interrupts: HistoryInterrupt[]
  surfaces: Array<{ surfaceId: string; messageId: string }>
}

export interface HistoryMessage {
  id: string
  role: string
  // content is the message's text. A user turn that carried images has AG-UI
  // content parts instead: its text and its images, in the order sent.
  content?: string | HistoryContentPart[]
  toolCalls?: Array<{
    id: string
    type: string
    function: { name: string; arguments: string }
  }>
  toolCallId?: string
}

// HistoryContentPart is one content part of a user turn: text, or an image
// inline as a data source, the shape a client sends.
export interface HistoryContentPart {
  type: string
  text?: string
  source?: { type: string; value?: string; mimeType?: string }
}

export interface HistoryInterrupt {
  id: string
  reason: string
  message?: string
}

// The AG-UI runtime's metadata namespace: it reads a reply's open
// Interrupts from metadata.custom.agui.interrupts.
const AGUI_METADATA_NAMESPACE = 'agui'

interface ToolCallPart {
  type: 'tool-call'
  toolCallId: string
  toolName: string
  args: Record<string, unknown>
  argsText: string
  result?: unknown
}

type AssistantPart =
  | ToolCallPart
  | { type: 'data'; name: string; data: A2UIEventValue }
  | { type: 'text'; text: string }

interface Draft {
  id: string
  role: 'user' | 'assistant'
  content: string | AssistantPart[]
  attachments?: CompleteAttachment[]
  status?: ThreadMessageLike['status']
  metadata?: ThreadMessageLike['metadata']
}

// threadRepository turns a thread's history into the messages the AG-UI
// runtime hydrates with, shaped as its live runs left them:
//   - a user turn's images are its attachments, as the composer sent them;
//   - a reply's tool calls carry their results;
//   - each restored surface sits in the reply that produced it, between the
//     tool calls and the text, where the live stream placed it;
//   - the last reply holds the open Interrupts, so prompts and forms answer
//     exactly as after a live run.
// placed lists the surfaces shown in a reply; the rest stay in the block of
// surfaces restored on their own.
export function threadRepository(
  history: ThreadHistory,
  snapshot: UISnapshot
): { repository: ExportedMessageRepository; placed: Set<string> } {
  const surfaces = new Map(
    (snapshot.surfaces ?? []).map((s) => [s.surfaceId, s] as const)
  )
  const byReply = new Map<string, SnapshotSurface[]>()
  for (const { surfaceId, messageId } of history.surfaces ?? []) {
    const surface = surfaces.get(surfaceId)
    if (
      !surface ||
      envelopeOp(surface.envelopes[0] ?? {}) !== 'createSurface'
    ) {
      continue
    }
    byReply.set(messageId, [...(byReply.get(messageId) ?? []), surface])
  }

  const placed = new Set<string>()
  const toolCalls = new Map<string, ToolCallPart>()
  const drafts: Draft[] = []
  for (const message of history.messages ?? []) {
    switch (message.role) {
      case 'user':
        drafts.push({ id: message.id, role: 'user', ...userTurn(message) })
        break
      case 'tool': {
        const call = toolCalls.get(message.toolCallId ?? '')
        if (call) call.result = parseJSON(textOf(message.content))
        break
      }
      case 'assistant': {
        const parts: AssistantPart[] = []
        for (const call of message.toolCalls ?? []) {
          const part: ToolCallPart = {
            type: 'tool-call',
            toolCallId: call.id,
            toolName: call.function.name,
            args: parseArgs(call.function.arguments),
            argsText: call.function.arguments,
          }
          toolCalls.set(call.id, part)
          parts.push(part)
        }
        for (const surface of byReply.get(message.id) ?? []) {
          placed.add(surface.surfaceId)
          parts.push({
            type: 'data',
            name: EVENT_NAME,
            data: createEvent(surface),
          })
        }
        const text = textOf(message.content)
        if (text) parts.push({ type: 'text', text })
        drafts.push({
          id: message.id,
          role: 'assistant',
          content: parts,
          status: { type: 'complete', reason: 'stop' },
        })
        break
      }
    }
  }

  if ((history.interrupts ?? []).length > 0) {
    let reply: Draft | undefined
    for (let i = drafts.length - 1; i >= 0 && !reply; i--) {
      if (drafts[i].role === 'assistant') reply = drafts[i]
    }
    if (!reply) {
      reply = {
        id: `${history.threadId}:interrupts`,
        role: 'assistant',
        content: [],
      }
      drafts.push(reply)
    }
    reply.status = { type: 'requires-action', reason: 'interrupt' }
    reply.metadata = {
      custom: { [AGUI_METADATA_NAMESPACE]: { interrupts: history.interrupts } },
    }
  }

  let parentId: string | null = null
  const messages: ExportedMessageRepository['messages'] = []
  for (const draft of drafts) {
    const message = fromThreadMessageLike(
      draft as ThreadMessageLike,
      draft.id,
      { type: 'complete', reason: 'unknown' }
    )
    messages.push({ message, parentId })
    parentId = message.id
  }
  return { repository: { headId: parentId, messages }, placed }
}

// userTurn is a user turn as the composer sends one: its text, and its images
// as attachments, so a reload shows the turn as it showed when it was sent.
// The server keeps no file names, so the images are numbered.
function userTurn(
  message: HistoryMessage
): Pick<Draft, 'content' | 'attachments'> {
  if (!Array.isArray(message.content)) return { content: message.content ?? '' }
  const attachments: CompleteAttachment[] = []
  for (const part of message.content) {
    const source = part.source
    if (
      part.type !== 'image' ||
      source?.type !== 'data' ||
      !source.value ||
      !source.mimeType
    ) {
      continue
    }
    const n = attachments.length + 1
    attachments.push({
      id: `${message.id}:image-${n}`,
      type: 'image',
      name: `Image ${n}`,
      contentType: source.mimeType,
      status: { type: 'complete' },
      content: [
        {
          type: 'image',
          image: `data:${source.mimeType};base64,${source.value}`,
        },
      ],
    })
  }
  // A turn of images alone has no text to show.
  const text = textOf(message.content)
  return { content: text ? text : [], attachments }
}

// textOf is the text of a message's content. Text parts read as the server
// writes a turn of text alone: one paragraph each.
function textOf(content: HistoryMessage['content']): string {
  if (!Array.isArray(content)) return content ?? ''
  return content
    .flatMap((part) => (part.type === 'text' && part.text ? [part.text] : []))
    .join('\n\n')
}

// createEvent is the butter.a2ui event that created surface, which is what a
// reply carries to show it.
function createEvent(surface: SnapshotSurface): A2UIEventValue {
  return {
    version: A2UI_VERSION,
    surfaceId: surface.surfaceId,
    kind: surface.kind,
    revision: surface.revision,
    seq: 0,
    envelope: surface.envelopes[0],
    fallback: surface.fallback,
    form: surface.form,
  }
}

function parseArgs(text: string): Record<string, unknown> {
  const value = parseJSON(text)
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {}
}

function parseJSON(text: string): unknown {
  try {
    return JSON.parse(text)
  } catch {
    return text
  }
}

// loadThread reads a thread's history and its UI snapshot together. A run
// holding the thread answers 409, so a busy read is retried while keepTrying
// holds; any other failure is final, and the page offers Retry. A server
// without the history endpoint (404) restores the surfaces alone.
export async function loadThread(
  agentId: string,
  threadId: string,
  keepTrying: () => boolean
): Promise<{ history: ThreadHistory; snapshot: UISnapshot }> {
  for (let attempt = 0; ; attempt++) {
    try {
      const [history, snapshot] = await Promise.all([
        fetchAGUIThreadHistory<ThreadHistory>(agentId, threadId).catch(
          (err: unknown) => {
            if (err instanceof ApiError && err.code === '404') {
              return emptyHistory(threadId)
            }
            throw err
          }
        ),
        fetchAGUIUISnapshot<UISnapshot>(agentId, threadId),
      ])
      return { history, snapshot }
    } catch (err) {
      if (!threadBusy(err) || attempt >= 5 || !keepTrying()) throw err
      await new Promise((resolve) => setTimeout(resolve, 500 * 2 ** attempt))
    }
  }
}

// threadBusy reports a read refused only for the moment: a run holds the
// thread (409), or its lease cannot be taken right now (503).
function threadBusy(err: unknown): boolean {
  return err instanceof ApiError && (err.code === '409' || err.code === '503')
}

function emptyHistory(threadId: string): ThreadHistory {
  return { threadId, messages: [], interrupts: [], surfaces: [] }
}
