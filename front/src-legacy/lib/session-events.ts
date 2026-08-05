import type { SessionEvent } from "@/types/api";

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

/** Compact summary of an event: joined text plus truncated tool-call/response previews. */
export function parseSessionEvent(evt: SessionEvent): ParsedEvent {
  const full = parseSessionEventFull(evt);
  return {
    eventId: full.eventId,
    author: full.author,
    role: full.role,
    text: full.textParts.map((part) => part.text).join("\n"),
    toolCalls: full.functionCalls.map((call) => ({ name: call.name, argsPreview: previewJson(call.args) })),
    toolResponses: full.functionResponses.map((resp) => ({
      name: resp.name,
      responsePreview: previewJson(resp.response),
    })),
    timestamp: full.timestamp,
    traceUrl: full.traceUrl,
    raw: full.raw,
  };
}

export function parseSessionEvents(events: SessionEvent[] | undefined): ParsedEvent[] {
  if (!events) return [];
  return events.map(parseSessionEvent);
}

/** Full parse of an event's content into text parts, tool calls, and tool responses (no truncation). */
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
