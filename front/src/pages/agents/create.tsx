import { useNavigate } from "react-router-dom";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCreateAgent } from "@/api/agents";
import { useMCPServers } from "@/api/mcp-servers";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { AgentModelSelect } from "./model-select";
import { AgentIconUpload } from "./icon-upload";
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
import type { AgentType } from "@/types/api";

const agentSchema = z.object({
  name: z.string().min(1, "Name is required").refine((v) => v !== "user", "Name cannot be 'user'"),
  description: z.string().optional(),
  type: z.string(),
  enable_a2a: z.boolean(),
  model: z.string().optional(),
  instruction: z.string().optional(),
  mcp_server_ids: z.array(z.string()).optional(),
  icon_url: z.string().optional(),
});

type AgentFormValues = z.infer<typeof agentSchema>;

export default function AgentCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateAgent();
  const { data: mcpData, isLoading: isLoadingMCPServers } = useMCPServers();

  const form = useForm<AgentFormValues>({
    resolver: zodResolver(agentSchema),
    defaultValues: { name: "", description: "", type: "AGENT_TYPE_LLM", enable_a2a: false, model: "", instruction: "", mcp_server_ids: [], icon_url: "" },
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
        metadata: values.icon_url ? { icon_url: values.icon_url } : undefined,
        config: {
          model: values.model,
          instruction: values.instruction,
          mcp_server_ids: values.mcp_server_ids ?? [],
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
      <h2 className="mb-6 text-2xl font-bold">Create Agent</h2>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
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
              <FormField control={form.control} name="enable_a2a" render={({ field }) => (
                <FormItem className="flex items-center gap-3">
                  <FormLabel>Enable A2A</FormLabel>
                  <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                </FormItem>
              )} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle>Icon</CardTitle></CardHeader>
            <CardContent>
              <AgentIconUpload
                agentName={agentName}
                value={iconUrl}
                onChange={(url) => form.setValue("icon_url", url, { shouldDirty: true })}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle>Model Configuration</CardTitle></CardHeader>
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
            <CardHeader><CardTitle>MCP Servers</CardTitle></CardHeader>
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
                              {server.transport ?? "transport unspecified"}{server.url ? ` · ${server.url}` : ""}
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

          <div className="flex gap-3">
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
