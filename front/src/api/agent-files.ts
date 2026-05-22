import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { AgentFile, AgentFileSearchResult, AgentFileSpace } from "@/types/api";

const SVC = "agents.v1.AgentFileService";

function listAgentFileSpaces() {
  return twirpFetch<object, { spaces?: AgentFileSpace[] }>(SVC, "ListAgentFileSpaces", {});
}

function createAgentFileSpace(space: AgentFileSpace) {
  return twirpFetch<{ space: AgentFileSpace }, { space?: AgentFileSpace }>(SVC, "CreateAgentFileSpace", { space });
}

function updateAgentFileSpace(space: AgentFileSpace) {
  return twirpFetch<{ space: AgentFileSpace }, { space?: AgentFileSpace }>(SVC, "UpdateAgentFileSpace", { space });
}

function deleteAgentFileSpace(id: string) {
  return twirpFetch<{ id: string }, object>(SVC, "DeleteAgentFileSpace", { id });
}

function listAgentFiles(spaceId: string, pathPrefix = "") {
  return twirpFetch<{ space_id: string; path_prefix?: string }, { files?: AgentFile[] }>(
    SVC,
    "ListAgentFiles",
    { space_id: spaceId, path_prefix: pathPrefix },
  );
}

function getAgentFile(spaceId: string, path: string, version?: number) {
  return twirpFetch<
    { space_id: string; path: string; version?: number },
    { file?: AgentFile; content?: string }
  >(SVC, "GetAgentFile", { space_id: spaceId, path, version });
}

interface WriteAgentFileInput {
  spaceId: string;
  path: string;
  content: string;
  contentType?: string;
  metadata?: Record<string, string>;
}

function writeAgentFile(input: WriteAgentFileInput) {
  return twirpFetch<
    { space_id: string; path: string; content: string; content_type?: string; metadata?: Record<string, string> },
    { file?: AgentFile }
  >(SVC, "WriteAgentFile", {
    space_id: input.spaceId,
    path: input.path,
    content: input.content,
    content_type: input.contentType,
    metadata: input.metadata,
  });
}

function deleteAgentFile({ spaceId, path }: { spaceId: string; path: string }) {
  return twirpFetch<{ space_id: string; path: string }, object>(SVC, "DeleteAgentFile", { space_id: spaceId, path });
}

function searchAgentFiles(spaceId: string, query: string, limit = 20) {
  return twirpFetch<
    { space_id: string; query: string; limit?: number },
    { results?: AgentFileSearchResult[] }
  >(SVC, "SearchAgentFiles", { space_id: spaceId, query, limit });
}

export function useAgentFileSpaces() {
  return useQuery({ queryKey: ["agent-file-spaces"], queryFn: listAgentFileSpaces });
}

export function useCreateAgentFileSpace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createAgentFileSpace,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agent-file-spaces"] }),
  });
}

export function useUpdateAgentFileSpace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateAgentFileSpace,
    onSuccess: (_data, space) => {
      qc.invalidateQueries({ queryKey: ["agent-file-spaces"] });
      qc.invalidateQueries({ queryKey: ["agent-files", space.id] });
    },
  });
}

export function useDeleteAgentFileSpace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteAgentFileSpace,
    onSuccess: (_data, id) => {
      qc.invalidateQueries({ queryKey: ["agent-file-spaces"] });
      qc.removeQueries({ queryKey: ["agent-files", id] });
    },
  });
}

export function useAgentFiles(spaceId: string, pathPrefix = "") {
  return useQuery({
    queryKey: ["agent-files", spaceId, pathPrefix],
    queryFn: () => listAgentFiles(spaceId, pathPrefix),
    enabled: !!spaceId,
  });
}

export function useAgentFile(spaceId: string, path: string) {
  return useQuery({
    queryKey: ["agent-file", spaceId, path],
    queryFn: () => getAgentFile(spaceId, path),
    enabled: !!spaceId && !!path,
  });
}

export function useWriteAgentFile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: writeAgentFile,
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: ["agent-files", input.spaceId] });
      qc.invalidateQueries({ queryKey: ["agent-file", input.spaceId, input.path] });
    },
  });
}

export function useDeleteAgentFile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteAgentFile,
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: ["agent-files", input.spaceId] });
      qc.removeQueries({ queryKey: ["agent-file", input.spaceId, input.path] });
    },
  });
}

export function useSearchAgentFiles(spaceId: string, query: string, limit = 20) {
  return useQuery({
    queryKey: ["agent-file-search", spaceId, query, limit],
    queryFn: () => searchAgentFiles(spaceId, query, limit),
    enabled: !!spaceId && query.trim().length > 0,
  });
}
