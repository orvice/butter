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
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import {
  HeadersEditor,
  entriesToRecord,
  recordToEntries,
  type HeaderEntry,
} from "./headers-editor";
import type { MCPServer, MCPServerAuthType, MCPServerTransport } from "@/types/api";

const schema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  transport: z.string(),
  command: z.string().optional(),
  url: z.string().optional(),
  timeout_seconds: z.string().regex(/^\d*$/, "Timeout must be a non-negative integer").optional(),
  auth_type: z.string().optional(),
  oauth_registration_mode: z.string().optional(),
  oauth_client_id: z.string().optional(),
  oauth_client_secret: z.string().optional(),
  oauth_scopes: z.string().optional(),
  oauth_authorization_url: z.string().optional(),
  oauth_token_url: z.string().optional(),
  oauth_resource_metadata_url: z.string().optional(),
  oauth_authorization_server_url: z.string().optional(),
  oauth_resource: z.string().optional(),
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
    defaultValues: {
      id: "",
      name: "",
      transport: "MCP_SERVER_TRANSPORT_STDIO",
      command: "",
      url: "",
      timeout_seconds: "",
      auth_type: "MCP_SERVER_AUTH_TYPE_NONE",
      oauth_registration_mode: "preconfigured",
      oauth_client_id: "",
      oauth_client_secret: "",
      oauth_scopes: "",
      oauth_authorization_url: "",
      oauth_token_url: "",
      oauth_resource_metadata_url: "",
      oauth_authorization_server_url: "",
      oauth_resource: "",
    },
  });

  const transport = useWatch({ control: form.control, name: "transport" });
  const authType = useWatch({ control: form.control, name: "auth_type" }) ?? "MCP_SERVER_AUTH_TYPE_NONE";
  const registrationMode = useWatch({ control: form.control, name: "oauth_registration_mode" }) ?? "preconfigured";
  const initialHeaders = useMemo(
    () => recordToEntries(data?.mcp_server?.headers),
    [data],
  );
  const [headersDraft, setHeadersDraft] = useState<HeaderEntry[] | null>(null);
  const headers = headersDraft ?? initialHeaders;

  useEffect(() => {
    if (data?.mcp_server) {
      const s = data.mcp_server;
      form.reset({
        id: s.id ?? "",
        name: s.name,
        transport: s.transport ?? "MCP_SERVER_TRANSPORT_STDIO",
        command: s.command ?? "",
        url: s.url ?? "",
        timeout_seconds: s.timeout_seconds ? String(s.timeout_seconds) : "",
        auth_type: authTypeFromServer(s),
        oauth_registration_mode: s.auth?.oauth2?.client_id ? "preconfigured" : "dynamic",
        oauth_client_id: s.auth?.oauth2?.client_id ?? "",
        oauth_client_secret: s.auth?.oauth2?.client_secret ?? "",
        oauth_scopes: s.auth?.oauth2?.scopes?.join(" ") ?? "",
        oauth_authorization_url: s.auth?.oauth2?.authorization_url ?? "",
        oauth_token_url: s.auth?.oauth2?.token_url ?? "",
        oauth_resource_metadata_url: s.auth?.oauth2?.resource_metadata_url ?? "",
        oauth_authorization_server_url: s.auth?.oauth2?.authorization_server_url ?? "",
        oauth_resource: s.auth?.oauth2?.resource ?? "",
      });
    }
  }, [data, form]);

  function onSubmit(values: FormValues) {
    const remote = isRemoteTransport(values.transport);
    const timeoutSeconds = Number(values.timeout_seconds || 0);
    const auth = remote ? buildAuth(values) : undefined;
    updateMutation.mutate(
      {
        ...data?.mcp_server,
        id: values.id,
        name: values.name,
        transport: values.transport as MCPServerTransport,
        command: values.command,
        url: values.url,
        headers: remote && authType === "MCP_SERVER_AUTH_TYPE_STATIC_HEADERS" ? entriesToRecord(headers) : undefined,
        timeout_seconds: timeoutSeconds > 0 ? timeoutSeconds : undefined,
        auth,
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
                  <FormField control={form.control} name="auth_type" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Authentication</FormLabel>
                      <Select onValueChange={field.onChange} value={field.value}>
                        <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                        <SelectContent>
                          <SelectItem value="MCP_SERVER_AUTH_TYPE_NONE">None</SelectItem>
                          <SelectItem value="MCP_SERVER_AUTH_TYPE_STATIC_HEADERS">Static headers</SelectItem>
                          <SelectItem value="MCP_SERVER_AUTH_TYPE_OAUTH2">OAuth 2.0</SelectItem>
                        </SelectContent>
                      </Select>
                    </FormItem>
                  )} />
                  {authType === "MCP_SERVER_AUTH_TYPE_OAUTH2" && (
                    <div className="grid gap-4 md:grid-cols-2">
                      <FormField control={form.control} name="oauth_registration_mode" render={({ field }) => (
                        <FormItem>
                          <FormLabel>Client Registration</FormLabel>
                          <Select onValueChange={field.onChange} value={field.value}>
                            <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                            <SelectContent>
                              <SelectItem value="preconfigured">Preconfigured client</SelectItem>
                              <SelectItem value="dynamic">Dynamic registration</SelectItem>
                            </SelectContent>
                          </Select>
                          <FormDescription>Dynamic registration is used when the authorization server advertises it.</FormDescription>
                        </FormItem>
                      )} />
                      {registrationMode === "preconfigured" && (
                        <>
                          <FormField control={form.control} name="oauth_client_id" render={({ field }) => (
                            <FormItem><FormLabel>Client ID</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                          )} />
                          <FormField control={form.control} name="oauth_client_secret" render={({ field }) => (
                            <FormItem><FormLabel>Client Secret</FormLabel><FormControl><Input type="password" autoComplete="new-password" {...field} /></FormControl></FormItem>
                          )} />
                        </>
                      )}
                      <FormField control={form.control} name="oauth_scopes" render={({ field }) => (
                        <FormItem><FormLabel>Scopes</FormLabel><FormControl><Input placeholder="read:tools offline_access" {...field} /></FormControl></FormItem>
                      )} />
                      <FormField control={form.control} name="oauth_resource" render={({ field }) => (
                        <FormItem><FormLabel>Resource</FormLabel><FormControl><Input placeholder="Defaults to the MCP URL" {...field} /></FormControl></FormItem>
                      )} />
                      <FormField control={form.control} name="oauth_authorization_server_url" render={({ field }) => (
                        <FormItem><FormLabel>Authorization Server</FormLabel><FormControl><Input placeholder="https://issuer.example.com" {...field} /></FormControl></FormItem>
                      )} />
                      <FormField control={form.control} name="oauth_resource_metadata_url" render={({ field }) => (
                        <FormItem><FormLabel>Resource Metadata</FormLabel><FormControl><Input placeholder="https://mcp.example.com/.well-known/oauth-protected-resource" {...field} /></FormControl></FormItem>
                      )} />
                      <FormField control={form.control} name="oauth_authorization_url" render={({ field }) => (
                        <FormItem><FormLabel>Authorization URL</FormLabel><FormControl><Input placeholder="https://issuer.example.com/authorize" {...field} /></FormControl></FormItem>
                      )} />
                      <FormField control={form.control} name="oauth_token_url" render={({ field }) => (
                        <FormItem><FormLabel>Token URL</FormLabel><FormControl><Input placeholder="https://issuer.example.com/token" {...field} /></FormControl></FormItem>
                      )} />
                    </div>
                  )}
                  <div className="space-y-2">
                    <FormLabel>Headers</FormLabel>
                    {authType === "MCP_SERVER_AUTH_TYPE_STATIC_HEADERS" ? (
                      <HeadersEditor value={headers} onChange={setHeadersDraft} />
                    ) : (
                      <p className="text-sm text-muted-foreground">Static headers are only sent when static header auth is selected.</p>
                    )}
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
            <Button type="submit" disabled={updateMutation.isPending}>{updateMutation.isPending ? "Saving..." : "Save"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}

function authTypeFromServer(server: MCPServer): MCPServerAuthType {
  if (server.auth?.type && server.auth.type !== "MCP_SERVER_AUTH_TYPE_UNSPECIFIED") {
    return server.auth.type;
  }
  return server.headers && Object.keys(server.headers).length > 0
    ? "MCP_SERVER_AUTH_TYPE_STATIC_HEADERS"
    : "MCP_SERVER_AUTH_TYPE_NONE";
}

function buildAuth(values: FormValues) {
  const type = (values.auth_type || "MCP_SERVER_AUTH_TYPE_NONE") as MCPServerAuthType;
  if (type !== "MCP_SERVER_AUTH_TYPE_OAUTH2") {
    return { type };
  }
  return {
    type,
    oauth2: {
      client_id: values.oauth_registration_mode === "dynamic" ? undefined : clean(values.oauth_client_id),
      client_secret: values.oauth_registration_mode === "dynamic" ? undefined : clean(values.oauth_client_secret),
      scopes: splitScopes(values.oauth_scopes),
      authorization_url: clean(values.oauth_authorization_url),
      token_url: clean(values.oauth_token_url),
      resource_metadata_url: clean(values.oauth_resource_metadata_url),
      authorization_server_url: clean(values.oauth_authorization_server_url),
      resource: clean(values.oauth_resource),
    },
  };
}

function splitScopes(value?: string): string[] {
  return (value ?? "").split(/[,\s]+/).map((s) => s.trim()).filter(Boolean);
}

function clean(value?: string): string | undefined {
  const next = value?.trim();
  return next || undefined;
}
