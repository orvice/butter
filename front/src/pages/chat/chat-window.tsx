import { memo, useEffect, useMemo, useRef, useState, type ComponentProps } from "react";
import ReactMarkdown, { type Components } from "react-markdown";
import remarkGfm from "remark-gfm";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useLayoutDensity } from "@/hooks/use-layout-density";
import { cn } from "@/lib/utils";
import { parseSessionEvent, parseSessionEvents, type ParsedEvent } from "@/lib/session-events";
import { useLiveSession, useReplySession } from "@/api/sessions";
import { cancelAgentInvocation, streamChat, type ChatStreamPayload } from "@/api/chat";
import { Bot, Send, User as UserIcon, Wrench, ExternalLink, Loader2, Square } from "lucide-react";
import { toast } from "sonner";
import type { SessionInfo } from "@/types/api";

const APP_NAME = "web-chat";
const EMPTY_STREAMING_EVENTS: ParsedEvent[] = [];
const MARKDOWN_REMARK_PLUGINS = [remarkGfm];

interface ChatWindowProps {
  session: SessionInfo | null;
  userId: string;
  agentName: string | null;
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

export function ChatWindow({ session, userId, agentName }: ChatWindowProps) {
  const sessionId = session?.session_id ?? "";
  const { isCompact } = useLayoutDensity();
  const [draft, setDraft] = useState("");
  const [runState, setRunState] = useState<ChatRunState>(() => emptyChatRunState(""));
  const abortRef = useRef<AbortController | null>(null);
  const activeRunIdRef = useRef<string | null>(null);
  const scrollRef = useRef<HTMLDivElement | null>(null);

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
      <div className="flex flex-1 flex-col items-center justify-center text-sm text-muted-foreground">
        <Bot className={cn("mb-2", isCompact ? "h-6 w-6" : "h-8 w-8")} />
        Select a chat on the left or start a new one.
      </div>
    );
  }

  async function handleSend() {
    const text = draft.trim();
    if (!text || !agentName || pending) return;
    const runId = newRunId();
    abortRef.current?.abort();
    activeRunIdRef.current = runId;
    setDraft("");
    setRunState({
      runId,
      sessionId,
      pending: true,
      pendingBaseEventIds: new Set(persistedEvents.map((evt) => evt.eventId)),
      pendingUserMessage: text,
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
          });
        } catch (fallbackErr) {
          toast.error(fallbackErr instanceof Error ? fallbackErr.message : "Failed to send message");
          setDraft(text);
        }
      } else {
        toast.error(err instanceof Error ? err.message : "Failed to send message");
      }
    } finally {
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
    if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      void handleSend();
    }
  }

  return (
    <div className="flex h-full flex-1 flex-col">
      <div className={cn("flex items-center justify-between border-b px-3 sm:px-4", isCompact ? "py-2" : "py-3")}>
        <div className="flex min-w-0 items-center gap-2">
          <Bot className={cn("text-muted-foreground", isCompact ? "h-3.5 w-3.5" : "h-4 w-4")} />
          <div className="min-w-0">
            <div className={cn("font-semibold", isCompact ? "text-xs" : "text-sm")}>{agentName ?? "Unknown agent"}</div>
            <div className="truncate font-mono text-[10px] leading-tight text-muted-foreground">{sessionId}</div>
          </div>
        </div>
      </div>

      <div
        ref={scrollRef}
        className={cn(
          "flex-1 overflow-y-auto px-3 sm:px-4",
          isCompact ? "space-y-2 py-2.5" : "space-y-3 py-4",
        )}
      >
        {liveQuery.isLoading ? (
          <>
            <Skeleton className={cn("w-2/3", isCompact ? "h-12" : "h-16")} />
            <Skeleton className={cn("ml-auto w-1/2", isCompact ? "h-12" : "h-16")} />
          </>
        ) : events.length === 0 ? (
          <div className="flex h-full items-center justify-center text-xs text-muted-foreground">
            No messages yet.
          </div>
        ) : (
          events.map((evt) => <MessageBubble key={evt.eventId} event={evt} isCompact={isCompact} />)
        )}
        {pending ? (
          <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
            <Loader2 className="h-3 w-3 animate-spin" /> Agent is thinking...
          </div>
        ) : null}
      </div>

      <div className={cn("border-t bg-background", isCompact ? "p-2" : "p-3")}>
        <div className={cn("flex items-end", isCompact ? "gap-1.5" : "gap-2")}>
          <Textarea
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={
              agentName
                ? "Message the agent..."
                : "This chat is missing an agent reference; cannot send."
            }
            disabled={!agentName || pending}
            rows={2}
            className={cn(
              "flex-1 resize-none",
              isCompact && "min-h-10 rounded-md px-2 py-1.5 text-sm leading-5",
            )}
          />
          <Button
            variant={pending ? "secondary" : "default"}
            size={isCompact ? "sm" : "default"}
            onClick={() => pending ? void handleStop() : void handleSend()}
            disabled={!agentName || (!pending && draft.trim().length === 0)}
          >
            {pending ? (
              <><Square className="mr-1 h-3 w-3" /> Stop</>
            ) : (
              <><Send className="mr-1 h-3 w-3" /> Send</>
            )}
          </Button>
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

function buildMarkdownComponents(isCompact: boolean): Components {
  return {
    a: MarkdownLink,
    code: MarkdownCode,
    pre: ({ children }) => <MarkdownPre isCompact={isCompact}>{children}</MarkdownPre>,
    table: ({ children }) => <MarkdownTable isCompact={isCompact}>{children}</MarkdownTable>,
    th: MarkdownTableHeader,
    td: MarkdownTableCell,
    p: ({ children }) => <p className={cn(isCompact ? "mb-1.5" : "mb-2", "last:mb-0")}>{children}</p>,
    ul: ({ children }) => (
      <ul className={cn(isCompact ? "mb-1.5 space-y-0.5" : "mb-2 space-y-1", "list-disc pl-5 last:mb-0")}>
        {children}
      </ul>
    ),
    ol: ({ children }) => (
      <ol className={cn(isCompact ? "mb-1.5 space-y-0.5" : "mb-2 space-y-1", "list-decimal pl-5 last:mb-0")}>
        {children}
      </ol>
    ),
    li: ({ children }) => <li className="pl-1">{children}</li>,
    blockquote: ({ children }) => (
      <blockquote className={cn(isCompact ? "mb-1.5" : "mb-2", "border-l-2 border-current/30 pl-3 italic opacity-90 last:mb-0")}>
        {children}
      </blockquote>
    ),
    hr: () => <hr className={cn(isCompact ? "my-2" : "my-3", "border-current/20")} />,
    h1: ({ children }) => <h1 className={cn(isCompact ? "mb-1.5 text-base" : "mb-2 text-lg", "font-semibold last:mb-0")}>{children}</h1>,
    h2: ({ children }) => <h2 className={cn(isCompact ? "mb-1.5 text-sm" : "mb-2 text-base", "font-semibold last:mb-0")}>{children}</h2>,
    h3: ({ children }) => <h3 className={cn(isCompact ? "mb-1 text-sm" : "mb-2 text-sm", "font-semibold last:mb-0")}>{children}</h3>,
  };
}

// Two stable component maps cover the only variable (layout density), so the
// closures aren't rebuilt on every streaming token.
const MARKDOWN_COMPONENTS_COMPACT = buildMarkdownComponents(true);
const MARKDOWN_COMPONENTS_REGULAR = buildMarkdownComponents(false);

function MarkdownMessage({ text, isUser, isCompact }: { text: string; isUser: boolean; isCompact: boolean }) {
  return (
    <div
      className={cn(
        "rounded-lg text-sm",
        isCompact ? "px-2.5 py-1.5 leading-6" : "px-3 py-2 leading-relaxed",
        isUser ? "bg-primary text-primary-foreground" : "bg-muted text-foreground",
      )}
    >
      <ReactMarkdown
        remarkPlugins={MARKDOWN_REMARK_PLUGINS}
        components={isCompact ? MARKDOWN_COMPONENTS_COMPACT : MARKDOWN_COMPONENTS_REGULAR}
      >
        {text}
      </ReactMarkdown>
    </div>
  );
}

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
      <code className="rounded bg-background/60 px-1 py-0.5 font-mono text-[0.85em] text-foreground">
        {children}
      </code>
    );
  }

  return <code className={cn("font-mono text-xs", className)}>{children}</code>;
}

