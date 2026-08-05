import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import { useAuth } from "@/hooks/use-auth";
import { useCreateSession, useDeleteSession, useSessions } from "@/api/sessions";
import { DeleteDialog } from "@/components/delete-dialog";
import { AgentSelector } from "./agent-selector";
import { ChatWindow } from "./chat-window";
import { sessionAgentName, sessionTitle } from "@/lib/session-title";
import type { SessionInfo } from "@/types/api";

const APP_NAME = "web-chat";

export default function ChatPage() {
  const { user, isAuthenticated, isLoading: isAuthLoading } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const userId = user?.id ?? "";

  const sessionsQuery = useSessions(
    {
      app_name: APP_NAME,
      user_id: userId || undefined,
      page_size: 100,
    },
    { enabled: !!userId },
  );
  const createMutation = useCreateSession();
  const deleteMutation = useDeleteSession();

  const sessions = useMemo(() => sessionsQuery.data?.sessions ?? [], [sessionsQuery.data]);

  const [deleteTarget, setDeleteTarget] = useState<SessionInfo | null>(null);

  const wantsNewChat = searchParams.get("new") === "1";
  const requestedSessionId = searchParams.get("session");
  const requestedAgent = searchParams.get("agent");

  // Quick-start links (/chat?new=1&agent=x) create the session immediately.
  // The guard ref makes this fire once; the redirect replaces the URL so a
  // refresh lands on the created session instead of creating another one.
  const autoCreatedRef = useRef(false);
  useEffect(() => {
    if (!wantsNewChat || !requestedAgent || !userId || autoCreatedRef.current) return;
    autoCreatedRef.current = true;
    void handleCreate(requestedAgent);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wantsNewChat, requestedAgent, userId]);

  const activeSession = useMemo(() => {
    if (wantsNewChat) return null;
    if (requestedSessionId) {
      return sessions.find((s) => s.session_id === requestedSessionId) ?? null;
    }
    return sessions[0] ?? null;
  }, [wantsNewChat, requestedSessionId, sessions]);

  const activeAgent = activeSession ? sessionAgentName(activeSession.state) ?? null : null;

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
      navigate(`/chat?session=${encodeURIComponent(resp.session.session_id)}`, { replace: true });
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
          setDeleteTarget(null);
          navigate("/chat?new=1", { replace: true });
        },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  if (!userId) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-sm text-muted-foreground">
          {isAuthenticated || isAuthLoading ? "Loading chat…" : "Sign-in required to use chat."}
        </p>
      </div>
    );
  }

  // A specific session was requested but the list is still loading — avoid
  // flashing the agent selector before we know whether it exists.
  if (!activeSession && (sessionsQuery.isLoading || (wantsNewChat && requestedAgent))) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-sm text-muted-foreground">Loading chat…</p>
      </div>
    );
  }

  // New-chat / empty state — centered agent selector
  if (!activeSession) {
    return (
      <div className="flex h-full flex-col">
        <div className="scrollbar-thin flex flex-1 items-center justify-center overflow-y-auto p-4">
          <AgentSelector onPick={(name) => void handleCreate(name)} busy={createMutation.isPending} />
        </div>
      </div>
    );
  }

  return (
    <>
      <ChatWindow
        session={activeSession}
        userId={userId}
        agentName={activeAgent}
        onDelete={() => setDeleteTarget(activeSession)}
      />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete chat"
        description={`Delete chat "${deleteTarget ? sessionTitle(deleteTarget) : ""}"? This cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}
