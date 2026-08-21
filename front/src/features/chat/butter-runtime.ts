import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  InvocationStatus,
  type Invocation,
} from '@/gen/agents/v1/agent_service_pb'
import {
  useExternalStoreRuntime,
  type AppendMessage,
  type ThreadMessageLike,
} from '@assistant-ui/react'
import { toast } from 'sonner'
import {
  cancelAgentInvocation,
  getActiveInvocationForSession,
  getAgentInvocation,
  getAgentInvocationInput,
  getLatestInvocationForSession,
  isTerminalInvocationStatus,
  submitAgentInvocation,
  watchChatInvocation,
  type ChatStreamPayload,
} from '@/api/chat'
import { useLiveSession } from '@/api/sessions'
import { newClientID } from '@/lib/client-id'
import { CHAT_APP_NAME } from '@/lib/constants'
import {
  buildInputParts,
  decodeInputParts,
  type InputPartInit,
} from '@/lib/image-attachments'
import {
  parseSessionEvent,
  parseSessionEvents,
  type ParsedEvent,
} from '@/lib/session-events'
import {
  convertSessionEvents,
  makeOptimisticUserMessage,
  makeStreamingAssistantMessage,
} from './message-converter'

interface ChatRunState {
  runId: string | null
  sessionId: string
  pending: boolean
  pendingBaseEventIds: Set<string> | null
  pendingUserMessage: string | null
  streamingEvents: ParsedEvent[]
  streamingResponse: string
  invocationId: string | null
}

export interface TerminalNotice {
  sessionId: string
  invocationId: string
  status: 'failed' | 'stopped'
  error: string
  input: string
}

function emptyChatRunState(sessionId: string): ChatRunState {
  return {
    runId: null,
    sessionId,
    pending: false,
    pendingBaseEventIds: null,
    pendingUserMessage: null,
    streamingEvents: [],
    streamingResponse: '',
    invocationId: null,
  }
}

function updateChatRun(
  prev: ChatRunState,
  sessionId: string,
  runId: string,
  update: (current: ChatRunState) => ChatRunState
): ChatRunState {
  if (prev.sessionId !== sessionId || prev.runId !== runId) return prev
  return update(prev)
}

function terminalNoticeFor(
  inv: Invocation | null,
  sessionId: string
): TerminalNotice | null {
  if (!inv?.id) return null
  if (inv.status === InvocationStatus.FAILED) {
    return {
      sessionId,
      invocationId: inv.id,
      status: 'failed',
      error: inv.error,
      input: inv.input,
    }
  }
  if (inv.status === InvocationStatus.CANCELLED) {
    return {
      sessionId,
      invocationId: inv.id,
      status: 'stopped',
      error: '',
      input: inv.input,
    }
  }
  return null
}

function isActiveInvocation(status: InvocationStatus): boolean {
  return (
    status === InvocationStatus.QUEUED || status === InvocationStatus.RUNNING
  )
}

function payloadToParsedEvent(payload: ChatStreamPayload): ParsedEvent | null {
  const evt = payload.event
  if (!evt?.event_id) return null
  const parsed = parseSessionEvent({
    event_id: evt.event_id,
    invocation_id: evt.invocation_id,
    author: evt.author,
    branch: evt.branch,
    content_json: evt.content_json,
    timestamp: evt.timestamp,
    trace_id: evt.invocation_id,
  })
  if (evt.partial && parsed.text) {
    return { ...parsed, text: '' }
  }
  return parsed
}

function isAbortError(err: unknown): boolean {
  return err instanceof DOMException && err.name === 'AbortError'
}

function appendUniqueEvent(
  out: ParsedEvent[],
  seen: Set<string>,
  event: ParsedEvent
) {
  if (seen.has(event.eventId)) return
  seen.add(event.eventId)
  out.push(event)
}

function makeOptimisticTextEvent(
  eventId: string,
  role: ParsedEvent['role'],
  text: string
): ParsedEvent {
  const author = role === 'assistant' ? 'agent' : role
  return {
    eventId,
    author,
    role,
    text,
    toolCalls: [],
    toolResponses: [],
    timestamp: new Date().toISOString(),
    raw: {
      event_id: eventId,
      author,
      timestamp: new Date().toISOString(),
      content_json: JSON.stringify({ role, parts: [{ text }] }),
    },
  }
}

