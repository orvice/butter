// Example grpc-web client that mirrors a subset of `agents.ts`. The
// existing Twirp-backed `useAgents` keeps working unchanged; pages can
// switch to `useAgentsGRPC` one at a time. The full migration touches
// both the API and the consuming components (hand-written snake_case
// types → generated camelCase messages), so this file is intentionally
// scoped to a single read path as a working integration test.

import { useQuery } from "@tanstack/react-query";

import { AgentService } from "@/gen/agents/v1/agent_service_pb";
import { makeClient } from "./grpc";

const client = makeClient(AgentService);

export interface ListAgentsParamsGRPC {
  pageSize?: number;
  pageToken?: string;
}

export function useAgentsGRPC(params: ListAgentsParamsGRPC = {}) {
  return useQuery({
    queryKey: ["agents-grpc", params],
    queryFn: () =>
      client.listAgents({
        pageSize: params.pageSize ?? 0,
        pageToken: params.pageToken ?? "",
      }),
  });
}
