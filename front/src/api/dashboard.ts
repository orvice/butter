import { useQuery } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import { WORKSPACE_KEY } from "@/lib/constants";
import type {
  ActivityEvent,
  CronExecutionBucket,
  CronTimeseriesRange,
  GetOverviewResponse,
} from "@/types/api";

const SVC = "agents.v1.DashboardService";

function getOverview(environment?: string) {
  return twirpFetch<{ environment?: string }, GetOverviewResponse>(
    SVC,
    "GetOverview",
    { environment },
  );
}

function getActivityFeed(limit = 20, pageToken?: string) {
  return twirpFetch<
    { limit?: number; page_token?: string },
    { events?: ActivityEvent[]; next_page_token?: string }
  >(SVC, "GetActivityFeed", { limit, page_token: pageToken });
}

function getCronTimeseries(range: CronTimeseriesRange, jobName?: string) {
  return twirpFetch<
    { range: CronTimeseriesRange; job_name?: string },
    { buckets?: CronExecutionBucket[] }
  >(SVC, "GetCronExecutionTimeseries", { range, job_name: jobName });
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
