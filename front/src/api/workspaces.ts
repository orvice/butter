import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "@/api/client";
import type { Workspace, WorkspaceMember } from "@/gen/agents/v1/workspace_pb";

const SVC = "agents.v1.WorkspaceService";

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

export function listWorkspaces() {
  return twirpFetch<object, { workspaces?: Workspace[] }>(SVC, "ListWorkspaces", {});
}

export function createWorkspaceRequest(input: CreateWorkspaceInput) {
  return twirpFetch<{ workspace: CreateWorkspaceInput }, { workspace?: Workspace }>(SVC, "CreateWorkspace", {
    workspace: input,
  });
}

function updateWorkspaceRequest(input: UpdateWorkspaceInput) {
  return twirpFetch<{ workspace: UpdateWorkspaceInput }, { workspace?: Workspace }>(SVC, "UpdateWorkspace", {
    workspace: input,
  });
}

function deleteWorkspaceRequest(id: string) {
  return twirpFetch<{ id: string }, object>(SVC, "DeleteWorkspace", { id });
}

function listWorkspaceMembersRequest(workspaceId: string) {
  return twirpFetch<{ workspace_id: string }, { members?: WorkspaceMember[] }>(SVC, "ListWorkspaceMembers", {
    workspace_id: workspaceId,
  });
}

function addWorkspaceMemberRequest(input: AddWorkspaceMemberInput) {
  return twirpFetch<AddWorkspaceMemberInput, { member?: WorkspaceMember }>(SVC, "AddWorkspaceMember", input);
}

function updateWorkspaceMemberRequest(input: UpdateWorkspaceMemberInput) {
  return twirpFetch<UpdateWorkspaceMemberInput, { member?: WorkspaceMember }>(SVC, "UpdateWorkspaceMember", input);
}

function removeWorkspaceMemberRequest(input: RemoveWorkspaceMemberInput) {
  return twirpFetch<RemoveWorkspaceMemberInput, object>(SVC, "RemoveWorkspaceMember", input);
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
    onSuccess: () => {
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
