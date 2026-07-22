import { useNavigate, useSearchParams } from "react-router-dom";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCreateAgent } from "@/api/agents";
import { useMCPServers } from "@/api/mcp-servers";
import { useRemoteAgents } from "@/api/remote-agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { AgentModelSelect } from "./model-select";
import { AgentIconUpload } from "./icon-upload";
import { AgentFileMountsField } from "./file-mounts-field";
import { AgentRemoteAgentsField } from "./remote-agents-field";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { enumLabel } from "@/lib/constants";
import type { AgentFileMountPermission, AgentType } from "@/types/api";

const MOUNT_PERMISSIONS = [
  "AGENT_FILE_MOUNT_PERMISSION_READ",
  "AGENT_FILE_MOUNT_PERMISSION_READ_WRITE",
  "AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE",
] as const satisfies readonly AgentFileMountPermission[];

const agentSchema = z.object({
  name: z.string().min(1, "Name is required").refine((v) => v !== "user", "Name cannot be 'user'"),
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

export default function AgentCreatePage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const initialRemoteAgentId = searchParams.get("remote_agent_id") ?? "";
  const createMutation = useCreateAgent();
  const { data: mcpData, isLoading: isLoadingMCPServers } = useMCPServers();
  const { data: remoteData, isLoading: isLoadingRemoteAgents } = useRemoteAgents();

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
      remote_agent_ids: initialRemoteAgentId ? [initialRemoteAgentId] : [],
      file_mounts: [],
      icon_url: "",
    },
  });
  const agentName = useWatch({ control: form.control, name: "name" });
  const iconUrl = useWatch({ control: form.control, name: "icon_url" });

  function onSubmit(values: AgentFormValues) {
    createMutation.mutate(
      {
        name: values.name,
        description: values.description,
        type: values.type as AgentType,
        enable_a2a: values.enable_a2a,
        enable_openai_api: values.enable_openai_api,
        metadata: values.icon_url ? { icon_url: values.icon_url } : undefined,
        config: {
          model: values.model,
          instruction: values.instruction,
          mcp_server_ids: values.mcp_server_ids ?? [],
          remote_agent_ids: values.remote_agent_ids ?? [],
          file_mounts: values.file_mounts ?? [],
        },
      },
      {
        onSuccess: () => { toast.success("Agent created"); navigate("/agents"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/agents">Agents</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <div className="mb-6 space-y-1">
        <h2 className="text-2xl font-bold">Create Agent</h2>
        <p className="text-sm text-muted-foreground">Start with identity and model settings, then optionally connect tools and file spaces.</p>
      </div>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Basic Info</CardTitle>
              <CardDescription>Name and describe how this agent appears across the workspace.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl><Input placeholder="my-agent" {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="description" render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl><Input placeholder="A helpful assistant" {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="type" render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      <SelectItem value="AGENT_TYPE_LLM">LLM</SelectItem>
                      <SelectItem value="AGENT_TYPE_LOOP">Loop</SelectItem>
                      <SelectItem value="AGENT_TYPE_SEQUENTIAL">Sequential</SelectItem>
                      <SelectItem value="AGENT_TYPE_PARALLEL">Parallel</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
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
              <CardDescription>Upload or link an avatar so the agent is easier to recognize in lists and chat.</CardDescription>
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
              <CardDescription>Pick the model and instruction the agent will use for LLM responses.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="model" render={({ field }) => (
                <FormItem>
                  <FormLabel>Model</FormLabel>
                  <AgentModelSelect value={field.value} onChange={field.onChange} />
                  <p className="text-xs text-muted-foreground">
                    Models are loaded from configured model providers. Agents use the model alias when available.
                  </p>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="instruction" render={({ field }) => (
                <FormItem>
                  <FormLabel>Instruction</FormLabel>
                  <FormControl><Textarea placeholder="You are a helpful assistant..." rows={5} {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>MCP Servers</CardTitle>
              <CardDescription>Optional tools the agent can call during runs.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <p className="text-sm text-muted-foreground">
                Select shared MCP servers this agent can use. Leave empty to disable shared MCP tools.
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

                if (isLoadingMCPServers) {
                  return <p className="text-sm text-muted-foreground">Loading MCP servers...</p>;
                }

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
                  <FormMessage />
                </FormItem>
              )} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Agent Files</CardTitle>
              <CardDescription>Mount file spaces only when the agent needs persistent workspace files.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <p className="text-sm text-muted-foreground">
                Mount workspace file spaces into this agent's built-in agent_files tools.
              </p>
              <FormField control={form.control} name="file_mounts" render={({ field }) => (
                <FormItem>
                  <AgentFileMountsField value={field.value} onChange={field.onChange} />
                  <FormMessage />
                </FormItem>
              )} />
            </CardContent>
          </Card>

          <div className="sticky bottom-0 z-10 -mx-1 flex gap-3 border-t bg-background/95 px-1 py-4 backdrop-blur supports-[backdrop-filter]:bg-background/80">
            <Button type="button" variant="outline" onClick={() => navigate("/agents")}>Cancel</Button>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? "Creating..." : "Create Agent"}
            </Button>
          </div>
        </form>
      </Form>
    </>
  );
}
