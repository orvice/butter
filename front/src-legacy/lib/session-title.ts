import type { SessionInfo } from "@/types/api";

export function sessionAgentName(state: SessionInfo["state"]): string | undefined {
  if (!state) return undefined;
  const v = state["agent_name"];
  return typeof v === "string" && v ? v : undefined;
}

/**
 * Effective display title, in precedence order: first-class title,
 * legacy state["title"], agent name, shortened session ID.
 */
export function sessionTitle(session: SessionInfo): string {
  if (session.title && session.title.trim()) return session.title.trim();
  const legacy = session.state?.["title"];
  if (typeof legacy === "string" && legacy.trim()) return legacy.trim();
  return sessionAgentName(session.state) ?? session.session_id.slice(0, 12);
}

export const MAX_TITLE_CODE_POINTS = 100;

/** Trims and collapses a manual title to a single line, mirroring the backend. */
export function normalizeTitleInput(raw: string): string {
  return raw.replace(/\r\n|\r|\n/g, " ").trim();
}

export function titleCodePointCount(s: string): number {
  return [...s].length;
}
