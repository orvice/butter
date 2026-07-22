import { useState, useEffect, useMemo } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm, useWatch } from "react-hook-form";
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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
import { AGENT_TYPE_LABELS, enumLabel } from "@/lib/constants";
import { useTheme } from "next-themes";
import { AgentModelSelect } from "./model-select";
import { AgentIconUpload } from "./icon-upload";
import { AgentFileMountsField } from "./file-mounts-field";
import { AgentRemoteAgentsField } from "./remote-agents-field";
import { agentIconUrl } from "./icon-utils";
import type { Agent, AgentFileMount, AgentFileMountPermission, AgentType } from "@/types/api";

const MOUNT_PERMISSIONS = [
  "AGENT_FILE_MOUNT_PERMISSION_READ",
  "AGENT_FILE_MOUNT_PERMISSION_READ_WRITE",
  "AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE",
] as const satisfies readonly AgentFileMountPermission[];

const agentSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  type: z.string(),
  enable_a2a: z.boolean(),
  enable_openai_api: z.boolean(),
  model: z.string().optional(),
  instruction: z.string().optional(),
  mcp_server_ids: z.array(z.string()).optional(),
  remote_agent_ids: z.array(z.string()).optional(),
  file_mounts: z.array(z.object({
    space_id: z.string().min(1),
    mount_path: z.string().optional(),
    permission: z.enum(MOUNT_PERMISSIONS).optional(),
  })).optional(),
  icon_url: z.string().optional(),
});

type AgentFormValues = z.infer<typeof agentSchema>;

function toAgentFileMountFormValues(mounts?: AgentFileMount[]): AgentFormValues["file_mounts"] {
  return (mounts ?? [])
    .filter((mount) => mount.space_id)
    .map((mount) => ({
      space_id: mount.space_id,
      mount_path: mount.mount_path,
      permission:
        mount.permission === "AGENT_FILE_MOUNT_PERMISSION_READ_WRITE" ||
        mount.permission === "AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE"
          ? mount.permission
          : "AGENT_FILE_MOUNT_PERMISSION_READ",
    }));
}

function mergeAgentIconMetadata(metadata: Agent["metadata"] | undefined, iconUrl: string | undefined) {
  const next = { ...(metadata ?? {}) };
  delete next.avatar_url;
  if (iconUrl) next.icon_url = iconUrl;
  else delete next.icon_url;
  return Object.keys(next).length > 0 ? next : undefined;
}

