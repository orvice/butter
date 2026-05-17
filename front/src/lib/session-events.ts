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

export function parseSessionEvent(evt: SessionEvent): ParsedEvent {
  const author = evt.author ?? "unknown";
  let role: ParsedEvent["role"] = "assistant";
  if (author === "user") role = "user";
  else if (author === "system") role = "system";

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
