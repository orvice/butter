import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import {
  GitHostSchema,
  GitHostService,
  type GitHost,
  type GitHostKind,
} from "@/gen/agents/v1/githost_pb";
import { makeClient } from "./transport";

const client = makeClient(GitHostService);

export interface GitHostInput {
  id?: string;
  name: string;
  kind: GitHostKind;
  apiBaseUrl: string;
  webBaseUrl?: string;
}

function toProto(input: GitHostInput): GitHost {
  return create(GitHostSchema, {
    id: input.id ?? "",
    name: input.name,
    kind: input.kind,
    apiBaseUrl: input.apiBaseUrl,
    webBaseUrl: input.webBaseUrl ?? "",
  });
}

async function listGitHosts(): Promise<{ hosts: GitHost[] }> {
  const res = await client.listGitHosts({});
  return { hosts: res.hosts };
}

async function createGitHost(input: GitHostInput): Promise<GitHost> {
  const res = await client.createGitHost({ host: toProto(input) });
  if (!res.host) throw new Error("create returned no host");
  return res.host;
}

async function updateGitHost(input: GitHostInput): Promise<GitHost> {
  const res = await client.updateGitHost({ host: toProto(input) });
  if (!res.host) throw new Error("update returned no host");
  return res.host;
}

async function deleteGitHost(id: string): Promise<void> {
  await client.deleteGitHost({ id });
}

export function useGitHosts() {
  return useQuery({ queryKey: ["git-hosts"], queryFn: listGitHosts });
}

export function useCreateGitHost() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createGitHost,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["git-hosts"] }),
  });
}

export function useUpdateGitHost() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateGitHost,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["git-hosts"] }),
  });
}

export function useDeleteGitHost() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteGitHost,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["git-hosts"] }),
  });
}
