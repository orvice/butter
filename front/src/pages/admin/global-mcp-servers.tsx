import { useMemo, useState } from "react";
import { toast } from "sonner";
import {
  useCreateGlobalMCPServer,
  useDeleteGlobalMCPServer,
  useGlobalMCPServers,
  useInstallGlobalMCPServer,
  useUpdateGlobalMCPServer,
} from "@/api/global-mcp-servers";
import { useWorkspace } from "@/hooks/use-workspace";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
  Cloud,
  KeyRound,
  Pencil,
  Plus,
  Radio,
  RefreshCw,
  Server,
  ShieldCheck,
  Trash2,
} from "lucide-react";
import { MCP_TRANSPORT_LABELS } from "@/lib/constants";
import type { MCPServer, MCPServerAuthType, MCPServerTransport } from "@/types/api";

type RegistrationMode = "dynamic" | "manual";

type FormValues = {
  id: string;
  name: string;
  transport: MCPServerTransport;
  url: string;
  timeout_seconds: string;
  auth_type: MCPServerAuthType;
  registration_mode: RegistrationMode;
  client_id: string;
  client_secret: string;
  scopes: string;
  resource: string;
  authorization_server_url: string;
  resource_metadata_url: string;
  authorization_url: string;
  token_url: string;
};

const EMPTY_FORM: FormValues = {
  id: "",
  name: "",
  transport: "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP",
  url: "",
  timeout_seconds: "10",
  auth_type: "MCP_SERVER_AUTH_TYPE_OAUTH2",
  registration_mode: "dynamic",
  client_id: "",
  client_secret: "",
  scopes: "",
  resource: "",
  authorization_server_url: "",
  resource_metadata_url: "",
  authorization_url: "",
  token_url: "",
};

const STARTERS: Array<{ label: string; values: Partial<FormValues> }> = [
  {
    label: "Linear",
    values: {
      id: "linear",
      name: "Linear",
      url: "https://mcp.linear.app/mcp",
      resource: "https://mcp.linear.app/mcp",
    },
  },
  {
    label: "Gmail",
    values: {
      id: "gmail",
      name: "Gmail",
      url: "https://gmailmcp.googleapis.com/mcp/v1",
      resource: "https://gmailmcp.googleapis.com/mcp/v1",
    },
  },
];

const TRANSPORT_ICON: Record<MCPServerTransport, typeof Server> = {
  MCP_SERVER_TRANSPORT_UNSPECIFIED: Server,
  MCP_SERVER_TRANSPORT_STREAMABLE_HTTP: Cloud,
  MCP_SERVER_TRANSPORT_SSE: Radio,
};

