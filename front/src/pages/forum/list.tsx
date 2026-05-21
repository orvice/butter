import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { MessageSquarePlus, Bot, Clock } from "lucide-react";
import { useForumThreads, useCreateForumThread } from "@/api/forum";
import { useAgents } from "@/api/agents";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

function fmtDate(v?: string) {
  return v ? new Date(v).toLocaleString() : "—";
}

export default function ForumListPage() {
  const navigate = useNavigate();
  const { data, isLoading } = useForumThreads({ page_size: 50 });
  const { data: agentsData } = useAgents({ page_size: 200 });
  const createMutation = useCreateForumThread();
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [agentName, setAgentName] = useState("");

  async function handleCreate() {
    const cleanTitle = title.trim();
    const cleanBody = body.trim();
    if (!cleanTitle || !cleanBody) {
      toast.error("Title and body are required");
      return;
    }
    try {
      const resp = await createMutation.mutateAsync({
        title: cleanTitle,
        body: cleanBody,
        agent_names: agentName ? [agentName] : [],
      });
      toast.success("Thread created");
      setOpen(false);
      setTitle("");
      setBody("");
      setAgentName("");
      navigate(`/forum/${resp.thread.id}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create thread");
    }
  }

  const threads = data?.threads ?? [];
  const agents = agentsData?.agents ?? [];

  return (
    <>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageHeader
          title="Forum"
          description="Shared threads where users and agents discuss together."
        />
        <Button onClick={() => setOpen(true)} className="w-full sm:w-auto">
          <MessageSquarePlus className="mr-2 h-4 w-4" /> New Thread
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          <Skeleton className="h-28 w-full" />
          <Skeleton className="h-28 w-full" />
        </div>
      ) : threads.length === 0 ? (
        <Card>
          <CardContent className="p-8 text-center text-sm text-muted-foreground">
            No forum threads yet. Create one to start a shared discussion.
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {threads.map((thread) => (
            <Link key={thread.id} to={`/forum/${thread.id}`}>
              <Card className="transition-colors hover:bg-muted/40">
                <CardHeader className="pb-2">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <CardTitle className="text-lg">{thread.title}</CardTitle>
                    <Badge variant={thread.status === "open" ? "default" : "secondary"}>{thread.status || "open"}</Badge>
                  </div>
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="line-clamp-2 text-sm text-muted-foreground">{thread.body}</p>
                  <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                    <span className="inline-flex items-center gap-1">
                      <Clock className="h-3.5 w-3.5" /> {fmtDate(thread.updated_at)}
                    </span>
                    {thread.agent_names?.length ? (
                      <span className="inline-flex items-center gap-1">
                        <Bot className="h-3.5 w-3.5" /> {thread.agent_names.join(", ")}
                      </span>
                    ) : null}
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Forum Thread</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="forum-title">Title</Label>
              <Input id="forum-title" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="What should we discuss?" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="forum-body">Body</Label>
              <Textarea id="forum-body" value={body} onChange={(e) => setBody(e.target.value)} rows={6} placeholder="Write the opening post..." />
            </div>
            <div className="space-y-2">
              <Label>Default agent</Label>
              <Select value={agentName || "none"} onValueChange={(v) => setAgentName(v && v !== "none" ? v : "")}>
                <SelectTrigger>
                  <SelectValue placeholder="Optional agent" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">No default agent</SelectItem>
                  {agents.map((agent) => (
                    <SelectItem key={agent.name} value={agent.name}>{agent.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)} disabled={createMutation.isPending}>Cancel</Button>
            <Button onClick={handleCreate} disabled={createMutation.isPending}>{createMutation.isPending ? "Creating..." : "Create"}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
