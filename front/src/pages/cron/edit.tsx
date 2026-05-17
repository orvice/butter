import { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCronJob, useUpdateCronJob } from "@/api/cron";
import { useAgents } from "@/api/agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
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
});

type FormValues = z.infer<typeof schema>;

export default function CronJobEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useCronJob(name ?? "");
  const { data: agentsData } = useAgents();
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
    },
  });
  const deliveryType = form.watch("delivery_type");

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
        delivery: { type: values.delivery_type as CronDeliveryType, webhook_url: values.webhook_url, channel_name: values.channel_name, chat_id: values.chat_id },
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
      <h2 className="mb-6 text-2xl font-bold">Edit Cron Job</h2>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input {...field} disabled /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="schedule" render={({ field }) => (
                <FormItem><FormLabel>Schedule</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
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
            <CardHeader><CardTitle>Delivery</CardTitle></CardHeader>
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
            </CardContent>
          </Card>
          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/cron")}>Cancel</Button>
            <Button type="submit" disabled={updateMutation.isPending}>{updateMutation.isPending ? "Saving..." : "Save"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