export default function AdminGlobalMCPServersPage() {
  const { data, isLoading } = useGlobalMCPServers();
  const { selectedWorkspaceId } = useWorkspace();
  const createMutation = useCreateGlobalMCPServer();
  const updateMutation = useUpdateGlobalMCPServer();
  const deleteMutation = useDeleteGlobalMCPServer();
  const installMutation = useInstallGlobalMCPServer();
  const [editingId, setEditingId] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<MCPServer | null>(null);
  const [form, setForm] = useState<FormValues>(EMPTY_FORM);

  const servers = data?.mcp_servers ?? [];
  const isEditing = editingId !== null;
  const isSaving = createMutation.isPending || updateMutation.isPending;

  const columns: Column<MCPServer>[] = [
    {
      header: "Preset",
      cell: (row) => {
        const Icon = TRANSPORT_ICON[row.transport ?? "MCP_SERVER_TRANSPORT_UNSPECIFIED"];
        return (
          <div className="flex items-center gap-2">
            <Icon className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="font-medium">{row.name}</div>
              <div className="max-w-sm truncate text-xs text-muted-foreground">{row.url || row.id}</div>
            </div>
          </div>
        );
      },
    },
    {
      header: "Transport",
      cell: (row) => (
        <Badge variant="outline" className="font-mono text-[10px]">
          {MCP_TRANSPORT_LABELS[row.transport ?? "MCP_SERVER_TRANSPORT_UNSPECIFIED"]}
        </Badge>
      ),
    },
    {
      header: "Auth",
      cell: (row) => <AuthBadge server={row} />,
    },
    {
      header: "OAuth Config",
      cell: (row) => <OAuthConfigSummary server={row} />,
    },
    {
      header: "",
      cell: (row) => (
        <div className="flex justify-end gap-2">
          <Button
            variant="ghost"
            size="sm"
            disabled={!row.id || !selectedWorkspaceId || installMutation.isPending}
            onClick={() => {
              if (!row.id) return;
              installMutation.mutate(
                { id: row.id, workspaceId: selectedWorkspaceId },
                {
                  onSuccess: () => toast.success(`${row.name} installed to workspace`),
                  onError: (err) => toast.error(err.message),
                },
              );
            }}
          >
            <Plus className="mr-1 h-3.5 w-3.5" />
            Install
          </Button>
          <Button variant="ghost" size="icon-sm" aria-label={`Edit ${row.name}`} onClick={() => startEdit(row)}>
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            className="text-destructive hover:text-destructive"
            aria-label={`Delete ${row.name}`}
            onClick={() => setDeleteTarget(row)}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ),
    },
  ];

  const helperText = useMemo(() => {
    if (form.auth_type !== "MCP_SERVER_AUTH_TYPE_OAUTH2") {
      return "Workspace users can install this preset without OAuth client setup.";
    }
    if (form.registration_mode === "dynamic") {
      return "OAuth discovery and dynamic client registration are used when the MCP resource advertises them.";
    }
    return "Manual OAuth clients are stored as admin-owned global preset configuration.";
  }, [form.auth_type, form.registration_mode]);

  function patch(values: Partial<FormValues>) {
    setForm((current) => ({ ...current, ...values }));
  }

  function resetForm() {
    setEditingId(null);
    setForm(EMPTY_FORM);
  }

  function applyStarter(values: Partial<FormValues>) {
    setEditingId(null);
    setForm({ ...EMPTY_FORM, ...values });
  }

  function startEdit(server: MCPServer) {
    const oauth = server.auth?.oauth2;
    setEditingId(server.id ?? "");
    setForm({
      id: server.id ?? "",
      name: server.name ?? "",
      transport: server.transport ?? "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP",
      url: server.url ?? "",
      timeout_seconds: server.timeout_seconds ? String(server.timeout_seconds) : "",
      auth_type: authTypeFromServer(server),
      registration_mode: oauth?.client_id ? "manual" : "dynamic",
      client_id: oauth?.client_id ?? "",
      client_secret: oauth?.client_secret ?? "",
      scopes: oauth?.scopes?.join(" ") ?? "",
      resource: oauth?.resource ?? "",
      authorization_server_url: oauth?.authorization_server_url ?? "",
      resource_metadata_url: oauth?.resource_metadata_url ?? "",
      authorization_url: oauth?.authorization_url ?? "",
      token_url: oauth?.token_url ?? "",
    });
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const server = formToServer(form);
    if (!server.id || !server.name || !server.url) {
      toast.error("ID, name, and URL are required");
      return;
    }
    const mutation = isEditing ? updateMutation : createMutation;
    mutation.mutate(server, {
      onSuccess: () => {
        toast.success(isEditing ? "Global MCP preset updated" : "Global MCP preset created");
        resetForm();
      },
      onError: (err) => toast.error(err.message),
    });
  }

  return (
    <>
      <div className="mb-6 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight text-foreground">Global MCP Presets</h2>
          <p className="mt-1 max-w-3xl text-sm text-muted-foreground">
            Configure admin-owned MCP connections that workspaces can install and OAuth from their MCP server list.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          {STARTERS.map((starter) => (
            <Button key={starter.label} type="button" variant="outline" size="sm" onClick={() => applyStarter(starter.values)}>
              <Plus className="mr-2 h-4 w-4" />
              {starter.label}
            </Button>
          ))}
        </div>
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_420px]">
        <DataTable
          columns={columns}
          data={servers}
          isLoading={isLoading}
          emptyMessage="No global MCP presets configured."
        />

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <ShieldCheck className="h-4 w-4" />
              {isEditing ? "Edit Preset" : "New Preset"}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid gap-3 sm:grid-cols-2">
                <Field label="ID">
                  <Input value={form.id} onChange={(e) => patch({ id: e.target.value })} disabled={isEditing || isSaving} placeholder="linear" />
                </Field>
                <Field label="Name">
                  <Input value={form.name} onChange={(e) => patch({ name: e.target.value })} disabled={isSaving} placeholder="Linear" />
                </Field>
              </div>

              <Field label="URL">
                <Input value={form.url} onChange={(e) => patch({ url: e.target.value })} disabled={isSaving} placeholder="https://mcp.linear.app/mcp" />
              </Field>

              <div className="grid gap-3 sm:grid-cols-2">
                <Field label="Transport">
                  <Select value={form.transport} onValueChange={(value) => patch({ transport: value as MCPServerTransport })} disabled={isSaving}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="MCP_SERVER_TRANSPORT_STREAMABLE_HTTP">HTTP</SelectItem>
                      <SelectItem value="MCP_SERVER_TRANSPORT_SSE">SSE</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <Field label="Timeout Seconds">
                  <Input
                    type="number"
                    min={0}
                    value={form.timeout_seconds}
                    onChange={(e) => patch({ timeout_seconds: e.target.value.replace(/\D/g, "") })}
                    disabled={isSaving}
                    placeholder="10"
                  />
                </Field>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <Field label="Authentication">
                  <Select value={form.auth_type} onValueChange={(value) => patch({ auth_type: value as MCPServerAuthType })} disabled={isSaving}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="MCP_SERVER_AUTH_TYPE_NONE">None</SelectItem>
                      <SelectItem value="MCP_SERVER_AUTH_TYPE_OAUTH2">OAuth 2.0</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                {form.auth_type === "MCP_SERVER_AUTH_TYPE_OAUTH2" ? (
                  <Field label="Client Registration">
                    <Select value={form.registration_mode} onValueChange={(value) => patch({ registration_mode: value as RegistrationMode })} disabled={isSaving}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="dynamic">Discovery / dynamic</SelectItem>
                        <SelectItem value="manual">Manual client</SelectItem>
                      </SelectContent>
                    </Select>
                  </Field>
                ) : null}
              </div>

              <p className="rounded-md border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">{helperText}</p>

              {form.auth_type === "MCP_SERVER_AUTH_TYPE_OAUTH2" ? (
                <div className="space-y-4">
                  {form.registration_mode === "manual" ? (
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label="Client ID">
                        <Input value={form.client_id} onChange={(e) => patch({ client_id: e.target.value })} disabled={isSaving} />
                      </Field>
                      <Field label="Client Secret">
                        <Input type="password" autoComplete="new-password" value={form.client_secret} onChange={(e) => patch({ client_secret: e.target.value })} disabled={isSaving} />
                      </Field>
                    </div>
                  ) : null}

                  <Field label="Scopes">
                    <Input value={form.scopes} onChange={(e) => patch({ scopes: e.target.value })} disabled={isSaving} placeholder="read:tools offline_access" />
                  </Field>
                  <Field label="Resource">
                    <Input value={form.resource} onChange={(e) => patch({ resource: e.target.value })} disabled={isSaving} placeholder="Defaults to the MCP URL" />
                  </Field>

                  <div className="grid gap-3 sm:grid-cols-2">
                    <Field label="Authorization Server">
                      <Input value={form.authorization_server_url} onChange={(e) => patch({ authorization_server_url: e.target.value })} disabled={isSaving} placeholder="https://issuer.example.com" />
                    </Field>
                    <Field label="Resource Metadata">
                      <Input value={form.resource_metadata_url} onChange={(e) => patch({ resource_metadata_url: e.target.value })} disabled={isSaving} placeholder="https://mcp.example.com/.well-known/oauth-protected-resource" />
                    </Field>
                    <Field label="Authorization URL">
                      <Input value={form.authorization_url} onChange={(e) => patch({ authorization_url: e.target.value })} disabled={isSaving} placeholder="https://issuer.example.com/authorize" />
                    </Field>
                    <Field label="Token URL">
                      <Input value={form.token_url} onChange={(e) => patch({ token_url: e.target.value })} disabled={isSaving} placeholder="https://issuer.example.com/token" />
                    </Field>
                  </div>
                </div>
              ) : null}

              <Field label="Notes">
                <Textarea
                  value="Workspace installs copy this preset into the selected workspace; admin-owned OAuth secrets are redacted from normal workspace reads."
                  readOnly
                  className="min-h-16 resize-none text-muted-foreground"
                />
              </Field>

              <div className="flex gap-2">
                <Button type="submit" disabled={isSaving}>
                  {isSaving ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <ShieldCheck className="mr-2 h-4 w-4" />}
                  {isEditing ? "Save Preset" : "Create Preset"}
                </Button>
                <Button type="button" variant="outline" onClick={resetForm} disabled={isSaving}>
                  Clear
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Global MCP Preset"
        description={`Delete ${deleteTarget?.name ?? "this preset"}? Existing workspace installs are not removed.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget?.id) return;
          deleteMutation.mutate(deleteTarget.id, {
            onSuccess: () => {
              toast.success("Global MCP preset deleted");
              setDeleteTarget(null);
              if (editingId === deleteTarget.id) resetForm();
            },
            onError: (err) => toast.error(err.message),
          });
        }}
      />
    </>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  const id = label.toLowerCase().replace(/[^a-z0-9]+/g, "-");
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      {children}
    </div>
  );
}

function AuthBadge({ server }: { server: MCPServer }) {
  const type = authTypeFromServer(server);
  if (type === "MCP_SERVER_AUTH_TYPE_OAUTH2") {
    return (
      <Badge className="bg-emerald-500/10 text-emerald-700">
        <KeyRound className="mr-1 h-3 w-3" />
        OAuth 2.0
      </Badge>
    );
  }
  return <span className="text-sm text-muted-foreground">None</span>;
}

function OAuthConfigSummary({ server }: { server: MCPServer }) {
  if (authTypeFromServer(server) !== "MCP_SERVER_AUTH_TYPE_OAUTH2") {
    return <span className="text-sm text-muted-foreground">-</span>;
  }
  const oauth = server.auth?.oauth2;
  const manual = !!oauth?.client_id;
  return (
    <div className="space-y-1 text-sm">
      <div>{manual ? "Manual client" : "Discovery / dynamic"}</div>
      {oauth?.scopes?.length ? (
        <div className="max-w-xs truncate text-xs text-muted-foreground">{oauth.scopes.join(" ")}</div>
      ) : null}
    </div>
  );
}

function authTypeFromServer(server: MCPServer): MCPServerAuthType {
  if (server.auth?.type && server.auth.type !== "MCP_SERVER_AUTH_TYPE_UNSPECIFIED") {
    return server.auth.type;
  }
  return "MCP_SERVER_AUTH_TYPE_NONE";
}

function formToServer(values: FormValues): MCPServer {
  const timeout = Number(values.timeout_seconds || 0);
  const authType = values.auth_type;
  return {
    id: clean(values.id),
    name: values.name.trim(),
    transport: values.transport,
    url: clean(values.url),
    timeout_seconds: timeout > 0 ? timeout : undefined,
    auth: authType === "MCP_SERVER_AUTH_TYPE_OAUTH2"
      ? {
          type: authType,
          oauth2: {
            client_id: values.registration_mode === "manual" ? clean(values.client_id) : undefined,
            client_secret: values.registration_mode === "manual" ? clean(values.client_secret) : undefined,
            scopes: splitScopes(values.scopes),
            authorization_url: clean(values.authorization_url),
            token_url: clean(values.token_url),
            resource_metadata_url: clean(values.resource_metadata_url),
            authorization_server_url: clean(values.authorization_server_url),
            resource: clean(values.resource),
          },
        }
      : { type: authType },
  };
}

function splitScopes(value?: string): string[] {
  return (value ?? "").split(/[,\s]+/).map((s) => s.trim()).filter(Boolean);
}

function clean(value?: string): string | undefined {
  const next = value?.trim();
  return next || undefined;
}
