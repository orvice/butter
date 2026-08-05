import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { WorkspaceService, type Workspace, type WorkspaceMember } from "@/gen/agents/v1/workspace_pb";
import { makeClient } from "./transport";

type WorkspacesCache = { workspaces?: Workspace[] };

export interface CreateWorkspaceInput {
  name: string;
  slug: string;
  description?: string;
}

export interface UpdateWorkspaceInput {
  id: string;
  name: string;
  slug: string;
  description?: string;
}

export interface AddWorkspaceMemberInput {
  workspace_id: string;
  user_id: string;
  role: string;
}

export interface UpdateWorkspaceMemberInput {
  workspace_id: string;
  user_id: string;
  role: string;
}

export interface RemoveWorkspaceMemberInput {
  workspace_id: string;
  user_id: string;
}

const client = makeClient(WorkspaceService);

export async function listWorkspaces(): Promise<{ workspaces?: Workspace[] }> {
  const res = await client.listWorkspaces({});
  return { workspaces: res.workspaces };
}

export async function createWorkspaceRequest(input: CreateWorkspaceInput): Promise<{ workspace?: Workspace }> {
  const res = await client.createWorkspace({
    workspace: { name: input.name, slug: input.slug, description: input.description ?? "" },
  });
  return { workspace: res.workspace };
}

async function updateWorkspaceRequest(input: UpdateWorkspaceInput): Promise<{ workspace?: Workspace }> {
  const res = await client.updateWorkspace({
    workspace: { id: input.id, name: input.name, slug: input.slug, description: input.description ?? "" },
  });
  return { workspace: res.workspace };
}

async function deleteWorkspaceRequest(id: string): Promise<void> {
  await client.deleteWorkspace({ id });
}

async function listWorkspaceMembersRequest(workspaceId: string): Promise<{ members?: WorkspaceMember[] }> {
  const res = await client.listWorkspaceMembers({ workspaceId });
  return { members: res.members };
}

async function addWorkspaceMemberRequest(input: AddWorkspaceMemberInput): Promise<{ member?: WorkspaceMember }> {
  const res = await client.addWorkspaceMember({
    workspaceId: input.workspace_id,
    userId: input.user_id,
    role: input.role,
  });
  return { member: res.member };
}

async function updateWorkspaceMemberRequest(input: UpdateWorkspaceMemberInput): Promise<{ member?: WorkspaceMember }> {
  const res = await client.updateWorkspaceMember({
    workspaceId: input.workspace_id,
    userId: input.user_id,
    role: input.role,
  });
  return { member: res.member };
}

async function removeWorkspaceMemberRequest(input: RemoveWorkspaceMemberInput): Promise<void> {
  await client.removeWorkspaceMember({
    workspaceId: input.workspace_id,
    userId: input.user_id,
  });
}

export function useWorkspaceMembers(workspaceId: string) {
  return useQuery({
    queryKey: ["workspace-members", workspaceId],
    queryFn: () => listWorkspaceMembersRequest(workspaceId),
    enabled: !!workspaceId,
  });
}

export function useUpdateWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateWorkspaceRequest,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}

export function useDeleteWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteWorkspaceRequest,
    onSuccess: (_, deletedId) => {
      qc.setQueryData<WorkspacesCache>(["workspaces"], (old) => ({
        ...old,
        workspaces: old?.workspaces?.filter((workspace) => workspace.id !== deletedId) ?? [],
      }));
      void qc.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}

export function useAddWorkspaceMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: addWorkspaceMemberRequest,
    onSuccess: (_, variables) => {
      void qc.invalidateQueries({ queryKey: ["workspace-members", variables.workspace_id] });
    },
  });
}

export function useUpdateWorkspaceMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateWorkspaceMemberRequest,
    onSuccess: (_, variables) => {
      void qc.invalidateQueries({ queryKey: ["workspace-members", variables.workspace_id] });
    },
  });
}

export function useRemoveWorkspaceMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: removeWorkspaceMemberRequest,
    onSuccess: (_, variables) => {
      void qc.invalidateQueries({ queryKey: ["workspace-members", variables.workspace_id] });
    },
  });
}
