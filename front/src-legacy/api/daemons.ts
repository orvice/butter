import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import { durationFromMs } from "@bufbuild/protobuf/wkt";
import { WORKSPACE_KEY } from "@/lib/constants";
import {
  DaemonRuntimeSchema,
  DaemonService,
  DaemonStatus_State,
  type BridgeDiagnostics as PbBridgeDiagnostics,
  type DaemonRuntime as PbDaemonRuntime,
  type DaemonStatus as PbDaemonStatus,
  type DaemonTaskInFlight as PbDaemonTaskInFlight,
  type LatencyPoint as PbLatencyPoint,
} from "@/gen/agents/v1/dashboard_pb";
import type {
  BridgeDiagnostics,
  CreateDaemonRuntimeTokenInput,
  CreateDaemonRuntimeTokenResult,
  DaemonRuntime,
  DaemonState,
  DaemonStatus,
  DaemonTaskInFlight,
  LatencyPoint,
} from "@/types/api";
import { toAPIToken } from "./apitokens";
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
    daemon_runtime_id: s.daemonRuntimeId,
    name: s.name,
    acp_runtimes: s.acpRuntimes,
    labels: s.labels,
    state: toDaemonState(s.state),
    connected_at: tsToISO(s.connectedAt),
    uptime: durationToString(s.uptime),
    active_tasks: s.activeTasks,
    version: s.version,
    os: s.os,
    executors: s.executors,
    remote_addr: s.remoteAddr,
    workspace_id: s.workspaceId,
  };
}

function toDaemonRuntime(d: PbDaemonRuntime): DaemonRuntime {
  return {
    id: d.id,
    name: d.name,
    description: d.description,
    labels: d.labels,
    created_at: tsToISO(d.createdAt),
    created_by: d.createdBy,
    workspace_id: d.workspaceId,
  };
}

function toDaemonRuntimeProto(d: DaemonRuntime): PbDaemonRuntime {
  return create(DaemonRuntimeSchema, {
    id: d.id,
    name: d.name,
    description: d.description ?? "",
    labels: d.labels ?? {},
  });
}

function toTaskInFlight(t: PbDaemonTaskInFlight): DaemonTaskInFlight {
  return {
    task_id: t.taskId,
    daemon_runtime_id: t.daemonRuntimeId,
    daemon_name: t.daemonName,
    acp_runtime: t.acpRuntime,
    started_at: tsToISO(t.startedAt),
    elapsed: durationToString(t.elapsed),
    current_step: t.currentStep,
    progress: t.progress,
    agent_name: t.agentName,
    workspace_id: t.workspaceId,
  };
}

async function listDaemonRuntimes(): Promise<{ runtimes: DaemonRuntime[] }> {
  const res = await client.listDaemonRuntimes({});
  return { runtimes: res.runtimes.map(toDaemonRuntime) };
}

async function getDaemonRuntime(id: string): Promise<{ runtime: DaemonRuntime }> {
  const res = await client.getDaemonRuntime({ id });
  if (!res.runtime) throw new Error("daemon runtime not found");
  return { runtime: toDaemonRuntime(res.runtime) };
}

async function createDaemonRuntime(runtime: DaemonRuntime): Promise<{ runtime: DaemonRuntime }> {
  const res = await client.createDaemonRuntime({ runtime: toDaemonRuntimeProto(runtime) });
  if (!res.runtime) throw new Error("server returned no daemon runtime");
  return { runtime: toDaemonRuntime(res.runtime) };
}

async function updateDaemonRuntime(runtime: DaemonRuntime): Promise<{ runtime: DaemonRuntime }> {
  const res = await client.updateDaemonRuntime({ runtime: toDaemonRuntimeProto(runtime) });
  if (!res.runtime) throw new Error("server returned no daemon runtime");
  return { runtime: toDaemonRuntime(res.runtime) };
}

async function deleteDaemonRuntime(id: string): Promise<void> {
  await client.deleteDaemonRuntime({ id });
}

async function createDaemonRuntimeToken(input: CreateDaemonRuntimeTokenInput): Promise<CreateDaemonRuntimeTokenResult> {
  const ttl = input.ttl_hours && input.ttl_hours > 0 ? durationFromMs(input.ttl_hours * 60 * 60 * 1000) : undefined;
  const res = await client.createDaemonRuntimeToken({
    daemonRuntimeId: input.daemon_runtime_id,
    name: input.name ?? "",
    ttl,
  });
  return {
    token: res.token ? toAPIToken(res.token) : undefined,
    secret: res.secret,
  };
}

async function listDaemons(): Promise<{ daemons: DaemonStatus[] }> {
  const res = await client.listDaemons({});
  return { daemons: res.daemons.map(toDaemonStatus) };
}

async function getDaemon(daemonRuntimeId: string): Promise<{ daemon: DaemonStatus }> {
  const res = await client.getDaemon({ daemonRuntimeId });
  if (!res.daemon) throw new Error("daemon not found");
  return { daemon: toDaemonStatus(res.daemon) };
}

async function listDaemonTasks(daemonRuntimeId?: string): Promise<{ tasks: DaemonTaskInFlight[] }> {
  const res = await client.listDaemonTasks({ daemonRuntimeId: daemonRuntimeId ?? "" });
  return { tasks: res.tasks.map(toTaskInFlight) };
}

async function cancelDaemonTask(taskId: string, daemonRuntimeId?: string): Promise<void> {
  await client.cancelDaemonTask({ taskId, daemonRuntimeId: daemonRuntimeId ?? "" });
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

export function useDaemonRuntimes() {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["daemon-runtimes", workspaceId],
    queryFn: listDaemonRuntimes,
    enabled: !!workspaceId,
  });
}

export function useDaemonRuntime(id: string) {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["daemon-runtimes", workspaceId, id],
    queryFn: () => getDaemonRuntime(id),
    enabled: !!workspaceId && !!id,
  });
}

export function useCreateDaemonRuntime() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createDaemonRuntime,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["daemon-runtimes"] }),
  });
}

export function useUpdateDaemonRuntime() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateDaemonRuntime,
    onSuccess: (_res, runtime) => {
      qc.invalidateQueries({ queryKey: ["daemon-runtimes"] });
      qc.invalidateQueries({ queryKey: ["daemon-runtimes", localStorage.getItem(WORKSPACE_KEY), runtime.id] });
    },
  });
}

export function useDeleteDaemonRuntime() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteDaemonRuntime,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["daemon-runtimes"] }),
  });
}

export function useCreateDaemonRuntimeToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createDaemonRuntimeToken,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["daemon-runtimes"] });
      qc.invalidateQueries({ queryKey: ["api-tokens"] });
    },
  });
}

export function useDaemon(daemonRuntimeId: string) {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["daemons", workspaceId, daemonRuntimeId],
    queryFn: () => getDaemon(daemonRuntimeId),
    enabled: !!workspaceId && !!daemonRuntimeId,
  });
}

export function useDaemonTasks(daemonRuntimeId?: string) {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["daemons", workspaceId, "tasks", daemonRuntimeId ?? "all"],
    queryFn: () => listDaemonTasks(daemonRuntimeId),
    enabled: !!workspaceId,
    refetchInterval: 5_000,
  });
}

export function useCancelDaemonTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, daemonRuntimeId }: { taskId: string; daemonRuntimeId?: string }) =>
      cancelDaemonTask(taskId, daemonRuntimeId),
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
