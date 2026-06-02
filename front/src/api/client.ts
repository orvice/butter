import { TOKEN_KEY, WORKSPACE_KEY } from "@/lib/constants";

export const BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

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
    const message = data?.error || data?.msg || "Unknown error";
    throw new ApiError(data?.code || "unknown", message);
  }

  return data as TRes;
}

export async function validateToken(token: string): Promise<boolean> {
  const url = `${BASE_URL}/api/agents.v1.AuthService/Me`;
  const res = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(),
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({}),
  });
  return res.ok;
}
