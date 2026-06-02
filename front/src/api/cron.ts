import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import {
  CronDeliverySchema,
  CronDeliveryType,
  CronExecutionStatus,
  CronJobSchema,
  CronJobService,
  type CronExecution as PbCronExecution,
  type CronJob as PbCronJob,
} from "@/gen/agents/v1/cron_pb";
import type {
  CronDelivery,
  CronDeliveryType as LegacyDeliveryType,
  CronExecution,
  CronExecutionStatus as LegacyExecStatus,
  CronJob,
} from "@/types/api";
import { tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(CronJobService);

const DELIVERY_TYPE_NAMES: LegacyDeliveryType[] = [
  "CRON_DELIVERY_TYPE_UNSPECIFIED",
  "CRON_DELIVERY_TYPE_LOG",
  "CRON_DELIVERY_TYPE_WEBHOOK",
  "CRON_DELIVERY_TYPE_CHANNEL",
  "CRON_DELIVERY_TYPE_NOTIFY_GROUP",
];

const EXECUTION_STATUS_NAMES: LegacyExecStatus[] = [
  "CRON_EXECUTION_STATUS_UNSPECIFIED",
  "CRON_EXECUTION_STATUS_SUCCESS",
  "CRON_EXECUTION_STATUS_ERROR",
];

function deliveryTypeFromProto(t: CronDeliveryType): LegacyDeliveryType {
  return DELIVERY_TYPE_NAMES[t] ?? "CRON_DELIVERY_TYPE_UNSPECIFIED";
}

function deliveryTypeToProto(t: LegacyDeliveryType | undefined): CronDeliveryType {
  const idx = DELIVERY_TYPE_NAMES.indexOf(t ?? "CRON_DELIVERY_TYPE_UNSPECIFIED");
  return idx < 0 ? CronDeliveryType.UNSPECIFIED : (idx as CronDeliveryType);
}

function execStatusFromProto(s: CronExecutionStatus): LegacyExecStatus {
  return EXECUTION_STATUS_NAMES[s] ?? "CRON_EXECUTION_STATUS_UNSPECIFIED";
}

function deliveryFromProto(d: PbCronJob["delivery"]): CronDelivery | undefined {
  if (!d) return undefined;
  return {
    type: deliveryTypeFromProto(d.type),
    webhook_url: d.webhookUrl,
    channel_name: d.channelName,
    chat_id: d.chatId,
    notify_group_name: d.notifyGroupName,
  };
}

function deliveryToProto(d: CronDelivery | undefined): PbCronJob["delivery"] | undefined {
  if (!d) return undefined;
  return create(CronDeliverySchema, {
    type: deliveryTypeToProto(d.type),
    webhookUrl: d.webhook_url ?? "",
    channelName: d.channel_name ?? "",
    chatId: d.chat_id ?? "",
    notifyGroupName: d.notify_group_name ?? "",
  });
}

function jobFromProto(j: PbCronJob): CronJob {
  return {
    name: j.name,
    schedule: j.schedule,
    agent_name: j.agentName,
    input: j.input,
    timezone: j.timezone,
    enabled: j.enabled,
    delivery: deliveryFromProto(j.delivery),
    metadata: j.metadata,
  };
}

function jobToProto(j: CronJob): PbCronJob {
  return create(CronJobSchema, {
    name: j.name,
    schedule: j.schedule,
    agentName: j.agent_name,
    input: j.input ?? "",
    timezone: j.timezone ?? "",
    enabled: j.enabled ?? false,
    delivery: deliveryToProto(j.delivery),
    metadata: j.metadata ?? {},
  });
}

function execFromProto(e: PbCronExecution): CronExecution {
  return {
    id: e.id,
    job_name: e.jobName,
    agent_name: e.agentName,
    status: execStatusFromProto(e.status),
    input: e.input,
    output: e.output,
    started_at: tsToISO(e.startedAt),
    finished_at: tsToISO(e.finishedAt),
  };
}

async function listCronJobs(): Promise<{ cron_jobs: CronJob[] }> {
  const res = await client.listCronJobs({});
  return { cron_jobs: res.cronJobs.map(jobFromProto) };
}

async function getCronJob(name: string): Promise<{ cron_job: CronJob }> {
  const res = await client.getCronJob({ name });
  if (!res.cronJob) throw new Error("not found");
  return { cron_job: jobFromProto(res.cronJob) };
}

async function createCronJob(job: CronJob): Promise<{ cron_job: CronJob }> {
  const res = await client.createCronJob({ cronJob: jobToProto(job) });
  if (!res.cronJob) throw new Error("create returned nothing");
  return { cron_job: jobFromProto(res.cronJob) };
}

async function updateCronJob(job: CronJob): Promise<{ cron_job: CronJob }> {
  const res = await client.updateCronJob({ cronJob: jobToProto(job) });
  if (!res.cronJob) throw new Error("update returned nothing");
  return { cron_job: jobFromProto(res.cronJob) };
}

async function deleteCronJob(name: string): Promise<{ cron_job: CronJob | undefined }> {
  const res = await client.deleteCronJob({ name });
  return { cron_job: res.cronJob ? jobFromProto(res.cronJob) : undefined };
}

async function runCronJobNow(name: string): Promise<{ execution: CronExecution }> {
  const res = await client.runCronJobNow({ name });
  if (!res.execution) throw new Error("no execution returned");
  return { execution: execFromProto(res.execution) };
}

interface ListExecutionsParams {
  job_name?: string;
  page_size?: number;
  page_token?: string;
}

async function listCronExecutions(
  params: ListExecutionsParams,
): Promise<{ executions: CronExecution[]; next_page_token: string }> {
  const res = await client.listCronExecutions({
    jobName: params.job_name ?? "",
    pageSize: params.page_size ?? 0,
    pageToken: params.page_token ?? "",
  });
  return {
    executions: res.executions.map(execFromProto),
    next_page_token: res.nextPageToken,
  };
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
