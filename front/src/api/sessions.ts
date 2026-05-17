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

interface CreateSessionParams {
  app_name: string;
  user_id: string;
  session_id?: string;
  state?: Record<string, unknown>;
}

interface ReplySessionParams {
  agent_name: string;
  app_name: string;
  user_id: string;
  session_id: string;
  message: string;
  model_override?: string;
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

function createSession(params: CreateSessionParams) {
  return twirpFetch<CreateSessionParams, { session: SessionInfo }>(SVC, "CreateSession", params);
}

function replySession(params: ReplySessionParams) {
  return twirpFetch<ReplySessionParams, { response: string }>(SVC, "ReplySession", params);
}

export function useSessions(params: ListSessionsParams = {}, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ["sessions", params],
    queryFn: () => listSessions(params),
    enabled: options?.enabled ?? true,
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

export function useCreateSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createSession,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["sessions"] });
    },
  });
}

export function useReplySession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: replySession,
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ["sessions"] });
      qc.invalidateQueries({
        queryKey: [
          "sessions",
          { appName: vars.app_name, userId: vars.user_id, sessionId: vars.session_id },
        ],
      });
    },
  });
}

export function useLiveSession(
  appName: string,
  userId: string,
  sessionId: string,
  enabled: boolean,
  pollIntervalMs = 1500,
) {
  return useQuery({
    queryKey: ["sessions", { appName, userId, sessionId }],
    queryFn: () => getSession({ app_name: appName, user_id: userId, session_id: sessionId }),
    enabled: !!appName && !!userId && !!sessionId,
    refetchInterval: enabled ? pollIntervalMs : false,
  });
}
