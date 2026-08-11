import { memo, useEffect, useMemo, useRef, useState, type ComponentProps } from "react";
import ReactMarkdown, { type Components } from "react-markdown";
import remarkGfm from "remark-gfm";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { AgentAvatar } from "@/components/butter/primitives";
import { cn } from "@/lib/utils";
import { parseSessionEvents, type ParsedEvent, type ToolCallSummary, type ToolResponseSummary } from "@/lib/session-events";
import { buildInputParts, type InputPartInit } from "@/lib/image-attachments";
import { useImageAttachments } from "@/hooks/use-image-attachments";
import { useLiveSession, useUpdateSessionTitle } from "@/api/sessions";
import { useAgents } from "@/api/agents";
import { InlineTitleInput } from "@/components/inline-title-input";
import { sessionTitle } from "@/lib/session-title";
import { CHAT_APP_NAME } from "@/lib/constants";
import { newClientID } from "@/lib/client-id";
import { cancelAgentInvocation, getAgentInvocation, submitAgentInvocation } from "@/api/chat";
import { InvocationStatus } from "@/gen/agents/v1/agent_service_pb";
import {
  ArrowUp,
  ChevronDown,
  Copy,
  ExternalLink,
  Loader2,
  MoreHorizontal,
  Paperclip,
  Pencil,
  Square,
  Trash2,
  Wrench,
  X,
} from "lucide-react";
import { toast } from "sonner";
import type { SessionInfo } from "@/types/api";

const APP_NAME = CHAT_APP_NAME;
const EMPTY_STREAMING_EVENTS: ParsedEvent[] = [];
const MARKDOWN_REMARK_PLUGINS = [remarkGfm];

interface ChatWindowProps {
  session: SessionInfo | null;
  userId: string;
  agentName: string | null;
  /** Immutable agent_id from session state; preferred over agentName when set. */
  agentId?: string | null;
  onDelete?: () => void;
  onInvocationAccepted?: (invocationId: string, message: string) => void;
  /** Message queued from the draft composer; auto-sent once the session is ready. */
  pendingMessage?: string;
  /** Invocation already accepted by the new-chat submit command. */
  initialInvocationId?: string;
}

interface ChatRunState {
  runId: string | null;
  sessionId: string;
  pending: boolean;
  pendingBaseEventIds: Set<string> | null;
  pendingUserMessage: string | null;
  streamingEvents: ParsedEvent[];
  streamingResponse: string;
  invocationId: string | null;
}

