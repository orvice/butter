import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { APITokenService, type APIToken as PbAPIToken } from "@/gen/agents/v1/api_token_pb";
import type { APIToken } from "@/types/api";
import { tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(APITokenService);

function toAPIToken(t: PbAPIToken): APIToken {
  return {
    id: t.id,
    name: t.name,
    prefix: t.prefix,
    created_at: tsToISO(t.createdAt),
    last_used_at: tsToISO(t.lastUsedAt),
    revoked: t.revoked,
  };
}

async function listAPITokens(): Promise<{ tokens?: APIToken[] }> {
  const res = await client.listAPITokens({});
  return { tokens: res.tokens.map(toAPIToken) };
}

async function createAPIToken(name: string): Promise<{ token: APIToken; secret: string }> {
  const res = await client.createAPIToken({ name });
  if (!res.token) throw new Error("server returned no token");
  return { token: toAPIToken(res.token), secret: res.secret };
}

async function revokeAPIToken(id: string): Promise<{ token: APIToken }> {
  const res = await client.revokeAPIToken({ id });
  if (!res.token) throw new Error("server returned no token");
  return { token: toAPIToken(res.token) };
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
