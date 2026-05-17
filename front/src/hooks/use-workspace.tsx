import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "@/api/client";
import { WORKSPACE_KEY } from "@/lib/constants";
import { useAuth } from "@/hooks/use-auth";
import type { Workspace } from "@/gen/agents/v1/workspace_pb";

const SVC = "agents.v1.WorkspaceService";

export interface CreateWorkspaceInput {
  name: string;
  slug: string;
  description?: string;
}

interface WorkspaceContextValue {
  workspaces: Workspace[];
  selectedWorkspaceId: string;
  selectedWorkspace: Workspace | null;
  isLoading: boolean;
  isCreating: boolean;
  error: unknown;
  setSelectedWorkspaceId: (id: string) => void;
  createWorkspace: (input: CreateWorkspaceInput) => Promise<Workspace>;
}

const WorkspaceContext = createContext<WorkspaceContextValue | null>(null);

function listWorkspaces() {
  return twirpFetch<object, { workspaces?: Workspace[] }>(SVC, "ListWorkspaces", {});
}

function createWorkspaceRequest(input: CreateWorkspaceInput) {
  return twirpFetch<{ workspace: CreateWorkspaceInput }, { workspace?: Workspace }>(SVC, "CreateWorkspace", {
    workspace: input,
  });
}

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth();
  const queryClient = useQueryClient();
  const [selectedWorkspaceId, setSelectedWorkspaceIdState] = useState<string>(() =>
    localStorage.getItem(WORKSPACE_KEY) ?? "",
  );

  const { data, isLoading, error } = useQuery({
    queryKey: ["workspaces"],
    queryFn: listWorkspaces,
    enabled: isAuthenticated,
    staleTime: 60_000,
  });

  const createMutation = useMutation({
    mutationFn: createWorkspaceRequest,
    onSuccess: (res) => {
      const workspace = res.workspace;
      if (!workspace?.id) return;
      localStorage.setItem(WORKSPACE_KEY, workspace.id);
      setSelectedWorkspaceIdState(workspace.id);
      void queryClient.invalidateQueries();
    },
  });

  const workspaces = useMemo(() => data?.workspaces ?? [], [data?.workspaces]);

  useEffect(() => {
    if (!isAuthenticated) {
      localStorage.removeItem(WORKSPACE_KEY);
      return;
    }
    if (workspaces.length === 0) return;

    const persisted = localStorage.getItem(WORKSPACE_KEY) ?? "";
    const persistedIsValid = persisted && workspaces.some((ws) => ws.id === persisted);
    const selectedIsValid = selectedWorkspaceId && workspaces.some((ws) => ws.id === selectedWorkspaceId);

    if (selectedIsValid) return;

    const next = persistedIsValid ? persisted : workspaces[0]?.id ?? "";

    if (next) {
      localStorage.setItem(WORKSPACE_KEY, next);
      queueMicrotask(() => {
        setSelectedWorkspaceIdState(next);
        void queryClient.invalidateQueries();
      });
    }
  }, [isAuthenticated, queryClient, selectedWorkspaceId, workspaces]);

  const setSelectedWorkspaceId = useCallback(
    (id: string) => {
      if (!id || id === selectedWorkspaceId) return;
      localStorage.setItem(WORKSPACE_KEY, id);
      setSelectedWorkspaceIdState(id);
      void queryClient.invalidateQueries();
    },
    [queryClient, selectedWorkspaceId],
  );

  const createWorkspace = useCallback(
    async (input: CreateWorkspaceInput) => {
      const res = await createMutation.mutateAsync(input);
      if (!res.workspace) {
        throw new Error("Workspace was not returned by the server");
      }
      return res.workspace;
    },
    [createMutation],
  );

  const selectedWorkspace = useMemo(
    () => workspaces.find((ws) => ws.id === selectedWorkspaceId) ?? null,
    [selectedWorkspaceId, workspaces],
  );

  return (
    <WorkspaceContext.Provider
      value={{
        workspaces,
        selectedWorkspaceId,
        selectedWorkspace,
        isLoading,
        isCreating: createMutation.isPending,
        error,
        setSelectedWorkspaceId,
        createWorkspace,
      }}
    >
      {children}
    </WorkspaceContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useWorkspace(): WorkspaceContextValue {
  const ctx = useContext(WorkspaceContext);
  if (!ctx) throw new Error("useWorkspace must be inside WorkspaceProvider");
  return ctx;
}
