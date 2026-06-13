import { useEffect } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useAgents } from "@/api/agents";
import { useNotifyGroups } from "@/api/notify-groups";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { ScheduleBuilder } from "@/components/schedule-builder";
import type { CronDeliveryType, CronJob } from "@/types/api";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  schedule: z.string().min(1, "Schedule is required"),
  agent_name: z.string().min(1, "Agent is required"),
  input: z.string().optional(),
  timezone: z.string().optional(),
  enabled: z.boolean(),
  delivery_type: z.string(),
  webhook_url: z.string().optional(),
  channel_name: z.string().optional(),
  chat_id: z.string().optional(),
  notify_group_name: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

type CronJobFormProps = {
  mode: "create" | "edit";
  initialValue?: CronJob;
  loading?: boolean;
  submitLabel: string;
  onCancel: () => void;
  onSubmit: (job: CronJob) => void;
};

export default function CronJobForm({
  mode,
  initialValue,
  loading,
  submitLabel,
  onCancel,
  onSubmit,
}: CronJobFormProps) {
  const { data: agentsData } = useAgents();
  const { data: notifyGroupsData } = useNotifyGroups();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      schedule: "",
      agent_name: "",
      input: "",
      timezone: "UTC",
      enabled: true,
      delivery_type: "CRON_DELIVERY_TYPE_LOG",
      webhook_url: "",
      channel_name: "",
      chat_id: "",
      notify_group_name: "",
    },
  });
  const deliveryType = useWatch({ control: form.control, name: "delivery_type" });
  const isEdit = mode === "edit";

  useEffect(() => {
    if (!initialValue) return;
    form.reset({
      name: initialValue.name,
      schedule: initialValue.schedule,
      agent_name: initialValue.agent_name,
      input: initialValue.input ?? "",
      timezone: initialValue.timezone ?? "UTC",
      enabled: initialValue.enabled ?? true,
      delivery_type: initialValue.delivery?.type ?? "CRON_DELIVERY_TYPE_LOG",
      webhook_url: initialValue.delivery?.webhook_url ?? "",
      channel_name: initialValue.delivery?.channel_name ?? "",
      chat_id: initialValue.delivery?.chat_id ?? "",
      notify_group_name: initialValue.delivery?.notify_group_name ?? "",
    });
  }, [form, initialValue]);

  function handleSubmit(values: FormValues) {
    onSubmit({
      name: values.name,
      schedule: values.schedule,
      agent_name: values.agent_name,
      input: values.input,
      timezone: values.timezone,
      enabled: values.enabled,
      delivery: {
        type: values.delivery_type as CronDeliveryType,
        webhook_url: values.webhook_url,
        channel_name: values.channel_name,
        chat_id: values.chat_id,
        notify_group_name: values.notify_group_name,
      },
    });
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Basic Info</CardTitle>
            <CardDescription>Pick the schedule, target agent, input, and timezone for automatic runs.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="name" render={({ field }) => (
              <FormItem><FormLabel>Name</FormLabel><FormControl><Input placeholder="daily-summary" {...field} disabled={isEdit} /></FormControl><FormMessage /></FormItem>
            )} />
            <FormField control={form.control} name="schedule" render={({ field }) => (
              <FormItem>
                <FormLabel>Schedule</FormLabel>
                <FormControl>
                  <ScheduleBuilder value={field.value} onChange={field.onChange} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="agent_name" render={({ field }) => (
              <FormItem>
                <FormLabel>Agent</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl><SelectTrigger><SelectValue placeholder="Select agent" /></SelectTrigger></FormControl>
                  <SelectContent>
                    {(agentsData?.agents ?? []).map((a) => (
                      <SelectItem key={a.name} value={a.name}>{a.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="input" render={({ field }) => (
              <FormItem><FormLabel>Input Message</FormLabel><FormControl><Textarea placeholder="Generate a daily summary" rows={3} {...field} /></FormControl></FormItem>
            )} />
            <FormField control={form.control} name="timezone" render={({ field }) => (
              <FormItem><FormLabel>Timezone</FormLabel><FormControl><Input placeholder="UTC" {...field} /></FormControl></FormItem>
            )} />
            <FormField control={form.control} name="enabled" render={({ field }) => (
              <FormItem className="flex items-center gap-3">
                <FormLabel>Enabled</FormLabel>
                <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
              </FormItem>
            )} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Delivery</CardTitle>
            <CardDescription>Choose whether results stay in logs or are sent to a webhook, channel, or notify group.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="delivery_type" render={({ field }) => (
              <FormItem>
                <FormLabel>Type</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                  <SelectContent>
                    <SelectItem value="CRON_DELIVERY_TYPE_LOG">Log</SelectItem>
                    <SelectItem value="CRON_DELIVERY_TYPE_WEBHOOK">Webhook</SelectItem>
                    <SelectItem value="CRON_DELIVERY_TYPE_CHANNEL">Channel</SelectItem>
                    <SelectItem value="CRON_DELIVERY_TYPE_NOTIFY_GROUP">Notify Group</SelectItem>
                  </SelectContent>
                </Select>
              </FormItem>
            )} />
            {deliveryType === "CRON_DELIVERY_TYPE_WEBHOOK" && (
              <FormField control={form.control} name="webhook_url" render={({ field }) => (
                <FormItem><FormLabel>Webhook URL</FormLabel><FormControl><Input placeholder="https://..." {...field} /></FormControl></FormItem>
              )} />
            )}
            {deliveryType === "CRON_DELIVERY_TYPE_CHANNEL" && (
              <>
                <FormField control={form.control} name="channel_name" render={({ field }) => (
                  <FormItem><FormLabel>Channel Name</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                )} />
                <FormField control={form.control} name="chat_id" render={({ field }) => (
                  <FormItem><FormLabel>Chat ID</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                )} />
              </>
            )}
            {deliveryType === "CRON_DELIVERY_TYPE_NOTIFY_GROUP" && (
              <FormField control={form.control} name="notify_group_name" render={({ field }) => (
                <FormItem>
                  <FormLabel>Notify Group</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl><SelectTrigger><SelectValue placeholder="Select notify group" /></SelectTrigger></FormControl>
                    <SelectContent>
                      {(notifyGroupsData?.notify_groups ?? []).map((group) => (
                        <SelectItem key={group.name} value={group.name}>{group.name}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormItem>
              )} />
            )}
          </CardContent>
        </Card>

        <div className="sticky bottom-0 z-10 -mx-1 flex gap-3 border-t bg-background/95 px-1 py-4 backdrop-blur supports-[backdrop-filter]:bg-background/80">
          <Button type="button" variant="outline" onClick={onCancel}>Cancel</Button>
          <Button type="submit" disabled={loading}>{loading ? "Saving..." : submitLabel}</Button>
        </div>
      </form>
    </Form>
  );
}
