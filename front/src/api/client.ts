import type { TwirpError } from "@/types/api";
import { TOKEN_KEY } from "@/lib/constants";

const BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

export class ApiError extends Error {
  code: string;
  constructor(code: string, msg: string) {
    super(msg);
    this.code = code;
    this.name = "ApiError";
  }
}

export async function twirpFetch<TReq, TRes>(
  service: string,
  method: string,
  body: TReq,
): Promise<TRes> {
  const token = localStorage.getItem(TOKEN_KEY);
  const url = `${BASE_URL}/api/${service}/${method}`;

  const res = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(body),
  });

  if (res.status === 401) {
    localStorage.removeItem(TOKEN_KEY);
    window.location.href = "/login";
    throw new ApiError("unauthenticated", "Invalid or expired token");
  }

  const data = await res.json();

  if (!res.ok) {
    const err = data as TwirpError;
    throw new ApiError(err.code || "unknown", err.msg || "Unknown error");
  }

  return data as TRes;
}

export async function validateToken(token: string): Promise<boolean> {
  const url = `${BASE_URL}/api/agents.v1.AuthService/Me`;
  const res = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({}),
  });
  return res.ok;
}
