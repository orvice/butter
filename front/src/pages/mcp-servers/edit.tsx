import { useEffect, useMemo, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useMCPServer, useUpdateMCPServer } from "@/api/mcp-servers";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import {
  HeadersEditor,
  entriesToRecord,
  recordToEntries,
  type HeaderEntry,
} from "./headers-editor";
import type { MCPServerTransport } from "@/types/api";

const schema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  transport: z.string(),
  command: z.string().optional(),
  url: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

function isRemoteTransport(t: string): boolean {
  return (
    t === "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP" ||
    t === "MCP_SERVER_TRANSPORT_SSE"
  );
}

export default function MCPServerEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useMCPServer(id ?? "");
  const updateMutation = useUpdateMCPServer();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { id: "", name: "", transport: "MCP_SERVER_TRANSPORT_STDIO", command: "", url: "" },
  });

  const transport = useWatch({ control: form.control, name: "transport" });
  const initialHeaders = useMemo(
    () => recordToEntries(data?.mcp_server?.headers),
    [data],
  );
  const [headersDraft, setHeadersDraft] = useState<HeaderEntry[] | null>(null);
  const headers = headersDraft ?? initialHeaders;

  useEffect(() => {
    if (data?.mcp_server) {
      const s = data.mcp_server;
      form.reset({ id: s.id ?? "", name: s.name, transport: s.transport ?? "MCP_SERVER_TRANSPORT_STDIO", command: s.command ?? "", url: s.url ?? "" });
    }
  }, [data, form]);

  function onSubmit(values: FormValues) {
    const remote = isRemoteTransport(values.transport);
    updateMutation.mutate(
      {
        ...data?.mcp_server,
        id: values.id,
        name: values.name,
        transport: values.transport as MCPServerTransport,
        command: values.command,
        url: values.url,
        headers: remote ? entriesToRecord(headers) : undefined,
      },
      {
        onSuccess: () => { toast.success("MCP server updated"); navigate("/mcp-servers"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/mcp-servers">MCP Servers</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{id}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit MCP Server</h2>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="id" render={({ field }) => (
                <FormItem><FormLabel>ID</FormLabel><FormControl><Input {...field} disabled /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="transport" render={({ field }) => (
                <FormItem>
                  <FormLabel>Transport</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
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
                  <FormItem><FormLabel>Command</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                )} />
              ) : (
                <>
                  <FormField control={form.control} name="url" render={({ field }) => (
                    <FormItem><FormLabel>URL</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                  )} />
                  <div className="space-y-2">
                    <FormLabel>Headers</FormLabel>
                    <HeadersEditor value={headers} onChange={setHeadersDraft} />
                  </div>
                </>
              )}
            </CardContent>
          </Card>
          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/mcp-servers")}>Cancel</Button>
            <Button type="submit" disabled={updateMutation.isPending}>{updateMutation.isPending ? "Saving..." : "Save"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
