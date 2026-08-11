import { Code, ConnectError } from "@connectrpc/connect";
import type { MessageInitShape } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import {
  AgentService,
  InvocationStatus,
  SessionService,
  type Invocation,
  type StreamAgentRunEvent,
} from "@/gen/agents/v1/agent_service_pb";
import type { InputPartSchema } from "@/gen/agents/v1/content_pb";
import { ApiError } from "./client";
import { makeClient } from "./transport";

type InputPartInit = MessageInitShape<typeof InputPartSchema>;

const sessionClient = makeClient(SessionService);
const agentClient = makeClient(AgentService);

export interface SendChatParams {
  /** Legacy agent name; fallback for agents without an agent_id. */
  agent_name?: string;
  /** Immutable agent_id of the agent to invoke. Preferred over agent_name. */
  agent_id?: string;
  app_name: string;
  user_id: string;
  session_id: string;
  message: string;
  model_override?: string;
  // Multimodal input. When non-empty the server uses `parts` and ignores
  // `message` (see docs/api.md, StreamAgent).
  parts?: InputPartInit[];
}

export interface ReplySessionResponse {
  response: string;
}

// ChatStreamRunEvent mirrors the legacy SSE payload shape so chat-window.tsx
// can keep parsing events into ParsedEvent via the same path it always has.
// The fields come from the proto StreamAgentRunEvent message.
export interface ChatStreamRunEvent {
  event_id?: string;
  invocation_id?: string;
  author?: string;
  branch?: string;
  partial?: boolean;
  final_response?: boolean;
  content_json?: string;
  timestamp?: string;
}

export interface ChatStreamPayload {
  invocation_id?: string;
  session_id?: string;
  agent_name?: string;
  agent_id?: string;
  response?: string;
  text_delta?: string;
  error?: string;
  event?: ChatStreamRunEvent;
}

// runEventPayload maps a proto StreamAgentRunEvent (shared by StreamAgent and
// WatchAgentInvocation frames) into the legacy callback payload shape.
function runEventPayload(v: StreamAgentRunEvent): ChatStreamPayload {
  return {
    invocation_id: v.invocationId,
    session_id: v.sessionId,
    agent_name: v.agentName,
    agent_id: v.agentId,
    event: {
      event_id: v.eventId,
      invocation_id: v.invocationId,
      author: v.author,
      branch: v.branch,
      partial: v.partial,
      final_response: v.finalResponse,
      content_json: v.contentJson,
      timestamp: v.timestamp ? timestampDate(v.timestamp).toISOString() : undefined,
    },
  };
}

export interface ChatStreamHandlers {
  onStarted?: (payload: ChatStreamPayload) => void;
  onAgentEvent?: (payload: ChatStreamPayload) => void;
  onTextDelta?: (payload: ChatStreamPayload) => void;
  onFinal?: (payload: ChatStreamPayload) => void;
  onError?: (payload: ChatStreamPayload) => void;
}

export async function replySession(params: SendChatParams): Promise<ReplySessionResponse> {
  const res = await sessionClient.replySession({
    agentName: params.agent_name ?? "",
    agentId: params.agent_id ?? "",
    appName: params.app_name,
    userId: params.user_id,
    sessionId: params.session_id,
    message: params.message,
    modelOverride: params.model_override ?? "",
    parts: params.parts,
  });
  return { response: res.response };
}

export async function cancelAgentInvocation(invocationId: string): Promise<{ cancelled: boolean }> {
  const res = await agentClient.cancelAgentInvocation({ invocationId });
  return { cancelled: res.cancelled };
}

// --- Asynchronous invocations (submit / lookup / observe) ---

export type InvocationStatusName =
  | "unspecified"
  | "queued"
  | "running"
  | "succeeded"
  | "failed"
  | "cancelled";

export function invocationStatusName(status: InvocationStatus): InvocationStatusName {
  switch (status) {
    case InvocationStatus.QUEUED:
      return "queued";
    case InvocationStatus.RUNNING:
      return "running";
    case InvocationStatus.SUCCEEDED:
      return "succeeded";
    case InvocationStatus.FAILED:
      return "failed";
    case InvocationStatus.CANCELLED:
      return "cancelled";
    default:
      return "unspecified";
  }
}

export function isTerminalInvocationStatus(status: InvocationStatusName): boolean {
  return status === "succeeded" || status === "failed" || status === "cancelled";
}

export interface InvocationSummary {
  invocation_id: string;
  session_id: string;
  status: InvocationStatusName;
  output?: string;
  error?: string;
}

function invocationSummary(inv: Invocation): InvocationSummary {
  return {
    invocation_id: inv.id,
    session_id: inv.sessionId,
    status: invocationStatusName(inv.status),
    output: inv.output || undefined,
    error: inv.error || undefined,
  };
}

export interface SubmitChatParams {
  /** Client-generated idempotency key. */
  request_id: string;
  agent_id: string;
  session_id?: string;
  message: string;
  parts?: InputPartInit[];
  model_override?: string;
}

export interface SubmitChatResult {
  session_id: string;
  invocation_id: string;
  status: InvocationStatusName;
  session_created: boolean;
}

// submitChatInvocation durably accepts one chat turn as an asynchronous
// Invocation. The server returns as soon as the input is durable; the agent
// runs independently of this browser connection.
export async function submitChatInvocation(params: SubmitChatParams): Promise<SubmitChatResult> {
  const res = await agentClient.submitAgentInvocation({
    requestId: params.request_id,
    agentId: params.agent_id,
    sessionId: params.session_id ?? "",
    message: params.message,
    parts: params.parts,
    modelOverride: params.model_override ?? "",
  });
  return {
    session_id: res.sessionId,
    invocation_id: res.invocationId,
    status: invocationStatusName(res.status),
    session_created: res.sessionCreated,
  };
}

