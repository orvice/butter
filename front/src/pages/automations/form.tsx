import { useEffect } from "react";
import { useFieldArray, useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useNavigate } from "react-router-dom";
import { useAgents } from "@/api/agents";
import { useNotifyGroups } from "@/api/notify-groups";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import type { Automation, AutomationConditionOperator, AutomationStepType, AutomationTriggerType } from "@/types/api";
import { Plus, Trash2 } from "lucide-react";

const conditionSchema = z.object({
  selector: z.string().min(1),
  operator: z.string(),
  value: z.string().optional(),
});

const stepSchema = z.object({
  name: z.string().min(1),
  type: z.string(),
  agent_name: z.string().optional(),
  input: z.string().optional(),
  webhook_url: z.string().optional(),
  webhook_method: z.string().optional(),
  webhook_payload_json: z.string().optional(),
  notify_group_name: z.string().optional(),
  notify_title: z.string().optional(),
  notify_message: z.string().optional(),
  forum_thread_id: z.string().optional(),
  forum_body: z.string().optional(),
  timeout_seconds: z.number().optional(),
  retry_attempts: z.number().optional(),
});

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  enabled: z.boolean(),
  trigger_type: z.string(),
  schedule: z.string().optional(),
  timezone: z.string().optional(),
  timeout_seconds: z.number().optional(),
  retry_attempts: z.number().optional(),
  backoff_seconds: z.number().optional(),
  max_output_bytes: z.number().optional(),
  concurrency: z.string(),
  conditions: z.array(conditionSchema),
  steps: z.array(stepSchema).min(1),
});

type FormValues = z.infer<typeof schema>;

const DEFAULT_STEP: FormValues["steps"][number] = {
  name: "invoke-agent",
  type: "AUTOMATION_STEP_TYPE_INVOKE_AGENT",
  agent_name: "",
  input: "",
};

function valuesFromAutomation(automation?: Automation): FormValues {
  const policy = automation?.policy;
  return {
    name: automation?.name ?? "",
    enabled: automation?.enabled ?? true,
    trigger_type: automation?.trigger?.type ?? "AUTOMATION_TRIGGER_TYPE_MANUAL",
    schedule: automation?.trigger?.schedule?.schedule ?? "@daily",
    timezone: automation?.trigger?.schedule?.timezone ?? "UTC",
    timeout_seconds: policy?.timeout_seconds ?? 0,
    retry_attempts: policy?.retry?.max_attempts ?? 0,
    backoff_seconds: policy?.retry?.backoff_seconds ?? 0,
    max_output_bytes: policy?.max_output_bytes ?? 4096,
    concurrency: policy?.concurrency ?? "AUTOMATION_CONCURRENCY_POLICY_ALLOW",
    conditions: automation?.conditions?.map((c) => ({
      selector: c.selector,
      operator: c.operator,
      value: c.value ?? "",
    })) ?? [],
    steps: automation?.steps?.map((s) => ({
      name: s.name,
      type: s.type,
      agent_name: s.invoke_agent?.agent_name ?? "",
      input: s.invoke_agent?.input ?? "",
      webhook_url: s.call_webhook?.url ?? "",
      webhook_method: s.call_webhook?.method || "POST",
      webhook_payload_json: s.call_webhook?.payload_json ?? "",
      notify_group_name: s.send_notify_group?.notify_group_name ?? "",
      notify_title: s.send_notify_group?.title ?? "",
      notify_message: s.send_notify_group?.message ?? "",
      forum_thread_id: s.create_forum_post?.thread_id ?? "",
      forum_body: s.create_forum_post?.body ?? "",
      timeout_seconds: s.policy?.timeout_seconds ?? 0,
      retry_attempts: s.policy?.retry?.max_attempts ?? 0,
    })) ?? [DEFAULT_STEP],
  };
}

