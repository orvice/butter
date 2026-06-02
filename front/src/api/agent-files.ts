import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import {
  AgentFileService,
  AgentFileSpaceSchema,
  type AgentFile as PbAgentFile,
  type AgentFileSearchResult as PbAgentFileSearchResult,
  type AgentFileSpace as PbAgentFileSpace,
} from "@/gen/agents/v1/agent_file_pb";
import type { AgentFile, AgentFileSearchResult, AgentFileSpace } from "@/types/api";
import { bigintToNumber, tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(AgentFileService);

function spaceFromProto(s: PbAgentFileSpace): AgentFileSpace {
  return {
    id: s.id,
    name: s.name,
    description: s.description,
    metadata: s.metadata,
    created_at: tsToISO(s.createdAt),
    updated_at: tsToISO(s.updatedAt),
    workspace_id: s.workspaceId,
  };
}

function spaceToProto(s: AgentFileSpace): PbAgentFileSpace {
  return create(AgentFileSpaceSchema, {
    id: s.id ?? "",
    name: s.name,
    description: s.description ?? "",
    metadata: s.metadata ?? {},
  });
}

function fileFromProto(f: PbAgentFile): AgentFile {
  return {
    id: f.id,
    space_id: f.spaceId,
    path: f.path,
    content_type: f.contentType,
    size_bytes: bigintToNumber(f.sizeBytes),
    version: bigintToNumber(f.version),
    metadata: f.metadata,
    created_at: tsToISO(f.createdAt),
    updated_at: tsToISO(f.updatedAt),
    workspace_id: f.workspaceId,
  };
}

function searchResultFromProto(r: PbAgentFileSearchResult): AgentFileSearchResult {
  return {
    file: r.file ? fileFromProto(r.file) : undefined,
    snippets: r.snippets,
  };
}

async function listAgentFileSpaces(): Promise<{ spaces?: AgentFileSpace[] }> {
  const res = await client.listAgentFileSpaces({});
  return { spaces: res.spaces.map(spaceFromProto) };
}

async function createAgentFileSpace(space: AgentFileSpace): Promise<{ space?: AgentFileSpace }> {
  const res = await client.createAgentFileSpace({ space: spaceToProto(space) });
  return { space: res.space ? spaceFromProto(res.space) : undefined };
}

async function updateAgentFileSpace(space: AgentFileSpace): Promise<{ space?: AgentFileSpace }> {
  const res = await client.updateAgentFileSpace({ space: spaceToProto(space) });
  return { space: res.space ? spaceFromProto(res.space) : undefined };
}

async function deleteAgentFileSpace(id: string): Promise<void> {
  await client.deleteAgentFileSpace({ id });
}

async function listAgentFiles(spaceId: string, pathPrefix = ""): Promise<{ files?: AgentFile[] }> {
  const res = await client.listAgentFiles({ spaceId, pathPrefix });
  return { files: res.files.map(fileFromProto) };
}

async function getAgentFile(
  spaceId: string,
  path: string,
  version?: number,
): Promise<{ file?: AgentFile; content?: string }> {
  const res = await client.getAgentFile({
    spaceId,
    path,
    version: version !== undefined ? BigInt(version) : 0n,
  });
  return {
    file: res.file ? fileFromProto(res.file) : undefined,
    content: res.content || undefined,
  };
}

interface WriteAgentFileInput {
  spaceId: string;
  path: string;
  content: string;
  contentType?: string;
  metadata?: Record<string, string>;
}

async function writeAgentFile(input: WriteAgentFileInput): Promise<{ file?: AgentFile }> {
  const res = await client.writeAgentFile({
    spaceId: input.spaceId,
    path: input.path,
    content: input.content,
    contentType: input.contentType ?? "",
    metadata: input.metadata ?? {},
  });
  return { file: res.file ? fileFromProto(res.file) : undefined };
}

async function deleteAgentFile({ spaceId, path }: { spaceId: string; path: string }): Promise<void> {
  await client.deleteAgentFile({ spaceId, path });
}

async function searchAgentFiles(
  spaceId: string,
  query: string,
  limit = 20,
): Promise<{ results?: AgentFileSearchResult[] }> {
  const res = await client.searchAgentFiles({ spaceId, query, limit });
  return { results: res.results.map(searchResultFromProto) };
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
