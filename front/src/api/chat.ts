import { Code, ConnectError } from "@connectrpc/connect";
import type { MessageInitShape } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import {
  AgentService,
  type Invocation,
  InvocationStatus,
  SessionService,
  type StreamAgentRunEvent,
} from "@/gen/agents/v1/agent_service_pb";
import type { InputPart, InputPartSchema } from "@/gen/agents/v1/content_pb";
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

export interface SubmitAgentInvocationParams {
  request_id: string;
  agent_id: string;
  session_id?: string;
  message: string;
  model_override?: string;
  parts?: InputPartInit[];
}

export interface SubmitAgentInvocationResult {
  session_id: string;
  invocation_id: string;
  status: InvocationStatus;
  session_created: boolean;
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

export interface ChatStreamHandlers {
  onStarted?: (payload: ChatStreamPayload) => void;
  onAgentEvent?: (payload: ChatStreamPayload) => void;
  onTextDelta?: (payload: ChatStreamPayload) => void;
  onFinal?: (payload: ChatStreamPayload) => void;
  onError?: (payload: ChatStreamPayload) => void;
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

export function isTerminalInvocationStatus(status: InvocationStatus): boolean {
  return (
    status === InvocationStatus.SUCCEEDED ||
    status === InvocationStatus.FAILED ||
    status === InvocationStatus.CANCELLED
  );
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

export async function submitAgentInvocation(
  params: SubmitAgentInvocationParams,
): Promise<SubmitAgentInvocationResult> {
  const res = await agentClient.submitAgentInvocation({
    requestId: params.request_id,
    agentId: params.agent_id,
    sessionId: params.session_id ?? "",
    message: params.message,
    modelOverride: params.model_override ?? "",
    parts: params.parts,
  });
  return {
    session_id: res.sessionId,
    invocation_id: res.invocationId,
    status: res.status,
    session_created: res.sessionCreated,
  };
}

export async function getAgentInvocation(invocationId: string): Promise<Invocation> {
  const res = await agentClient.getAgentInvocation({ invocationId });
  if (!res.invocation) throw new Error("invocation not found");
  return res.invocation;
}

async function findInvocationForSession(sessionId: string, latest: boolean): Promise<Invocation | null> {
  try {
    const res = await agentClient.getAgentInvocation({ sessionId, latest });
    return res.invocation ?? null;
  } catch (err) {
    if (err instanceof ConnectError && err.code === Code.NotFound) return null;
    throw err;
  }
}

// getActiveInvocationForSession returns the session's QUEUED/RUNNING
// invocation, or null when the session is idle. Clients re-entering a
// session call this after loading persisted events to decide whether to
// attach a watch stream.
export function getActiveInvocationForSession(sessionId: string): Promise<Invocation | null> {
  return findInvocationForSession(sessionId, false);
}

// getLatestInvocationForSession returns the session's most recent invocation
// regardless of status, or null when the session has none. Used to render a
// failed or stopped last turn inline after reload or navigation.
export function getLatestInvocationForSession(sessionId: string): Promise<Invocation | null> {
  return findInvocationForSession(sessionId, true);
}

export interface InvocationInput {
  invocation: Invocation;
  parts: InputPart[];
}

// getAgentInvocationInput fetches an invocation together with its retained
// Input Parts. Parts are kept for FAILED and CANCELLED invocations so the
// original input can be restored into the composer for explicit review and
// resubmission; a successful run's parts were cleaned up and come back empty.
export async function getAgentInvocationInput(invocationId: string): Promise<InvocationInput> {
  const res = await agentClient.getAgentInvocation({ invocationId, includeInputParts: true });
  if (!res.invocation) throw new Error("invocation not found");
  return { invocation: res.invocation, parts: res.inputParts };
}

export interface WatchChatHandlers {
  /** Authoritative invocation snapshots (first frame, RUNNING transition). */
  onState?: (inv: Invocation) => void;
  onAgentEvent?: (payload: ChatStreamPayload) => void;
  onTextDelta?: (payload: ChatStreamPayload) => void;
}

// watchChatInvocation attaches a read-only observer stream to a running
// invocation. It never owns the run: aborting the signal only detaches this
// observer. Resolves with the terminal Invocation, or null when the stream
// ended without one (e.g. this observer lagged and was disconnected — reload
// persisted state and re-attach). Rejects with AbortError when the signal
// aborts, mirroring streamChat.
export async function watchChatInvocation(
  invocationId: string,
  handlers: WatchChatHandlers,
  signal?: AbortSignal,
): Promise<Invocation | null> {
  try {
    const stream = agentClient.watchAgentInvocation({ invocationId }, { signal });
    for await (const msg of stream) {
      switch (msg.event.case) {
        case "state": {
          const inv = msg.event.value.invocation;
          if (!inv) break;
          handlers.onState?.(inv);
          if (isTerminalInvocationStatus(inv.status)) return inv;
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
