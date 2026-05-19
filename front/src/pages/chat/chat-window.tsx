import { useEffect, useMemo, useRef, useState, type ComponentProps } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { parseSessionEvent, parseSessionEvents, type ParsedEvent } from "@/lib/session-events";
import { useLiveSession, useReplySession } from "@/api/sessions";
import { cancelAgentInvocation, streamChat, type ChatStreamPayload } from "@/api/chat";
import { Bot, Send, User as UserIcon, Wrench, ExternalLink, Loader2, Square } from "lucide-react";
import { toast } from "sonner";
import type { SessionInfo } from "@/types/api";

const APP_NAME = "web-chat";

interface ChatWindowProps {
  session: SessionInfo | null;
  userId: string;
  agentName: string | null;
}

export function ChatWindow({ session, userId, agentName }: ChatWindowProps) {
  const [draft, setDraft] = useState("");
  const [pending, setPending] = useState(false);
  const [pendingUserMessage, setPendingUserMessage] = useState<string | null>(null);
  const [streamingEvents, setStreamingEvents] = useState<ParsedEvent[]>([]);
  const [streamingResponse, setStreamingResponse] = useState("");
  const [invocationId, setInvocationId] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  const scrollRef = useRef<HTMLDivElement | null>(null);

  const sessionId = session?.session_id ?? "";
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
    const out = [...persistedEvents];
    if (optimisticUserEvent) out.push(optimisticUserEvent);
    out.push(...streamingEvents);
    if (streamingResponse.trim()) {
      out.push(makeOptimisticTextEvent("streaming-assistant", "assistant", streamingResponse));
    }
    return out;
  }, [persistedEvents, optimisticUserEvent, streamingEvents, streamingResponse]);

  useEffect(() => {
    setStreamingEvents([]);
    setStreamingResponse("");
    setInvocationId(null);
    abortRef.current?.abort();
    abortRef.current = null;
  }, [sessionId]);

  useEffect(() => {
    const node = scrollRef.current;
    if (node) node.scrollTop = node.scrollHeight;
  }, [events.length, pending]);

  if (!session) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center text-sm text-muted-foreground">
        <Bot className="mb-2 h-8 w-8" />
        Select a chat on the left or start a new one.
      </div>
    );
  }

  async function handleSend() {
    const text = draft.trim();
    if (!text || !agentName || pending) return;
    setDraft("");
    setPending(true);
    setPendingUserMessage(text);
    setStreamingEvents([]);
    setStreamingResponse("");
    setInvocationId(null);

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
            if (payload.invocation_id) setInvocationId(payload.invocation_id);
          },
          onAgentEvent: (payload) => {
            const event = payloadToParsedEvent(payload);
            if (event) {
              setStreamingEvents((prev) => [...prev, event]);
            }
          },
          onFinal: (payload) => {
            setStreamingResponse(payload.response ?? "");
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
      setPending(false);
      setPendingUserMessage(null);
      setStreamingEvents([]);
      setStreamingResponse("");
      setInvocationId(null);
      abortRef.current = null;
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
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      void handleSend();
    }
  }

  return (
    <div className="flex h-full flex-1 flex-col">
      <div className="flex items-center justify-between border-b px-3 py-3 sm:px-4">
        <div className="flex min-w-0 items-center gap-2">
          <Bot className="h-4 w-4 text-muted-foreground" />
          <div className="min-w-0">
            <div className="text-sm font-semibold">{agentName ?? "Unknown agent"}</div>
            <div className="truncate font-mono text-[10px] text-muted-foreground">{sessionId}</div>
          </div>
        </div>
      </div>

      <div ref={scrollRef} className="flex-1 space-y-3 overflow-y-auto px-3 py-4 sm:px-4">
        {liveQuery.isLoading ? (
          <>
            <Skeleton className="h-16 w-2/3" />
            <Skeleton className="ml-auto h-16 w-1/2" />
          </>
        ) : events.length === 0 ? (
          <div className="flex h-full items-center justify-center text-xs text-muted-foreground">
            No messages yet.
          </div>
        ) : (
          events.map((evt) => <MessageBubble key={evt.eventId} event={evt} />)
        )}
        {pending ? (
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Loader2 className="h-3 w-3 animate-spin" /> Agent is thinking...
          </div>
        ) : null}
      </div>

      <div className="border-t bg-background p-3">
        <div className="flex items-end gap-2">
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
            className="flex-1 resize-none"
          />
          <Button
            variant={pending ? "secondary" : "default"}
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

function payloadToParsedEvent(payload: ChatStreamPayload): ParsedEvent | null {
  const evt = payload.event;
  if (!evt?.event_id) return null;
  return parseSessionEvent({
    event_id: evt.event_id,
    invocation_id: evt.invocation_id,
    author: evt.author,
    branch: evt.branch,
    content_json: evt.content_json,
    timestamp: evt.timestamp,
    trace_id: evt.invocation_id,
  });
}

function isAbortError(err: unknown): boolean {
  return err instanceof DOMException && err.name === "AbortError";
}

function MarkdownMessage({ text, isUser }: { text: string; isUser: boolean }) {
  return (
    <div
      className={cn(
        "rounded-lg px-3 py-2 text-sm leading-relaxed",
        isUser ? "bg-primary text-primary-foreground" : "bg-muted text-foreground",
      )}
    >
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          a: MarkdownLink,
          code: MarkdownCode,
          pre: MarkdownPre,
          table: MarkdownTable,
          th: MarkdownTableHeader,
          td: MarkdownTableCell,
          p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
          ul: ({ children }) => <ul className="mb-2 list-disc space-y-1 pl-5 last:mb-0">{children}</ul>,
          ol: ({ children }) => <ol className="mb-2 list-decimal space-y-1 pl-5 last:mb-0">{children}</ol>,
          li: ({ children }) => <li className="pl-1">{children}</li>,
          blockquote: ({ children }) => (
            <blockquote className="mb-2 border-l-2 border-current/30 pl-3 italic opacity-90 last:mb-0">
              {children}
            </blockquote>
          ),
          hr: () => <hr className="my-3 border-current/20" />,
          h1: ({ children }) => <h1 className="mb-2 text-lg font-semibold last:mb-0">{children}</h1>,
          h2: ({ children }) => <h2 className="mb-2 text-base font-semibold last:mb-0">{children}</h2>,
          h3: ({ children }) => <h3 className="mb-2 text-sm font-semibold last:mb-0">{children}</h3>,
        }}
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

function MarkdownPre({ children }: ComponentProps<"pre">) {
  return (
    <pre className="mb-2 overflow-x-auto rounded-md bg-background/80 p-3 text-foreground last:mb-0">
      {children}
    </pre>
  );
}

function MarkdownTable({ children }: ComponentProps<"table">) {
  return (
    <div className="mb-2 overflow-x-auto last:mb-0">
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

function MessageBubble({ event }: { event: ParsedEvent }) {
  const isUser = event.role === "user";
  const hasText = event.text.trim().length > 0;
  const hasTools = event.toolCalls.length > 0 || event.toolResponses.length > 0;
  if (!hasText && !hasTools) return null;

  return (
    <div className={cn("flex gap-2", isUser ? "justify-end" : "justify-start")}>
      {!isUser ? (
        <div className="mt-1 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground">
          <Bot className="h-3 w-3" />
        </div>
      ) : null}
      <div className={cn("max-w-[88%] space-y-1.5 sm:max-w-[75%]", isUser && "items-end")}>
        {hasText ? (
          <MarkdownMessage text={event.text} isUser={isUser} />
        ) : null}
        {event.toolCalls.map((tc, i) => (
          <Card key={`call-${i}`} className="border-dashed">
            <CardContent className="flex items-start gap-2 p-2 text-xs">
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
            <CardContent className="flex items-start gap-2 p-2 text-xs">
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
        <div className="mt-1 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground">
          <UserIcon className="h-3 w-3" />
        </div>
      ) : null}
    </div>
  );
}
