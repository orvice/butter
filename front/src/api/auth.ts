import { twirpFetch } from "./client";

export interface AuthUser {
  id: string;
  username: string;
  display_name?: string;
  role?: string;
  disabled?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface LoginResponse {
  token: string;
  user?: AuthUser;
  expires_at?: string;
}

const SVC = "agents.v1.AuthService";

export function login(username: string, password: string) {
  return twirpFetch<{ username: string; password: string }, LoginResponse>(SVC, "Login", { username, password });
}

export function me() {
  return twirpFetch<object, { user?: AuthUser }>(SVC, "Me", {});
}

export function logout() {
  return twirpFetch<object, object>(SVC, "Logout", {});
}
