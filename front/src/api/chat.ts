import { ApiError, BASE_URL, authHeaders } from "./client";
import { AgentService } from "@/gen/agents/v1/agent_service_pb";
import { SessionService } from "@/gen/agents/v1/agent_service_pb";
import { TOKEN_KEY } from "@/lib/constants";
import { makeClient } from "./transport";

const sessionClient = makeClient(SessionService);
const agentClient = makeClient(AgentService);

export interface SendChatParams {
  agent_name: string;
  app_name: string;
  user_id: string;
  session_id: string;
  message: string;
  model_override?: string;
}

export interface ReplySessionResponse {
  response: string;
}

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

export type ChatStreamEvent = "invocation_started" | "agent_event" | "text_delta" | "final" | "error" | string;

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
  });
  return { response: res.response };
}

export async function cancelAgentInvocation(invocationId: string): Promise<{ cancelled: boolean }> {
  const res = await agentClient.cancelAgentInvocation({ invocationId });
  return { cancelled: res.cancelled };
}

export async function streamChat(
  params: SendChatParams,
  handlers: ChatStreamHandlers,
  signal?: AbortSignal,
): Promise<ChatStreamPayload | null> {
  const res = await fetch(`${BASE_URL}/api/chat/stream`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "text/event-stream",
      ...authHeaders(),
    },
    body: JSON.stringify(params),
    signal,
  });

  if (res.status === 401) {
    localStorage.removeItem(TOKEN_KEY);
    window.location.href = "/login";
    throw new ApiError("unauthenticated", "Invalid or expired token");
  }

  if (!res.ok) {
    throw new ApiError("stream_failed", await responseErrorMessage(res));
  }
  if (!res.body) {
    throw new ApiError("stream_failed", "Streaming response body is empty");
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let finalPayload: ChatStreamPayload | null = null;

  for (;;) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    const chunks = buffer.split("\n\n");
    buffer = chunks.pop() ?? "";
    for (const chunk of chunks) {
      const parsed = parseSSEChunk(chunk);
      if (!parsed) continue;
      const { event, data } = parsed;
      if (event === "invocation_started") handlers.onStarted?.(data);
      else if (event === "agent_event") handlers.onAgentEvent?.(data);
      else if (event === "text_delta") handlers.onTextDelta?.(data);
      else if (event === "final") {
        finalPayload = data;
        handlers.onFinal?.(data);
      } else if (event === "error") {
        handlers.onError?.(data);
        throw new ApiError("stream_error", data.error || "Chat stream failed");
      }
    }
  }

  return finalPayload;
}

async function responseErrorMessage(res: Response): Promise<string> {
  try {
    const data = await res.json() as { error?: string; msg?: string };
    return data.error || data.msg || `Request failed with status ${res.status}`;
  } catch {
    return `Request failed with status ${res.status}`;
  }
}

function parseSSEChunk(chunk: string): null | { event: ChatStreamEvent; data: ChatStreamPayload } {
  let event: ChatStreamEvent = "message";
  const dataLines: string[] = [];

  for (const rawLine of chunk.split("\n")) {
    const line = rawLine.endsWith("\r") ? rawLine.slice(0, -1) : rawLine;
    if (line.startsWith("event:")) {
      event = line.slice("event:".length).trim();
    } else if (line.startsWith("data:")) {
      dataLines.push(line.slice("data:".length).trimStart());
    }
  }

  if (dataLines.length === 0) return null;
  return { event, data: JSON.parse(dataLines.join("\n")) as ChatStreamPayload };
}
