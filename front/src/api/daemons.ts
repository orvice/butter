import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import { WORKSPACE_KEY } from "@/lib/constants";
import type {
  BridgeDiagnostics,
  DaemonState,
  DaemonStatus,
  DaemonTaskInFlight,
  LatencyPoint,
} from "@/types/api";

const SVC = "agents.v1.DaemonService";

type UnknownRecord = Record<string, unknown>;

function asRecord(value: unknown): UnknownRecord {
  return value && typeof value === "object" ? (value as UnknownRecord) : {};
}

function asString(value: unknown): string | undefined {
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "bigint") return String(value);
  return undefined;
}

function asNumber(value: unknown): number | undefined {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "bigint") return Number(value);
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : undefined;
  }
  return undefined;
}

function asStringArray(value: unknown): string[] | undefined {
  return Array.isArray(value) ? value.map(asString).filter((v): v is string => !!v) : undefined;
}

function asDaemonState(value: unknown): DaemonState | undefined {
  if (typeof value === "string") return value as DaemonState;
  if (typeof value === "number") {
    return (["STATE_UNSPECIFIED", "STATE_ONLINE", "STATE_IDLE", "STATE_OFFLINE"] as const)[value];
  }
  return undefined;
}

function normalizeLatencyPoint(value: unknown): LatencyPoint {
  const r = asRecord(value);
  return {
    timestamp: asString(r.timestamp),
    latency_ms: asNumber(r.latency_ms ?? r.latencyMs),
  };
}

function normalizeDiagnostics(value: unknown): BridgeDiagnostics {
  const r = asRecord(value);
  return {
    cpu_percent: asNumber(r.cpu_percent ?? r.cpuPercent),
    memory_used_bytes: asNumber(r.memory_used_bytes ?? r.memoryUsedBytes),
    memory_limit_bytes: asNumber(r.memory_limit_bytes ?? r.memoryLimitBytes),
    goroutines: asNumber(r.goroutines),
    checked_at: asString(r.checked_at ?? r.checkedAt),
    latency: Array.isArray(r.latency) ? r.latency.map(normalizeLatencyPoint) : [],
  };
}

function normalizeDaemon(value: unknown): DaemonStatus {
  const r = asRecord(value);
  return {
    daemon_id: asString(r.daemon_id ?? r.daemonId) ?? "",
    name: asString(r.name),
    capabilities: asStringArray(r.capabilities) ?? [],
    labels: asRecord(r.labels) as Record<string, string>,
    state: asDaemonState(r.state),
    connected_at: asString(r.connected_at ?? r.connectedAt),
    uptime: asString(r.uptime),
    active_tasks: asNumber(r.active_tasks ?? r.activeTasks),
    version: asString(r.version),
    os: asString(r.os),
    executors: asStringArray(r.executors) ?? [],
    remote_addr: asString(r.remote_addr ?? r.remoteAddr),
    host: asString(r.host),
  };
}

function normalizeTask(value: unknown): DaemonTaskInFlight {
  const r = asRecord(value);
  return {
    task_id: asString(r.task_id ?? r.taskId) ?? "",
    daemon_id: asString(r.daemon_id ?? r.daemonId),
    daemon_name: asString(r.daemon_name ?? r.daemonName),
    capability: asString(r.capability),
    started_at: asString(r.started_at ?? r.startedAt),
    elapsed: asString(r.elapsed),
    current_step: asString(r.current_step ?? r.currentStep),
    progress: asNumber(r.progress),
    agent_name: asString(r.agent_name ?? r.agentName),
  };
}

async function listDaemons() {
  const res = await twirpFetch<object, { daemons?: unknown[] }>(SVC, "ListDaemons", {});
  return { daemons: (res.daemons ?? []).map(normalizeDaemon) };
}

async function getDaemon(daemonId: string) {
  const res = await twirpFetch<{ daemon_id: string }, { daemon?: unknown }>(
    SVC,
    "GetDaemon",
    { daemon_id: daemonId },
  );
  return { daemon: normalizeDaemon(res.daemon) };
}

async function listDaemonTasks(daemonId?: string) {
  const res = await twirpFetch<{ daemon_id?: string }, { tasks?: unknown[] }>(
    SVC,
    "ListDaemonTasks",
    { daemon_id: daemonId },
  );
  return { tasks: (res.tasks ?? []).map(normalizeTask) };
}

function cancelDaemonTask(taskId: string, daemonId?: string) {
  return twirpFetch<
    { task_id: string; daemon_id?: string },
    { daemon_id?: string; daemonId?: string }
  >(SVC, "CancelDaemonTask", { task_id: taskId, daemon_id: daemonId });
}

async function getBridgeDiagnostics() {
  const res = await twirpFetch<object, { diagnostics?: unknown }>(
    SVC,
    "GetBridgeDiagnostics",
    {},
  );
  return { diagnostics: normalizeDiagnostics(res.diagnostics) };
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
