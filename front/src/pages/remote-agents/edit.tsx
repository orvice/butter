import { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useDaemonRuntimes } from "@/api/daemons";
import { useRemoteAgent, useUpdateRemoteAgent } from "@/api/remote-agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { RemoteAgentProtocol } from "@/types/api";

const schema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  protocol: z.enum([
    "REMOTE_AGENT_PROTOCOL_A2A",
    "REMOTE_AGENT_PROTOCOL_DAEMON",
  ]),
  url: z.string().optional(),
  daemon_runtime_id: z.string().optional(),
  acp_runtime: z.enum(["opencode", "codex"]).optional(),
}).superRefine((values, ctx) => {
  if (values.protocol === "REMOTE_AGENT_PROTOCOL_A2A") {
    if (!values.url) {
      ctx.addIssue({ code: "custom", path: ["url"], message: "URL is required" });
      return;
    }
    if (!z.string().url().safeParse(values.url).success) {
      ctx.addIssue({ code: "custom", path: ["url"], message: "Must be a valid URL" });
    }
  }
  if (values.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON" && !values.daemon_runtime_id) {
    ctx.addIssue({ code: "custom", path: ["daemon_runtime_id"], message: "Daemon runtime is required" });
  }
});

type FormValues = z.infer<typeof schema>;

export default function RemoteAgentEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useRemoteAgent(id ?? "");
  const { data: runtimeData } = useDaemonRuntimes();
  const updateMutation = useUpdateRemoteAgent();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      id: "",
      name: "",
      protocol: "REMOTE_AGENT_PROTOCOL_A2A",
      url: "",
      daemon_runtime_id: "",
      acp_runtime: "opencode",
    },
  });
  const protocol = form.watch("protocol") as RemoteAgentProtocol;
  const runtimes = runtimeData?.runtimes ?? [];

  useEffect(() => {
    if (data?.remote_agent) {
      const ra = data.remote_agent;
      form.reset({
        id: ra.id,
        name: ra.name,
        protocol: ra.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON" ? ra.protocol : "REMOTE_AGENT_PROTOCOL_A2A",
        url: ra.url,
        daemon_runtime_id: ra.daemon_runtime_id ?? "",
        acp_runtime: ra.acp_runtime === "codex" ? "codex" : "opencode",
      });
    }
  }, [data, form]);

  function onSubmit(values: FormValues) {
    updateMutation.mutate(
      {
        id: values.id,
        name: values.name,
        url: values.protocol === "REMOTE_AGENT_PROTOCOL_A2A" ? values.url ?? "" : "",
        protocol: values.protocol,
        daemon_runtime_id: values.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON" ? values.daemon_runtime_id : "",
        acp_runtime: values.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON" ? values.acp_runtime ?? "opencode" : "",
      },
      {
        onSuccess: () => { toast.success("Remote agent updated"); navigate("/remote-agents"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/remote-agents">Remote Agents</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{id}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit Remote Agent</h2>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Remote Agent</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="id" render={({ field }) => (
                <FormItem><FormLabel>ID</FormLabel><FormControl><Input {...field} disabled /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="protocol" render={({ field }) => (
                <FormItem>
                  <FormLabel>Protocol</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl><SelectTrigger className="w-full"><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      <SelectItem value="REMOTE_AGENT_PROTOCOL_A2A">A2A</SelectItem>
                      <SelectItem value="REMOTE_AGENT_PROTOCOL_DAEMON">Daemon Runtime</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )} />
              {protocol === "REMOTE_AGENT_PROTOCOL_A2A" ? (
                <FormField control={form.control} name="url" render={({ field }) => (
                  <FormItem><FormLabel>URL</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                )} />
              ) : (
                <>
                  <FormField control={form.control} name="daemon_runtime_id" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Daemon Runtime</FormLabel>
                      <Select onValueChange={field.onChange} value={field.value}>
                        <FormControl><SelectTrigger className="w-full"><SelectValue placeholder="Select runtime" /></SelectTrigger></FormControl>
                        <SelectContent>
                          {runtimes.map((runtime) => (
                            <SelectItem key={runtime.id} value={runtime.id}>
                              {runtime.name || runtime.id}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )} />
                  <FormField control={form.control} name="acp_runtime" render={({ field }) => (
                    <FormItem>
                      <FormLabel>ACP Runtime</FormLabel>
                      <Select onValueChange={field.onChange} value={field.value}>
                        <FormControl><SelectTrigger className="w-full"><SelectValue /></SelectTrigger></FormControl>
                        <SelectContent>
                          <SelectItem value="opencode">opencode</SelectItem>
                          <SelectItem value="codex">codex</SelectItem>
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )} />
                </>
              )}
            </CardContent>
          </Card>
          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/remote-agents")}>Cancel</Button>
            <Button type="submit" disabled={updateMutation.isPending}>{updateMutation.isPending ? "Saving..." : "Save"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
