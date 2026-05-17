import { useMemo, useState } from "react";
import { toast } from "sonner";
import { useAuth } from "@/hooks/use-auth";
import { useCreateSession, useDeleteSession, useSessions } from "@/api/sessions";
import { DeleteDialog } from "@/components/delete-dialog";
import { AgentPicker } from "./agent-picker";
import { ChatSidebar } from "./chat-sidebar";
import { ChatWindow } from "./chat-window";
import type { SessionInfo } from "@/types/api";

const APP_NAME = "web-chat";

function agentNameOf(state: SessionInfo["state"]): string | null {
  if (!state) return null;
  const v = state["agent_name"];
  return typeof v === "string" ? v : null;
}

export default function ChatPage() {
  const { user } = useAuth();
  const userId = user?.id ?? "";

  const sessionsQuery = useSessions({
    app_name: APP_NAME,
    user_id: userId || undefined,
    page_size: 100,
  });
  const createMutation = useCreateSession();
  const deleteMutation = useDeleteSession();

  const sessions = useMemo(() => sessionsQuery.data?.sessions ?? [], [sessionsQuery.data]);

  const [preferredSessionId, setPreferredSessionId] = useState<string | null>(null);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<SessionInfo | null>(null);

  const activeSessionId = useMemo(() => {
    if (preferredSessionId && sessions.some((s) => s.session_id === preferredSessionId)) {
      return preferredSessionId;
    }
    return sessions[0]?.session_id ?? null;
  }, [preferredSessionId, sessions]);

  const activeSession = sessions.find((s) => s.session_id === activeSessionId) ?? null;
  const activeAgent = activeSession ? agentNameOf(activeSession.state) : null;

  async function handleCreate(agentName: string) {
    if (!userId) {
      toast.error("Missing user context; please re-login.");
      return;
    }
    try {
      const resp = await createMutation.mutateAsync({
        app_name: APP_NAME,
        user_id: userId,
        state: { agent_name: agentName },
      });
      setPickerOpen(false);
      setPreferredSessionId(resp.session.session_id);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create chat");
    }
  }

  function handleDeleteConfirm() {
    if (!deleteTarget) return;
    deleteMutation.mutate(
      {
        app_name: deleteTarget.app_name,
        user_id: deleteTarget.user_id,
        session_id: deleteTarget.session_id,
      },
      {
        onSuccess: () => {
          toast.success("Chat deleted");
          if (preferredSessionId === deleteTarget.session_id) {
            setPreferredSessionId(null);
          }
          setDeleteTarget(null);
        },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  if (!userId) {
    return (
      <p className="text-sm text-muted-foreground">
        Sign-in required to use chat.
      </p>
    );
  }

  return (
    <div className="-m-6 flex h-[calc(100vh-3.5rem)]">
      <ChatSidebar
        sessions={sessions}
        isLoading={sessionsQuery.isLoading}
        activeSessionId={activeSessionId}
        onSelect={setPreferredSessionId}
        onNewChat={() => setPickerOpen(true)}
        onDelete={(s) => setDeleteTarget(s)}
      />
      <ChatWindow session={activeSession} userId={userId} agentName={activeAgent} />

      <AgentPicker
        open={pickerOpen}
        onOpenChange={setPickerOpen}
        onConfirm={handleCreate}
        busy={createMutation.isPending}
      />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete chat"
        description={`Delete chat "${deleteTarget?.session_id}"? This cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={handleDeleteConfirm}
      />
    </div>
  );
}
