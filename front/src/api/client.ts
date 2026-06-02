import { TOKEN_KEY, WORKSPACE_KEY } from "@/lib/constants";

export const BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

// authHeaders is still needed by chat.ts (Server-Sent Events stream) and
// uploads.ts (multipart). The typed Connect clients in api/transport.ts add
// these headers via an interceptor and don't go through this helper.
export function authHeaders(): Record<string, string> {
  const token = localStorage.getItem(TOKEN_KEY);
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return {
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(workspaceId ? { "X-Workspace-ID": workspaceId } : {}),
  };
}

export class ApiError extends Error {
  code: string;
  constructor(code: string, msg: string) {
    super(msg);
    this.code = code;
    this.name = "ApiError";
  }
}

// apiFetch is reserved for the non-Connect REST endpoints: multipart uploads
// and the SSE chat stream. Connect endpoints go through transport.ts.
export async function apiFetch<TRes>(path: string, init: RequestInit = {}): Promise<TRes> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      ...(init.body ? { "Content-Type": "application/json" } : {}),
      ...authHeaders(),
      ...init.headers,
    },
  });

  if (res.status === 401) {
    localStorage.removeItem(TOKEN_KEY);
    window.location.href = "/login";
    throw new ApiError("unauthenticated", "Invalid or expired token");
  }

  if (res.status === 204) {
    return undefined as TRes;
  }

  const data = await res.json();

  if (!res.ok) {
    const message = data?.error || "Unknown error";
    throw new ApiError(data?.code || "unknown", message);
  }

  return data as TRes;
}
