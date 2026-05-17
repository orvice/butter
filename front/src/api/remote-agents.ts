import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { RemoteAgent, RemoteAgentStatus } from "@/types/api";

const SVC = "agents.v1.RemoteAgentService";

function listRemoteAgents() {
  return twirpFetch<object, { remote_agents: RemoteAgent[] }>(SVC, "ListRemoteAgents", {});
}

function getRemoteAgent(id: string) {
  return twirpFetch<{ id: string }, { remote_agent: RemoteAgent }>(SVC, "GetRemoteAgent", { id });
}

function createRemoteAgent(remote_agent: RemoteAgent) {
  return twirpFetch<{ remote_agent: RemoteAgent }, { remote_agent: RemoteAgent }>(SVC, "CreateRemoteAgent", { remote_agent });
}

function updateRemoteAgent(remote_agent: RemoteAgent) {
  return twirpFetch<{ remote_agent: RemoteAgent }, { remote_agent: RemoteAgent }>(SVC, "UpdateRemoteAgent", { remote_agent });
}

function deleteRemoteAgent(id: string) {
  return twirpFetch<{ id: string }, object>(SVC, "DeleteRemoteAgent", { id });
}

export function useRemoteAgents() {
  return useQuery({ queryKey: ["remote-agents"], queryFn: listRemoteAgents });
}

export function useRemoteAgent(id: string) {
  return useQuery({ queryKey: ["remote-agents", id], queryFn: () => getRemoteAgent(id), enabled: !!id });
}

export function useCreateRemoteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createRemoteAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["remote-agents"] }),
  });
}

export function useUpdateRemoteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateRemoteAgent,
    onSuccess: (_data, agent) => {
      qc.invalidateQueries({ queryKey: ["remote-agents"] });
      qc.invalidateQueries({ queryKey: ["remote-agents", agent.id] });
    },
  });
}

export function useDeleteRemoteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteRemoteAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["remote-agents"] }),
  });
}

function getRemoteAgentStatus(id: string) {
  return twirpFetch<{ id: string }, { status: RemoteAgentStatus }>(
    SVC,
    "GetRemoteAgentStatus",
    { id },
  );
}

export function useRemoteAgentStatus(id: string) {
  return useQuery({
    queryKey: ["remote-agents", id, "status"],
    queryFn: () => getRemoteAgentStatus(id),
    enabled: !!id,
    refetchInterval: 30_000,
  });
}
