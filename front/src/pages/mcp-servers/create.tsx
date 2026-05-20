import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCreateMCPServer } from "@/api/mcp-servers";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import {
  HeadersEditor,
  entriesToRecord,
  type HeaderEntry,
} from "./headers-editor";
import type { MCPServerTransport } from "@/types/api";

const schema = z.object({
  id: z.string().min(1, "ID is required"),
  name: z.string().min(1, "Name is required"),
  transport: z.string(),
  command: z.string().optional(),
  url: z.string().optional(),
  timeout_seconds: z.string().regex(/^\d*$/, "Timeout must be a non-negative integer").optional(),
});

type FormValues = z.infer<typeof schema>;

function isRemoteTransport(t: string): boolean {
  return (
    t === "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP" ||
    t === "MCP_SERVER_TRANSPORT_SSE"
  );
}

export default function MCPServerCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateMCPServer();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { id: "", name: "", transport: "MCP_SERVER_TRANSPORT_STDIO", command: "", url: "", timeout_seconds: "" },
  });

  const transport = useWatch({ control: form.control, name: "transport" });
  const [headers, setHeaders] = useState<HeaderEntry[]>([]);

  function onSubmit(values: FormValues) {
    const remote = isRemoteTransport(values.transport);
    const timeoutSeconds = Number(values.timeout_seconds || 0);
    createMutation.mutate(
      {
        id: values.id,
        name: values.name,
        transport: values.transport as MCPServerTransport,
        command: values.command,
        url: values.url,
        headers: remote ? entriesToRecord(headers) : undefined,
        timeout_seconds: timeoutSeconds > 0 ? timeoutSeconds : undefined,
      },
      {
        onSuccess: () => { toast.success("MCP server created"); navigate("/mcp-servers"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/mcp-servers">MCP Servers</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Create MCP Server</h2>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="id" render={({ field }) => (
                <FormItem><FormLabel>ID</FormLabel><FormControl><Input placeholder="my-mcp-server" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input placeholder="My MCP Server" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="transport" render={({ field }) => (
                <FormItem>
                  <FormLabel>Transport</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      <SelectItem value="MCP_SERVER_TRANSPORT_STDIO">Stdio</SelectItem>
                      <SelectItem value="MCP_SERVER_TRANSPORT_STREAMABLE_HTTP">HTTP</SelectItem>
                      <SelectItem value="MCP_SERVER_TRANSPORT_SSE">SSE</SelectItem>
                    </SelectContent>
                  </Select>
                </FormItem>
              )} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle>Connection</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              {transport === "MCP_SERVER_TRANSPORT_STDIO" ? (
                <FormField control={form.control} name="command" render={({ field }) => (
                  <FormItem><FormLabel>Command</FormLabel><FormControl><Input placeholder="npx @modelcontextprotocol/server" {...field} /></FormControl></FormItem>
                )} />
              ) : (
                <>
                  <FormField control={form.control} name="url" render={({ field }) => (
                    <FormItem><FormLabel>URL</FormLabel><FormControl><Input placeholder="https://..." {...field} /></FormControl></FormItem>
                  )} />
                  <div className="space-y-2">
                    <FormLabel>Headers</FormLabel>
                    <HeadersEditor value={headers} onChange={setHeaders} />
                  </div>
                  <FormField control={form.control} name="timeout_seconds" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Timeout Seconds</FormLabel>
                      <FormControl><Input type="number" min={0} placeholder="5" {...field} /></FormControl>
                      <FormMessage />
                    </FormItem>
                  )} />
                </>
              )}
            </CardContent>
          </Card>

          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/mcp-servers")}>Cancel</Button>
            <Button type="submit" disabled={createMutation.isPending}>{createMutation.isPending ? "Creating..." : "Create"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
