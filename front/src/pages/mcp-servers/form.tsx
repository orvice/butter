import { useEffect, useMemo, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import {
  HeadersEditor,
  entriesToRecord,
  recordToEntries,
  type HeaderEntry,
} from "./headers-editor";
import type { MCPServer, MCPServerAuthType, MCPServerTransport } from "@/types/api";

const schema = z.object({
  id: z.string().min(1, "ID is required"),
  name: z.string().min(1, "Name is required"),
  transport: z.string(),
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

type MCPServerFormProps = {
  mode: "create" | "edit";
  initialValue?: MCPServer;
  loading?: boolean;
  submitLabel: string;
  onCancel: () => void;
  onSubmit: (server: MCPServer) => void;
};

function isRemoteTransport(t: string): boolean {
  return (
    t === "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP" ||
    t === "MCP_SERVER_TRANSPORT_SSE"
  );
}

export default function MCPServerForm({
  mode,
  initialValue,
  loading,
  submitLabel,
  onCancel,
  onSubmit,
}: MCPServerFormProps) {
  const isEdit = mode === "edit";
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      id: "",
      name: "",
      transport: "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP",
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

  const authType = useWatch({ control: form.control, name: "auth_type" }) ?? "MCP_SERVER_AUTH_TYPE_NONE";
  const registrationMode = useWatch({ control: form.control, name: "oauth_registration_mode" }) ?? "preconfigured";
  const initialHeaders = useMemo(() => recordToEntries(initialValue?.headers), [initialValue]);
  const [headersDraft, setHeadersDraft] = useState<HeaderEntry[] | null>(null);
  const headers = headersDraft ?? initialHeaders;

  useEffect(() => {
    if (!initialValue) return;
    form.reset({
      id: initialValue.id ?? "",
      name: initialValue.name,
      transport: initialValue.transport ?? "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP",
      url: initialValue.url ?? "",
      timeout_seconds: initialValue.timeout_seconds ? String(initialValue.timeout_seconds) : "",
      auth_type: authTypeFromServer(initialValue),
      oauth_registration_mode: initialValue.auth?.oauth2?.client_id ? "preconfigured" : "dynamic",
      oauth_client_id: initialValue.auth?.oauth2?.client_id ?? "",
      oauth_client_secret: initialValue.auth?.oauth2?.client_secret ?? "",
      oauth_scopes: initialValue.auth?.oauth2?.scopes?.join(" ") ?? "",
      oauth_authorization_url: initialValue.auth?.oauth2?.authorization_url ?? "",
      oauth_token_url: initialValue.auth?.oauth2?.token_url ?? "",
      oauth_resource_metadata_url: initialValue.auth?.oauth2?.resource_metadata_url ?? "",
      oauth_authorization_server_url: initialValue.auth?.oauth2?.authorization_server_url ?? "",
      oauth_resource: initialValue.auth?.oauth2?.resource ?? "",
    });
  }, [form, initialValue]);

  function handleSubmit(values: FormValues) {
    const remote = isRemoteTransport(values.transport);
    const timeoutSeconds = Number(values.timeout_seconds || 0);
    const auth = remote ? buildAuth(values) : undefined;
    onSubmit({
      id: values.id,
      name: values.name,
      transport: values.transport as MCPServerTransport,
      url: values.url,
      headers: remote && authType === "MCP_SERVER_AUTH_TYPE_STATIC_HEADERS" ? entriesToRecord(headers) : undefined,
      timeout_seconds: timeoutSeconds > 0 ? timeoutSeconds : undefined,
      auth,
    });
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Basic Info</CardTitle>
            <CardDescription>
              {isEdit
                ? "The ID is fixed; update the display name or transport as needed."
                : "Use a stable ID for references and a readable name for operators."}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="id" render={({ field }) => (
              <FormItem><FormLabel>ID</FormLabel><FormControl><Input placeholder="my-mcp-server" {...field} disabled={isEdit} /></FormControl><FormMessage /></FormItem>
            )} />
            <FormField control={form.control} name="name" render={({ field }) => (
              <FormItem><FormLabel>Name</FormLabel><FormControl><Input placeholder="My MCP Server" {...field} /></FormControl><FormMessage /></FormItem>
            )} />
            <FormField control={form.control} name="transport" render={({ field }) => (
              <FormItem>
                <FormLabel>Transport</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                  <SelectContent>
                    <SelectItem value="MCP_SERVER_TRANSPORT_STREAMABLE_HTTP">HTTP</SelectItem>
                    <SelectItem value="MCP_SERVER_TRANSPORT_SSE">SSE</SelectItem>
                  </SelectContent>
                </Select>
              </FormItem>
            )} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Connection</CardTitle>
            <CardDescription>
              {isEdit
                ? "Endpoint, auth, headers, and timeout used for tool discovery and execution."
                : "Provide the endpoint, auth details, and timeout used when agents call tools."}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="url" render={({ field }) => (
              <FormItem><FormLabel>URL</FormLabel><FormControl><Input placeholder="https://..." {...field} /></FormControl></FormItem>
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
          </CardContent>
        </Card>

        <div className="sticky bottom-0 z-10 -mx-1 flex gap-3 border-t bg-background/95 px-1 py-4 backdrop-blur supports-[backdrop-filter]:bg-background/80">
          <Button type="button" variant="outline" onClick={onCancel}>Cancel</Button>
          <Button type="submit" disabled={loading}>{loading ? "Saving..." : submitLabel}</Button>
        </div>
      </form>
    </Form>
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
