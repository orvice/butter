import { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCronJob, useUpdateCronJob } from "@/api/cron";
import { useAgents } from "@/api/agents";
import { useNotifyGroups } from "@/api/notify-groups";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import { ScheduleBuilder } from "@/components/schedule-builder";
import type { CronDeliveryType } from "@/types/api";

const schema = z.object({
  name: z.string().min(1),
  schedule: z.string().min(1),
  agent_name: z.string().min(1),
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

export default function CronJobEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useCronJob(name ?? "");
  const { data: agentsData } = useAgents();
  const { data: notifyGroupsData } = useNotifyGroups();
  const updateMutation = useUpdateCronJob();
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

  useEffect(() => {
    if (data?.cron_job) {
      const j = data.cron_job;
      form.reset({
        name: j.name,
        schedule: j.schedule,
        agent_name: j.agent_name,
        input: j.input ?? "",
        timezone: j.timezone ?? "UTC",
        enabled: j.enabled ?? true,
        delivery_type: j.delivery?.type ?? "CRON_DELIVERY_TYPE_LOG",
        webhook_url: j.delivery?.webhook_url ?? "",
        channel_name: j.delivery?.channel_name ?? "",
        chat_id: j.delivery?.chat_id ?? "",
        notify_group_name: j.delivery?.notify_group_name ?? "",
      });
    }
  }, [data, form]);

  function onSubmit(values: FormValues) {
    updateMutation.mutate(
      {
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
      },
      {
        onSuccess: () => { toast.success("Cron job updated"); navigate("/cron"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/cron">Cron Jobs</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <div className="mb-6 space-y-1">
        <h2 className="text-2xl font-bold">Edit Cron Job</h2>
        <p className="text-sm text-muted-foreground">Review schedule, agent input, and delivery settings before the next automatic run.</p>
      </div>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Basic Info</CardTitle>
              <CardDescription>Update when the job runs, which agent it invokes, and the message it sends.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input {...field} disabled /></FormControl></FormItem>
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
                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      {(agentsData?.agents ?? []).map((a) => (
                        <SelectItem key={a.name} value={a.name}>{a.name}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormItem>
              )} />
              <FormField control={form.control} name="input" render={({ field }) => (
                <FormItem><FormLabel>Input Message</FormLabel><FormControl><Textarea rows={3} {...field} /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="timezone" render={({ field }) => (
                <FormItem><FormLabel>Timezone</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
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
              <CardDescription>Choose where execution results are written or sent after each run.</CardDescription>
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
                  <FormItem><FormLabel>Webhook URL</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
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
            <Button type="button" variant="outline" onClick={() => navigate("/cron")}>Cancel</Button>
            <Button type="submit" disabled={updateMutation.isPending}>{updateMutation.isPending ? "Saving..." : "Save"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
