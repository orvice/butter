import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import {
  WorkspaceRepoBindingSchema,
  WorkspaceRepoBindingService,
  type RepoBindingOverlap,
  type RepoBindingWriteMode,
  type WorkspaceRepoBinding,
} from "@/gen/agents/v1/repobinding_pb";
import { makeClient } from "./transport";

const client = makeClient(WorkspaceRepoBindingService);

export interface RepoBindingInput {
  gitHostId: string;
  repository: string;
  branch: string;
  rootPath?: string;
  writeMode: RepoBindingWriteMode;
}

async function getRepoBinding(): Promise<{
  binding?: WorkspaceRepoBinding;
  overlaps: RepoBindingOverlap[];
}> {
  const res = await client.getWorkspaceRepoBinding({});
  return { binding: res.binding, overlaps: res.overlaps };
}

async function putRepoBinding(input: RepoBindingInput): Promise<WorkspaceRepoBinding> {
  const res = await client.putWorkspaceRepoBinding({
    binding: create(WorkspaceRepoBindingSchema, {
      gitHostId: input.gitHostId,
      repository: input.repository,
      branch: input.branch,
      rootPath: input.rootPath ?? "",
      writeMode: input.writeMode,
      contentSchemaVersion: 1,
    }),
  });
  if (!res.binding) throw new Error("put returned no binding");
  return res.binding;
}

async function deleteRepoBinding(): Promise<void> {
  await client.deleteWorkspaceRepoBinding({});
}

async function setRepoBindingCredential(pat: string): Promise<WorkspaceRepoBinding> {
  const res = await client.setWorkspaceRepoBindingCredential({ pat });
  if (!res.binding) throw new Error("set credential returned no binding");
  return res.binding;
}

async function validateRepoBinding(): Promise<WorkspaceRepoBinding> {
  const res = await client.validateWorkspaceRepoBinding({});
  if (!res.binding) throw new Error("validate returned no binding");
  return res.binding;
}

export function useRepoBinding() {
  return useQuery({ queryKey: ["repo-binding"], queryFn: getRepoBinding });
}

export function usePutRepoBinding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: putRepoBinding,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["repo-binding"] }),
  });
}

export function useDeleteRepoBinding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteRepoBinding,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["repo-binding"] }),
  });
}

export function useSetRepoBindingCredential() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: setRepoBindingCredential,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["repo-binding"] }),
  });
}

export function useValidateRepoBinding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: validateRepoBinding,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["repo-binding"] }),
  });
}
