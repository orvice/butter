import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import { durationFromMs } from "@bufbuild/protobuf/wkt";
import { WORKSPACE_KEY } from "@/lib/constants";
import {
  DaemonConfigSchema,
  DaemonService,
  DaemonStatus_State,
  type BridgeDiagnostics as PbBridgeDiagnostics,
  type DaemonConfig as PbDaemonConfig,
  type DaemonStatus as PbDaemonStatus,
  type DaemonTaskInFlight as PbDaemonTaskInFlight,
  type LatencyPoint as PbLatencyPoint,
} from "@/gen/agents/v1/dashboard_pb";
import type {
  BridgeDiagnostics,
  CreateDaemonCredentialInput,
  CreateDaemonCredentialResult,
  DaemonConfig,
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
    workspace_id: s.workspaceId,
  };
}

function toDaemonConfig(d: PbDaemonConfig): DaemonConfig {
  return {
    id: d.id,
    name: d.name,
    description: d.description,
    allowed_capabilities: d.allowedCapabilities,
    labels: d.labels,
    created_at: tsToISO(d.createdAt),
    created_by: d.createdBy,
    workspace_id: d.workspaceId,
  };
}

function toDaemonConfigProto(d: DaemonConfig): PbDaemonConfig {
  return create(DaemonConfigSchema, {
    id: d.id,
    name: d.name,
    description: d.description ?? "",
    allowedCapabilities: d.allowed_capabilities ?? [],
    labels: d.labels ?? {},
  });
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
    workspace_id: t.workspaceId,
  };
}

async function listDaemonConfigs(): Promise<{ daemons: DaemonConfig[] }> {
  const res = await client.listDaemonConfigs({});
  return { daemons: res.daemons.map(toDaemonConfig) };
}

async function getDaemonConfig(id: string): Promise<{ daemon: DaemonConfig }> {
  const res = await client.getDaemonConfig({ id });
  if (!res.daemon) throw new Error("daemon config not found");
  return { daemon: toDaemonConfig(res.daemon) };
}

async function createDaemonConfig(daemon: DaemonConfig): Promise<{ daemon: DaemonConfig }> {
  const res = await client.createDaemonConfig({ daemon: toDaemonConfigProto(daemon) });
  if (!res.daemon) throw new Error("server returned no daemon config");
  return { daemon: toDaemonConfig(res.daemon) };
}

async function updateDaemonConfig(daemon: DaemonConfig): Promise<{ daemon: DaemonConfig }> {
  const res = await client.updateDaemonConfig({ daemon: toDaemonConfigProto(daemon) });
  if (!res.daemon) throw new Error("server returned no daemon config");
  return { daemon: toDaemonConfig(res.daemon) };
}

async function deleteDaemonConfig(id: string): Promise<void> {
  await client.deleteDaemonConfig({ id });
}

async function createDaemonCredential(input: CreateDaemonCredentialInput): Promise<CreateDaemonCredentialResult> {
  const ttl = input.ttl_hours && input.ttl_hours > 0 ? durationFromMs(input.ttl_hours * 60 * 60 * 1000) : undefined;
  const res = await client.createDaemonCredential({
    daemonId: input.daemon_id,
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

export function useDaemonConfigs() {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["daemon-configs", workspaceId],
    queryFn: listDaemonConfigs,
    enabled: !!workspaceId,
  });
}

export function useDaemonConfig(id: string) {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["daemon-configs", workspaceId, id],
    queryFn: () => getDaemonConfig(id),
    enabled: !!workspaceId && !!id,
  });
}

export function useCreateDaemonConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createDaemonConfig,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["daemon-configs"] }),
  });
}

export function useUpdateDaemonConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateDaemonConfig,
    onSuccess: (_res, daemon) => {
      qc.invalidateQueries({ queryKey: ["daemon-configs"] });
      qc.invalidateQueries({ queryKey: ["daemon-configs", localStorage.getItem(WORKSPACE_KEY), daemon.id] });
    },
  });
}

export function useDeleteDaemonConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteDaemonConfig,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["daemon-configs"] }),
  });
}

export function useCreateDaemonCredential() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createDaemonCredential,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["daemon-configs"] });
      qc.invalidateQueries({ queryKey: ["api-tokens"] });
    },
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
