import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  useChannels,
  usePauseChannel,
  useResumeChannel,
  useRestartChannel,
  useDeleteChannel,
} from "@/api/channels";
import { useAPITokens } from "@/api/apitokens";
import { PageHeader } from "@/components/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { MoreVertical, Send, MessageCircle, Pause, Play, RotateCcw, Pencil, Trash2, Copy, KeyRound, Workflow } from "lucide-react";
import type { AgentChannel, AgentChannelPlatform } from "@/types/api";

const PLATFORM_ICON: Record<string, typeof Send> = {
  AGENT_CHANNEL_PLATFORM_TELEGRAM: Send,
  AGENT_CHANNEL_PLATFORM_DISCORD: MessageCircle,
};

const PLATFORM_LABEL: Record<string, string> = {
  AGENT_CHANNEL_PLATFORM_TELEGRAM: "Telegram Poller",
  AGENT_CHANNEL_PLATFORM_DISCORD: "Discord Gateway",
};

async function copyText(text: string, message: string) {
  try {
    await navigator.clipboard.writeText(text);
    toast.success(message);
  } catch {
    toast.error("Copy failed");
  }
}

function maskToken(token: string | undefined): string {
  if (!token) return "-";
  return "•".repeat(Math.min(token.length, 32));
}

function ChannelCard({ channel }: { channel: AgentChannel }) {
  const navigate = useNavigate();
  const pause = usePauseChannel();
  const resume = useResumeChannel();
  const restart = useRestartChannel();
  const del = useDeleteChannel();

  const platform = (channel.platform ?? "AGENT_CHANNEL_PLATFORM_UNSPECIFIED") as AgentChannelPlatform;
  const Icon = PLATFORM_ICON[platform] ?? MessageCircle;
  const label = PLATFORM_LABEL[platform] ?? channel.name;
  const enabled = channel.enabled === true;

  const triggers = channel.triggers ?? [];
  const triggerKeywords = triggers.flatMap((t) => [...(t.commands ?? []), ...(t.prefixes ?? [])]);

  const botToken =
    platform === "AGENT_CHANNEL_PLATFORM_TELEGRAM"
      ? channel.telegram?.bot_token
      : channel.discord?.bot_token;

  return (
    <Card className="relative">
      <div className={`absolute left-0 top-0 h-full w-1 ${platform === "AGENT_CHANNEL_PLATFORM_TELEGRAM" ? "bg-sky-500" : "bg-indigo-500"}`} />
      <CardHeader className="flex flex-row items-start justify-between gap-2 pb-2">
        <div className="flex min-w-0 flex-wrap items-center gap-2">
          <Icon className="h-4 w-4 text-muted-foreground" />
          <CardTitle className="text-base">{label}</CardTitle>
          {enabled ? (
            <Badge className="bg-emerald-500/10 text-emerald-700">
              <span className="h-1.5 w-1.5 rounded-full bg-current" />
              Live
            </Badge>
          ) : (
            <Badge variant="outline">Paused</Badge>
          )}
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreVertical className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/channels/${encodeURIComponent(channel.name)}/edit`)}>
              <Pencil className="mr-2 h-4 w-4" /> Edit settings
            </DropdownMenuItem>
            <DropdownMenuItem
              className="text-destructive"
              onClick={() => del.mutate(channel.name, { onSuccess: () => toast.success("Channel deleted") })}
            >
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="break-all text-xs text-muted-foreground">ID: {channel.name}</div>
        <div>
          <div className="text-xs text-muted-foreground">Bot Token</div>
          <div className="flex items-center justify-between rounded-md bg-muted px-3 py-2">
            <span className="break-all font-mono text-sm">{maskToken(botToken)}</span>
            <Button
              size="icon-xs"
              variant="ghost"
              aria-label="Copy token"
              disabled={!botToken}
              onClick={() => botToken && void copyText(botToken, "Bot token copied")}
            >
              <Copy className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
        {triggerKeywords.length > 0 && (
          <div>
            <div className="text-xs text-muted-foreground">Trigger Keywords</div>
            <div className="mt-1 flex flex-wrap gap-1">
              {triggerKeywords.map((k) => (
                <Badge key={k} variant="secondary">{k}</Badge>
              ))}
            </div>
          </div>
        )}
        <div>
          <div className="text-xs text-muted-foreground">Agent</div>
          <div className="text-sm font-medium">{channel.agent_name}</div>
        </div>
        <div className="flex flex-col gap-2 pt-2 sm:flex-row">
          {enabled ? (
            <Button
              size="sm"
              variant="outline"
              onClick={() =>
                pause.mutate(channel.name, { onSuccess: () => toast.success("Paused") })
              }
            >
              <Pause className="mr-1 h-3 w-3" /> Pause
            </Button>
          ) : (
            <Button
              size="sm"
              onClick={() =>
                resume.mutate(channel.name, { onSuccess: () => toast.success("Resumed") })
              }
            >
              <Play className="mr-1 h-3 w-3" /> Resume
            </Button>
          )}
          <Button
            size="sm"
            variant="outline"
            onClick={() => restart.mutate(channel.name, { onSuccess: () => toast.success("Restart triggered") })}
          >
            <RotateCcw className="mr-1 h-3 w-3" /> Restart
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

export default function ChannelListPage() {
  const navigate = useNavigate();
  const { data, isLoading } = useChannels();
  const { data: tokenData } = useAPITokens();
  const channels = data?.channels ?? [];
  const tokens = tokenData?.tokens ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Channel Configuration"
        description="Manage external entry points, communication bridges, API execution, and webhook authentication."
        createLabel="Add Channel"
        createTo="/channels/create"
      />

      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {Array.from({ length: 2 }).map((_, i) => <Skeleton key={i} className="h-64" />)}
        </div>
      ) : channels.length === 0 ? (
        <div className="rounded-lg border border-dashed p-6 text-center text-muted-foreground sm:p-12">
          No channels configured yet.
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {channels.map((c) => (
            <ChannelCard key={c.name} channel={c} />
          ))}
        </div>
      )}

      <Card>
        <CardHeader className="border-b pb-4">
          <CardTitle className="flex items-center gap-2">
            <Workflow className="h-4 w-4 text-primary" />
            API & Webhooks
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6 pt-5">
          <div className="rounded-lg border p-4">
            <div className="mb-4 flex items-start justify-between gap-3">
              <div>
                <h3 className="text-lg font-semibold">Twirp/HTTP API</h3>
                <p className="text-sm text-muted-foreground">Synchronous agent execution endpoint.</p>
              </div>
              <Badge className="bg-emerald-500/10 text-emerald-700">
                <span className="h-1.5 w-1.5 rounded-full bg-current" />
                Live
              </Badge>
            </div>
            <div className="relative rounded-lg bg-[#111827] p-4 font-mono text-[13px] text-gray-300">
              <Button
                size="icon-sm"
                variant="ghost"
                className="absolute right-3 top-3 text-gray-400 hover:bg-gray-800 hover:text-white"
                aria-label="Copy endpoint"
                onClick={() => void copyText("/api/agents.v1.AgentService/InvokeAgent", "Endpoint copied")}
              >
                <Copy className="h-4 w-4" />
              </Button>
              <span className="mr-4 text-gray-500">1</span>
              <span className="text-emerald-400">POST</span>
              <span className="ml-2 text-white">/api/agents.v1.AgentService/InvokeAgent</span>
            </div>
          </div>

          <div>
            <div className="mb-4 flex items-center justify-between gap-3">
              <div>
                <h3 className="text-lg font-semibold">APITokenAuthMiddleware</h3>
                <p className="text-sm text-muted-foreground">Bearer tokens for external webhook ingestion.</p>
              </div>
              <Button variant="outline" onClick={() => navigate("/api-tokens")}>
                <KeyRound className="mr-2 h-4 w-4" />
                Manage Tokens
              </Button>
            </div>
            {tokens.length === 0 ? (
              <div className="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">
                No API tokens created yet.
              </div>
            ) : (
              <div className="overflow-hidden rounded-lg border">
                <div className="grid grid-cols-[1.2fr_0.8fr_0.8fr] bg-muted/60 px-4 py-3 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground">
                  <span>Token Name</span>
                  <span>Prefix</span>
                  <span>Created</span>
                </div>
                {tokens.slice(0, 4).map((token) => (
                  <div key={token.id} className="grid grid-cols-[1.2fr_0.8fr_0.8fr] border-t px-4 py-3 text-sm">
                    <span className="font-medium">{token.name}</span>
                    <code className="text-muted-foreground">{token.prefix || "-"}</code>
                    <span className="text-muted-foreground">{token.created_at ? new Date(token.created_at).toLocaleDateString() : "-"}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
