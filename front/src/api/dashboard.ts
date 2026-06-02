import { useQuery } from "@tanstack/react-query";
import { WORKSPACE_KEY } from "@/lib/constants";
import {
  ComponentHealth_Status,
  DashboardService,
  GetCronExecutionTimeseriesRequest_Range,
  type ActivityEvent as PbActivityEvent,
  type ComponentHealth as PbComponentHealth,
  type CronExecutionBucket as PbCronExecutionBucket,
  type DaemonHandshake as PbDaemonHandshake,
  type GetOverviewResponse as PbGetOverviewResponse,
  type HealthSummary as PbHealthSummary,
  type OverviewCounts as PbOverviewCounts,
} from "@/gen/agents/v1/dashboard_pb";
import type {
  ActivityEvent,
  ComponentHealth,
  ComponentHealthStatus,
  CronExecutionBucket,
  CronTimeseriesRange,
  DaemonHandshake,
  GetOverviewResponse,
  HealthSummary,
  OverviewCounts,
} from "@/types/api";
import { tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(DashboardService);

const HEALTH_STATUS_NAMES: ComponentHealthStatus[] = [
  "STATUS_UNSPECIFIED",
  "STATUS_HEALTHY",
  "STATUS_DEGRADED",
  "STATUS_DOWN",
];

const RANGE_NAMES: CronTimeseriesRange[] = [
  "RANGE_UNSPECIFIED",
  "RANGE_1D",
  "RANGE_7D",
  "RANGE_30D",
];

function healthStatusFromProto(s: ComponentHealth_Status): ComponentHealthStatus {
  return HEALTH_STATUS_NAMES[s] ?? "STATUS_UNSPECIFIED";
}

function rangeToProto(r: CronTimeseriesRange): GetCronExecutionTimeseriesRequest_Range {
  const idx = RANGE_NAMES.indexOf(r);
  return idx < 0
    ? GetCronExecutionTimeseriesRequest_Range.RANGE_UNSPECIFIED
    : (idx as GetCronExecutionTimeseriesRequest_Range);
}

function countsFromProto(c: PbOverviewCounts | undefined): OverviewCounts | undefined {
  if (!c) return undefined;
  return {
    active_agents: c.activeAgents,
    mcp_servers: c.mcpServers,
    connected_daemons: c.connectedDaemons,
    remote_agents: c.remoteAgents,
    channels: c.channels,
    cron_jobs: c.cronJobs,
    active_sessions: c.activeSessions,
  };
}

function componentHealthFromProto(h: PbComponentHealth | undefined): ComponentHealth | undefined {
  if (!h) return undefined;
  return {
    status: healthStatusFromProto(h.status),
    detail: h.detail,
    checked_at: tsToISO(h.checkedAt),
    latency_ms: h.latencyMs,
  };
}

function healthFromProto(h: PbHealthSummary | undefined): HealthSummary | undefined {
  if (!h) return undefined;
  return {
    mongodb: componentHealthFromProto(h.mongodb),
    redis: componentHealthFromProto(h.redis),
    runner: componentHealthFromProto(h.runner),
  };
}

function handshakeFromProto(d: PbDaemonHandshake | undefined): DaemonHandshake | undefined {
  if (!d) return undefined;
  return {
    daemon_id: d.daemonId,
    name: d.name,
    capabilities: d.capabilities,
    connected_at: tsToISO(d.connectedAt),
    os: d.os,
  };
}

function overviewFromProto(r: PbGetOverviewResponse): GetOverviewResponse {
  return {
    counts: countsFromProto(r.counts),
    health: healthFromProto(r.health),
    latest_daemon_handshake: handshakeFromProto(r.latestDaemonHandshake),
  };
}

function activityFromProto(e: PbActivityEvent): ActivityEvent {
  return {
    id: e.id,
    kind: e.kind,
    actor: e.actor,
    message: e.message,
    timestamp: tsToISO(e.timestamp),
  };
}

function bucketFromProto(b: PbCronExecutionBucket): CronExecutionBucket {
  return {
    start: tsToISO(b.start),
    success: b.success,
    error: b.error,
  };
}

async function getOverview(environment?: string): Promise<GetOverviewResponse> {
  const res = await client.getOverview({ environment: environment ?? "" });
  return overviewFromProto(res);
}

async function getActivityFeed(
  limit = 20,
  pageToken?: string,
): Promise<{ events?: ActivityEvent[]; next_page_token?: string }> {
  const res = await client.getActivityFeed({ limit, pageToken: pageToken ?? "" });
  return {
    events: res.events.map(activityFromProto),
    next_page_token: res.nextPageToken,
  };
}

async function getCronTimeseries(
  range: CronTimeseriesRange,
  jobName?: string,
): Promise<{ buckets?: CronExecutionBucket[] }> {
  const res = await client.getCronExecutionTimeseries({
    range: rangeToProto(range),
    jobName: jobName ?? "",
  });
  return { buckets: res.buckets.map(bucketFromProto) };
}

export function useOverview(environment?: string) {
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  return useQuery({
    queryKey: ["dashboard", "overview", workspaceId, environment ?? ""],
    queryFn: () => getOverview(environment),
    enabled: !!workspaceId,
    refetchInterval: 30_000,
  });
}

export function useActivityFeed(limit = 20) {
  return useQuery({
    queryKey: ["dashboard", "activity", limit],
    queryFn: () => getActivityFeed(limit),
    refetchInterval: 30_000,
  });
}

export function useCronTimeseries(range: CronTimeseriesRange = "RANGE_1D", jobName?: string) {
  return useQuery({
    queryKey: ["dashboard", "cron-timeseries", range, jobName],
    queryFn: () => getCronTimeseries(range, jobName),
    refetchInterval: 60_000,
  });
}