function MarkdownPre({ children, isCompact }: ComponentProps<"pre"> & { isCompact?: boolean }) {
  return (
    <pre className={cn("overflow-x-auto rounded-md bg-background/80 text-foreground last:mb-0", isCompact ? "mb-1.5 p-2" : "mb-2 p-3")}>
      {children}
    </pre>
  );
}

function MarkdownTable({ children, isCompact }: ComponentProps<"table"> & { isCompact?: boolean }) {
  return (
    <div className={cn("overflow-x-auto last:mb-0", isCompact ? "mb-1.5" : "mb-2")}>
      <table className="w-full border-collapse text-left text-xs">{children}</table>
    </div>
  );
}

function MarkdownTableHeader({ children }: ComponentProps<"th">) {
  return <th className="border border-current/20 px-2 py-1 font-semibold">{children}</th>;
}

function MarkdownTableCell({ children }: ComponentProps<"td">) {
  return <td className="border border-current/20 px-2 py-1 align-top">{children}</td>;
}

const MessageBubble = memo(function MessageBubble({ event, isCompact }: { event: ParsedEvent; isCompact: boolean }) {
  const isUser = event.role === "user";
  const hasText = event.text.trim().length > 0;
  const hasTools = event.toolCalls.length > 0 || event.toolResponses.length > 0;
  if (!hasText && !hasTools) return null;

  return (
    <div className={cn("flex", isCompact ? "gap-1.5" : "gap-2", isUser ? "justify-end" : "justify-start")}>
      {!isUser ? (
        <div
          className={cn(
            "mt-1 flex shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground",
            isCompact ? "h-5 w-5" : "h-6 w-6",
          )}
        >
          <Bot className={cn(isCompact ? "h-2.5 w-2.5" : "h-3 w-3")} />
        </div>
      ) : null}
      <div className={cn(isCompact ? "max-w-[94%] space-y-1 sm:max-w-[86%]" : "max-w-[88%] space-y-1.5 sm:max-w-[75%]", isUser && "items-end")}>
        {hasText ? (
          <MarkdownMessage text={event.text} isUser={isUser} isCompact={isCompact} />
        ) : null}
        {event.toolCalls.map((tc, i) => (
          <Card key={`call-${i}`} className="border-dashed">
            <CardContent className={cn("flex items-start gap-2 text-xs", isCompact ? "p-1.5" : "p-2")}>
              <Wrench className="mt-0.5 h-3 w-3 shrink-0 text-muted-foreground" />
              <div className="min-w-0">
                <div className="font-medium">Tool call: {tc.name}</div>
                {tc.argsPreview ? (
                  <div className="truncate font-mono text-[10px] text-muted-foreground">
                    {tc.argsPreview}
                  </div>
                ) : null}
              </div>
            </CardContent>
          </Card>
        ))}
        {event.toolResponses.map((tr, i) => (
          <Card key={`resp-${i}`} className="border-dashed">
            <CardContent className={cn("flex items-start gap-2 text-xs", isCompact ? "p-1.5" : "p-2")}>
              <Wrench className="mt-0.5 h-3 w-3 shrink-0 text-muted-foreground" />
              <div className="min-w-0">
                <div className="font-medium">Tool response: {tr.name}</div>
                {tr.responsePreview ? (
                  <div className="truncate font-mono text-[10px] text-muted-foreground">
                    {tr.responsePreview}
                  </div>
                ) : null}
              </div>
            </CardContent>
          </Card>
        ))}
        {event.traceUrl ? (
          <a
            href={event.traceUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center text-[10px] text-muted-foreground hover:text-primary"
          >
            <ExternalLink className="mr-1 h-2.5 w-2.5" /> trace
          </a>
        ) : null}
      </div>
      {isUser ? (
        <div
          className={cn(
            "mt-1 flex shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground",
            isCompact ? "h-5 w-5" : "h-6 w-6",
          )}
        >
          <UserIcon className={cn(isCompact ? "h-2.5 w-2.5" : "h-3 w-3")} />
        </div>
      ) : null}
    </div>
  );
});