export default function AgentEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { theme } = useTheme();
  const { data, isLoading } = useAgent(name ?? "");
  const { data: mcpData } = useMCPServers();
  const { data: remoteData, isLoading: isLoadingRemoteAgents } = useRemoteAgents();
  const updateMutation = useUpdateAgent();
  const initialJsonValue = useMemo(() => (data?.agent ? JSON.stringify(data.agent, null, 2) : ""), [data]);
  const [jsonValue, setJsonValue] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState("form");

  const form = useForm<AgentFormValues>({
    resolver: zodResolver(agentSchema),
    defaultValues: {
      name: "",
      description: "",
      type: "AGENT_TYPE_LLM",
      enable_a2a: false,
      enable_openai_api: false,
      model: "",
      instruction: "",
      mcp_server_ids: [],
      remote_agent_ids: [],
      file_mounts: [],
      icon_url: "",
    },
  });
  const agentName = useWatch({ control: form.control, name: "name" });
  const iconUrl = useWatch({ control: form.control, name: "icon_url" });

  useEffect(() => {
    if (data?.agent) {
      const a = data.agent;
      form.reset({
        name: a.name,
        description: a.description ?? "",
        type: a.type ?? "AGENT_TYPE_LLM",
        enable_a2a: a.enable_a2a ?? false,
        enable_openai_api: a.enable_openai_api ?? false,
        model: a.config?.model ?? "",
        instruction: a.config?.instruction ?? "",
        mcp_server_ids: a.config?.mcp_server_ids ?? [],
        remote_agent_ids: a.config?.remote_agent_ids ?? [],
        file_mounts: toAgentFileMountFormValues(a.config?.file_mounts),
        icon_url: agentIconUrl(a),
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
      enable_openai_api: values.enable_openai_api,
      metadata: mergeAgentIconMetadata(data?.agent?.metadata, values.icon_url),
      config: {
        ...data?.agent?.config,
        model: values.model,
        instruction: values.instruction,
        mcp_server_ids: values.mcp_server_ids ?? [],
        remote_agent_ids: values.remote_agent_ids ?? [],
        file_mounts: values.file_mounts ?? [],
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
        enable_openai_api: values.enable_openai_api,
        metadata: mergeAgentIconMetadata(data?.agent?.metadata, values.icon_url),
        config: {
          ...data?.agent?.config,
          model: values.model,
          instruction: values.instruction,
          mcp_server_ids: values.mcp_server_ids ?? [],
          remote_agent_ids: values.remote_agent_ids ?? [],
          file_mounts: values.file_mounts ?? [],
        },
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
          enable_openai_api: agent.enable_openai_api ?? false,
          model: agent.config?.model ?? "",
          instruction: agent.config?.instruction ?? "",
          mcp_server_ids: agent.config?.mcp_server_ids ?? [],
          remote_agent_ids: agent.config?.remote_agent_ids ?? [],
          file_mounts: toAgentFileMountFormValues(agent.config?.file_mounts),
          icon_url: agentIconUrl(agent),
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
      <div className="mb-6 space-y-1">
        <h2 className="text-2xl font-bold">Edit Agent</h2>
        <p className="text-sm text-muted-foreground">Use the guided form for common settings or JSON mode for advanced agent topology.</p>
      </div>

      <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList className="mb-4">
          <TabsTrigger value="form">Form</TabsTrigger>
          <TabsTrigger value="json">JSON</TabsTrigger>
        </TabsList>

        <TabsContent value="form">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onFormSubmit)} className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Basic Info</CardTitle>
                  <CardDescription>Update the visible description and orchestration type.</CardDescription>
                </CardHeader>
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
                  <div className="space-y-3 border-t pt-4">
                    <div>
                      <h3 className="text-sm font-medium">External Access</h3>
                      <p className="text-xs text-muted-foreground">Choose which external protocols may invoke this agent.</p>
                    </div>
                    <FormField control={form.control} name="enable_a2a" render={({ field }) => (
                      <FormItem className="flex items-center justify-between gap-4 rounded-md border px-3 py-3">
                        <div className="space-y-0.5">
                          <FormLabel>A2A</FormLabel>
                          <p className="text-xs text-muted-foreground">Expose this agent through the A2A endpoint.</p>
                        </div>
                        <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                      </FormItem>
                    )} />
                    <FormField control={form.control} name="enable_openai_api" render={({ field }) => (
                      <FormItem className="flex items-center justify-between gap-4 rounded-md border px-3 py-3">
                        <div className="space-y-0.5">
                          <FormLabel>OpenAI API</FormLabel>
                          <p className="text-xs text-muted-foreground">Expose this agent as a model through the OpenAI-compatible API.</p>
                        </div>
                        <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                      </FormItem>
                    )} />
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Icon</CardTitle>
                  <CardDescription>Keep the agent visually identifiable in chat, sessions, and lists.</CardDescription>
                </CardHeader>
                <CardContent>
                  <AgentIconUpload
                    agentName={agentName}
                    value={iconUrl}
                    onChange={(url) => form.setValue("icon_url", url, { shouldDirty: true })}
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Model Configuration</CardTitle>
                  <CardDescription>Adjust the model alias and system instruction used for LLM responses.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <FormField control={form.control} name="model" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Model</FormLabel>
                      <AgentModelSelect value={field.value} onChange={field.onChange} />
                      <p className="text-xs text-muted-foreground">
                        Models are loaded from configured model providers. Agents use the model alias when available.
                      </p>
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

              <Card>
                <CardHeader>
                  <CardTitle>MCP Servers</CardTitle>
                  <CardDescription>Shared tool servers available to this agent. Inline servers remain in JSON mode.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="text-sm text-muted-foreground">
                    Select shared MCP servers this agent can use. Inline MCP servers can still be managed in JSON mode.
                  </p>
                  <FormField control={form.control} name="mcp_server_ids" render={({ field }) => {
                    const selected = field.value ?? [];
                    const servers = mcpData?.mcp_servers ?? [];
                    const toggle = (id: string) => {
                      field.onChange(
                        selected.includes(id)
                          ? selected.filter((selectedId) => selectedId !== id)
                          : [...selected, id],
                      );
                    };

                    if (servers.length === 0) {
                      return <p className="text-sm text-muted-foreground">No shared MCP servers configured yet.</p>;
                    }

                    return (
                      <FormItem>
                        <div className="grid gap-2 md:grid-cols-2">
                          {servers.map((server) => {
                            const id = server.id ?? "";
                            const isSelected = selected.includes(id);
                            return (
                              <button
                                key={id || server.name}
                                type="button"
                                disabled={!id}
                                onClick={() => id && toggle(id)}
                                className={`rounded-md border p-3 text-left transition-colors ${
                                  isSelected ? "border-primary bg-primary/10" : "hover:bg-muted"
                                } ${!id ? "cursor-not-allowed opacity-60" : ""}`}
                              >
                                <div className="flex items-center justify-between gap-2">
                                  <span className="font-medium">{server.name}</span>
                                  {isSelected && <Badge>Selected</Badge>}
                                </div>
                                <div className="mt-1 text-xs text-muted-foreground">
                                  {enumLabel(server.transport, "Transport unspecified")}{server.url ? ` · ${server.url}` : ""}
                                </div>
                              </button>
                            );
                          })}
                        </div>
                      </FormItem>
                    );
                  }} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Remote Agents</CardTitle>
                  <CardDescription>Daemon and A2A agents this agent can delegate work to.</CardDescription>
                </CardHeader>
                <CardContent>
                  <FormField control={form.control} name="remote_agent_ids" render={({ field }) => (
                    <FormItem>
                      <AgentRemoteAgentsField
                        value={field.value}
                        onChange={field.onChange}
                        remoteAgents={remoteData?.remote_agents}
                        isLoading={isLoadingRemoteAgents}
                      />
                    </FormItem>
                  )} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Agent Files</CardTitle>
                  <CardDescription>Review file-space mounts and permissions before saving.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="text-sm text-muted-foreground">
                    Mount workspace file spaces into this agent's built-in agent_files tools.
                  </p>
                  <FormField control={form.control} name="file_mounts" render={({ field }) => (
                    <FormItem>
                      <AgentFileMountsField value={field.value} onChange={field.onChange} />
                    </FormItem>
                  )} />
                </CardContent>
              </Card>

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

              <div className="sticky bottom-0 z-10 -mx-1 flex gap-3 border-t bg-background/95 px-1 py-4 backdrop-blur supports-[backdrop-filter]:bg-background/80">
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
          <div className="sticky bottom-0 z-10 mt-6 -mx-1 flex gap-3 border-t bg-background/95 px-1 py-4 backdrop-blur supports-[backdrop-filter]:bg-background/80">
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
