import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { Workspace } from "@/gen/agents/v1/workspace_pb";
import {
  AuthService,
  type User as PbUser,
  type LoginResponse as PbLoginResponse,
  type CompleteOAuthFlowResponse as PbCompleteOAuthFlowResponse,
} from "@/gen/agents/v1/auth_pb";
import { tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

// The exported interfaces here keep the snake_case/camelCase tolerance the
// rest of the dashboard already relies on. Proto-generated types are
// camelCase-only, so we map them at this boundary instead of forcing every
// caller to use the proto types directly.
export interface AuthUser {
  id: string;
  username: string;
  display_name?: string;
  displayName?: string;
  avatar_url?: string;
  avatarUrl?: string;
  role?: string;
  disabled?: boolean;
  created_at?: string;
  createdAt?: string;
  updated_at?: string;
  updatedAt?: string;
}

export interface LoginResponse {
  token: string;
  user?: AuthUser;
  expires_at?: string;
  expiresAt?: string;
  workspaces?: Workspace[];
}

export interface CreateUserInput {
  username: string;
  password: string;
  display_name?: string;
  role?: string;
  disabled?: boolean;
}

export interface UpdateUserPasswordInput {
  id: string;
  password: string;
}

export interface SetUserDisabledInput {
  id: string;
  disabled: boolean;
}

export interface UpdateProfileInput {
  display_name: string;
  avatar_url?: string;
}

export interface ChangePasswordInput {
  current_password: string;
  new_password: string;
}

export interface OAuthProviderInfo {
  name: string;
  display_name?: string;
  displayName?: string;
}

export interface BeginOAuthFlowResponse {
  authorize_url?: string;
  authorizeUrl?: string;
  state: string;
}

export type CompleteOAuthFlowResponse = LoginResponse;

const client = makeClient(AuthService);

function toAuthUser(u: PbUser | undefined): AuthUser | undefined {
  if (!u) return undefined;
  return {
    id: u.id,
    username: u.username,
    displayName: u.displayName,
    avatarUrl: u.avatarUrl,
    role: u.role,
    disabled: u.disabled,
  };
}

function toLoginResponse(r: PbLoginResponse | PbCompleteOAuthFlowResponse): LoginResponse {
  return {
    token: r.token,
    user: toAuthUser(r.user),
    expiresAt: tsToISO(r.expiresAt),
    workspaces: r.workspaces,
  };
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  return toLoginResponse(await client.login({ username, password }));
}

export async function listOAuthProviders(): Promise<{ providers?: OAuthProviderInfo[] }> {
  const res = await client.listOAuthProviders({});
  return {
    providers: res.providers.map((p) => ({ name: p.name, displayName: p.displayName })),
  };
}

export async function beginOAuthFlow(
  provider: string,
  redirectUri: string,
): Promise<BeginOAuthFlowResponse> {
  const res = await client.beginOAuthFlow({ provider, redirectUri });
  return { authorizeUrl: res.authorizeUrl, state: res.state };
}

export async function completeOAuthFlow(
  provider: string,
  code: string,
  state: string,
): Promise<CompleteOAuthFlowResponse> {
  return toLoginResponse(await client.completeOAuthFlow({ provider, code, state }));
}

export async function me(): Promise<{ user?: AuthUser }> {
  const res = await client.me({});
  return { user: toAuthUser(res.user) };
}

export async function logout(): Promise<void> {
  await client.logout({});
}

async function listUsers(): Promise<{ users?: AuthUser[] }> {
  const res = await client.listUsers({});
  return { users: res.users.map((u) => toAuthUser(u)!) };
}

async function createUser(input: CreateUserInput): Promise<{ user?: AuthUser }> {
  const res = await client.createUser({
    username: input.username,
    password: input.password,
    displayName: input.display_name ?? "",
    role: input.role ?? "",
    disabled: input.disabled ?? false,
  });
  return { user: toAuthUser(res.user) };
}

async function updateUserPassword(input: UpdateUserPasswordInput): Promise<{ user?: AuthUser }> {
  const res = await client.updateUserPassword(input);
  return { user: toAuthUser(res.user) };
}

async function setUserDisabled(input: SetUserDisabledInput): Promise<{ user?: AuthUser }> {
  const res = await client.setUserDisabled(input);
  return { user: toAuthUser(res.user) };
}

async function updateProfile(input: UpdateProfileInput): Promise<{ user?: AuthUser }> {
  const res = await client.updateProfile({
    displayName: input.display_name,
    avatarUrl: input.avatar_url,
  });
  return { user: toAuthUser(res.user) };
}

async function changePassword(input: ChangePasswordInput): Promise<{ user?: AuthUser }> {
  const res = await client.changePassword({
    currentPassword: input.current_password,
    newPassword: input.new_password,
  });
  return { user: toAuthUser(res.user) };
}

export function isAdmin(user: AuthUser | null | undefined) {
  return user?.role === "admin";
}

export function useUsers(options?: { enabled?: boolean }) {
  return useQuery({ queryKey: ["users"], queryFn: listUsers, enabled: options?.enabled ?? true });
}

export function useCreateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createUser,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export function useUpdateUserPassword() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateUserPassword,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export function useSetUserDisabled() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: setUserDisabled,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users"] }),
  });
}

export function useUpdateProfile() {
  return useMutation({ mutationFn: updateProfile });
}

export function useChangePassword() {
  return useMutation({ mutationFn: changePassword });
}
