import { Cpu, Link as LinkIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { enumLabel } from "@/lib/constants";
import type { RemoteAgent } from "@/types/api";

interface AgentRemoteAgentsFieldProps {
  value?: string[];
  onChange: (value: string[]) => void;
  remoteAgents?: RemoteAgent[];
  isLoading?: boolean;
}

function remoteAgentDetail(agent: RemoteAgent): string {
  if (agent.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON") {
    return `${agent.daemon_runtime_id || "-"} / ${agent.acp_runtime || "-"}`;
  }
  return agent.url || "-";
}

export function AgentRemoteAgentsField({
  value,
  onChange,
  remoteAgents,
  isLoading,
}: AgentRemoteAgentsFieldProps) {
  const selected = value ?? [];
  const agents = remoteAgents ?? [];

  if (isLoading) {
    return <p className="text-sm text-muted-foreground">Loading remote agents...</p>;
  }

  if (agents.length === 0) {
    return (
      <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
        No remote agents registered.
      </div>
    );
  }

  const toggle = (id: string) => {
    onChange(
      selected.includes(id)
        ? selected.filter((selectedId) => selectedId !== id)
        : [...selected, id],
    );
  };

  return (
    <div className="grid gap-2 md:grid-cols-2">
      {agents.map((agent) => {
        const id = agent.id ?? "";
        const isDaemon = agent.protocol === "REMOTE_AGENT_PROTOCOL_DAEMON";
        const Icon = isDaemon ? Cpu : LinkIcon;
        const isSelected = selected.includes(id);
        return (
          <button
            key={id || agent.name}
            type="button"
            disabled={!id}
            onClick={() => id && toggle(id)}
            className={`min-h-24 rounded-md border p-3 text-left transition-colors ${
              isSelected ? "border-primary bg-primary/10" : "hover:bg-muted"
            } ${!id ? "cursor-not-allowed opacity-60" : ""}`}
          >
            <div className="flex items-start gap-2">
              <Icon className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
              <div className="min-w-0 flex-1">
                <div className="flex items-start justify-between gap-2">
                  <span className="truncate font-medium">{agent.name}</span>
                  {isSelected && <Badge>Selected</Badge>}
                </div>
                <div className="mt-1 truncate text-xs text-muted-foreground">
                  {remoteAgentDetail(agent)}
                </div>
              </div>
            </div>
            <Badge variant="outline" className="mt-3 font-mono text-[10px]">
              {enumLabel(agent.protocol, "Unknown")}
            </Badge>
          </button>
        );
      })}
    </div>
  );
}