export function ChatWindow({ session, userId, agentName, agentId, onDelete, onInvocationAccepted, pendingMessage, initialInvocationId }: ChatWindowProps) {
  const sessionId = session?.session_id ?? "";
  const [draft, setDraft] = useState("");
  const {
    attachments, previewUrls, isDragOver,
    removeAttachment, clearAttachments,
    validateForSend,
    handleDragOver, handleDragEnter, handleDragLeave, handleDrop,
    handlePaste, fileInputRef, openFilePicker, handleFileInputChange, fileAccept,
  } = useImageAttachments(sessionId);
  const [runState, setRunState] = useState<ChatRunState>(() => emptyChatRunState(""));
  // Synchronous re-entry guard: `pending` flips only after the async file
  // read in handleSend, so two rapid clicks could otherwise both pass the
  // guard and start two runs.
  const sendingRef = useRef(false);
  const observationRef = useRef(0);
  const scrollRef = useRef<HTMLDivElement | null>(null);
  const taRef = useRef<HTMLTextAreaElement | null>(null);

  const isRunForCurrentSession = runState.sessionId === sessionId;
  const pending = isRunForCurrentSession && runState.pending;
  const pendingBaseEventIds = isRunForCurrentSession ? runState.pendingBaseEventIds : null;
  const pendingUserMessage = isRunForCurrentSession ? runState.pendingUserMessage : null;
  const streamingEvents = isRunForCurrentSession ? runState.streamingEvents : EMPTY_STREAMING_EVENTS;
  const streamingResponse = isRunForCurrentSession ? runState.streamingResponse : "";
  const invocationId = isRunForCurrentSession ? runState.invocationId : null;

  const liveQuery = useLiveSession(APP_NAME, userId, sessionId, pending);
  // Sessions created before the Agent ID migration only stored agent_name.
  // The backend now requires agent_id on ReplySession/StreamAgent, so resolve
  // the stored name to its agent_id via the loaded agents list.
  const { data: agentsData } = useAgents({ page_size: 200 });
  const agents = agentsData?.agents;
  const resolvedAgentId = useMemo(() => {
    if (agentId) return agentId;
    if (!agentName || !agents) return undefined;
    return agents.find((a) => a.name === agentName)?.agent_id;
  }, [agentId, agentName, agents]);
  const renameMutation = useUpdateSessionTitle();
  // Keyed by session so switching chats implicitly leaves edit mode without
  // an effect-driven state reset.
  const [editingTitleFor, setEditingTitleFor] = useState<string | null>(null);
  const editingTitle = editingTitleFor === sessionId;

  const persistedEvents = useMemo<ParsedEvent[]>(
    () => parseSessionEvents(liveQuery.data?.session_detail.events),
    [liveQuery.data],
  );
  const optimisticUserEvent = useMemo<ParsedEvent | null>(() => {
    if (!pendingUserMessage) return null;
    return makeOptimisticTextEvent("pending-user", "user", pendingUserMessage);
  }, [pendingUserMessage]);
  const events = useMemo<ParsedEvent[]>(() => {
    const out: ParsedEvent[] = [];
    const seen = new Set<string>();
    const baseEvents = pendingBaseEventIds
      ? persistedEvents.filter((evt) => pendingBaseEventIds.has(evt.eventId))
      : persistedEvents;

    for (const event of baseEvents) appendUniqueEvent(out, seen, event);
    if (optimisticUserEvent) appendUniqueEvent(out, seen, optimisticUserEvent);
    for (const event of streamingEvents) appendUniqueEvent(out, seen, event);
    if (streamingResponse.trim()) {
      appendUniqueEvent(
        out,
        seen,
        makeOptimisticTextEvent("streaming-assistant", "assistant", streamingResponse),
      );
    }
    return out;
  }, [persistedEvents, pendingBaseEventIds, optimisticUserEvent, streamingEvents, streamingResponse]);
  const visibleEvents = useMemo(() => events.filter(isRenderableEvent), [events]);
  const timelineEvents = useMemo(() => groupToolEvents(visibleEvents), [visibleEvents]);

  // Observe an Invocation accepted by the new-chat composer. Cleanup only
  // detaches this browser view; it never calls cancellation.
  useEffect(() => {
    observationRef.current += 1;
    if (!initialInvocationId || !sessionId) return;

    const runId = newClientID();
    const frame = globalThis.requestAnimationFrame(() => {
      setRunState({
        runId,
        sessionId,
        pending: true,
        pendingBaseEventIds: new Set(persistedEvents.map((evt) => evt.eventId)),
        pendingUserMessage: pendingMessage?.trim() || null,
        streamingEvents: [],
        streamingResponse: "",
        invocationId: initialInvocationId,
      });
      void observeInvocation(initialInvocationId, sessionId, runId, pendingMessage?.trim() || "");
    });

    return () => {
      globalThis.cancelAnimationFrame(frame);
      observationRef.current += 1;
    };
    // persistedEvents are deliberately captured at acceptance time only.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialInvocationId, sessionId]);

  useEffect(() => {
    const node = scrollRef.current;
    if (node) node.scrollTop = node.scrollHeight;
  }, [events.length, pending]);

  if (!session) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        Select a chat in the sidebar or start a new one.
      </div>
    );
  }

  async function handleSend() {
    const text = draft.trim();
    const images = attachments;
    if ((!text && images.length === 0) || !agentName || pending || sendingRef.current) return;
    if (!resolvedAgentId) {
      toast.error(
        agents
          ? `Agent "${agentName}" is no longer available; cannot resume this chat.`
          : "Agents are still loading; please try again in a moment.",
      );
      return;
    }
    sendingRef.current = true;

    if (images.length > 0) {
      const validationErrors = validateForSend();
      if (validationErrors.length > 0) {
        validationErrors.forEach((msg) => toast.error(msg));
        sendingRef.current = false;
        return;
      }
    }

    let parts: InputPartInit[] | undefined;
    try {
      parts = images.length > 0 ? await buildInputParts(text, images) : undefined;
    } catch {
      toast.error("Failed to read attached images");
      sendingRef.current = false;
      return;
    }

    const runId = newClientID();
    setDraft("");
    if (taRef.current) taRef.current.style.height = "auto";
    setRunState({
      runId,
      sessionId,
      pending: true,
      pendingBaseEventIds: new Set(persistedEvents.map((evt) => evt.eventId)),
      pendingUserMessage: text || `(${images.length} image${images.length > 1 ? "s" : ""})`,
      streamingEvents: [],
      streamingResponse: "",
      invocationId: null,
    });

    try {
      const accepted = await submitAgentInvocation({
        request_id: newClientID(),
        agent_id: resolvedAgentId,
        session_id: sessionId,
        message: text,
        parts,
      });
      setRunState((prev) => updateChatRun(prev, sessionId, runId, (current) => ({
        ...current,
        invocationId: accepted.invocation_id,
      })));
      clearAttachments();
      sendingRef.current = false;
      if (onInvocationAccepted) {
        onInvocationAccepted(accepted.invocation_id, text);
      } else {
        void observeInvocation(accepted.invocation_id, sessionId, runId, text);
      }
    } catch (err) {
      sendingRef.current = false;
      setRunState((prev) => prev.runId === runId ? emptyChatRunState(prev.sessionId) : prev);
      setDraft(text);
      toast.error(err instanceof Error ? err.message : "Failed to send message");
    }
  }

  async function handleStop() {
    const id = invocationId;
    if (id) {
      try {
        const result = await cancelAgentInvocation(id);
        if (!result.cancelled) {
          await observeInvocation(id, sessionId, runState.runId ?? newClientID(), pendingUserMessage ?? "");
        }
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Failed to cancel invocation");
      }
    }
  }

  async function observeInvocation(id: string, observedSessionId: string, runId: string, retryText: string) {
    const token = ++observationRef.current;
    for (;;) {
      if (token !== observationRef.current) return;
      try {
        const inv = await getAgentInvocation(id);
        if (isActiveInvocation(inv.status)) {
          await waitForStatusPoll();
          continue;
        }
        await liveQuery.refetch();
        if (token !== observationRef.current) return;
        setRunState((prev) => prev.runId === runId ? emptyChatRunState(observedSessionId) : prev);
        if (inv.status === InvocationStatus.CANCELLED) {
          if (retryText) setDraft(retryText);
          toast.info("Chat stopped");
        } else if (inv.status === InvocationStatus.FAILED) {
          if (retryText) setDraft(retryText);
          toast.error(inv.error || "Agent invocation failed");
        }
        return;
      } catch (err) {
        if (token !== observationRef.current) return;
        setRunState((prev) => prev.runId === runId ? emptyChatRunState(observedSessionId) : prev);
        if (retryText) setDraft(retryText);
        toast.error(err instanceof Error ? err.message : "Failed to load invocation status");
        return;
      }
    }
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (
      e.key === "Enter" &&
      !e.shiftKey &&
      !e.nativeEvent.isComposing &&
      e.keyCode !== 229
    ) {
      e.preventDefault();
      void handleSend();
    }
  }

  const canSend = !!agentName && (draft.trim().length > 0 || attachments.length > 0);

  return (
    <div
      className={cn("flex min-h-0 flex-1 flex-col bg-background", isDragOver && "ring-2 ring-inset ring-ring/50")}
      onDragOver={handleDragOver}
      onDragEnter={handleDragEnter}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      {/* Chat header */}
      <header className="flex h-12 shrink-0 items-center justify-between gap-2 border-b border-border/60 bg-background/95 px-2.5 sm:px-4 md:px-6">
        <div className="flex min-w-0 flex-1 items-center gap-2.5">
          <AgentAvatar name={agentName ?? "?"} size="sm" />
          <span className="flex min-w-0 max-w-md flex-1 flex-col">
            {editingTitle ? (
              <InlineTitleInput
                initial={sessionTitle(session)}
                onSave={async (title) => {
                  await renameMutation.mutateAsync({
                    app_name: session.app_name,
                    user_id: session.user_id,
                    session_id: session.session_id,
                    title,
                  });
                }}
                onClose={() => setEditingTitleFor(null)}
                className="text-sm font-medium"
              />
            ) : (
              <span className="flex min-w-0 items-center gap-1">
                <span className="truncate text-sm font-semibold leading-tight">
                  {sessionTitle(session)}
                </span>
                <button
                  type="button"
                  onClick={() => setEditingTitleFor(sessionId)}
                  aria-label="Rename chat"
                  title="Rename chat"
                  className="inline-flex size-9 shrink-0 touch-manipulation items-center justify-center rounded-md text-muted-foreground transition-[color,background-color,scale] hover:bg-muted hover:text-foreground active:scale-[0.96] motion-reduce:active:scale-100"
                >
                  <Pencil className="size-3.5" />
                </button>
              </span>
            )}
            <span
              title={sessionId}
              className="truncate font-mono text-[0.65rem] leading-tight text-muted-foreground/80"
            >
              {agentName ?? "Unknown agent"}
              <span className="hidden sm:inline"> / {sessionId.slice(0, 8)}</span>
            </span>
          </span>
        </div>

        <DropdownMenu>
          <DropdownMenuTrigger
            aria-label="Chat options"
            className="inline-flex size-9 touch-manipulation items-center justify-center rounded-md text-muted-foreground transition-[color,background-color,scale] hover:bg-muted hover:text-foreground active:scale-[0.96] motion-reduce:active:scale-100"
          >
            <MoreHorizontal className="size-4" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" sideOffset={6}>
            <DropdownMenuItem variant="destructive" onClick={onDelete}>
              <Trash2 />
              Delete chat
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </header>

      {/* Messages */}
      <div ref={scrollRef} className="scrollbar-thin min-h-0 flex-1 overflow-y-auto">
        <div className="mx-auto w-full max-w-4xl px-3.5 sm:px-5 lg:px-6">
          {liveQuery.isLoading ? (
            <div className="space-y-6 py-8">
              <div className="flex gap-3">
                <Skeleton className="size-8 shrink-0 rounded-md" />
                <div className="w-full max-w-xl space-y-2">
                  <Skeleton className="h-3 w-28" />
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-4/5" />
                </div>
              </div>
              <Skeleton className="ml-auto h-14 w-1/2 max-w-md rounded-lg" />
            </div>
          ) : visibleEvents.length === 0 ? (
            <div className="flex min-h-[50vh] flex-col items-center justify-center text-center">
              <AgentAvatar name={agentName ?? "?"} size="lg" />
              <h2 className="mt-3 text-lg font-semibold">{agentName ?? "Unknown agent"}</h2>
              <p className="mt-1 max-w-sm text-sm text-muted-foreground text-pretty">
                Send a message below to start the conversation.
              </p>
            </div>
          ) : (
            <div className="py-4 sm:py-6">
              {timelineEvents.map((evt, index) => {
                const previous = timelineEvents[index - 1];
                const showIdentity =
                  evt.role !== "user" &&
                  (!previous || previous.role === "user" || previous.author !== evt.author);
                return (
                  <MessageRow
                    key={evt.eventId}
                    event={evt}
                    agentName={agentName ?? "agent"}
                    showIdentity={showIdentity}
                  />
                );
              })}
              {pending && (
                <div className="grid grid-cols-[2rem_minmax(0,1fr)] gap-3 py-2">
                  <span />
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <Loader2 className="size-3 animate-spin" /> Thinking…
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Composer */}
      <div className="shrink-0 border-t border-border/60 bg-background/95 backdrop-blur-sm">
        <div className="mx-auto w-full max-w-4xl px-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] pt-2.5 sm:px-5 md:pb-3.5">
          {attachments.length > 0 && (
            <div className="mb-2 flex flex-wrap gap-1.5">
              {attachments.map((file, index) => (
                <span
                  key={`${file.name}-${index}`}
                  className="group relative inline-flex"
                >
                  <img
                    src={previewUrls[index]}
                    alt={file.name}
                    title={file.name}
                    className="h-14 w-14 rounded-md object-cover outline outline-1 -outline-offset-1 outline-black/10 dark:outline-white/10"
                  />
                  <button
                    type="button"
                    onClick={() => removeAttachment(index)}
                    aria-label={`Remove ${file.name}`}
                    className="absolute -right-2 -top-2 inline-flex size-6 touch-manipulation items-center justify-center rounded-full border border-border bg-background text-muted-foreground shadow-sm transition-[color,background-color,scale] duration-150 ease-out hover:bg-muted hover:text-foreground active:scale-[0.96] motion-reduce:active:scale-100"
                  >
                    <X className="size-3" />
                  </button>
                </span>
              ))}
            </div>
          )}
          <div
            className={cn(
              "flex items-end gap-1.5 rounded-lg border border-border/70 bg-card p-1.5 shadow-sm transition-[border-color,box-shadow] focus-within:border-foreground/25 focus-within:ring-2 focus-within:ring-ring/10 sm:gap-2 sm:p-2",
              !agentName && "opacity-60",
            )}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept={fileAccept}
              multiple
              className="hidden"
              onChange={handleFileInputChange}
            />
            <button
              type="button"
              disabled={!agentName || pending}
              onClick={openFilePicker}
              aria-label="Attach images"
              className="inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md text-muted-foreground transition-[color,background-color,scale] duration-150 ease-out hover:bg-muted hover:text-foreground active:scale-[0.96] motion-reduce:active:scale-100 disabled:pointer-events-none"
            >
              <Paperclip className="size-4" />
            </button>
            <textarea
              ref={taRef}
              rows={1}
              value={draft}
              disabled={!agentName}
              onChange={(e) => {
                setDraft(e.target.value);
                e.target.style.height = "auto";
                e.target.style.height = `${Math.min(e.target.scrollHeight, 160)}px`;
              }}
              onKeyDown={handleKeyDown}
              onPaste={handlePaste}
              placeholder={
                agentName
                  ? `Message ${agentName}...`
                  : "This chat is missing an agent reference; cannot send."
              }
              className="max-h-40 min-h-10 flex-1 resize-none bg-transparent py-2.5 text-[0.9rem] leading-5 outline-none placeholder:text-muted-foreground/75"
            />
            {pending && invocationId ? (
              <button
                type="button"
                onClick={() => void handleStop()}
                aria-label="Stop generating"
                className="inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md bg-secondary text-secondary-foreground transition-[background-color,scale] duration-150 ease-out hover:bg-secondary/80 active:scale-[0.96] motion-reduce:active:scale-100"
              >
                <Square className="size-4 fill-current" />
              </button>
            ) : pending ? (
              <button
                type="button"
                disabled
                aria-label="Submitting message"
                className="inline-flex size-10 shrink-0 items-center justify-center rounded-md bg-secondary text-secondary-foreground"
              >
                <Loader2 className="size-4 animate-spin" />
              </button>
            ) : (
              <button
                type="button"
                onClick={() => void handleSend()}
                disabled={!canSend}
                aria-label="Send message"
                className="inline-flex size-10 shrink-0 touch-manipulation items-center justify-center rounded-md bg-primary text-primary-foreground transition-[background-color,opacity,scale] duration-150 ease-out hover:bg-primary/90 active:scale-[0.96] motion-reduce:active:scale-100 disabled:opacity-35"
              >
                <ArrowUp className="size-4" />
              </button>
            )}
          </div>
          <p className="mt-1.5 px-2 text-center text-[0.7rem] leading-4 text-muted-foreground/80">
            Butter can make mistakes. Verify important actions before running them.
          </p>
        </div>
      </div>
    </div>
  );
}

function makeOptimisticTextEvent(
  eventId: string,
  role: ParsedEvent["role"],
  text: string,
): ParsedEvent {
  const author = role === "assistant" ? "agent" : role;
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
  };
}

function emptyChatRunState(sessionId: string): ChatRunState {
  return {
    runId: null,
    sessionId,
    pending: false,
    pendingBaseEventIds: null,
    pendingUserMessage: null,
    streamingEvents: [],
    streamingResponse: "",
    invocationId: null,
  };
}

function updateChatRun(
  prev: ChatRunState,
  sessionId: string,
  runId: string,
  update: (current: ChatRunState) => ChatRunState,
): ChatRunState {
  if (prev.sessionId !== sessionId || prev.runId !== runId) return prev;
  return update(prev);
}

function appendUniqueEvent(out: ParsedEvent[], seen: Set<string>, event: ParsedEvent) {
  if (seen.has(event.eventId)) return;
  seen.add(event.eventId);
  out.push(event);
}

function isRenderableEvent(event: ParsedEvent): boolean {
  return (
    event.text.trim().length > 0 ||
    event.toolCalls.length > 0 ||
    event.toolResponses.length > 0
  );
}

function groupToolEvents(events: ParsedEvent[]): ParsedEvent[] {
  const grouped: ParsedEvent[] = [];

  for (const event of events) {
    const previous = grouped[grouped.length - 1];
    const isToolOnly =
      event.role !== "user" &&
      event.text.trim().length === 0 &&
      (event.toolCalls.length > 0 || event.toolResponses.length > 0);
    const previousIsToolOnly =
      previous &&
      previous.role !== "user" &&
      previous.text.trim().length === 0 &&
      (previous.toolCalls.length > 0 || previous.toolResponses.length > 0);

    if (
      isToolOnly &&
      previousIsToolOnly &&
      previous.author === event.author &&
      previous.traceUrl === event.traceUrl
    ) {
      grouped[grouped.length - 1] = {
        ...previous,
        eventId: `${previous.eventId}:${event.eventId}`,
        toolCalls: [...previous.toolCalls, ...event.toolCalls],
        toolResponses: [...previous.toolResponses, ...event.toolResponses],
      };
      continue;
    }

    grouped.push(event);
  }

  return grouped;
}

function isActiveInvocation(status: InvocationStatus): boolean {
  return status === InvocationStatus.QUEUED || status === InvocationStatus.RUNNING;
}

function waitForStatusPoll(): Promise<void> {
  return new Promise((resolve) => globalThis.setTimeout(resolve, 500));
}

const MARKDOWN_COMPONENTS: Components = {
  a: MarkdownLink,
  code: MarkdownCode,
  pre: MarkdownPre,
  table: MarkdownTable,
  th: MarkdownTableHeader,
  td: MarkdownTableCell,
  p: ({ children }) => (
    <p className="my-2 max-w-[72ch] first:mt-0 last:mb-0">{children}</p>
  ),
  ul: ({ children }) => (
    <ul className="my-2 max-w-[72ch] list-disc space-y-1 pl-5 first:mt-0 last:mb-0">
      {children}
    </ul>
  ),
  ol: ({ children }) => (
    <ol className="my-2 max-w-[72ch] list-decimal space-y-1 pl-5 first:mt-0 last:mb-0">
      {children}
    </ol>
  ),
  li: ({ children }) => <li className="pl-1">{children}</li>,
  blockquote: ({ children }) => (
    <blockquote className="my-2 max-w-[72ch] border-l-2 border-border pl-3 italic text-muted-foreground first:mt-0 last:mb-0">
      {children}
    </blockquote>
  ),
  hr: () => <hr className="my-3 border-border" />,
  h1: ({ children }) => <h1 className="my-3 max-w-[72ch] text-lg font-semibold first:mt-0 last:mb-0">{children}</h1>,
  h2: ({ children }) => <h2 className="my-3 max-w-[72ch] text-base font-semibold first:mt-0 last:mb-0">{children}</h2>,
  h3: ({ children }) => <h3 className="my-2 max-w-[72ch] text-sm font-semibold first:mt-0 last:mb-0">{children}</h3>,
};

function MarkdownLink(props: ComponentProps<"a">) {
  return (
    <a
      {...props}
      target="_blank"
      rel="noopener noreferrer"
      className="font-medium underline underline-offset-2 hover:opacity-80"
    />
  );
}

function MarkdownCode({ children, className }: ComponentProps<"code">) {
  const isInline = !className;
  if (isInline) {
    return (
      <code className="rounded bg-muted px-1 py-0.5 font-mono text-[0.85em] text-foreground">
        {children}
      </code>
    );
  }
  return <code className={cn("font-mono text-xs", className)}>{children}</code>;
}

function MarkdownPre({ children }: ComponentProps<"pre">) {
  return (
    <pre className="scrollbar-thin my-2 max-w-full overflow-x-auto rounded-md border border-border/70 bg-muted/35 p-3 text-foreground first:mt-0 last:mb-0">
      {children}
    </pre>
  );
}

function MarkdownTable({ children }: ComponentProps<"table">) {
  return (
    <div className="scrollbar-thin my-3 max-w-full overflow-x-auto rounded-md border border-border/70 first:mt-0 last:mb-0">
      <table className="w-full min-w-[42rem] border-separate border-spacing-0 text-left text-xs">
        {children}
      </table>
    </div>
  );
}

function MarkdownTableHeader({ children }: ComponentProps<"th">) {
  return (
    <th className="whitespace-nowrap border-b border-border bg-muted/55 px-3 py-2 font-semibold text-foreground">
      {children}
    </th>
  );
}

function MarkdownTableCell({ children }: ComponentProps<"td">) {
  return (
    <td className="border-b border-border/50 px-3 py-2 align-top leading-5 last:border-r-0">
      {children}
    </td>
  );
}

function ToolActivity({
  calls,
  responses,
}: {
  calls: ToolCallSummary[];
  responses: ToolResponseSummary[];
}) {
  const steps = [
    ...calls.map((item) => ({ kind: "Arguments", name: item.name, preview: item.argsPreview })),
    ...responses.map((item) => ({ kind: "Response", name: item.name, preview: item.responsePreview })),
  ];
  const names = Array.from(new Set(steps.map((step) => step.name)));

  return (
    <details className="group/tool my-2 max-w-[72ch] overflow-hidden rounded-md border border-border/60 bg-muted/20">
      <summary className="flex cursor-pointer list-none items-center gap-2 px-2.5 py-1.5 text-left transition-colors hover:bg-muted/45 [&::-webkit-details-marker]:hidden">
        <Wrench className="size-3.5 shrink-0 text-muted-foreground" />
        <span className="truncate font-mono text-xs text-foreground/85">
          {names.join(", ")}
        </span>
        <span className="shrink-0 text-[0.7rem] text-muted-foreground">
          {steps.length} {steps.length === 1 ? "step" : "steps"}
        </span>
        <ChevronDown className="ml-auto size-3.5 shrink-0 text-muted-foreground transition-transform group-open/tool:rotate-180" />
      </summary>
      <div className="space-y-2 border-t border-border/60 px-2.5 py-2">
        {steps.map((step, index) => (
          <div key={`${step.kind}-${step.name}-${index}`}>
            <div className="mb-1 flex items-center gap-2 text-[0.68rem] text-muted-foreground">
              <span>{step.kind}</span>
              <span className="font-mono text-foreground/75">{step.name}</span>
            </div>
            {step.preview && (
              <pre className="scrollbar-thin max-h-64 overflow-auto rounded bg-muted/55 p-2 font-mono text-xs leading-5">
                {step.preview}
              </pre>
            )}
          </div>
        ))}
      </div>
    </details>
  );
}

function MessageActions({ event }: { event: ParsedEvent }) {
  const hasText = event.text.trim().length > 0;
  if (!hasText && !event.traceUrl) return null;

  return (
    <div className="mt-1 flex min-h-6 items-center gap-0.5 text-muted-foreground opacity-100 transition-opacity sm:opacity-0 sm:focus-within:opacity-100 sm:group-hover/message:opacity-100">
      {hasText && (
        <button
          type="button"
          title="Copy message"
          aria-label="Copy message"
          onClick={() => {
            void navigator.clipboard.writeText(event.text).then(
              () => toast.success("Message copied"),
              () => toast.error("Failed to copy message"),
            );
          }}
          className="inline-flex size-8 items-center justify-center rounded-md transition-colors hover:bg-muted hover:text-foreground"
        >
          <Copy className="size-3.5" />
        </button>
      )}
      {event.traceUrl && (
        <a
          href={event.traceUrl}
          target="_blank"
          rel="noopener noreferrer"
          title="Open trace"
          aria-label="Open trace"
          className="inline-flex size-8 items-center justify-center rounded-md transition-colors hover:bg-muted hover:text-foreground"
        >
          <ExternalLink className="size-3.5" />
        </a>
      )}
    </div>
  );
}

const MessageRow = memo(function MessageRow({
  event,
  agentName,
  showIdentity,
}: {
  event: ParsedEvent;
  agentName: string;
  showIdentity: boolean;
}) {
  const isUser = event.role === "user";
  const hasText = event.text.trim().length > 0;
  const hasTools = event.toolCalls.length > 0 || event.toolResponses.length > 0;

  if (isUser) {
    return (
      <div className="group/message flex flex-col items-end py-3 sm:py-4">
        <div className="max-w-[92%] rounded-lg rounded-tr-sm bg-secondary px-3.5 py-2.5 text-[0.9rem] leading-relaxed text-secondary-foreground sm:max-w-[min(80%,48rem)]">
          <ReactMarkdown remarkPlugins={MARKDOWN_REMARK_PLUGINS} components={MARKDOWN_COMPONENTS}>
            {event.text}
          </ReactMarkdown>
        </div>
      </div>
    );
  }

  return (
    <div
      className={cn(
        "group/message grid grid-cols-[1.75rem_minmax(0,1fr)] gap-2.5 sm:grid-cols-[2rem_minmax(0,1fr)] sm:gap-3",
        showIdentity ? "pt-4" : "pt-1",
        hasText ? "pb-2" : "pb-0.5",
      )}
    >
      <div className="pt-0.5">
        {showIdentity && <AgentAvatar name={event.author || agentName} size="sm" />}
      </div>
      <div className="min-w-0">
        {showIdentity && (
          <div className="mb-1 flex min-h-5 items-center gap-2">
            <span className="text-sm font-medium">{event.author || agentName}</span>
            {event.timestamp && (
              <span className="text-[0.7rem] text-muted-foreground/80">
                {new Date(event.timestamp).toLocaleTimeString([], {
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </span>
            )}
          </div>
        )}

        {hasTools && (
          <ToolActivity calls={event.toolCalls} responses={event.toolResponses} />
        )}

        {hasText && (
          <div className="text-[0.9rem] leading-6 text-foreground">
            <ReactMarkdown remarkPlugins={MARKDOWN_REMARK_PLUGINS} components={MARKDOWN_COMPONENTS}>
              {event.text}
            </ReactMarkdown>
          </div>
        )}

        <MessageActions event={event} />
      </div>
    </div>
  );
});
