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
import { parseSessionEvent, parseSessionEvents, type ParsedEvent, type ToolCallSummary, type ToolResponseSummary } from "@/lib/session-events";
import { buildInputParts, type InputPartInit } from "@/lib/image-attachments";
import { useImageAttachments } from "@/hooks/use-image-attachments";
import { useLiveSession, useReplySession } from "@/api/sessions";
import { cancelAgentInvocation, streamChat, type ChatStreamPayload } from "@/api/chat";
import {
  ArrowUp,
  ChevronDown,
  ExternalLink,
  Loader2,
  MoreHorizontal,
  Paperclip,
  Square,
  Trash2,
  Wrench,
  X,
} from "lucide-react";
import { toast } from "sonner";
import type { SessionInfo } from "@/types/api";

const APP_NAME = "web-chat";
const EMPTY_STREAMING_EVENTS: ParsedEvent[] = [];
const MARKDOWN_REMARK_PLUGINS = [remarkGfm];

interface ChatWindowProps {
  session: SessionInfo | null;
  userId: string;
  agentName: string | null;
  onDelete?: () => void;
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

export function ChatWindow({ session, userId, agentName, onDelete }: ChatWindowProps) {
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
  const abortRef = useRef<AbortController | null>(null);
  const activeRunIdRef = useRef<string | null>(null);
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
  const reply = useReplySession();

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

  useEffect(() => {
    abortRef.current?.abort();
  }, [sessionId]);

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

    // Attachments stay visible (and in state) while the run is in flight —
    // the picker is disabled during pending, and they are only cleared once
    // the send actually succeeds, so any failure path keeps them for retry.
    const runId = newRunId();
    abortRef.current?.abort();
    activeRunIdRef.current = runId;
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

    const controller = new AbortController();
    abortRef.current = controller;
    let streamStarted = false;

    try {
      await streamChat(
        {
          agent_name: agentName,
          app_name: APP_NAME,
          user_id: userId,
          session_id: sessionId,
          message: text,
          parts,
        },
        {
          onStarted: (payload) => {
            streamStarted = true;
            if (payload.invocation_id) {
              setRunState((prev) => updateChatRun(prev, sessionId, runId, (current) => ({
                ...current,
                invocationId: payload.invocation_id ?? current.invocationId,
              })));
            }
          },
          onAgentEvent: (payload) => {
            const event = payloadToParsedEvent(payload);
            if (event) {
              setRunState((prev) => updateChatRun(prev, sessionId, runId, (current) => ({
                ...current,
                streamingEvents: [...current.streamingEvents, event],
              })));
            }
          },
          onTextDelta: (payload) => {
            if (payload.text_delta) {
              setRunState((prev) => updateChatRun(prev, sessionId, runId, (current) => ({
                ...current,
                streamingResponse: current.streamingResponse + payload.text_delta,
              })));
            }
          },
          onFinal: (payload) => {
            setRunState((prev) => updateChatRun(prev, sessionId, runId, (current) => ({
              ...current,
              streamingResponse: payload.response ?? "",
            })));
          },
          onError: (payload) => {
            if (payload.error) toast.error(payload.error);
          },
        },
        controller.signal,
      );
      await liveQuery.refetch();
      clearAttachments();
    } catch (err) {
      if (isAbortError(err)) {
        toast.info("Chat stopped");
      } else if (!streamStarted) {
        // Preserve old behavior as a fallback when the SSE endpoint cannot be opened.
        try {
          await reply.mutateAsync({
            agent_name: agentName,
            app_name: APP_NAME,
            user_id: userId,
            session_id: sessionId,
            message: text,
            parts,
          });
          clearAttachments();
        } catch (fallbackErr) {
          toast.error(fallbackErr instanceof Error ? fallbackErr.message : "Failed to send message");
          setDraft(text);
        }
      } else {
        toast.error(err instanceof Error ? err.message : "Failed to send message");
      }
    } finally {
      sendingRef.current = false;
      setRunState((prev) => prev.runId === runId ? emptyChatRunState(prev.sessionId) : prev);
      if (activeRunIdRef.current === runId) {
        activeRunIdRef.current = null;
        abortRef.current = null;
      }
    }
  }

