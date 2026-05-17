import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { SessionInfo, SessionDetail } from "@/types/api";

const SVC = "agents.v1.SessionService";

interface ListSessionsParams {
  app_name?: string;
  user_id?: string;
  start_time?: string;
  end_time?: string;
  page_size?: number;
  page_token?: string;
}

interface ListSessionsResponse {
  sessions?: SessionInfo[];
  next_page_token?: string;
  total?: number;
}

interface GetSessionParams {
  app_name: string;
  user_id: string;
  session_id: string;
  num_recent_events?: number;
}

interface DeleteSessionParams {
  app_name: string;
  user_id: string;
  session_id: string;
}

function listSessions(params: ListSessionsParams) {
  return twirpFetch<ListSessionsParams, ListSessionsResponse>(SVC, "ListSessions", params);
}

function getSession(params: GetSessionParams) {
  return twirpFetch<GetSessionParams, { session_detail: SessionDetail }>(
    SVC,
    "GetSession",
    params,
  );
}

function deleteSession(params: DeleteSessionParams) {
  return twirpFetch<DeleteSessionParams, object>(SVC, "DeleteSession", params);
}

export function useSessions(params: ListSessionsParams = {}) {
  return useQuery({
    queryKey: ["sessions", params],
    queryFn: () => listSessions(params),
  });
}

export function useSession(appName: string, userId: string, sessionId: string) {
  return useQuery({
    queryKey: ["sessions", { appName, userId, sessionId }],
    queryFn: () => getSession({ app_name: appName, user_id: userId, session_id: sessionId }),
    enabled: !!appName && !!userId && !!sessionId,
  });
}

export function useDeleteSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteSession,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["sessions"] });
    },
  });
}
