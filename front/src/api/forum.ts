import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { ForumPost, ForumThread } from "@/types/api";

const SVC = "agents.v1.ForumService";

interface ListThreadsParams {
  status?: string;
  page_size?: number;
  page_token?: string;
}

interface ListThreadsResponse {
  threads?: ForumThread[];
  next_page_token?: string;
  total?: number;
}

interface GetThreadResponse {
  thread?: ForumThread;
  posts?: ForumPost[];
  next_post_page_token?: string;
  post_total?: number;
}

interface CreateThreadParams {
  title: string;
  body: string;
  agent_names?: string[];
  metadata?: Record<string, string>;
}

interface CreateThreadResponse {
  thread: ForumThread;
  first_post?: ForumPost;
}

interface CreatePostParams {
  thread_id: string;
  body: string;
  parent_post_id?: string;
}

interface CreatePostResponse {
  post: ForumPost;
}

interface InvokeAgentParams {
  thread_id: string;
  agent_name: string;
  message?: string;
  model_override?: string;
  recent_post_limit?: number;
}

interface InvokeAgentResponse {
  post: ForumPost;
  response?: string;
}

export function listForumThreads(params: ListThreadsParams = {}) {
  return twirpFetch<ListThreadsParams, ListThreadsResponse>(SVC, "ListThreads", params);
}

export function getForumThread(id: string) {
  return twirpFetch<{ id: string; post_page_size: number }, GetThreadResponse>(SVC, "GetThread", {
    id,
    post_page_size: 200,
  });
}

export function createForumThread(params: CreateThreadParams) {
  return twirpFetch<CreateThreadParams, CreateThreadResponse>(SVC, "CreateThread", params);
}

export function createForumPost(params: CreatePostParams) {
  return twirpFetch<CreatePostParams, CreatePostResponse>(SVC, "CreatePost", params);
}

export function invokeAgentInThread(params: InvokeAgentParams) {
  return twirpFetch<InvokeAgentParams, InvokeAgentResponse>(SVC, "InvokeAgentInThread", params);
}

export function deleteForumThread(id: string) {
  return twirpFetch<{ id: string }, object>(SVC, "DeleteThread", { id });
}

export function deleteForumPost(threadId: string, postId: string) {
  return twirpFetch<{ thread_id: string; post_id: string }, object>(SVC, "DeletePost", {
    thread_id: threadId,
    post_id: postId,
  });
}

export function useForumThreads(params: ListThreadsParams = {}) {
  return useQuery({
    queryKey: ["forum", "threads", params],
    queryFn: () => listForumThreads(params),
  });
}

export function useForumThread(id: string) {
  return useQuery({
    queryKey: ["forum", "thread", id],
    queryFn: () => getForumThread(id),
    enabled: !!id,
    refetchInterval: (query) => (query.state.data?.thread?.status === "processing" ? 3000 : false),
  });
}

export function useCreateForumThread() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createForumThread,
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: ["forum", "threads"] });
      if (data.thread?.id) qc.invalidateQueries({ queryKey: ["forum", "thread", data.thread.id] });
    },
  });
}

export function useCreateForumPost() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createForumPost,
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ["forum", "threads"] });
      qc.invalidateQueries({ queryKey: ["forum", "thread", vars.thread_id] });
    },
  });
}

export function useInvokeAgentInThread() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: invokeAgentInThread,
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ["forum", "threads"] });
      qc.invalidateQueries({ queryKey: ["forum", "thread", vars.thread_id] });
    },
  });
}