function automationFromValues(values: FormValues): Automation {
  return {
    name: values.name,
    enabled: values.enabled,
    trigger: {
      type: values.trigger_type as AutomationTriggerType,
      schedule:
        values.trigger_type === "AUTOMATION_TRIGGER_TYPE_SCHEDULE"
          ? { schedule: values.schedule, timezone: values.timezone || "UTC" }
          : undefined,
    },
    conditions: values.conditions.map((condition) => ({
      selector: condition.selector,
      operator: condition.operator as AutomationConditionOperator,
      value: condition.value,
    })),
    policy: {
      timeout_seconds: values.timeout_seconds || undefined,
      retry: values.retry_attempts ? { max_attempts: values.retry_attempts, backoff_seconds: values.backoff_seconds || undefined } : undefined,
      max_output_bytes: values.max_output_bytes || undefined,
      concurrency: values.concurrency as NonNullable<Automation["policy"]>["concurrency"],
    },
    steps: values.steps.map((step) => {
      const type = step.type as AutomationStepType;
      return {
        name: step.name,
        type,
        invoke_agent:
          type === "AUTOMATION_STEP_TYPE_INVOKE_AGENT"
            ? { agent_name: step.agent_name ?? "", input: step.input }
            : undefined,
        call_webhook:
          type === "AUTOMATION_STEP_TYPE_CALL_WEBHOOK"
            ? { url: step.webhook_url ?? "", method: step.webhook_method || "POST", payload_json: step.webhook_payload_json }
            : undefined,
        send_notify_group:
          type === "AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP"
            ? { notify_group_name: step.notify_group_name ?? "", title: step.notify_title, message: step.notify_message }
            : undefined,
        create_forum_post:
          type === "AUTOMATION_STEP_TYPE_CREATE_FORUM_POST"
            ? { thread_id: step.forum_thread_id ?? "", body: step.forum_body ?? "" }
            : undefined,
        policy:
          step.timeout_seconds || step.retry_attempts
            ? {
                timeout_seconds: step.timeout_seconds || undefined,
                retry: step.retry_attempts ? { max_attempts: step.retry_attempts } : undefined,
              }
            : undefined,
      };
    }),
  };
}