  async function handleStop() {
    abortRef.current?.abort();
    const id = invocationId;
    if (id) {
      try {
        await cancelAgentInvocation(id);
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Failed to cancel invocation");
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
      className={cn("flex h-full flex-col", isDragOver && "ring-2 ring-inset ring-ring/50")}
      onDragOver={handleDragOver}
      onDragEnter={handleDragEnter}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      {/* Chat header */}
      <header className="flex h-14 shrink-0 items-center justify-between gap-3 border-b border-border px-4">
        <div className="flex min-w-0 items-center gap-2.5 md:pl-8">
          <AgentAvatar name={agentName ?? "?"} size="sm" />
          <span className="flex min-w-0 flex-col">
            <span className="truncate text-sm font-semibold leading-tight">
              {agentName ?? "Unknown agent"}
            </span>
            <span className="truncate font-mono text-[0.65rem] leading-tight text-muted-foreground">
              {sessionId}
            </span>
          </span>
        </div>

        <DropdownMenu>
          <DropdownMenuTrigger className="rounded-md p-1.5 text-muted-foreground hover:bg-muted">
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
        <div className="mx-auto w-full max-w-3xl px-4">
          {liveQuery.isLoading ? (
            <div className="space-y-3 py-4">
              <Skeleton className="h-16 w-2/3" />
              <Skeleton className="ml-auto h-16 w-1/2" />
            </div>
          ) : events.length === 0 ? (
            <div className="flex min-h-[50vh] flex-col items-center justify-center text-center">
              <AgentAvatar name={agentName ?? "?"} size="lg" />
              <h2 className="mt-3 text-lg font-semibold">{agentName ?? "Unknown agent"}</h2>
              <p className="mt-1 max-w-sm text-sm text-muted-foreground text-pretty">
                Send a message below to start the conversation.
              </p>
            </div>
          ) : (
            <div className="py-2">
              {events.map((evt) => (
                <MessageRow key={evt.eventId} event={evt} agentName={agentName ?? "agent"} />
              ))}
              {pending && (
                <div className="flex items-center gap-2 py-2 text-xs text-muted-foreground">
                  <Loader2 className="size-3 animate-spin" /> Agent is thinking…
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Composer */}
      <div className="shrink-0 border-t border-border bg-background">
        <div className="mx-auto w-full max-w-3xl px-3 pb-3 pt-3 md:px-4 md:pb-4">
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
                    className="h-14 w-14 rounded-md border border-border object-cover"
                  />
                  <button
                    type="button"
                    onClick={() => removeAttachment(index)}
                    aria-label={`Remove ${file.name}`}
                    className="absolute -right-1.5 -top-1.5 rounded-full border border-border bg-background p-0.5 text-muted-foreground shadow-sm hover:text-foreground"
                  >
                    <X className="size-3" />
                  </button>
                </span>
              ))}
            </div>
          )}
          <div
            className={cn(
              "flex items-end gap-2 rounded-xl border border-border bg-card p-2 shadow-sm transition-colors focus-within:border-ring",
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
              className="shrink-0 rounded-md p-2 text-muted-foreground hover:bg-muted hover:text-foreground disabled:pointer-events-none"
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
                  ? `Message ${agentName}…`
                  : "This chat is missing an agent reference; cannot send."
              }
              className="max-h-40 min-h-6 flex-1 resize-none bg-transparent py-1.5 text-[0.9rem] leading-relaxed outline-none placeholder:text-muted-foreground"
            />
            {pending ? (
              <button
                type="button"
                onClick={() => void handleStop()}
                aria-label="Stop generating"
                className="shrink-0 rounded-md bg-secondary p-2 text-secondary-foreground hover:bg-secondary/80"
              >
                <Square className="size-4 fill-current" />
              </button>
            ) : (
              <button
                type="button"
                onClick={() => void handleSend()}
                disabled={!canSend}
                aria-label="Send message"
                className="shrink-0 rounded-md bg-primary p-2 text-primary-foreground transition-opacity hover:bg-primary/90 disabled:opacity-40"
              >
                <ArrowUp className="size-4" />
              </button>
            )}
          </div>
          <p className="mt-1.5 text-center text-xs text-muted-foreground">
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

function newRunId(): string {
  return globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

function appendUniqueEvent(out: ParsedEvent[], seen: Set<string>, event: ParsedEvent) {
  if (seen.has(event.eventId)) return;
  seen.add(event.eventId);
  out.push(event);
}

function payloadToParsedEvent(payload: ChatStreamPayload): ParsedEvent | null {
  const evt = payload.event;
  if (!evt?.event_id) return null;
  const parsed = parseSessionEvent({
    event_id: evt.event_id,
    invocation_id: evt.invocation_id,
    author: evt.author,
    branch: evt.branch,
    content_json: evt.content_json,
    timestamp: evt.timestamp,
    trace_id: evt.invocation_id,
  });
  if (evt.partial && parsed.text) {
    return { ...parsed, text: "" };
  }
  return parsed;
}

function isAbortError(err: unknown): boolean {
  return err instanceof DOMException && err.name === "AbortError";
}

const MARKDOWN_COMPONENTS: Components = {
  a: MarkdownLink,
  code: MarkdownCode,
  pre: MarkdownPre,
  table: MarkdownTable,
  th: MarkdownTableHeader,
  td: MarkdownTableCell,
  p: ({ children }) => <p className="my-1.5 first:mt-0 last:mb-0">{children}</p>,
  ul: ({ children }) => (
    <ul className="my-1.5 list-disc space-y-1 pl-5 first:mt-0 last:mb-0">{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className="my-1.5 list-decimal space-y-1 pl-5 first:mt-0 last:mb-0">{children}</ol>
  ),
  li: ({ children }) => <li className="pl-1">{children}</li>,
  blockquote: ({ children }) => (
    <blockquote className="my-1.5 border-l-2 border-border pl-3 italic opacity-90 first:mt-0 last:mb-0">
      {children}
    </blockquote>
  ),
  hr: () => <hr className="my-3 border-border" />,
  h1: ({ children }) => <h1 className="my-2 text-lg font-semibold first:mt-0 last:mb-0">{children}</h1>,
  h2: ({ children }) => <h2 className="my-2 text-base font-semibold first:mt-0 last:mb-0">{children}</h2>,
  h3: ({ children }) => <h3 className="my-2 text-sm font-semibold first:mt-0 last:mb-0">{children}</h3>,
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
    <pre className="scrollbar-thin my-2 overflow-x-auto rounded-md border border-border bg-card p-3 text-foreground first:mt-0 last:mb-0">
      {children}
    </pre>
  );
}

function MarkdownTable({ children }: ComponentProps<"table">) {
  return (
    <div className="scrollbar-thin my-2 overflow-x-auto first:mt-0 last:mb-0">
      <table className="w-full border-collapse text-left text-xs">{children}</table>
    </div>
  );
}

function MarkdownTableHeader({ children }: ComponentProps<"th">) {
  return <th className="border border-border px-2 py-1 font-semibold">{children}</th>;
}

function MarkdownTableCell({ children }: ComponentProps<"td">) {
  return <td className="border border-border px-2 py-1 align-top">{children}</td>;
}

function ToolBlock({
  kind,
  name,
  preview,
}: {
  kind: "call" | "response";
  name: string;
  preview?: string;
}) {
  const [open, setOpen] = useState(false);
  return (
    <div className="my-2 overflow-hidden rounded-md border border-border bg-card/60">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-muted/50"
      >
        <Wrench className="size-3.5 text-muted-foreground" />
        <span className="font-mono text-xs">{name}</span>
        <span className="text-xs font-medium text-muted-foreground">
          {kind === "call" ? "Tool call" : "Tool response"}
        </span>
        {preview && (
          <ChevronDown
            className={cn(
              "ml-auto size-4 text-muted-foreground transition-transform",
              open && "rotate-180",
            )}
          />
        )}
      </button>
      {open && preview && (
        <div className="border-t border-border px-3 py-2">
          <p className="mb-1 text-[0.7rem] font-medium uppercase tracking-wide text-muted-foreground">
            {kind === "call" ? "Arguments" : "Response"}
          </p>
          <pre className="scrollbar-thin overflow-x-auto rounded bg-muted/60 p-2 font-mono text-xs">
            {preview}
          </pre>
        </div>
      )}
    </div>
  );
}

const MessageRow = memo(function MessageRow({
  event,
  agentName,
}: {
  event: ParsedEvent;
  agentName: string;
}) {
  const isUser = event.role === "user";
  const hasText = event.text.trim().length > 0;
  const hasTools = event.toolCalls.length > 0 || event.toolResponses.length > 0;
  if (!hasText && !hasTools) return null;

  if (isUser) {
    return (
      <div className="flex flex-col items-end gap-1 py-3">
        <div className="max-w-[80%] rounded-lg rounded-tr-sm bg-secondary px-3.5 py-2.5 text-[0.9rem] leading-relaxed text-secondary-foreground">
          <ReactMarkdown remarkPlugins={MARKDOWN_REMARK_PLUGINS} components={MARKDOWN_COMPONENTS}>
            {event.text}
          </ReactMarkdown>
        </div>
      </div>
    );
  }

  return (
    <div className="flex gap-3 py-3">
      <AgentAvatar name={event.author || agentName} size="sm" className="mt-0.5" />
      <div className="min-w-0 flex-1">
        <div className="mb-0.5 flex items-center gap-2">
          <span className="text-sm font-medium">{event.author || agentName}</span>
          {event.timestamp && (
            <span className="text-xs text-muted-foreground">
              {new Date(event.timestamp).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
            </span>
          )}
        </div>

        {event.toolCalls.map((tc: ToolCallSummary, i: number) => (
          <ToolBlock key={`call-${i}`} kind="call" name={tc.name} preview={tc.argsPreview} />
        ))}
        {event.toolResponses.map((tr: ToolResponseSummary, i: number) => (
          <ToolBlock key={`resp-${i}`} kind="response" name={tr.name} preview={tr.responsePreview} />
        ))}

        {hasText && (
          <div className="text-[0.9rem] leading-relaxed text-foreground">
            <ReactMarkdown remarkPlugins={MARKDOWN_REMARK_PLUGINS} components={MARKDOWN_COMPONENTS}>
              {event.text}
            </ReactMarkdown>
          </div>
        )}

        {event.traceUrl && (
          <a
            href={event.traceUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-1 inline-flex items-center gap-1 text-[0.7rem] text-muted-foreground hover:text-foreground"
          >
            <ExternalLink className="size-2.5" /> trace
          </a>
        )}
      </div>
    </div>
  );
});