// getInvocation returns the authoritative state of one invocation.
export async function getInvocation(invocationId: string): Promise<InvocationSummary | null> {
  try {
    const res = await agentClient.getAgentInvocation({ invocationId });
    return res.invocation ? invocationSummary(res.invocation) : null;
  } catch (err) {
    if (err instanceof ConnectError && err.code === Code.NotFound) return null;
    throw err;
  }
}

// getActiveInvocationForSession returns the session's QUEUED/RUNNING
// invocation, or null when the session is idle. Clients re-entering a
// session call this after loading persisted events to decide whether to
// attach a watch stream.
export async function getActiveInvocationForSession(sessionId: string): Promise<InvocationSummary | null> {
  try {
    const res = await agentClient.getAgentInvocation({ sessionId });
    return res.invocation ? invocationSummary(res.invocation) : null;
  } catch (err) {
    if (err instanceof ConnectError && err.code === Code.NotFound) return null;
    throw err;
  }
}

export interface WatchChatHandlers {
  /** Authoritative invocation snapshots (first frame, RUNNING, terminal). */
  onState?: (inv: InvocationSummary) => void;
  onAgentEvent?: (payload: ChatStreamPayload) => void;
  onTextDelta?: (payload: ChatStreamPayload) => void;
}

// watchChatInvocation attaches a read-only observer stream to a running
// invocation. It never owns the run: aborting the signal only detaches this
// observer. Resolves with the terminal invocation state, or null when the
// stream ended without one (e.g. this observer lagged and was disconnected —
// reload persisted state and re-attach). Rejects with AbortError when the
// signal aborts, mirroring streamChat.
export async function watchChatInvocation(
  invocationId: string,
  handlers: WatchChatHandlers,
  signal?: AbortSignal,
): Promise<InvocationSummary | null> {
  try {
    const stream = agentClient.watchAgentInvocation({ invocationId }, { signal });
    for await (const msg of stream) {
      switch (msg.event.case) {
        case "state": {
          const inv = msg.event.value.invocation;
          if (!inv) break;
          const summary = invocationSummary(inv);
          handlers.onState?.(summary);
          if (isTerminalInvocationStatus(summary.status)) return summary;
          break;
        }
        case "textDelta": {
          const v = msg.event.value;
          handlers.onTextDelta?.({
            invocation_id: v.invocationId,
            session_id: v.sessionId,
            agent_name: v.agentName,
            agent_id: v.agentId,
            text_delta: v.text,
          });
          break;
        }
        case "runEvent": {
          handlers.onAgentEvent?.(runEventPayload(msg.event.value));
          break;
        }
      }
    }
    return null;
  } catch (err) {
    if (signal?.aborted) {
      throw new DOMException("aborted", "AbortError");
    }
    if (err instanceof ConnectError && err.code === Code.ResourceExhausted) {
      // Lagged observer: the run is unaffected; the caller reconciles from
      // persisted state and re-attaches.
      return null;
    }
    const message = err instanceof ConnectError ? err.message : err instanceof Error ? err.message : "Watch stream failed";
    throw new ApiError("watch_error", message);
  }
}

// streamChat invokes the AgentService.StreamAgent server-stream and
// dispatches each event to the matching handler, mirroring the callback
// shape that the chat window used during the SSE era. Returns the final
// payload (or null if the stream ended without one). The caller can pass
// an AbortSignal to cancel the stream cleanly.
export async function streamChat(
  params: SendChatParams,
  handlers: ChatStreamHandlers,
  signal?: AbortSignal,
): Promise<ChatStreamPayload | null> {
  let finalPayload: ChatStreamPayload | null = null;

  try {
    const stream = agentClient.streamAgent(
      {
        agentName: params.agent_name ?? "",
        agentId: params.agent_id ?? "",
        appName: params.app_name,
        userId: params.user_id,
        sessionId: params.session_id,
        message: params.message,
        modelOverride: params.model_override ?? "",
        parts: params.parts,
      },
      { signal },
    );

    for await (const msg of stream) {
      switch (msg.event.case) {
        case "started": {
          const v = msg.event.value;
          handlers.onStarted?.({
            invocation_id: v.invocationId,
            session_id: v.sessionId,
            agent_name: v.agentName,
            agent_id: v.agentId,
          });
          break;
        }
        case "textDelta": {
          const v = msg.event.value;
          handlers.onTextDelta?.({
            invocation_id: v.invocationId,
            session_id: v.sessionId,
            agent_name: v.agentName,
            agent_id: v.agentId,
            text_delta: v.text,
          });
          break;
        }
        case "runEvent": {
          handlers.onAgentEvent?.(runEventPayload(msg.event.value));
          break;
        }
        case "final": {
          const v = msg.event.value;
          finalPayload = {
            invocation_id: v.invocationId,
            session_id: v.sessionId,
            agent_name: v.agentName,
            agent_id: v.agentId,
            response: v.response,
          };
          handlers.onFinal?.(finalPayload);
          break;
        }
      }
    }
  } catch (err) {
    if (signal?.aborted) {
      // Re-throw as DOMException so existing isAbortError() check in
      // chat-window.tsx still fires.
      throw new DOMException("aborted", "AbortError");
    }
    const message = err instanceof ConnectError ? err.message : err instanceof Error ? err.message : "Chat stream failed";
    handlers.onError?.({
      session_id: params.session_id,
      agent_name: params.agent_name,
      agent_id: params.agent_id,
      error: message,
    });
    throw new ApiError("stream_error", message);
  }

  return finalPayload;
}
