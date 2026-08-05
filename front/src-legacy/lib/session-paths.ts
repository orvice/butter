import type { SessionInfo } from "@/types/api";

/** Route to the dedicated session detail page for a given session. */
export function sessionDetailPath(session: Pick<SessionInfo, "app_name" | "user_id" | "session_id">): string {
  const params = new URLSearchParams({
    app: session.app_name,
    user: session.user_id,
    sid: session.session_id,
  });
  return `/sessions/detail?${params.toString()}`;
}
