import { useQuery } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { AgentFileSpace } from "@/types/api";

const SVC = "agents.v1.AgentFileService";

function listAgentFileSpaces() {
  return twirpFetch<object, { spaces?: AgentFileSpace[] }>(SVC, "ListAgentFileSpaces", {});
}

export function useAgentFileSpaces() {
  return useQuery({ queryKey: ["agent-file-spaces"], queryFn: listAgentFileSpaces });
}