export interface UseButterRuntimeOptions {
  sessionId: string
  userId: string
  agentId: string | null
  onInvocationAccepted?: (invocationId: string, message: string) => void
  pendingMessage?: string
  initialInvocationId?: string
  attachmentsRef?: React.RefObject<File[]>
  validateForSend?: () => string[]
  clearAttachments?: () => void
  addFiles?: (files: File[]) => void
}

export interface UseButterRuntimeResult {
  runtime: ReturnType<typeof useExternalStoreRuntime>
  isRunning: boolean
  invocationId: string | null
  notice: TerminalNotice | null
  persistedEvents: ParsedEvent[]
  liveQuery: ReturnType<typeof useLiveSession>
  restoreInput: () => Promise<string>
}

export function useButterRuntime({
  sessionId,
  userId,
  agentId,
  onInvocationAccepted,
  pendingMessage,
  initialInvocationId,
  attachmentsRef,
  validateForSend,
  clearAttachments,
  addFiles,
}: UseButterRuntimeOptions): UseButterRuntimeResult {
  const [runState, setRunState] = useState<ChatRunState>(() =>
    emptyChatRunState('')
  )
  const [notice, setNotice] = useState<TerminalNotice | null>(null)
  const sendingRef = useRef(false)
  const observationRef = useRef(0)
  const observeAbortRef = useRef<AbortController | null>(null)

  const isRunForCurrentSession = runState.sessionId === sessionId
  const pending = isRunForCurrentSession && runState.pending
  const pendingBaseEventIds = isRunForCurrentSession
    ? runState.pendingBaseEventIds
    : null
  const pendingUserMessage = isRunForCurrentSession
    ? runState.pendingUserMessage
    : null
  const streamingEvents = isRunForCurrentSession
    ? runState.streamingEvents
    : EMPTY_STREAMING_EVENTS
  const streamingResponse = isRunForCurrentSession
    ? runState.streamingResponse
    : ''
  const invocationId = isRunForCurrentSession ? runState.invocationId : null
  const activeNotice = notice && notice.sessionId === sessionId ? notice : null

  const liveQuery = useLiveSession(
    CHAT_APP_NAME,
    userId,
    sessionId,
    pending,
    15000
  )

  const persistedEvents = useMemo<ParsedEvent[]>(
    () => parseSessionEvents(liveQuery.data?.session_detail.events),
    [liveQuery.data]
  )

  const composedEvents = useMemo<ParsedEvent[]>(() => {
    const out: ParsedEvent[] = []
    const seen = new Set<string>()
    const baseEvents = pendingBaseEventIds
      ? persistedEvents.filter((evt) => pendingBaseEventIds.has(evt.eventId))
      : persistedEvents

    for (const event of baseEvents) appendUniqueEvent(out, seen, event)
    if (pendingUserMessage) {
      appendUniqueEvent(
        out,
        seen,
        makeOptimisticTextEvent('pending-user', 'user', pendingUserMessage)
      )
    }
    for (const event of streamingEvents) appendUniqueEvent(out, seen, event)
    if (streamingResponse.trim()) {
      appendUniqueEvent(
        out,
        seen,
        makeOptimisticTextEvent(
          'streaming-assistant',
          'assistant',
          streamingResponse
        )
      )
    }
    return out
  }, [
    persistedEvents,
    pendingBaseEventIds,
    pendingUserMessage,
    streamingEvents,
    streamingResponse,
  ])

  const messages = useMemo<ThreadMessageLike[]>(() => {
    const converted = convertSessionEvents(composedEvents)

    if (
      pendingUserMessage &&
      !composedEvents.some((e) => e.eventId === 'pending-user')
    ) {
      converted.push(
        makeOptimisticUserMessage('pending-user', pendingUserMessage)
      )
    }
    if (
      streamingResponse.trim() &&
      !composedEvents.some((e) => e.eventId === 'streaming-assistant')
    ) {
      converted.push(
        makeStreamingAssistantMessage('streaming-assistant', streamingResponse)
      )
    }
    return converted
  }, [composedEvents, pendingUserMessage, streamingResponse])

  const observeInvocation = useCallback(
    async (
      id: string,
      observedSessionId: string,
      runId: string,
      _retryText: string
    ) => {
      const token = ++observationRef.current
      observeAbortRef.current?.abort()
      const controller = new AbortController()
      observeAbortRef.current = controller
      const detached = () => token !== observationRef.current

      const finish = async (inv: Invocation) => {
        await liveQuery.refetch()
        if (detached()) return
        setRunState((prev) =>
          prev.runId === runId ? emptyChatRunState(observedSessionId) : prev
        )
        setNotice(terminalNoticeFor(inv, observedSessionId))
        if (inv.status === InvocationStatus.CANCELLED) {
          toast.info('Chat stopped')
        } else if (inv.status === InvocationStatus.FAILED) {
          toast.error(inv.error || 'Agent invocation failed')
        }
      }

      try {
        for (;;) {
          if (detached()) return
          const inv = await getAgentInvocation(id)
          if (detached()) return
          if (!isActiveInvocation(inv.status)) {
            await finish(inv)
            return
          }
          const terminal = await watchChatInvocation(
            id,
            {
              onAgentEvent: (payload) => {
                if (detached()) return
                const event = payloadToParsedEvent(payload)
                if (event) {
                  setRunState((prev) =>
                    updateChatRun(
                      prev,
                      observedSessionId,
                      runId,
                      (current) => ({
                        ...current,
                        streamingEvents: [...current.streamingEvents, event],
                      })
                    )
                  )
                }
              },
              onTextDelta: (payload) => {
                if (detached() || !payload.text_delta) return
                setRunState((prev) =>
                  updateChatRun(prev, observedSessionId, runId, (current) => ({
                    ...current,
                    streamingResponse:
                      current.streamingResponse + payload.text_delta,
                  }))
                )
              },
            },
            controller.signal
          )
          if (detached()) return
          if (terminal) {
            await finish(terminal)
            return
          }
          await liveQuery.refetch()
          if (detached()) return
          setRunState((prev) =>
            updateChatRun(prev, observedSessionId, runId, (current) => ({
              ...current,
              pendingBaseEventIds: null,
              streamingEvents: [],
              streamingResponse: '',
            }))
          )
        }
      } catch (err) {
        if (isAbortError(err) || detached()) return
        setRunState((prev) =>
          prev.runId === runId ? emptyChatRunState(observedSessionId) : prev
        )
        toast.error(
          err instanceof Error ? err.message : 'Failed to observe invocation'
        )
      }
    },
    // liveQuery.refetch is stable across renders
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  )

  // Observe an Invocation accepted by the new-chat composer.
  useEffect(() => {
    observationRef.current += 1
    observeAbortRef.current?.abort()
    if (!initialInvocationId || !sessionId) return

    const runId = newClientID()
    const frame = globalThis.requestAnimationFrame(() => {
      setRunState({
        runId,
        sessionId,
        pending: true,
        pendingBaseEventIds: new Set(persistedEvents.map((evt) => evt.eventId)),
        pendingUserMessage: pendingMessage?.trim() || null,
        streamingEvents: [],
        streamingResponse: '',
        invocationId: initialInvocationId,
      })
      void observeInvocation(
        initialInvocationId,
        sessionId,
        runId,
        pendingMessage?.trim() || ''
      )
    })

    return () => {
      globalThis.cancelAnimationFrame(frame)
      observationRef.current += 1
      observeAbortRef.current?.abort()
    }
    // persistedEvents captured at acceptance time only.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialInvocationId, sessionId])

  // Re-entering a session with a run still in flight.
  useEffect(() => {
    if (!sessionId || initialInvocationId) return
    let stale = false
    void (async () => {
      let active: Invocation | null
      try {
        active = await getActiveInvocationForSession(sessionId)
      } catch {
        return
      }
      if (stale) return
      if (!active?.id || isTerminalInvocationStatus(active.status)) {
        try {
          const latest = await getLatestInvocationForSession(sessionId)
          if (stale || sendingRef.current) return
          setNotice(terminalNoticeFor(latest, sessionId))
        } catch {
          /* lookup unavailable */
        }
        return
      }
      if (sendingRef.current) return
      await liveQuery.refetch()
      if (stale || sendingRef.current) return
      const runId = newClientID()
      setRunState({
        runId,
        sessionId,
        pending: true,
        pendingBaseEventIds: null,
        pendingUserMessage: null,
        streamingEvents: [],
        streamingResponse: '',
        invocationId: active.id,
      })
      void observeInvocation(active.id, sessionId, runId, '')
    })()
    return () => {
      stale = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sessionId, initialInvocationId])

  const onNew = useCallback(
    async (message: AppendMessage) => {
      const text =
        message.content
          .filter((p): p is { type: 'text'; text: string } => p.type === 'text')
          .map((p) => p.text)
          .join('\n')
          .trim() || ''

      const images = attachmentsRef?.current ?? []
      if (
        (!text && images.length === 0) ||
        !agentId ||
        pending ||
        sendingRef.current
      )
        return

      if (images.length > 0 && validateForSend) {
        const errors = validateForSend()
        if (errors.length > 0) {
          errors.forEach((msg) => toast.error(msg))
          return
        }
      }

      sendingRef.current = true

      let parts: InputPartInit[] | undefined
      try {
        parts =
          images.length > 0 ? await buildInputParts(text, images) : undefined
      } catch {
        toast.error('Failed to read attached images')
        sendingRef.current = false
        return
      }

      const displayMessage =
        text || `(${images.length} image${images.length > 1 ? 's' : ''})`

      const runId = newClientID()
      setRunState({
        runId,
        sessionId,
        pending: true,
        pendingBaseEventIds: new Set(persistedEvents.map((evt) => evt.eventId)),
        pendingUserMessage: displayMessage,
        streamingEvents: [],
        streamingResponse: '',
        invocationId: null,
      })

      try {
        const accepted = await submitAgentInvocation({
          request_id: newClientID(),
          agent_id: agentId,
          session_id: sessionId,
          message: text,
          parts,
        })
        setRunState((prev) =>
          updateChatRun(prev, sessionId, runId, (current) => ({
            ...current,
            invocationId: accepted.invocation_id,
          }))
        )
        setNotice(null)
        clearAttachments?.()
        sendingRef.current = false
        if (onInvocationAccepted) {
          onInvocationAccepted(accepted.invocation_id, text)
        } else {
          void observeInvocation(accepted.invocation_id, sessionId, runId, text)
        }
      } catch (err) {
        sendingRef.current = false
        setRunState((prev) =>
          prev.runId === runId ? emptyChatRunState(prev.sessionId) : prev
        )
        toast.error(
          err instanceof Error ? err.message : 'Failed to send message'
        )
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [agentId, sessionId, pending, persistedEvents, onInvocationAccepted]
  )

  const onCancel = useCallback(async () => {
    const id = invocationId
    if (id) {
      try {
        const result = await cancelAgentInvocation(id)
        if (!result.cancelled) {
          await observeInvocation(
            id,
            sessionId,
            runState.runId ?? newClientID(),
            pendingUserMessage ?? ''
          )
        }
      } catch (err) {
        toast.error(
          err instanceof Error ? err.message : 'Failed to cancel invocation'
        )
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [invocationId, sessionId, runState.runId, pendingUserMessage])

  const restoreInput = useCallback(async (): Promise<string> => {
    if (!activeNotice) return ''
    try {
      const { invocation: inv, parts } = await getAgentInvocationInput(
        activeNotice.invocationId
      )
      const { text: restoredText, files } = decodeInputParts(parts)
      clearAttachments?.()
      if (files.length > 0) addFiles?.(files)
      return restoredText || inv.input
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Failed to restore input'
      )
      return ''
    }
  }, [activeNotice, clearAttachments, addFiles])

  const runtime = useExternalStoreRuntime({
    messages,
    convertMessage: identityConvert,
    isRunning: pending,
    onNew,
    onCancel,
    isDisabled: !agentId,
  })

  return {
    runtime,
    isRunning: pending,
    invocationId,
    notice: activeNotice,
    persistedEvents,
    liveQuery,
    restoreInput,
  }
}

const EMPTY_STREAMING_EVENTS: ParsedEvent[] = []

// Workaround for an @assistant-ui/core bug: when convertMessage is absent,
// ThreadMessageLike objects bypass fromThreadMessageLike and land in the
// internal repository without metadata, crashing when the state getter
// accesses message.metadata.unstable_state.  An identity converter forces
// every message through fromThreadMessageLike which initialises metadata.
const identityConvert = (m: ThreadMessageLike): ThreadMessageLike => m
