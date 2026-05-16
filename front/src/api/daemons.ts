import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type {
  BridgeDiagnostics,
  DaemonStatus,
  DaemonTaskInFlight,
} from "@/types/api";

const SVC = "agents.v1.DaemonService";

function listDaemons() {
  return twirpFetch<object, { daemons?: DaemonStatus[] }>(SVC, "ListDaemons", {});
}

function getDaemon(daemonId: string) {
  return twirpFetch<{ daemon_id: string }, { daemon: DaemonStatus }>(
    SVC,
    "GetDaemon",
    { daemon_id: daemonId },
  );
}

function listDaemonTasks(daemonId?: string) {
  return twirpFetch<{ daemon_id?: string }, { tasks?: DaemonTaskInFlight[] }>(
    SVC,
    "ListDaemonTasks",
    { daemon_id: daemonId },
  );
}

function cancelDaemonTask(taskId: string, daemonId?: string) {
  return twirpFetch<
    { task_id: string; daemon_id?: string },
    { daemon_id: string }
  >(SVC, "CancelDaemonTask", { task_id: taskId, daemon_id: daemonId });
}

function getBridgeDiagnostics() {
  return twirpFetch<object, { diagnostics: BridgeDiagnostics }>(
    SVC,
    "GetBridgeDiagnostics",
    {},
  );
}

export function useDaemons() {
  return useQuery({
    queryKey: ["daemons"],
    queryFn: listDaemons,
    refetchInterval: 15_000,
  });
}

export function useDaemon(daemonId: string) {
  return useQuery({
    queryKey: ["daemons", daemonId],
    queryFn: () => getDaemon(daemonId),
    enabled: !!daemonId,
  });
}

export function useDaemonTasks(daemonId?: string) {
  return useQuery({
    queryKey: ["daemons", "tasks", daemonId ?? "all"],
    queryFn: () => listDaemonTasks(daemonId),
    refetchInterval: 5_000,
  });
}

export function useCancelDaemonTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, daemonId }: { taskId: string; daemonId?: string }) =>
      cancelDaemonTask(taskId, daemonId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["daemons", "tasks"] }),
  });
}

export function useBridgeDiagnostics() {
  return useQuery({
    queryKey: ["daemons", "bridge"],
    queryFn: getBridgeDiagnostics,
    refetchInterval: 5_000,
  });
}
