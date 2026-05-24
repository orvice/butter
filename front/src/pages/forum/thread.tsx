import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { toast } from "sonner";
import { Bot, Loader2, Send, User } from "lucide-react";
import { useAgents } from "@/api/agents";
import { useCreateForumPost, useForumThread, useInvokeAgentInThread } from "@/api/forum";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
  const { data, isLoading } = useForumThread(id);
  const { data: agentsData } = useAgents({ page_size: 200 });
  const createPost = useCreateForumPost();
  const invokeAgent = useInvokeAgentInThread();
  const [body, setBody] = useState("");
  const [agentMessage, setAgentMessage] = useState("");
  const [agentName, setAgentName] = useState("");

  const thread = data?.thread;
  const posts = data?.posts ?? [];
  const agents = useMemo(() => agentsData?.agents ?? [], [agentsData?.agents]);
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
                    {post.invocation_id ? <Badge variant="outline" className="font-mono text-[10px]">{post.invocation_id.slice(0, 8)}</Badge> : null}
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
    </>
  );
}
