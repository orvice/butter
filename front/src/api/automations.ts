import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import { durationFromMs, type Duration } from "@bufbuild/protobuf/wkt";
import {
  AutomationCallWebhookStepSchema,
  AutomationConcurrencyPolicy,
  AutomationConditionOperator,
  AutomationConditionSchema,
  AutomationCreateForumPostStepSchema,
  AutomationInvokeAgentStepSchema,
  AutomationPolicySchema,
  AutomationRetryPolicySchema,
  AutomationSchema,
  AutomationScheduleTriggerSchema,
  AutomationSendNotifyGroupStepSchema,
  AutomationService,
  AutomationStepSchema,
  AutomationStepType,
  AutomationTriggerSchema,
  AutomationTriggerType,
  type Automation as PbAutomation,
  type AutomationRun as PbAutomationRun,
  type AutomationStep as PbAutomationStep,
  type AutomationStepRun as PbAutomationStepRun,
} from "@/gen/agents/v1/automation_pb";
import type {
  Automation,
  AutomationCondition,
  AutomationConditionOperator as LegacyConditionOperator,
  AutomationConcurrencyPolicy as LegacyConcurrencyPolicy,
  AutomationPolicy,
  AutomationRun,
  AutomationRunStatus as LegacyRunStatus,
  AutomationStep,
  AutomationStepRun,
  AutomationStepRunStatus as LegacyStepRunStatus,
  AutomationStepType as LegacyStepType,
  AutomationTriggerType as LegacyTriggerType,
} from "@/types/api";
import { tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(AutomationService);

const TRIGGER_NAMES: LegacyTriggerType[] = [
  "AUTOMATION_TRIGGER_TYPE_UNSPECIFIED",
  "AUTOMATION_TRIGGER_TYPE_MANUAL",
  "AUTOMATION_TRIGGER_TYPE_SCHEDULE",
  "AUTOMATION_TRIGGER_TYPE_WEBHOOK",
  "AUTOMATION_TRIGGER_TYPE_FORUM_EVENT",
  "AUTOMATION_TRIGGER_TYPE_CHANNEL_EVENT",
  "AUTOMATION_TRIGGER_TYPE_DAEMON_EVENT",
];

const CONDITION_NAMES: LegacyConditionOperator[] = [
  "AUTOMATION_CONDITION_OPERATOR_UNSPECIFIED",
  "AUTOMATION_CONDITION_OPERATOR_EQUALS",
  "AUTOMATION_CONDITION_OPERATOR_NOT_EQUALS",
  "AUTOMATION_CONDITION_OPERATOR_CONTAINS",
  "AUTOMATION_CONDITION_OPERATOR_REGEX_MATCH",
  "AUTOMATION_CONDITION_OPERATOR_EXISTS",
  "AUTOMATION_CONDITION_OPERATOR_NOT_EXISTS",
];

const STEP_NAMES: LegacyStepType[] = [
  "AUTOMATION_STEP_TYPE_UNSPECIFIED",
  "AUTOMATION_STEP_TYPE_INVOKE_AGENT",
  "AUTOMATION_STEP_TYPE_CALL_WEBHOOK",
  "AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP",
  "AUTOMATION_STEP_TYPE_CREATE_FORUM_POST",
];

const CONCURRENCY_NAMES: LegacyConcurrencyPolicy[] = [
  "AUTOMATION_CONCURRENCY_POLICY_UNSPECIFIED",
  "AUTOMATION_CONCURRENCY_POLICY_SKIP",
  "AUTOMATION_CONCURRENCY_POLICY_QUEUE",
  "AUTOMATION_CONCURRENCY_POLICY_REPLACE",
  "AUTOMATION_CONCURRENCY_POLICY_ALLOW",
];

const RUN_STATUS_NAMES: LegacyRunStatus[] = [
  "AUTOMATION_RUN_STATUS_UNSPECIFIED",
  "AUTOMATION_RUN_STATUS_RUNNING",
  "AUTOMATION_RUN_STATUS_SUCCEEDED",
  "AUTOMATION_RUN_STATUS_FAILED",
  "AUTOMATION_RUN_STATUS_SKIPPED",
  "AUTOMATION_RUN_STATUS_CANCELLED",
];

const STEP_RUN_STATUS_NAMES: LegacyStepRunStatus[] = [
  "AUTOMATION_STEP_RUN_STATUS_UNSPECIFIED",
  "AUTOMATION_STEP_RUN_STATUS_RUNNING",
  "AUTOMATION_STEP_RUN_STATUS_SUCCEEDED",
  "AUTOMATION_STEP_RUN_STATUS_FAILED",
  "AUTOMATION_STEP_RUN_STATUS_SKIPPED",
  "AUTOMATION_STEP_RUN_STATUS_CANCELLED",
];

function enumName<T extends string>(names: T[], value: number, fallback: T): T {
  return names[value] ?? fallback;
}

function enumValue<T extends string>(names: T[], value: T | undefined, fallback = 0): number {
  const idx = names.indexOf(value ?? names[fallback]);
  return idx < 0 ? fallback : idx;
}

function durationToSeconds(d?: Duration): number | undefined {
  if (!d) return undefined;
  return Number(d.seconds ?? 0) + d.nanos / 1_000_000_000;
}

function secondsToDuration(seconds?: number): Duration | undefined {
  if (!seconds || seconds <= 0) return undefined;
  return durationFromMs(seconds * 1000);
}

function policyFromProto(p: PbAutomation["policy"]): AutomationPolicy | undefined {
  if (!p) return undefined;
  return {
    timeout_seconds: durationToSeconds(p.timeout),
    retry: p.retry
      ? { max_attempts: p.retry.maxAttempts, backoff_seconds: durationToSeconds(p.retry.backoff) }
      : undefined,
    concurrency: enumName(CONCURRENCY_NAMES, p.concurrency, "AUTOMATION_CONCURRENCY_POLICY_UNSPECIFIED"),
    max_output_bytes: p.maxOutputBytes,
  };
}

function policyToProto(p?: AutomationPolicy): PbAutomation["policy"] | undefined {
  if (!p) return undefined;
  return create(AutomationPolicySchema, {
    timeout: secondsToDuration(p.timeout_seconds),
    retry: p.retry
      ? create(AutomationRetryPolicySchema, {
          maxAttempts: p.retry.max_attempts ?? 0,
          backoff: secondsToDuration(p.retry.backoff_seconds),
        })
      : undefined,
    concurrency: enumValue(CONCURRENCY_NAMES, p.concurrency) as AutomationConcurrencyPolicy,
    maxOutputBytes: p.max_output_bytes ?? 0,
  });
}

function conditionFromProto(c: PbAutomation["conditions"][number]): AutomationCondition {
  return {
    selector: c.selector,
    operator: enumName(CONDITION_NAMES, c.operator, "AUTOMATION_CONDITION_OPERATOR_UNSPECIFIED"),
    value: c.value,
  };
}

function conditionToProto(c: AutomationCondition) {
  return create(AutomationConditionSchema, {
    selector: c.selector,
    operator: enumValue(CONDITION_NAMES, c.operator) as AutomationConditionOperator,
    value: c.value ?? "",
  });
}

function stepFromProto(s: PbAutomationStep): AutomationStep {
  return {
    name: s.name,
    type: enumName(STEP_NAMES, s.type, "AUTOMATION_STEP_TYPE_UNSPECIFIED"),
    invoke_agent: s.invokeAgent
      ? { agent_name: s.invokeAgent.agentName, input: s.invokeAgent.input, model_override: s.invokeAgent.modelOverride }
      : undefined,
    call_webhook: s.callWebhook
      ? { url: s.callWebhook.url, method: s.callWebhook.method, payload_json: s.callWebhook.payloadJson, headers: s.callWebhook.headers }
      : undefined,
    send_notify_group: s.sendNotifyGroup
      ? { notify_group_name: s.sendNotifyGroup.notifyGroupName, title: s.sendNotifyGroup.title, message: s.sendNotifyGroup.message }
      : undefined,
    create_forum_post: s.createForumPost
      ? { thread_id: s.createForumPost.threadId, body: s.createForumPost.body }
      : undefined,
    policy: policyFromProto(s.policy),
  };
}

function stepToProto(s: AutomationStep): PbAutomationStep {
  return create(AutomationStepSchema, {
    name: s.name,
    type: enumValue(STEP_NAMES, s.type) as AutomationStepType,
    invokeAgent: s.invoke_agent
      ? create(AutomationInvokeAgentStepSchema, {
          agentName: s.invoke_agent.agent_name,
          input: s.invoke_agent.input ?? "",
          modelOverride: s.invoke_agent.model_override ?? "",
        })
      : undefined,
    callWebhook: s.call_webhook
      ? create(AutomationCallWebhookStepSchema, {
          url: s.call_webhook.url,
          method: s.call_webhook.method ?? "",
          payloadJson: s.call_webhook.payload_json ?? "",
          headers: s.call_webhook.headers ?? {},
        })
      : undefined,
    sendNotifyGroup: s.send_notify_group
      ? create(AutomationSendNotifyGroupStepSchema, {
          notifyGroupName: s.send_notify_group.notify_group_name,
          title: s.send_notify_group.title ?? "",
          message: s.send_notify_group.message ?? "",
        })
      : undefined,
    createForumPost: s.create_forum_post
      ? create(AutomationCreateForumPostStepSchema, {
          threadId: s.create_forum_post.thread_id,
          body: s.create_forum_post.body,
        })
      : undefined,
    policy: policyToProto(s.policy),
  });
}

function automationFromProto(a: PbAutomation): Automation {
  return {
    name: a.name,
    enabled: a.enabled,
    trigger: a.trigger
      ? {
          type: enumName(TRIGGER_NAMES, a.trigger.type, "AUTOMATION_TRIGGER_TYPE_UNSPECIFIED"),
          schedule: a.trigger.schedule ? { schedule: a.trigger.schedule.schedule, timezone: a.trigger.schedule.timezone } : undefined,
        }
      : undefined,
    conditions: a.conditions.map(conditionFromProto),
    steps: a.steps.map(stepFromProto),
    policy: policyFromProto(a.policy),
    metadata: a.metadata,
    created_at: tsToISO(a.createdAt),
    updated_at: tsToISO(a.updatedAt),
    workspace_id: a.workspaceId,
  };
}

function automationToProto(a: Automation): PbAutomation {
  return create(AutomationSchema, {
    name: a.name,
    enabled: a.enabled ?? false,
    trigger: create(AutomationTriggerSchema, {
      type: enumValue(TRIGGER_NAMES, a.trigger?.type, 1) as AutomationTriggerType,
      schedule:
        a.trigger?.type === "AUTOMATION_TRIGGER_TYPE_SCHEDULE"
          ? create(AutomationScheduleTriggerSchema, {
              schedule: a.trigger.schedule?.schedule ?? "",
              timezone: a.trigger.schedule?.timezone ?? "UTC",
            })
          : undefined,
    }),
    conditions: (a.conditions ?? []).map(conditionToProto),
    steps: (a.steps ?? []).map(stepToProto),
    policy: policyToProto(a.policy),
    metadata: a.metadata ?? {},
  });
}

function runFromProto(r: PbAutomationRun): AutomationRun {
  return {
    id: r.id,
    automation_name: r.automationName,
    trigger_type: enumName(TRIGGER_NAMES, r.triggerType, "AUTOMATION_TRIGGER_TYPE_UNSPECIFIED"),
    status: enumName(RUN_STATUS_NAMES, r.status, "AUTOMATION_RUN_STATUS_UNSPECIFIED"),
    trigger_payload_json: r.triggerPayloadJson,
    error: r.error,
    started_at: tsToISO(r.startedAt),
    finished_at: tsToISO(r.finishedAt),
    duration_ms: Number(r.durationMs),
    workspace_id: r.workspaceId,
  };
}

function stepRunFromProto(s: PbAutomationStepRun): AutomationStepRun {
  return {
    id: s.id,
    run_id: s.runId,
    automation_name: s.automationName,
    step_name: s.stepName,
    step_type: enumName(STEP_NAMES, s.stepType, "AUTOMATION_STEP_TYPE_UNSPECIFIED"),
    status: enumName(STEP_RUN_STATUS_NAMES, s.status, "AUTOMATION_STEP_RUN_STATUS_UNSPECIFIED"),
    attempt_count: s.attemptCount,
    input_json: s.inputJson,
    output_json: s.outputJson,
    error: s.error,
    invocation_id: s.invocationId,
    started_at: tsToISO(s.startedAt),
    finished_at: tsToISO(s.finishedAt),
    duration_ms: Number(s.durationMs),
    order: s.order,
    truncated: s.truncated,
    workspace_id: s.workspaceId,
  };
}

async function listAutomations(): Promise<{ automations: Automation[] }> {
  const res = await client.listAutomations({});
  return { automations: res.automations.map(automationFromProto) };
}

async function getAutomation(name: string): Promise<{ automation: Automation }> {
  const res = await client.getAutomation({ name });
  if (!res.automation) throw new Error("not found");
  return { automation: automationFromProto(res.automation) };
}

async function createAutomation(automation: Automation): Promise<{ automation: Automation }> {
  const res = await client.createAutomation({ automation: automationToProto(automation) });
  if (!res.automation) throw new Error("create returned nothing");
  return { automation: automationFromProto(res.automation) };
}

async function updateAutomation(automation: Automation): Promise<{ automation: Automation }> {
  const res = await client.updateAutomation({ automation: automationToProto(automation) });
  if (!res.automation) throw new Error("update returned nothing");
  return { automation: automationFromProto(res.automation) };
}

async function deleteAutomation(name: string): Promise<{ automation: Automation | undefined }> {
  const res = await client.deleteAutomation({ name });
  return { automation: res.automation ? automationFromProto(res.automation) : undefined };
}

async function runAutomationNow(params: { name: string; trigger_payload_json?: string }): Promise<{ run: AutomationRun }> {
  const res = await client.runAutomationNow({ name: params.name, triggerPayloadJson: params.trigger_payload_json ?? "" });
  if (!res.run) throw new Error("run returned nothing");
  return { run: runFromProto(res.run) };
}

async function listAutomationRuns(params: { automation_name?: string; page_size?: number; page_token?: string }): Promise<{ runs: AutomationRun[]; next_page_token: string }> {
  const res = await client.listAutomationRuns({
    automationName: params.automation_name ?? "",
    pageSize: params.page_size ?? 0,
    pageToken: params.page_token ?? "",
  });
  return { runs: res.runs.map(runFromProto), next_page_token: res.nextPageToken };
}

async function getAutomationRun(id: string): Promise<{ run: AutomationRun }> {
  const res = await client.getAutomationRun({ id });
  if (!res.run) throw new Error("run not found");
  return { run: runFromProto(res.run) };
}

async function listAutomationStepRuns(runId: string): Promise<{ step_runs: AutomationStepRun[] }> {
  const res = await client.listAutomationStepRuns({ runId });
  return { step_runs: res.stepRuns.map(stepRunFromProto) };
}

export function useAutomations() {
  return useQuery({ queryKey: ["automations"], queryFn: listAutomations });
}

export function useAutomation(name: string) {
  return useQuery({ queryKey: ["automations", name], queryFn: () => getAutomation(name), enabled: !!name });
}

export function useCreateAutomation() {
  const qc = useQueryClient();
  return useMutation({ mutationFn: createAutomation, onSuccess: () => qc.invalidateQueries({ queryKey: ["automations"] }) });
}

export function useUpdateAutomation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateAutomation,
    onSuccess: (_data, automation) => {
      qc.invalidateQueries({ queryKey: ["automations"] });
      qc.invalidateQueries({ queryKey: ["automations", automation.name] });
    },
  });
}

export function useDeleteAutomation() {
  const qc = useQueryClient();
  return useMutation({ mutationFn: deleteAutomation, onSuccess: () => qc.invalidateQueries({ queryKey: ["automations"] }) });
}

export function useRunAutomationNow() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: runAutomationNow,
    onSuccess: (_data, params) => {
      qc.invalidateQueries({ queryKey: ["automation-runs"] });
      qc.invalidateQueries({ queryKey: ["automation-runs", params.name] });
    },
  });
}

export function useAutomationRuns(automationName?: string, pageSize?: number, pageToken?: string) {
  return useQuery({
    queryKey: ["automation-runs", automationName ?? "", pageToken ?? ""],
    queryFn: () => listAutomationRuns({ automation_name: automationName, page_size: pageSize, page_token: pageToken }),
  });
}

export function useAutomationRun(id: string) {
  return useQuery({ queryKey: ["automation-run", id], queryFn: () => getAutomationRun(id), enabled: !!id });
}

export function useAutomationStepRuns(runId: string) {
  return useQuery({
    queryKey: ["automation-step-runs", runId],
    queryFn: () => listAutomationStepRuns(runId),
    enabled: !!runId,
  });
}
