import { ConnectError } from "@connectrpc/connect";
import type { MessageInitShape } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { AgentService, SessionService } from "@/gen/agents/v1/agent_service_pb";
import type { InputPartSchema } from "@/gen/agents/v1/content_pb";
import { ApiError } from "./client";
import { makeClient } from "./transport";

type InputPartInit = MessageInitShape<typeof InputPartSchema>;

const sessionClient = makeClient(SessionService);
const agentClient = makeClient(AgentService);

export interface SendChatParams {
  agent_name: string;
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

export async function replySession(params: SendChatParams): Promise<ReplySessionResponse> {
  const res = await sessionClient.replySession({
    agentName: params.agent_name,
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
        agentName: params.agent_name,
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
          });
          break;
        }
        case "textDelta": {
          const v = msg.event.value;
          handlers.onTextDelta?.({
            invocation_id: v.invocationId,
            session_id: v.sessionId,
            agent_name: v.agentName,
            text_delta: v.text,
          });
          break;
        }
        case "runEvent": {
          const v = msg.event.value;
          handlers.onAgentEvent?.({
            invocation_id: v.invocationId,
            session_id: v.sessionId,
            agent_name: v.agentName,
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
          });
          break;
        }
        case "final": {
          const v = msg.event.value;
          finalPayload = {
            invocation_id: v.invocationId,
            session_id: v.sessionId,
            agent_name: v.agentName,
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
      error: message,
    });
    throw new ApiError("stream_error", message);
  }

  return finalPayload;
}
