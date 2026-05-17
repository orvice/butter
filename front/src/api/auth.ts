import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";

export interface AuthUser {
  id: string;
  username: string;
  display_name?: string;
  displayName?: string;
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

function listUsers() {
  return twirpFetch<object, { users?: AuthUser[] }>(SVC, "ListUsers", {});
}

function createUser(input: CreateUserInput) {
  return twirpFetch<CreateUserInput, { user?: AuthUser }>(SVC, "CreateUser", input);
}

function updateUserPassword(input: UpdateUserPasswordInput) {
  return twirpFetch<UpdateUserPasswordInput, { user?: AuthUser }>(SVC, "UpdateUserPassword", input);
}

function setUserDisabled(input: SetUserDisabledInput) {
  return twirpFetch<SetUserDisabledInput, { user?: AuthUser }>(SVC, "SetUserDisabled", input);
}

export function useUsers() {
  return useQuery({ queryKey: ["users"], queryFn: listUsers });
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
