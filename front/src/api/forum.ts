import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ForumService,
  type ForumPost as PbForumPost,
  type ForumThread as PbForumThread,
} from "@/gen/agents/v1/forum_pb";
import type { ForumPost, ForumThread } from "@/types/api";
import { tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(ForumService);

function threadFromProto(t: PbForumThread): ForumThread {
  return {
    id: t.id,
    title: t.title,
    body: t.body,
    created_by: t.createdBy,
    status: t.status,
    agent_names: t.agentNames,
    labels: t.labels,
    metadata: t.metadata,
    created_at: tsToISO(t.createdAt),
    updated_at: tsToISO(t.updatedAt),
    workspace_id: t.workspaceId,
  };
}

function postFromProto(p: PbForumPost): ForumPost {
  return {
    id: p.id,
    thread_id: p.threadId,
    body: p.body,
    author_user_id: p.authorUserId,
    author_agent_name: p.authorAgentName,
    author_kind: p.authorKind,
    invocation_id: p.invocationId,
    parent_post_id: p.parentPostId,
    created_at: tsToISO(p.createdAt),
    updated_at: tsToISO(p.updatedAt),
    workspace_id: p.workspaceId,
  };
}

interface ListThreadsParams {
  status?: string;
  label?: string;
  page_size?: number;
  page_token?: string;
}

interface ListThreadsResponse {
  threads?: ForumThread[];
  next_page_token?: string;
  total?: number;
}

interface ListThreadLabelsResponse {
  labels?: string[];
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
  labels?: string[];
  metadata?: Record<string, string>;
}

interface CreateThreadResponse {
  thread: ForumThread;
  first_post?: ForumPost;
}

export interface UpdateThreadParams {
  id: string;
  title?: string;
  body?: string;
  status?: string;
  agent_names?: string[];
  labels?: string[];
  metadata?: Record<string, string>;
}

interface UpdateThreadResponse {
  thread: ForumThread;
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

export async function listForumThreads(params: ListThreadsParams = {}): Promise<ListThreadsResponse> {
  const res = await client.listThreads({
    status: params.status ?? "",
    label: params.label ?? "",
    pageSize: params.page_size ?? 0,
    pageToken: params.page_token ?? "",
  });
  return {
    threads: res.threads.map(threadFromProto),
    next_page_token: res.nextPageToken,
    total: res.total,
  };
}

export async function listForumThreadLabels(): Promise<ListThreadLabelsResponse> {
  const res = await client.listThreadLabels({});
  return { labels: res.labels };
}

export async function getForumThread(id: string): Promise<GetThreadResponse> {
  const res = await client.getThread({ id, postPageSize: 200, postPageToken: "" });
  return {
    thread: res.thread ? threadFromProto(res.thread) : undefined,
    posts: res.posts.map(postFromProto),
    next_post_page_token: res.nextPostPageToken,
    post_total: res.postTotal,
  };
}

export async function createForumThread(params: CreateThreadParams): Promise<CreateThreadResponse> {
  const res = await client.createThread({
    title: params.title,
    body: params.body,
    agentNames: params.agent_names ?? [],
    labels: params.labels ?? [],
    metadata: params.metadata ?? {},
  });
  if (!res.thread) throw new Error("create returned no thread");
  return {
    thread: threadFromProto(res.thread),
    first_post: res.firstPost ? postFromProto(res.firstPost) : undefined,
  };
}

export async function updateForumThread(params: UpdateThreadParams): Promise<UpdateThreadResponse> {
  const res = await client.updateThread({
    id: params.id,
    title: params.title ?? "",
    body: params.body ?? "",
    status: params.status ?? "",
    agentNames: params.agent_names ?? [],
    labels: params.labels ?? [],
    metadata: params.metadata ?? {},
  });
  if (!res.thread) throw new Error("update returned no thread");
  return { thread: threadFromProto(res.thread) };
}

export async function createForumPost(params: CreatePostParams): Promise<CreatePostResponse> {
  const res = await client.createPost({
    threadId: params.thread_id,
    body: params.body,
    parentPostId: params.parent_post_id ?? "",
  });
  if (!res.post) throw new Error("create returned no post");
  return { post: postFromProto(res.post) };
}

export async function invokeAgentInThread(params: InvokeAgentParams): Promise<InvokeAgentResponse> {
  const res = await client.invokeAgentInThread({
    threadId: params.thread_id,
    agentName: params.agent_name,
    message: params.message ?? "",
    modelOverride: params.model_override ?? "",
    recentPostLimit: params.recent_post_limit ?? 0,
  });
  if (!res.post) throw new Error("invoke returned no post");
  return { post: postFromProto(res.post), response: res.response };
}

export async function deleteForumThread(id: string): Promise<void> {
  await client.deleteThread({ id });
}

export async function deleteForumPost(threadId: string, postId: string): Promise<void> {
  await client.deletePost({ threadId, postId });
}

export function useForumThreads(params: ListThreadsParams = {}) {
  return useQuery({
    queryKey: ["forum", "threads", params],
    queryFn: () => listForumThreads(params),
  });
}

export function useForumThreadLabels() {
  return useQuery({
    queryKey: ["forum", "labels"],
    queryFn: listForumThreadLabels,
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
      qc.invalidateQueries({ queryKey: ["forum", "labels"] });
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

export function useUpdateForumThread() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateForumThread,
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: ["forum", "threads"] });
      qc.invalidateQueries({ queryKey: ["forum", "labels"] });
      if (data.thread?.id) qc.invalidateQueries({ queryKey: ["forum", "thread", data.thread.id] });
    },
  });
}

export function useDeleteForumThread() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteForumThread,
    onSuccess: (_data, id) => {
      qc.invalidateQueries({ queryKey: ["forum", "threads"] });
      qc.invalidateQueries({ queryKey: ["forum", "labels"] });
      qc.removeQueries({ queryKey: ["forum", "thread", id] });
    },
  });
}

export function useDeleteForumPost() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ threadId, postId }: { threadId: string; postId: string }) =>
      deleteForumPost(threadId, postId),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ["forum", "threads"] });
      qc.invalidateQueries({ queryKey: ["forum", "thread", vars.threadId] });
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
