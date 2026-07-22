import { useState } from "react";
import { toast } from "sonner";
import { useAPITokens, useCreateAPIToken, useRevokeAPIToken } from "@/api/apitokens";
import { BASE_URL } from "@/api/transport";
import { PageHeader } from "@/components/page-header";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { DataTable, type Column } from "@/components/data-table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Copy, KeyRound, Trash2 } from "lucide-react";
import type { APIToken } from "@/types/api";

function openAISDKBaseURL(): string {
  const configuredBaseURL = BASE_URL.replace(/\/+$/, "");
  return `${configuredBaseURL || window.location.origin}/api/v1`;
}

export default function APITokenListPage() {
  const { data, isLoading } = useAPITokens();
  const create = useCreateAPIToken();
  const revoke = useRevokeAPIToken();

  const [createOpen, setCreateOpen] = useState(false);
  const [newName, setNewName] = useState("");
  const [revealedSecret, setRevealedSecret] = useState<{ name: string; secret: string } | null>(null);

  const tokens = data?.tokens ?? [];
  const sdkBaseURL = openAISDKBaseURL();

  function handleCreate() {
    if (!newName.trim()) {
      toast.error("Name is required");
      return;
    }
    create.mutate(newName.trim(), {
      onSuccess: (res) => {
        setCreateOpen(false);
        setNewName("");
        setRevealedSecret({ name: res.token.name, secret: res.secret });
      },
      onError: (e) => toast.error(e.message),
    });
  }

  const columns: Column<APIToken>[] = [
    { header: "Name", accessorKey: "name" },
    {
      header: "Prefix",
      cell: (t) => <span className="font-mono text-xs">{t.prefix ?? "-"}…</span>,
    },
    {
      header: "Created",
      cell: (t) => (
        <span className="text-xs text-muted-foreground">
          {t.created_at ? new Date(t.created_at).toLocaleDateString() : "-"}
        </span>
      ),
    },
    {
      header: "Last used",
      cell: (t) => (
        <span className="text-xs text-muted-foreground">
          {t.last_used_at ? new Date(t.last_used_at).toLocaleString() : "Never"}
        </span>
      ),
    },
    {
      header: "Status",
      cell: (t) => (t.revoked ? <Badge variant="destructive">Revoked</Badge> : <Badge>Active</Badge>),
    },
    {
      header: "Actions",
      cell: (t) =>
        t.revoked ? null : (
          <Button
            variant="ghost"
            size="sm"
            onClick={() =>
              revoke.mutate(t.id, { onSuccess: () => toast.success("Token revoked") })
            }
          >
            <Trash2 className="mr-1 h-3 w-3" /> Revoke
          </Button>
        ),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader title="API Tokens" />

      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle>Bearer Tokens</CardTitle>
            <CardDescription>Manage tokens used by external integrations to call ConnectRPC and OpenAI-compatible APIs.</CardDescription>
          </div>
          <Button className="w-full sm:w-auto" onClick={() => setCreateOpen(true)}>
            <KeyRound className="mr-2 h-4 w-4" /> Generate New Token
          </Button>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="border-b pb-5">
            <div className="mb-2 space-y-0.5">
              <Label htmlFor="openai-sdk-base-url">OpenAI SDK Base URL</Label>
              <p className="text-xs text-muted-foreground">Use a generated bearer token to authenticate requests to this endpoint.</p>
            </div>
            <div className="flex items-center gap-2">
              <Input id="openai-sdk-base-url" readOnly value={sdkBaseURL} className="font-mono text-xs" />
              <Button
                type="button"
                size="icon"
                variant="outline"
                aria-label="Copy OpenAI SDK Base URL"
                title="Copy OpenAI SDK Base URL"
                onClick={() => {
                  navigator.clipboard.writeText(sdkBaseURL);
                  toast.success("OpenAI SDK Base URL copied");
                }}
              >
                <Copy className="h-4 w-4" />
              </Button>
            </div>
          </div>
          <DataTable
            columns={columns}
            data={tokens}
            isLoading={isLoading}
            emptyMessage="No tokens yet. Generate one to enable external API access."
          />
        </CardContent>
      </Card>

      {/* Create dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New API Token</DialogTitle>
            <DialogDescription>
              The secret will be shown once after creation. Save it somewhere safe.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="token-name">Name</Label>
            <Input
              id="token-name"
              placeholder="e.g. Zendesk Ingestion"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} disabled={create.isPending}>
              {create.isPending ? "Creating…" : "Generate"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Reveal secret once */}
      <Dialog open={!!revealedSecret} onOpenChange={(o) => !o && setRevealedSecret(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Token created: {revealedSecret?.name}</DialogTitle>
            <DialogDescription>
              Copy this secret now. You will not be able to see it again.
            </DialogDescription>
          </DialogHeader>
          <div className="flex items-center gap-2 rounded-md border bg-muted p-3 font-mono text-xs">
            <span className="flex-1 break-all">{revealedSecret?.secret}</span>
            <Button
              size="icon"
              variant="ghost"
              onClick={() => {
                if (revealedSecret?.secret) {
                  navigator.clipboard.writeText(revealedSecret.secret);
                  toast.success("Copied to clipboard");
                }
              }}
            >
              <Copy className="h-4 w-4" />
            </Button>
          </div>
          <DialogFooter>
            <Button onClick={() => setRevealedSecret(null)}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
