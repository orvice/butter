import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  SessionService,
  type SessionDetail as PbSessionDetail,
  type SessionEvent as PbSessionEvent,
  type SessionInfo as PbSessionInfo,
} from "@/gen/agents/v1/agent_service_pb";
import type { SessionDetail, SessionEvent, SessionInfo } from "@/types/api";
import { replySession } from "./chat";
import { durationToString, tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(SessionService);

function infoFromProto(s: PbSessionInfo): SessionInfo {
  return {
    session_id: s.sessionId,
    app_name: s.appName,
    user_id: s.userId,
    state: s.state,
    last_update_time: tsToISO(s.lastUpdateTime),
    turn_count: s.turnCount,
  };
}

function eventFromProto(e: PbSessionEvent): SessionEvent {
  return {
    event_id: e.eventId,
    invocation_id: e.invocationId,
    author: e.author,
    branch: e.branch,
    content_json: e.contentJson,
    timestamp: tsToISO(e.timestamp),
    trace_id: e.traceId,
    trace_url: e.traceUrl,
  };
}

function detailFromProto(d: PbSessionDetail | undefined): SessionDetail | undefined {
  if (!d || !d.session) return undefined;
  return {
    session: infoFromProto(d.session),
    events: d.events.map(eventFromProto),
    duration: durationToString(d.duration),
  };
}

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

function parseTimestamp(s: string | undefined) {
  if (!s) return undefined;
  const date = new Date(s);
  if (Number.isNaN(date.getTime())) return undefined;
  const millis = date.getTime();
  return {
    seconds: BigInt(Math.trunc(millis / 1000)),
    nanos: (millis % 1000) * 1_000_000,
  };
}

async function listSessions(params: ListSessionsParams): Promise<ListSessionsResponse> {
  const startTs = parseTimestamp(params.start_time);
  const endTs = parseTimestamp(params.end_time);
  const res = await client.listSessions({
    appName: params.app_name ?? "",
    userId: params.user_id ?? "",
    startTime: startTs
      ? { $typeName: "google.protobuf.Timestamp", seconds: startTs.seconds, nanos: startTs.nanos }
      : undefined,
    endTime: endTs
      ? { $typeName: "google.protobuf.Timestamp", seconds: endTs.seconds, nanos: endTs.nanos }
      : undefined,
    pageSize: params.page_size ?? 0,
    pageToken: params.page_token ?? "",
  });
  return {
    sessions: res.sessions.map(infoFromProto),
    next_page_token: res.nextPageToken,
    total: res.total,
  };
}

async function getSession(params: GetSessionParams): Promise<{ session_detail: SessionDetail }> {
  const res = await client.getSession({
    appName: params.app_name,
    userId: params.user_id,
    sessionId: params.session_id,
    numRecentEvents: params.num_recent_events ?? 0,
  });
  const detail = detailFromProto(res.sessionDetail);
  if (!detail) throw new Error("session not found");
  return { session_detail: detail };
}

async function deleteSession(params: DeleteSessionParams): Promise<void> {
  await client.deleteSession({
    appName: params.app_name,
    userId: params.user_id,
    sessionId: params.session_id,
  });
}

async function createSession(params: CreateSessionParams): Promise<{ session: SessionInfo }> {
  const res = await client.createSession({
    appName: params.app_name,
    userId: params.user_id,
    sessionId: params.session_id ?? "",
    state: (params.state ?? {}) as Record<string, unknown> as never,
  });
  if (!res.session) throw new Error("create returned nothing");
  return { session: infoFromProto(res.session) };
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
