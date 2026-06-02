import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { WORKSPACE_KEY } from "@/lib/constants";
import {
  DaemonService,
  DaemonStatus_State,
  type BridgeDiagnostics as PbBridgeDiagnostics,
  type DaemonStatus as PbDaemonStatus,
  type DaemonTaskInFlight as PbDaemonTaskInFlight,
  type LatencyPoint as PbLatencyPoint,
} from "@/gen/agents/v1/dashboard_pb";
import type {
  BridgeDiagnostics,
  DaemonState,
  DaemonStatus,
  DaemonTaskInFlight,
  LatencyPoint,
} from "@/types/api";
import { bigintToNumber, durationToString, tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(DaemonService);

// Proto enum values are numeric (0..3); the legacy frontend type uses the
// proto enum value names as string literals. Indexed lookup mirrors the
// enum declaration order in the .proto.
const DAEMON_STATE_NAMES: DaemonState[] = [
  "STATE_UNSPECIFIED",
  "STATE_ONLINE",
  "STATE_IDLE",
  "STATE_OFFLINE",
];

function toDaemonState(s: DaemonStatus_State): DaemonState {
  return DAEMON_STATE_NAMES[s] ?? "STATE_UNSPECIFIED";
}

function toLatencyPoint(p: PbLatencyPoint): LatencyPoint {
  return {
    timestamp: tsToISO(p.timestamp),
    latency_ms: bigintToNumber(p.latencyMs),
  };
}

function toDiagnostics(d: PbBridgeDiagnostics | undefined): BridgeDiagnostics {
  if (!d) return { latency: [] };
  return {
    cpu_percent: d.cpuPercent,
    memory_used_bytes: bigintToNumber(d.memoryUsedBytes),
    memory_limit_bytes: bigintToNumber(d.memoryLimitBytes),
    goroutines: d.goroutines,
    checked_at: tsToISO(d.checkedAt),
    latency: d.latency.map(toLatencyPoint),
  };
}

function toDaemonStatus(s: PbDaemonStatus): DaemonStatus {
  return {
    daemon_id: s.daemonId,
    name: s.name,
    capabilities: s.capabilities,
    labels: s.labels,
    state: toDaemonState(s.state),
    connected_at: tsToISO(s.connectedAt),
    uptime: durationToString(s.uptime),
    active_tasks: s.activeTasks,
    version: s.version,
    os: s.os,
    executors: s.executors,
    remote_addr: s.remoteAddr,
  };
}

function toTaskInFlight(t: PbDaemonTaskInFlight): DaemonTaskInFlight {
  return {
    task_id: t.taskId,
    daemon_id: t.daemonId,
    daemon_name: t.daemonName,
    capability: t.capability,
    started_at: tsToISO(t.startedAt),
    elapsed: durationToString(t.elapsed),
    current_step: t.currentStep,
    progress: t.progress,
    agent_name: t.agentName,
  };
}

async function listDaemons(): Promise<{ daemons: DaemonStatus[] }> {
  const res = await client.listDaemons({});
  return { daemons: res.daemons.map(toDaemonStatus) };
}

async function getDaemon(daemonId: string): Promise<{ daemon: DaemonStatus }> {
  const res = await client.getDaemon({ daemonId });
  if (!res.daemon) throw new Error("daemon not found");
  return { daemon: toDaemonStatus(res.daemon) };
}

async function listDaemonTasks(daemonId?: string): Promise<{ tasks: DaemonTaskInFlight[] }> {
  const res = await client.listDaemonTasks({ daemonId: daemonId ?? "" });
  return { tasks: res.tasks.map(toTaskInFlight) };
}

async function cancelDaemonTask(taskId: string, daemonId?: string): Promise<void> {
  await client.cancelDaemonTask({ taskId, daemonId: daemonId ?? "" });
}

async function getBridgeDiagnostics(): Promise<{ diagnostics: BridgeDiagnostics }> {
  const res = await client.getBridgeDiagnostics({});
  return { diagnostics: toDiagnostics(res.diagnostics) };
}

export function useDaemons() {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["daemons", workspaceId],
    queryFn: listDaemons,
    enabled: !!workspaceId,
    refetchInterval: 15_000,
  });
}

export function useDaemon(daemonId: string) {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["daemons", workspaceId, daemonId],
    queryFn: () => getDaemon(daemonId),
    enabled: !!workspaceId && !!daemonId,
  });
}

export function useDaemonTasks(daemonId?: string) {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["daemons", workspaceId, "tasks", daemonId ?? "all"],
    queryFn: () => listDaemonTasks(daemonId),
    enabled: !!workspaceId,
    refetchInterval: 5_000,
  });
}

export function useCancelDaemonTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, daemonId }: { taskId: string; daemonId?: string }) =>
      cancelDaemonTask(taskId, daemonId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["daemons"] }),
  });
}

export function useBridgeDiagnostics() {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["daemons", workspaceId, "bridge"],
    queryFn: getBridgeDiagnostics,
    enabled: !!workspaceId,
    refetchInterval: 5_000,
  });
}
