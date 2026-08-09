import {
  useMutation,
  useQuery,
  useQueryClient,
  type QueryClient,
} from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import {
  WorkspaceRepoBindingSchema,
  WorkspaceRepoBindingService,
  type GetRepositoryFileResponse,
  type ListRepositoryEntriesResponse,
  type OnboardWorkspaceRepositoryResponse,
  type RepoBindingOnboardingMode,
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

async function syncRepository(): Promise<WorkspaceRepoBinding> {
  const res = await client.syncWorkspaceRepository({});
  if (!res.binding) throw new Error("sync returned no binding");
  return res.binding;
}

async function onboardWorkspaceRepository(
  mode: RepoBindingOnboardingMode,
): Promise<OnboardWorkspaceRepositoryResponse> {
  return client.onboardWorkspaceRepository({ mode });
}

async function listRepositoryEntries(path: string): Promise<ListRepositoryEntriesResponse> {
  return client.listRepositoryEntries({ path });
}

async function getRepositoryFile(path: string): Promise<GetRepositoryFileResponse> {
  return client.getRepositoryFile({ path });
}

function invalidateRepositoryQueries(qc: QueryClient) {
  return Promise.all([
    qc.invalidateQueries({ queryKey: ["repo-binding"] }),
    qc.invalidateQueries({ queryKey: ["repo-entries"] }),
    qc.invalidateQueries({ queryKey: ["repo-file"] }),
  ]);
}

export function useRepoBinding() {
  return useQuery({ queryKey: ["repo-binding"], queryFn: getRepoBinding });
}

export function usePutRepoBinding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: putRepoBinding,
    onSuccess: () => {
      qc.removeQueries({ queryKey: ["repo-entries"] });
      qc.removeQueries({ queryKey: ["repo-file"] });
      return qc.invalidateQueries({ queryKey: ["repo-binding"] });
    },
  });
}

export function useDeleteRepoBinding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteRepoBinding,
    onSuccess: () => {
      qc.removeQueries({ queryKey: ["repo-entries"] });
      qc.removeQueries({ queryKey: ["repo-file"] });
      return qc.invalidateQueries({ queryKey: ["repo-binding"] });
    },
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

export function useSyncRepository() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: syncRepository,
    onSuccess: () => invalidateRepositoryQueries(qc),
  });
}

export function useOnboardWorkspaceRepository() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: onboardWorkspaceRepository,
    onSettled: async () => {
      await Promise.all([
        invalidateRepositoryQueries(qc),
        qc.invalidateQueries({ queryKey: ["agents"] }),
      ]);
    },
  });
}

export function useRepositoryEntries(path: string) {
  return useQuery({
    queryKey: ["repo-entries", path],
    queryFn: () => listRepositoryEntries(path),
  });
}

export function useRepositoryFile(path: string) {
  return useQuery({
    queryKey: ["repo-file", path],
    queryFn: () => getRepositoryFile(path),
    enabled: path.length > 0,
  });
}
