import { Badge } from "@/components/ui/badge";
import { useMCPServerStatus } from "@/api/mcp-servers";
import type { MCPServerState } from "@/types/api";

const PALETTE: Record<MCPServerState, { cls: string; label: string }> = {
  STATE_UNSPECIFIED: { cls: "bg-muted text-muted-foreground", label: "Unknown" },
  STATE_CONFIGURED: { cls: "bg-muted text-muted-foreground", label: "Configured" },
  STATE_CONNECTED: { cls: "bg-emerald-500/10 text-emerald-700", label: "Connected" },
  STATE_DISCONNECTED: { cls: "bg-rose-500/10 text-rose-700", label: "Disconnected" },
  STATE_ERROR: { cls: "bg-rose-500/10 text-rose-700", label: "Error" },
};

export function ServerStatusBadge({ id }: { id: string }) {
  const { data, isLoading } = useMCPServerStatus(id);
  if (isLoading || !data) return <Badge variant="outline" className="text-xs">…</Badge>;
  const state = (data.status.state ?? "STATE_UNSPECIFIED") as MCPServerState;
  const p = PALETTE[state];
  return (
    <Badge className={p.cls}>
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
      {p.label}
    </Badge>
  );
}

export function ServerStatusInline({ id }: { id: string }) {
  const { data } = useMCPServerStatus(id);
  const count = data?.status.tool_count ?? 0;
  const state = (data?.status.state ?? "STATE_UNSPECIFIED") as MCPServerState;
  if (state === "STATE_CONNECTED") {
    return <span className="text-sm">{count}</span>;
  }
  return <span className="text-sm text-muted-foreground">—</span>;
}
