import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { Agent } from "@/types/api";

const SVC = "agents.v1.AgentService";

function listAgents() {
  return twirpFetch<object, { agents: Agent[] }>(SVC, "ListAgents", {});
}

function getAgent(name: string) {
  return twirpFetch<{ name: string }, { agent: Agent }>(SVC, "GetAgent", { name });
}

function createAgent(agent: Agent) {
  return twirpFetch<{ agent: Agent }, { agent: Agent }>(SVC, "CreateAgent", { agent });
}

function updateAgent(agent: Agent) {
  return twirpFetch<{ agent: Agent }, { agent: Agent }>(SVC, "UpdateAgent", { agent });
}

function deleteAgent(name: string) {
  return twirpFetch<{ name: string }, object>(SVC, "DeleteAgent", { name });
}

export function useAgents() {
  return useQuery({ queryKey: ["agents"], queryFn: listAgents });
}

export function useAgent(name: string) {
  return useQuery({ queryKey: ["agents", name], queryFn: () => getAgent(name), enabled: !!name });
}

export function useCreateAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agents"] }),
  });
}

export function useUpdateAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateAgent,
    onSuccess: (_data, agent) => {
      qc.invalidateQueries({ queryKey: ["agents"] });
      qc.invalidateQueries({ queryKey: ["agents", agent.name] });
    },
  });
}

export function useDeleteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agents"] }),
  });
}
