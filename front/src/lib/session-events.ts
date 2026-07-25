import type { SessionEvent, SessionInfo } from "@/types/api";

export function sessionDetailPath(session: Pick<SessionInfo, "app_name" | "user_id" | "session_id">): string {
  const params = new URLSearchParams({
    app: session.app_name,
    user: session.user_id,
    sid: session.session_id,
  });
  return `/sessions/detail?${params.toString()}`;
}

export interface ParsedEvent {
  eventId: string;
  author: string;
  role: "user" | "assistant" | "system";
  text: string;
  toolCalls: ToolCallSummary[];
  toolResponses: ToolResponseSummary[];
  timestamp?: string;
  traceUrl?: string;
  raw: SessionEvent;
}

export interface ToolCallSummary {
  name: string;
  argsPreview: string;
}

export interface ToolResponseSummary {
  name: string;
  responsePreview: string;
}

export interface ParsedTextPart {
  text: string;
}

export interface ParsedFunctionCall {
  name: string;
  args: unknown;
}

export interface ParsedFunctionResponse {
  name: string;
  response: unknown;
}

export interface FullParsedEvent {
  eventId: string;
  author: string;
  role: "user" | "assistant" | "system";
  textParts: ParsedTextPart[];
  functionCalls: ParsedFunctionCall[];
  functionResponses: ParsedFunctionResponse[];
  timestamp?: string;
  traceUrl?: string;
  traceId?: string;
  raw: SessionEvent;
}

interface GenaiPart {
  text?: string;
  functionCall?: { name?: string; args?: unknown };
  function_call?: { name?: string; args?: unknown };
  functionResponse?: { name?: string; response?: unknown };
  function_response?: { name?: string; response?: unknown };
}

interface GenaiContent {
  role?: string;
  parts?: GenaiPart[];
}

function previewJson(value: unknown, max = 120): string {
  if (value === undefined || value === null) return "";
  let s: string;
  try {
    s = typeof value === "string" ? value : JSON.stringify(value);
  } catch {
    s = String(value);
  }
  if (s.length > max) s = s.slice(0, max) + "…";
  return s;
}

function roleFromAuthor(author: string): ParsedEvent["role"] {
  if (author === "user") return "user";
  if (author === "system") return "system";
  return "assistant";
}

export function parseSessionEvent(evt: SessionEvent): ParsedEvent {
  const author = evt.author ?? "unknown";
  const role = roleFromAuthor(author);

  const out: ParsedEvent = {
    eventId: evt.event_id,
    author,
    role,
    text: "",
    toolCalls: [],
    toolResponses: [],
    timestamp: evt.timestamp,
    traceUrl: evt.trace_url,
    raw: evt,
  };

  if (!evt.content_json) return out;

  let content: GenaiContent;
  try {
    content = JSON.parse(evt.content_json) as GenaiContent;
  } catch {
    out.text = evt.content_json;
    return out;
  }

  const texts: string[] = [];
  for (const part of content.parts ?? []) {
    if (typeof part.text === "string" && part.text.length > 0) {
      texts.push(part.text);
    }
    const call = part.functionCall ?? part.function_call;
    if (call?.name) {
      out.toolCalls.push({ name: call.name, argsPreview: previewJson(call.args) });
    }
    const resp = part.functionResponse ?? part.function_response;
    if (resp?.name) {
      out.toolResponses.push({ name: resp.name, responsePreview: previewJson(resp.response) });
    }
  }
  out.text = texts.join("\n");
  return out;
}

export function parseSessionEvents(events: SessionEvent[] | undefined): ParsedEvent[] {
  if (!events) return [];
  return events.map(parseSessionEvent);
}

export function parseSessionEventFull(evt: SessionEvent): FullParsedEvent {
  const author = evt.author ?? "unknown";
  const out: FullParsedEvent = {
    eventId: evt.event_id,
    author,
    role: roleFromAuthor(author),
    textParts: [],
    functionCalls: [],
    functionResponses: [],
    timestamp: evt.timestamp,
    traceUrl: evt.trace_url,
    traceId: evt.trace_id,
    raw: evt,
  };

  if (!evt.content_json) return out;

  let content: GenaiContent;
  try {
    content = JSON.parse(evt.content_json) as GenaiContent;
  } catch {
    out.textParts = [{ text: evt.content_json }];
    return out;
  }

  for (const part of content.parts ?? []) {
    if (typeof part.text === "string" && part.text.length > 0) {
      out.textParts.push({ text: part.text });
    }
    const call = part.functionCall ?? part.function_call;
    if (call?.name) {
      out.functionCalls.push({ name: call.name, args: call.args });
    }
    const resp = part.functionResponse ?? part.function_response;
    if (resp?.name) {
      out.functionResponses.push({ name: resp.name, response: resp.response });
    }
  }
  return out;
}

export function parseSessionEventsFull(events: SessionEvent[] | undefined): FullParsedEvent[] {
  if (!events) return [];
  return events.map(parseSessionEventFull);
}