export function AutomationForm({
  automation,
  isSaving,
  onSubmit,
}: {
  automation?: Automation;
  isSaving: boolean;
  onSubmit: (automation: Automation) => void;
}) {
  const navigate = useNavigate();
  const { data: agentsData } = useAgents();
  const { data: notifyGroupsData } = useNotifyGroups();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: valuesFromAutomation(automation),
  });
  const triggerType = useWatch({ control: form.control, name: "trigger_type" });
  const stepValues = useWatch({ control: form.control, name: "steps" });
  const conditions = useFieldArray({ control: form.control, name: "conditions" });
  const steps = useFieldArray({ control: form.control, name: "steps" });

  useEffect(() => {
    if (automation) form.reset(valuesFromAutomation(automation));
  }, [automation, form]);

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit((values) => onSubmit(automationFromValues(values)))} className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Definition</CardTitle>
            <CardDescription>Name, trigger, and execution policy for the workflow.</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-2">
            <FormField control={form.control} name="name" render={({ field }) => (
              <FormItem><FormLabel>Name</FormLabel><FormControl><Input placeholder="daily-summary" {...field} disabled={!!automation} /></FormControl><FormMessage /></FormItem>
            )} />
            <FormField control={form.control} name="trigger_type" render={({ field }) => (
              <FormItem>
                <FormLabel>Trigger</FormLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                  <SelectContent>
                    <SelectItem value="AUTOMATION_TRIGGER_TYPE_MANUAL">Manual</SelectItem>
                    <SelectItem value="AUTOMATION_TRIGGER_TYPE_SCHEDULE">Schedule</SelectItem>
                  </SelectContent>
                </Select>
              </FormItem>
            )} />
            {triggerType === "AUTOMATION_TRIGGER_TYPE_SCHEDULE" ? (
              <>
                <FormField control={form.control} name="schedule" render={({ field }) => (
                  <FormItem><FormLabel>Schedule</FormLabel><FormControl><Input placeholder="0 9 * * *" {...field} /></FormControl><FormMessage /></FormItem>
                )} />
                <FormField control={form.control} name="timezone" render={({ field }) => (
                  <FormItem><FormLabel>Timezone</FormLabel><FormControl><Input placeholder="UTC" {...field} /></FormControl></FormItem>
                )} />
              </>
            ) : null}
            <FormField control={form.control} name="enabled" render={({ field }) => (
              <FormItem className="flex items-center gap-3 self-end">
                <FormLabel>Enabled</FormLabel>
                <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
              </FormItem>
            )} />
            <FormField control={form.control} name="concurrency" render={({ field }) => (
              <FormItem>
                <FormLabel>Concurrency</FormLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                  <SelectContent>
                    <SelectItem value="AUTOMATION_CONCURRENCY_POLICY_ALLOW">Allow</SelectItem>
                    <SelectItem value="AUTOMATION_CONCURRENCY_POLICY_SKIP">Skip</SelectItem>
                    <SelectItem value="AUTOMATION_CONCURRENCY_POLICY_QUEUE">Queue</SelectItem>
                    <SelectItem value="AUTOMATION_CONCURRENCY_POLICY_REPLACE">Replace</SelectItem>
                  </SelectContent>
                </Select>
              </FormItem>
            )} />
            <FormField control={form.control} name="timeout_seconds" render={({ field }) => (
              <FormItem><FormLabel>Timeout Seconds</FormLabel><FormControl><Input type="number" min={0} value={field.value ?? 0} onChange={(event) => field.onChange(event.currentTarget.valueAsNumber || 0)} /></FormControl></FormItem>
            )} />
            <FormField control={form.control} name="retry_attempts" render={({ field }) => (
              <FormItem><FormLabel>Retry Attempts</FormLabel><FormControl><Input type="number" min={0} value={field.value ?? 0} onChange={(event) => field.onChange(event.currentTarget.valueAsNumber || 0)} /></FormControl></FormItem>
            )} />
            <FormField control={form.control} name="max_output_bytes" render={({ field }) => (
              <FormItem><FormLabel>Max Output Bytes</FormLabel><FormControl><Input type="number" min={0} value={field.value ?? 0} onChange={(event) => field.onChange(event.currentTarget.valueAsNumber || 0)} /></FormControl></FormItem>
            )} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-start justify-between">
            <div>
              <CardTitle>Conditions</CardTitle>
              <CardDescription>All conditions must match before steps execute.</CardDescription>
            </div>
            <Button type="button" variant="outline" onClick={() => conditions.append({ selector: "payload.kind", operator: "AUTOMATION_CONDITION_OPERATOR_EQUALS", value: "" })}>
              <Plus className="mr-2 h-4 w-4" />
              Add
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            {conditions.fields.length === 0 ? <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">No conditions.</div> : null}
            {conditions.fields.map((condition, index) => (
              <div key={condition.id} className="grid gap-3 rounded-md border p-3 md:grid-cols-[1fr_220px_1fr_auto]">
                <FormField control={form.control} name={`conditions.${index}.selector`} render={({ field }) => (
                  <FormItem><FormLabel>Selector</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                )} />
                <FormField control={form.control} name={`conditions.${index}.operator`} render={({ field }) => (
                  <FormItem>
                    <FormLabel>Operator</FormLabel>
                    <Select value={field.value} onValueChange={field.onChange}>
                      <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                      <SelectContent>
                        <SelectItem value="AUTOMATION_CONDITION_OPERATOR_EQUALS">Equals</SelectItem>
                        <SelectItem value="AUTOMATION_CONDITION_OPERATOR_NOT_EQUALS">Not equals</SelectItem>
                        <SelectItem value="AUTOMATION_CONDITION_OPERATOR_CONTAINS">Contains</SelectItem>
                        <SelectItem value="AUTOMATION_CONDITION_OPERATOR_REGEX_MATCH">Regex</SelectItem>
                        <SelectItem value="AUTOMATION_CONDITION_OPERATOR_EXISTS">Exists</SelectItem>
                        <SelectItem value="AUTOMATION_CONDITION_OPERATOR_NOT_EXISTS">Not exists</SelectItem>
                      </SelectContent>
                    </Select>
                  </FormItem>
                )} />
                <FormField control={form.control} name={`conditions.${index}.value`} render={({ field }) => (
                  <FormItem><FormLabel>Value</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                )} />
                <Button type="button" variant="ghost" size="icon" className="self-end" onClick={() => conditions.remove(index)} aria-label="Remove condition">
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-start justify-between">
            <div>
              <CardTitle>Steps</CardTitle>
              <CardDescription>Steps execute in order and stop on the first failure.</CardDescription>
            </div>
            <Button type="button" variant="outline" onClick={() => steps.append({ ...DEFAULT_STEP, name: `step-${steps.fields.length + 1}` })}>
              <Plus className="mr-2 h-4 w-4" />
              Add
            </Button>
          </CardHeader>
          <CardContent className="space-y-4">
            {steps.fields.map((step, index) => {
              const type = stepValues?.[index]?.type;
              return (
                <div key={step.id} className="space-y-4 rounded-md border p-4">
                  <div className="grid gap-3 md:grid-cols-[1fr_260px_auto]">
                    <FormField control={form.control} name={`steps.${index}.name`} render={({ field }) => (
                      <FormItem><FormLabel>Name</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                    )} />
                    <FormField control={form.control} name={`steps.${index}.type`} render={({ field }) => (
                      <FormItem>
                        <FormLabel>Action</FormLabel>
                        <Select value={field.value} onValueChange={field.onChange}>
                          <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                          <SelectContent>
                            <SelectItem value="AUTOMATION_STEP_TYPE_INVOKE_AGENT">Invoke Agent</SelectItem>
                            <SelectItem value="AUTOMATION_STEP_TYPE_CALL_WEBHOOK">Call Webhook</SelectItem>
                            <SelectItem value="AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP">Notify Group</SelectItem>
                            <SelectItem value="AUTOMATION_STEP_TYPE_CREATE_FORUM_POST">Forum Post</SelectItem>
                          </SelectContent>
                        </Select>
                      </FormItem>
                    )} />
                    <Button type="button" variant="ghost" size="icon" className="self-end" onClick={() => steps.remove(index)} aria-label="Remove step">
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                  {type === "AUTOMATION_STEP_TYPE_INVOKE_AGENT" ? (
                    <div className="grid gap-3 md:grid-cols-2">
                      <FormField control={form.control} name={`steps.${index}.agent_name`} render={({ field }) => (
                        <FormItem>
                          <FormLabel>Agent</FormLabel>
                          <Select value={field.value} onValueChange={field.onChange}>
                            <FormControl><SelectTrigger><SelectValue placeholder="Select agent" /></SelectTrigger></FormControl>
                            <SelectContent>
                              {(agentsData?.agents ?? []).map((agent) => <SelectItem key={agent.name} value={agent.name}>{agent.name}</SelectItem>)}
                            </SelectContent>
                          </Select>
                        </FormItem>
                      )} />
                      <FormField control={form.control} name={`steps.${index}.input`} render={({ field }) => (
                        <FormItem><FormLabel>Input</FormLabel><FormControl><Textarea rows={3} {...field} /></FormControl></FormItem>
                      )} />
                    </div>
                  ) : null}
                  {type === "AUTOMATION_STEP_TYPE_CALL_WEBHOOK" ? (
                    <div className="grid gap-3 md:grid-cols-2">
                      <FormField control={form.control} name={`steps.${index}.webhook_url`} render={({ field }) => (
                        <FormItem><FormLabel>URL</FormLabel><FormControl><Input placeholder="https://..." {...field} /></FormControl></FormItem>
                      )} />
                      <FormField control={form.control} name={`steps.${index}.webhook_method`} render={({ field }) => (
                        <FormItem><FormLabel>Method</FormLabel><FormControl><Input placeholder="POST" {...field} /></FormControl></FormItem>
                      )} />
                      <FormField control={form.control} name={`steps.${index}.webhook_payload_json`} render={({ field }) => (
                        <FormItem className="md:col-span-2"><FormLabel>Payload JSON</FormLabel><FormControl><Textarea rows={4} {...field} /></FormControl></FormItem>
                      )} />
                    </div>
                  ) : null}
                  {type === "AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP" ? (
                    <div className="grid gap-3 md:grid-cols-3">
                      <FormField control={form.control} name={`steps.${index}.notify_group_name`} render={({ field }) => (
                        <FormItem>
                          <FormLabel>Notify Group</FormLabel>
                          <Select value={field.value} onValueChange={field.onChange}>
                            <FormControl><SelectTrigger><SelectValue placeholder="Select group" /></SelectTrigger></FormControl>
                            <SelectContent>
                              {(notifyGroupsData?.notify_groups ?? []).map((group) => <SelectItem key={group.name} value={group.name}>{group.name}</SelectItem>)}
                            </SelectContent>
                          </Select>
                        </FormItem>
                      )} />
                      <FormField control={form.control} name={`steps.${index}.notify_title`} render={({ field }) => (
                        <FormItem><FormLabel>Title</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                      )} />
                      <FormField control={form.control} name={`steps.${index}.notify_message`} render={({ field }) => (
                        <FormItem><FormLabel>Message</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                      )} />
                    </div>
                  ) : null}
                  {type === "AUTOMATION_STEP_TYPE_CREATE_FORUM_POST" ? (
                    <div className="grid gap-3 md:grid-cols-2">
                      <FormField control={form.control} name={`steps.${index}.forum_thread_id`} render={({ field }) => (
                        <FormItem><FormLabel>Thread ID</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                      )} />
                      <FormField control={form.control} name={`steps.${index}.forum_body`} render={({ field }) => (
                        <FormItem><FormLabel>Body</FormLabel><FormControl><Textarea rows={3} {...field} /></FormControl></FormItem>
                      )} />
                    </div>
                  ) : null}
                </div>
              );
            })}
          </CardContent>
        </Card>

        <div className="sticky bottom-0 z-10 -mx-1 flex gap-3 border-t bg-background/95 px-1 py-4 backdrop-blur supports-[backdrop-filter]:bg-background/80">
          <Button type="button" variant="outline" onClick={() => navigate("/automations")}>Cancel</Button>
          <Button type="submit" disabled={isSaving}>{isSaving ? "Saving..." : "Save"}</Button>
        </div>
      </form>
    </Form>
  );
}
