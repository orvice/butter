import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useDaemonRuntimes } from "@/api/daemons";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { RemoteAgent, RemoteAgentProtocol } from "@/types/api";

const schema = z.object({
  id: z.string().min(1, "ID is required"),
  name: z.string().min(1, "Name is required"),
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
  if (values.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON") {
    if (!values.daemon_runtime_id) {
      ctx.addIssue({ code: "custom", path: ["daemon_runtime_id"], message: "Daemon runtime is required" });
    }
    if (!values.acp_runtime) {
      ctx.addIssue({ code: "custom", path: ["acp_runtime"], message: "ACP runtime is required" });
    }
  }
});

type FormValues = z.infer<typeof schema>;

type RemoteAgentFormProps = {
  mode: "create" | "edit";
  initialValue?: RemoteAgent;
  initialDaemonRuntimeId?: string;
  initialAcpRuntime?: "opencode" | "codex";
  loading?: boolean;
  submitLabel: string;
  onCancel: () => void;
  onSubmit: (agent: RemoteAgent) => void;
};

export default function RemoteAgentForm({
  mode,
  initialValue,
  initialDaemonRuntimeId = "",
  initialAcpRuntime = "opencode",
  loading,
  submitLabel,
  onCancel,
  onSubmit,
}: RemoteAgentFormProps) {
  const { data: runtimeData } = useDaemonRuntimes();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      id: "",
      name: "",
      protocol: initialDaemonRuntimeId ? "REMOTE_AGENT_PROTOCOL_DAEMON" : "REMOTE_AGENT_PROTOCOL_A2A",
      url: "",
      daemon_runtime_id: initialDaemonRuntimeId,
      acp_runtime: initialAcpRuntime,
    },
  });
  const protocol = form.watch("protocol") as RemoteAgentProtocol;
  const runtimes = runtimeData?.runtimes ?? [];
  const isEdit = mode === "edit";

  useEffect(() => {
    if (!initialValue) return;
    form.reset({
      id: initialValue.id,
      name: initialValue.name,
      protocol: initialValue.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON"
        ? initialValue.protocol
        : "REMOTE_AGENT_PROTOCOL_A2A",
      url: initialValue.url,
      daemon_runtime_id: initialValue.daemon_runtime_id ?? "",
      acp_runtime: initialValue.acp_runtime === "codex" ? "codex" : "opencode",
    });
  }, [form, initialValue]);

  function handleSubmit(values: FormValues) {
    onSubmit({
      id: values.id,
      name: values.name,
      url: values.protocol === "REMOTE_AGENT_PROTOCOL_A2A" ? values.url ?? "" : "",
      protocol: values.protocol,
      daemon_runtime_id: values.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON" ? values.daemon_runtime_id : "",
      acp_runtime: values.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON" ? values.acp_runtime ?? "opencode" : "",
    });
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
        <Card>
          <CardHeader><CardTitle>Remote Agent</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="id" render={({ field }) => (
              <FormItem>
                <FormLabel>ID</FormLabel>
                <FormControl><Input placeholder="remote-assistant" {...field} disabled={isEdit} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="name" render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl><Input placeholder="Remote Assistant" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
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
                <FormItem>
                  <FormLabel>URL</FormLabel>
                  <FormControl><Input placeholder="https://agent.example.com/a2a" {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
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
          <Button type="button" variant="outline" onClick={onCancel}>Cancel</Button>
          <Button type="submit" disabled={loading}>{loading ? "Saving..." : submitLabel}</Button>
        </div>
      </form>
    </Form>
  );
}
