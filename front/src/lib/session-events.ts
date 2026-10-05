import type { SessionEvent } from "@/types/api";

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

function roleFromAuthor(author: string): FullParsedEvent["role"] {
  if (author === "user") return "user";
  if (author === "system") return "system";
  return "assistant";
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
