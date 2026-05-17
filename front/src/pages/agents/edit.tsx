import { useState, useEffect, useMemo } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import Editor from "@monaco-editor/react";
import { useAgent, useUpdateAgent } from "@/api/agents";
import { useMCPServers } from "@/api/mcp-servers";
import { useRemoteAgents } from "@/api/remote-agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
} from "@/components/ui/form";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { AGENT_TYPE_LABELS } from "@/lib/constants";
import { useTheme } from "next-themes";
import type { Agent, AgentType } from "@/types/api";

const agentSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  type: z.string(),
  enable_a2a: z.boolean(),
  model: z.string().optional(),
  instruction: z.string().optional(),
});

type AgentFormValues = z.infer<typeof agentSchema>;

export default function AgentEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { theme } = useTheme();
  const { data, isLoading } = useAgent(name ?? "");
  const { data: mcpData } = useMCPServers();
  const { data: remoteData } = useRemoteAgents();
  const updateMutation = useUpdateAgent();
  const initialJsonValue = useMemo(() => (data?.agent ? JSON.stringify(data.agent, null, 2) : ""), [data]);
  const [jsonValue, setJsonValue] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState("form");

  const form = useForm<AgentFormValues>({
    resolver: zodResolver(agentSchema),
    defaultValues: { name: "", description: "", type: "AGENT_TYPE_LLM", enable_a2a: false, model: "", instruction: "" },
  });

  useEffect(() => {
    if (data?.agent) {
      const a = data.agent;
      form.reset({
        name: a.name,
        description: a.description ?? "",
        type: a.type ?? "AGENT_TYPE_LLM",
        enable_a2a: a.enable_a2a ?? false,
        model: a.config?.model ?? "",
        instruction: a.config?.instruction ?? "",
      });
    }
  }, [data, form]);

  function onFormSubmit(values: AgentFormValues) {
    const agent: Agent = {
      ...data?.agent,
      name: values.name,
      description: values.description,
      type: values.type as AgentType,
      enable_a2a: values.enable_a2a,
      config: {
        ...data?.agent?.config,
        model: values.model,
        instruction: values.instruction,
      },
    };
    submitUpdate(agent);
  }

  function onJsonSubmit() {
    try {
      const agent = JSON.parse(jsonValue ?? initialJsonValue) as Agent;
      submitUpdate(agent);
    } catch {
      toast.error("Invalid JSON");
    }
  }

  function submitUpdate(agent: Agent) {
    updateMutation.mutate(agent, {
      onSuccess: () => { toast.success("Agent updated"); navigate("/agents"); },
      onError: (err) => toast.error(err.message),
    });
  }

  function handleTabChange(tab: string) {
    if (tab === "json") {
      const values = form.getValues();
      const agent: Agent = {
        ...data?.agent,
        name: values.name,
        description: values.description,
        type: values.type as AgentType,
        enable_a2a: values.enable_a2a,
        config: { ...data?.agent?.config, model: values.model, instruction: values.instruction },
      };
      setJsonValue(JSON.stringify(agent, null, 2));
    } else if (tab === "form") {
      try {
        const agent = JSON.parse(jsonValue ?? initialJsonValue) as Agent;
        form.reset({
          name: agent.name,
          description: agent.description ?? "",
          type: agent.type ?? "AGENT_TYPE_LLM",
          enable_a2a: agent.enable_a2a ?? false,
          model: agent.config?.model ?? "",
          instruction: agent.config?.instruction ?? "",
        });
      } catch { /* keep current form values if JSON is invalid */ }
    }
    setActiveTab(tab);
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/agents">Agents</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit Agent</h2>

      <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList className="mb-4">
          <TabsTrigger value="form">Form</TabsTrigger>
          <TabsTrigger value="json">JSON</TabsTrigger>
        </TabsList>

        <TabsContent value="form">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onFormSubmit)} className="space-y-6">
              <Card>
                <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
                <CardContent className="space-y-4">
                  <FormField control={form.control} name="name" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Name</FormLabel>
                      <FormControl><Input {...field} disabled /></FormControl>
                    </FormItem>
                  )} />
                  <FormField control={form.control} name="description" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Description</FormLabel>
                      <FormControl><Input {...field} /></FormControl>
                    </FormItem>
                  )} />
                  <FormField control={form.control} name="type" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Type</FormLabel>
                      <Select onValueChange={field.onChange} value={field.value}>
                        <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                        <SelectContent>
                          <SelectItem value="AGENT_TYPE_LLM">LLM</SelectItem>
                          <SelectItem value="AGENT_TYPE_LOOP">Loop</SelectItem>
                          <SelectItem value="AGENT_TYPE_SEQUENTIAL">Sequential</SelectItem>
                          <SelectItem value="AGENT_TYPE_PARALLEL">Parallel</SelectItem>
                        </SelectContent>
                      </Select>
                    </FormItem>
                  )} />
                  <FormField control={form.control} name="enable_a2a" render={({ field }) => (
                    <FormItem className="flex items-center gap-3">
                      <FormLabel>Enable A2A</FormLabel>
                      <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                    </FormItem>
                  )} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader><CardTitle>Model Configuration</CardTitle></CardHeader>
                <CardContent className="space-y-4">
                  <FormField control={form.control} name="model" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Model</FormLabel>
                      <FormControl><Input {...field} /></FormControl>
                    </FormItem>
                  )} />
                  <FormField control={form.control} name="instruction" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Instruction</FormLabel>
                      <FormControl><Textarea rows={5} {...field} /></FormControl>
                    </FormItem>
                  )} />
                </CardContent>
              </Card>

              {/* MCP Server IDs - read-only display */}
              {data?.agent?.config?.mcp_server_ids && data.agent.config.mcp_server_ids.length > 0 && (
                <Card>
                  <CardHeader><CardTitle>MCP Servers</CardTitle></CardHeader>
                  <CardContent className="flex flex-wrap gap-2">
                    {data.agent.config.mcp_server_ids.map((id) => {
                      const server = mcpData?.mcp_servers?.find((s) => s.id === id);
                      return <Badge key={id} variant="secondary">{server?.name ?? id}</Badge>;
                    })}
                  </CardContent>
                </Card>
              )}

              {/* Remote Agent IDs - read-only display */}
              {data?.agent?.config?.remote_agent_ids && data.agent.config.remote_agent_ids.length > 0 && (
                <Card>
                  <CardHeader><CardTitle>Remote Agents</CardTitle></CardHeader>
                  <CardContent className="flex flex-wrap gap-2">
                    {data.agent.config.remote_agent_ids.map((id) => {
                      const ra = remoteData?.remote_agents?.find((r) => r.id === id);
                      return <Badge key={id} variant="secondary">{ra?.name ?? id}</Badge>;
                    })}
                  </CardContent>
                </Card>
              )}

              {/* Sub-agents - read-only list */}
              {data?.agent?.sub_agents && data.agent.sub_agents.length > 0 && (
                <Card>
                  <CardHeader><CardTitle>Sub-Agents (read-only, edit in JSON mode)</CardTitle></CardHeader>
                  <CardContent className="space-y-2">
                    {data.agent.sub_agents.map((sa) => (
                      <div key={sa.name} className="flex items-center gap-2">
                        <span className="text-sm font-medium">{sa.name}</span>
                        <Badge variant="outline">{AGENT_TYPE_LABELS[sa.type ?? "AGENT_TYPE_UNSPECIFIED"]}</Badge>
                      </div>
                    ))}
                  </CardContent>
                </Card>
              )}

              <div className="flex gap-3">
                <Button type="button" variant="outline" onClick={() => navigate("/agents")}>Cancel</Button>
                <Button type="submit" disabled={updateMutation.isPending}>
                  {updateMutation.isPending ? "Saving..." : "Save"}
                </Button>
              </div>
            </form>
          </Form>
        </TabsContent>

        <TabsContent value="json">
          <Card>
            <CardContent className="pt-6">
              <Editor
                height="500px"
                language="json"
                theme={theme === "dark" ? "vs-dark" : "light"}
                value={jsonValue ?? initialJsonValue}
                onChange={(v) => setJsonValue(v ?? "")}
                options={{ minimap: { enabled: false }, formatOnPaste: true }}
              />
            </CardContent>
          </Card>
          <div className="mt-6 flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/agents")}>Cancel</Button>
            <Button onClick={onJsonSubmit} disabled={updateMutation.isPending}>
              {updateMutation.isPending ? "Saving..." : "Save"}
            </Button>
          </div>
        </TabsContent>
      </Tabs>
    </>
  );
}
