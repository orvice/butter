import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { APIToken } from "@/types/api";

const SVC = "agents.v1.APITokenService";

function listAPITokens() {
  return twirpFetch<object, { tokens?: APIToken[] }>(SVC, "ListAPITokens", {});
}

function createAPIToken(name: string) {
  return twirpFetch<{ name: string }, { token: APIToken; secret: string }>(
    SVC,
    "CreateAPIToken",
    { name },
  );
}

function revokeAPIToken(id: string) {
  return twirpFetch<{ id: string }, { token: APIToken }>(SVC, "RevokeAPIToken", { id });
}

export function useAPITokens() {
  return useQuery({ queryKey: ["api-tokens"], queryFn: listAPITokens });
}

export function useCreateAPIToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createAPIToken,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["api-tokens"] }),
  });
}

export function useRevokeAPIToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: revokeAPIToken,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["api-tokens"] }),
  });
}
