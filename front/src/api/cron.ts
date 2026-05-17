import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { CronJob, CronExecution } from "@/types/api";

const SVC = "agents.v1.CronJobService";

function listCronJobs() {
  return twirpFetch<object, { cron_jobs: CronJob[] }>(SVC, "ListCronJobs", {});
}

function getCronJob(name: string) {
  return twirpFetch<{ name: string }, { cron_job: CronJob }>(SVC, "GetCronJob", { name });
}

function createCronJob(cron_job: CronJob) {
  return twirpFetch<{ cron_job: CronJob }, { cron_job: CronJob }>(SVC, "CreateCronJob", { cron_job });
}

function updateCronJob(cron_job: CronJob) {
  return twirpFetch<{ cron_job: CronJob }, { cron_job: CronJob }>(SVC, "UpdateCronJob", { cron_job });
}

function deleteCronJob(name: string) {
  return twirpFetch<{ name: string }, { cron_job: CronJob }>(SVC, "DeleteCronJob", { name });
}

function runCronJobNow(name: string) {
  return twirpFetch<{ name: string }, { execution: CronExecution }>(
    SVC,
    "RunCronJobNow",
    { name },
  );
}

interface ListExecutionsParams {
  job_name?: string;
  page_size?: number;
  page_token?: string;
}

function listCronExecutions(params: ListExecutionsParams) {
  return twirpFetch<ListExecutionsParams, { executions: CronExecution[]; next_page_token: string }>(
    SVC,
    "ListCronExecutions",
    params,
  );
}

export function useCronJobs() {
  return useQuery({ queryKey: ["cron-jobs"], queryFn: listCronJobs });
}

export function useCronJob(name: string) {
  return useQuery({ queryKey: ["cron-jobs", name], queryFn: () => getCronJob(name), enabled: !!name });
}

export function useCreateCronJob() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createCronJob,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["cron-jobs"] }),
  });
}

export function useUpdateCronJob() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateCronJob,
    onSuccess: (_data, job) => {
      qc.invalidateQueries({ queryKey: ["cron-jobs"] });
      qc.invalidateQueries({ queryKey: ["cron-jobs", job.name] });
    },
  });
}

export function useDeleteCronJob() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteCronJob,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["cron-jobs"] }),
  });
}

export function useRunCronJobNow() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: runCronJobNow,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["cron-executions"] });
    },
  });
}

export function useCronExecutions(jobName?: string, pageSize?: number, pageToken?: string) {
  return useQuery({
    queryKey: ["cron-executions", { jobName, page: pageToken }],
    queryFn: () => listCronExecutions({ job_name: jobName, page_size: pageSize, page_token: pageToken }),
  });
}

export function useDashboardExecutions() {
  return useQuery({
    queryKey: ["cron-executions", { jobName: "", page: "" }],
    queryFn: () => listCronExecutions({ page_size: 100 }),
    refetchInterval: 60_000,
  });
}
