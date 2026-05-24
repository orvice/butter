import { useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import { Bot, Loader2, Pencil, Send, Trash2, User } from "lucide-react";
import { useAgents } from "@/api/agents";
import {
  useCreateForumPost,
  useDeleteForumPost,
  useDeleteForumThread,
  useForumThread,
  useInvokeAgentInThread,
  useUpdateForumThread,
} from "@/api/forum";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DeleteDialog } from "@/components/delete-dialog";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import type { ForumPost } from "@/types/api";
import { MarkdownContent } from "@/components/markdown-content";

function fmtDate(v?: string) {
  return v ? new Date(v).toLocaleString() : "";
}

function authorLabel(post: ForumPost) {
  if (post.author_kind === "system") return "system";
  if (post.author_kind === "agent" || post.author_agent_name) return post.author_agent_name || "agent";
  return post.author_user_id || "user";
}

export default function ForumThreadPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const { data, isLoading } = useForumThread(id);
  const { data: agentsData } = useAgents({ page_size: 200 });
  const createPost = useCreateForumPost();
  const updateThread = useUpdateForumThread();
  const deleteThread = useDeleteForumThread();
  const deletePost = useDeleteForumPost();
  const invokeAgent = useInvokeAgentInThread();
  const [body, setBody] = useState("");
  const [agentMessage, setAgentMessage] = useState("");
  const [agentName, setAgentName] = useState("");
  const [editOpen, setEditOpen] = useState(false);
  const [editTitle, setEditTitle] = useState("");
  const [editBody, setEditBody] = useState("");
  const [deleteThreadOpen, setDeleteThreadOpen] = useState(false);
  const [deletePostTarget, setDeletePostTarget] = useState<ForumPost | null>(null);

  const thread = data?.thread;
  const posts = data?.posts ?? [];
  const agents = agentsData?.agents ?? [];
  const defaultAgent = useMemo(() => thread?.agent_names?.[0] || agents[0]?.name || "", [thread, agents]);
  const selectedAgent = agentName || defaultAgent;
  const isProcessing = thread?.status === "processing";
  const processingAgent = thread?.metadata?.processing_agent;

  async function handlePost() {
    const clean = body.trim();
    if (!clean) return;
    try {
      await createPost.mutateAsync({ thread_id: id, body: clean });
      setBody("");
      toast.success("Posted");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to post");
    }
  }

  async function handleAskAgent() {
    if (!selectedAgent) {
      toast.error("Select an agent first");
      return;
    }
    try {
      await invokeAgent.mutateAsync({
        thread_id: id,
        agent_name: selectedAgent,
        message: agentMessage.trim(),
        recent_post_limit: 30,
      });
      setAgentMessage("");
      toast.success(`${selectedAgent} is processing this thread`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to queue agent");
    }
  }

  function openEditDialog() {
    if (!thread) return;
    setEditTitle(thread.title ?? "");
    setEditBody(thread.body ?? "");
    setEditOpen(true);
  }

  async function handleUpdateThread() {
    if (!thread) return;
    const cleanTitle = editTitle.trim();
    const cleanBody = editBody.trim();
    if (!cleanTitle || !cleanBody) {
      toast.error("Title and body are required");
      return;
    }
    try {
      await updateThread.mutateAsync({
        id: thread.id,
        title: cleanTitle,
        body: cleanBody,
        status: thread.status,
        agent_names: thread.agent_names ?? [],
        metadata: thread.metadata ?? {},
      });
      setEditOpen(false);
      toast.success("Thread updated");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to update thread");
    }
  }

  async function handleDeleteThread() {
    if (!thread) return;
    try {
      await deleteThread.mutateAsync(thread.id);
      toast.success("Thread deleted");
      navigate("/forum");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to delete thread");
    }
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;
  if (!thread) return <p className="text-sm text-muted-foreground">Thread not found.</p>;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem>
            <Link to="/forum" className="text-sm hover:underline">Forum</Link>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem className="max-w-lg truncate text-sm">{thread.title}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <div className="mb-6 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">{thread.title}</h2>
          <p className="mt-1 text-sm text-muted-foreground">Updated {fmtDate(thread.updated_at)}</p>
        </div>
        <div className="flex items-center gap-2">
          {isProcessing ? (
            <Badge variant="secondary" className="gap-1">
              <Loader2 className="h-3 w-3 animate-spin" />
              {processingAgent ? `${processingAgent} processing` : "Processing"}
            </Badge>
          ) : null}
          <Badge>{thread.status || "open"}</Badge>
          <Button size="sm" variant="outline" onClick={openEditDialog} disabled={updateThread.isPending}>
            <Pencil className="mr-1 h-3.5 w-3.5" />
            Edit
          </Button>
          <Button size="sm" variant="destructive" onClick={() => setDeleteThreadOpen(true)} disabled={deleteThread.isPending}>
            <Trash2 className="mr-1 h-3.5 w-3.5" />
            Delete
          </Button>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
        <div className="space-y-3">
          {posts.map((post) => {
            const isSystem = post.author_kind === "system";
            const isAgent = !isSystem && (post.author_kind === "agent" || !!post.author_agent_name);
            const Icon = isAgent || isSystem ? Bot : User;
            return (
              <Card key={post.id} className={isAgent ? "border-primary/30 bg-primary/5" : isSystem ? "border-destructive/30 bg-destructive/5" : undefined}>
                <CardContent className="p-4">
                  <div className="mb-3 flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2">
                      <span className="flex h-8 w-8 items-center justify-center rounded-full border bg-background">
                        <Icon className="h-4 w-4" />
                      </span>
                      <div>
                        <div className="text-sm font-medium">{authorLabel(post)}</div>
                        <div className="text-xs text-muted-foreground">{fmtDate(post.created_at)}</div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {post.invocation_id ? <Badge variant="outline" className="font-mono text-[10px]">{post.invocation_id.slice(0, 8)}</Badge> : null}
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        aria-label="Delete post"
                        disabled={deletePost.isPending}
                        onClick={() => setDeletePostTarget(post)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                  <MarkdownContent content={post.body ?? ""} />
                </CardContent>
              </Card>
            );
          })}
          {isProcessing ? (
            <Card className="border-dashed">
              <CardContent className="flex items-center gap-3 p-4 text-sm text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>{processingAgent ? `${processingAgent} is working on a reply.` : "Agent is working on a reply."}</span>
              </CardContent>
            </Card>
          ) : null}
          {posts.length === 0 ? <p className="text-sm text-muted-foreground">No posts yet.</p> : null}
        </div>

        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Add post</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Textarea value={body} onChange={(e) => setBody(e.target.value)} rows={5} placeholder="Write a reply..." />
              <Button className="w-full" onClick={handlePost} disabled={createPost.isPending || !body.trim()}>
                <Send className="mr-2 h-4 w-4" /> {createPost.isPending ? "Posting..." : "Post"}
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base"><Bot className="h-4 w-4" /> Ask agent</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Select value={selectedAgent || undefined} onValueChange={(v) => setAgentName(v ?? "")}>
                <SelectTrigger>
                  <SelectValue placeholder="Select an agent" />
                </SelectTrigger>
                <SelectContent>
                  {agents.map((agent) => (
                    <SelectItem key={agent.name} value={agent.name}>{agent.name}</SelectItem>
                  ))}
                  {agents.length === 0 ? <div className="px-3 py-2 text-xs text-muted-foreground">No agents configured.</div> : null}
                </SelectContent>
              </Select>
              <Textarea value={agentMessage} onChange={(e) => setAgentMessage(e.target.value)} rows={4} placeholder="Optional instruction for the agent..." />
              <Button className="w-full" onClick={handleAskAgent} disabled={invokeAgent.isPending || isProcessing || !selectedAgent}>
                <Bot className="mr-2 h-4 w-4" /> {invokeAgent.isPending ? "Queueing..." : isProcessing ? "Agent processing" : "Ask Agent"}
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit thread</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="thread-title">Title</Label>
              <Input id="thread-title" value={editTitle} onChange={(e) => setEditTitle(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="thread-body">Body</Label>
              <Textarea id="thread-body" value={editBody} onChange={(e) => setEditBody(e.target.value)} rows={6} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)} disabled={updateThread.isPending}>Cancel</Button>
            <Button onClick={handleUpdateThread} disabled={updateThread.isPending}>
              {updateThread.isPending ? "Saving..." : "Save"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <DeleteDialog
        open={deleteThreadOpen}
        onOpenChange={setDeleteThreadOpen}
        title="Delete thread"
        description={`Delete "${thread.title}"? This cannot be undone.`}
        loading={deleteThread.isPending}
        onConfirm={handleDeleteThread}
      />

      <DeleteDialog
        open={!!deletePostTarget}
        onOpenChange={(open) => !open && setDeletePostTarget(null)}
        title="Delete post"
        description="Delete this forum post? This cannot be undone."
        loading={deletePost.isPending}
        onConfirm={() => {
          if (!deletePostTarget) return;
          deletePost.mutate(
            { threadId: id, postId: deletePostTarget.id },
            {
              onSuccess: () => {
                toast.success("Post deleted");
                setDeletePostTarget(null);
              },
              onError: (err) => toast.error(err.message),
            },
          );
        }}
      />
    </>
  );
}
