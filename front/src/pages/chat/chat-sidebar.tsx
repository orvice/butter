import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { Plus, Trash2, MessageCircle } from "lucide-react";
import type { SessionInfo } from "@/types/api";

interface ChatSidebarProps {
  sessions: SessionInfo[];
  isLoading: boolean;
  activeSessionId: string | null;
  onSelect: (sessionId: string) => void;
  onNewChat: () => void;
  onDelete: (session: SessionInfo) => void;
}

function agentNameOf(state: SessionInfo["state"]): string | undefined {
  if (!state) return undefined;
  const v = state["agent_name"];
  return typeof v === "string" ? v : undefined;
}

export function ChatSidebar({
  sessions,
  isLoading,
  activeSessionId,
  onSelect,
  onNewChat,
  onDelete,
}: ChatSidebarProps) {
  return (
    <aside className="flex max-h-64 w-full shrink-0 flex-col border-b bg-card/40 md:h-full md:max-h-none md:w-72 md:border-b-0 md:border-r">
      <div className="flex items-center justify-between gap-2 border-b px-3 py-3">
        <div className="flex items-center gap-2 text-sm font-semibold">
          <MessageCircle className="h-4 w-4" /> Chats
        </div>
        <Button size="sm" onClick={onNewChat}>
          <Plus className="mr-1 h-3 w-3" /> New
        </Button>
      </div>
      <div className="flex-1 overflow-y-auto p-2">
        {isLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        ) : sessions.length === 0 ? (
          <div className="px-3 py-8 text-center text-xs text-muted-foreground">
            No chats yet. Click <span className="font-medium">New</span> to start one.
          </div>
        ) : (
          <ul className="space-y-1">
            {sessions.map((s) => {
              const active = s.session_id === activeSessionId;
              const agent = agentNameOf(s.state);
              const updated = s.last_update_time ? new Date(s.last_update_time) : null;
              return (
                <li key={s.session_id}>
                  <div
                    className={cn(
                      "group flex cursor-pointer items-start gap-2 rounded-md px-2 py-2 text-xs transition-colors",
                      active
                        ? "bg-accent text-accent-foreground"
                        : "hover:bg-accent/50",
                    )}
                    onClick={() => onSelect(s.session_id)}
                  >
                    <div className="min-w-0 flex-1">
                      <div className="truncate font-medium">
                        {agent ?? "Unknown agent"}
                      </div>
                      <div className="flex items-center justify-between text-[10px] text-muted-foreground">
                        <span className="truncate font-mono">
                          {s.session_id.slice(0, 12)}…
                        </span>
                        {updated ? <span>{updated.toLocaleDateString()}</span> : null}
                      </div>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      className="opacity-100 md:opacity-0 md:group-hover:opacity-100"
                      onClick={(e) => {
                        e.stopPropagation();
                        onDelete(s);
                      }}
                      aria-label="Delete chat"
                    >
                      <Trash2 className="h-3 w-3" />
                    </Button>
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </aside>
  );
}
