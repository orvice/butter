import { useState } from "react";
import { Navigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { GitBranch, Pencil, RefreshCw, Trash2 } from "lucide-react";
import {
  useCreateGitHost,
  useDeleteGitHost,
  useGitHosts,
  useUpdateGitHost,
} from "@/api/git-hosts";
import { GitHostKind, type GitHost } from "@/gen/agents/v1/githost_pb";
import { useAuth } from "@/hooks/use-auth";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

type FormValues = {
  name: string;
  kind: "github" | "gitlab";
  api_base_url: string;
  web_base_url: string;
};

const EMPTY_FORM: FormValues = {
  name: "",
  kind: "github",
  api_base_url: "",
  web_base_url: "",
};

const KIND_LABELS: Partial<Record<GitHostKind, string>> = {
  [GitHostKind.GITHUB]: "GitHub",
  [GitHostKind.GITLAB]: "GitLab",
};

export function GitHostKindBadge({ kind }: { kind: GitHostKind }) {
  const label = KIND_LABELS[kind];
  if (!label) return <span className="text-sm text-muted-foreground">Unknown</span>;
  return (
    <Badge variant="outline" className="font-mono text-[10px]">
      {label}
    </Badge>
  );
}

export function AdminGitHostsPage() {
  const { user: currentUser, isAdmin, isLoading: isAuthLoading } = useAuth();
  const { data, isLoading } = useGitHosts();
  const createMutation = useCreateGitHost();
  const updateMutation = useUpdateGitHost();
  const deleteMutation = useDeleteGitHost();

  const [editingId, setEditingId] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<GitHost | null>(null);
  const [form, setForm] = useState<FormValues>(EMPTY_FORM);

  const hosts = data?.hosts ?? [];
  const isEditing = editingId !== null;
  const isSaving = createMutation.isPending || updateMutation.isPending;

  if (isAuthLoading || !currentUser) {
    return (
      <Page>
        <PageHeader title="Git Hosts" />
        <PageScroll>
          <Card>
            <CardContent className="py-10 text-center text-sm text-muted-foreground">Loading…</CardContent>
          </Card>
        </PageScroll>
      </Page>
    );
  }

  if (!isAdmin) return <Navigate to="/profile" replace />;

  function patch(values: Partial<FormValues>) {
    setForm((current) => ({ ...current, ...values }));
  }

  function resetForm() {
    setEditingId(null);
    setForm(EMPTY_FORM);
  }

  function startEdit(host: GitHost) {
    setEditingId(host.id);
    setForm({
      name: host.name,
      kind: host.kind === GitHostKind.GITLAB ? "gitlab" : "github",
      api_base_url: host.apiBaseUrl,
      web_base_url: host.webBaseUrl,
    });
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const name = form.name.trim();
    const apiBaseUrl = form.api_base_url.trim();
    if (!name || !apiBaseUrl) {
      toast.error("Name and API base URL are required");
      return;
    }
    const input = {
      id: editingId ?? undefined,
      name,
      kind: form.kind === "gitlab" ? GitHostKind.GITLAB : GitHostKind.GITHUB,
      apiBaseUrl,
      webBaseUrl: form.web_base_url.trim(),
    };
    const mutation = isEditing ? updateMutation : createMutation;
    mutation.mutate(input, {
      onSuccess: () => {
        toast.success(isEditing ? "Git host updated" : "Git host created");
        resetForm();
      },
      onError: (err) => toast.error(err.message),
    });
  }

  const columns: Column<GitHost>[] = [
    {
      header: "Host",
      cell: (row) => (
        <div className="flex items-center gap-2">
          <GitBranch className="h-4 w-4 text-muted-foreground" />
          <div className="font-medium">{row.name}</div>
        </div>
      ),
    },
    {
      header: "Kind",
      cell: (row) => <GitHostKindBadge kind={row.kind} />,
    },
    {
      header: "API Base URL",
      cell: (row) => (
        <span className="max-w-sm truncate font-mono text-xs text-muted-foreground">{row.apiBaseUrl}</span>
      ),
    },
    {
      header: "",
      cell: (row) => (
        <div className="flex justify-end gap-2">
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            aria-label={`Edit ${row.name}`}
            onClick={() => startEdit(row)}
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-8 text-destructive hover:text-destructive"
            aria-label={`Delete ${row.name}`}
            onClick={() => setDeleteTarget(row)}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ),
    },
  ];

  return (
    <Page>
      <PageHeader
        title="Git Hosts"
        subtitle="Configure the platform-level allowlist of Git hosts that workspaces can bind repositories from."
      />
      <PageScroll>
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_420px]">
          <DataTable
            columns={columns}
            data={hosts}
            isLoading={isLoading}
            emptyMessage="No Git hosts configured."
            emptyDescription="Add a host so workspaces can bind repositories."
          />

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <GitBranch className="h-4 w-4" />
                {isEditing ? "Edit Git Host" : "New Git Host"}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-4">
                <Field label="Name">
                  <Input
                    value={form.name}
                    onChange={(e) => patch({ name: e.target.value })}
                    disabled={isSaving}
                    placeholder="GitHub.com"
                  />
                </Field>
                <Field label="Kind">
                  <Select
                    value={form.kind}
                    onValueChange={(value) => patch({ kind: value as FormValues["kind"] })}
                    disabled={isSaving}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="github">GitHub</SelectItem>
                      <SelectItem value="gitlab">GitLab</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <Field label="API Base URL">
                  <Input
                    value={form.api_base_url}
                    onChange={(e) => patch({ api_base_url: e.target.value })}
                    disabled={isSaving}
                    placeholder={form.kind === "gitlab" ? "https://gitlab.com/api/v4" : "https://api.github.com"}
                  />
                </Field>
                <Field label="Web Base URL (optional)">
                  <Input
                    value={form.web_base_url}
                    onChange={(e) => patch({ web_base_url: e.target.value })}
                    disabled={isSaving}
                    placeholder={form.kind === "gitlab" ? "https://gitlab.com" : "https://github.com"}
                  />
                </Field>
                <p className="rounded-md border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
                  The API base URL is the full REST API root the adapter joins request paths onto. The web
                  base URL is only used to build repository links in the dashboard.
                </p>
                <div className="flex gap-2">
                  <Button type="submit" disabled={isSaving}>
                    {isSaving ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : null}
                    {isEditing ? "Save Host" : "Create Host"}
                  </Button>
                  <Button type="button" variant="outline" onClick={resetForm} disabled={isSaving}>
                    Clear
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        </div>
      </PageScroll>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Git Host"
        description={`Delete ${deleteTarget?.name ?? "this host"}? Workspace bindings referencing it will stop working.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget?.id) return;
          deleteMutation.mutate(deleteTarget.id, {
            onSuccess: () => {
              toast.success("Git host deleted");
              setDeleteTarget(null);
              if (editingId === deleteTarget.id) resetForm();
            },
            onError: (err) => toast.error(err.message),
          });
        }}
      />
    </Page>
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
