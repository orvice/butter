import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useLayoutDensity } from "@/hooks/use-layout-density";
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
  const { isCompact } = useLayoutDensity();

  return (
    <aside
      className={cn(
        "flex w-full shrink-0 flex-col border-b bg-card/40 md:h-full md:max-h-none md:border-b-0 md:border-r",
        isCompact ? "max-h-52 md:w-64" : "max-h-64 md:w-72",
      )}
    >
      <div className={cn("flex items-center justify-between gap-2 border-b px-3", isCompact ? "py-2" : "py-3")}>
        <div className={cn("flex items-center gap-2 font-semibold", isCompact ? "text-xs" : "text-sm")}>
          <MessageCircle className={cn(isCompact ? "h-3.5 w-3.5" : "h-4 w-4")} /> Chats
        </div>
        <Button size={isCompact ? "xs" : "sm"} onClick={onNewChat}>
          <Plus className="mr-1 h-3 w-3" /> New
        </Button>
      </div>
      <div className={cn("flex-1 overflow-y-auto", isCompact ? "p-1.5" : "p-2")}>
        {isLoading ? (
          <div className={cn(isCompact ? "space-y-1.5" : "space-y-2")}>
            <Skeleton className={cn("w-full", isCompact ? "h-10" : "h-12")} />
            <Skeleton className={cn("w-full", isCompact ? "h-10" : "h-12")} />
            <Skeleton className={cn("w-full", isCompact ? "h-10" : "h-12")} />
          </div>
        ) : sessions.length === 0 ? (
          <div className={cn("px-3 text-center text-xs text-muted-foreground", isCompact ? "py-5" : "py-8")}>
            No chats yet. Click <span className="font-medium">New</span> to start one.
          </div>
        ) : (
          <ul className={cn(isCompact ? "space-y-0.5" : "space-y-1")}>
            {sessions.map((s) => {
              const active = s.session_id === activeSessionId;
              const agent = agentNameOf(s.state);
              const updated = s.last_update_time ? new Date(s.last_update_time) : null;
              return (
                <li key={s.session_id}>
                  <div
                    className={cn(
                      "group flex cursor-pointer items-start gap-2 rounded-md px-2 text-xs transition-colors",
                      isCompact ? "py-1.5" : "py-2",
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
                      <div className="flex items-center justify-between gap-2 text-[10px] leading-tight text-muted-foreground">
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
