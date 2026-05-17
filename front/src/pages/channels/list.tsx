import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  useChannels,
  usePauseChannel,
  useResumeChannel,
  useRestartChannel,
  useDeleteChannel,
} from "@/api/channels";
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
import { MoreVertical, Send, MessageCircle, Pause, Play, RotateCcw, Pencil, Trash2 } from "lucide-react";
import type { AgentChannel, AgentChannelPlatform } from "@/types/api";

const PLATFORM_ICON: Record<string, typeof Send> = {
  AGENT_CHANNEL_PLATFORM_TELEGRAM: Send,
  AGENT_CHANNEL_PLATFORM_DISCORD: MessageCircle,
};

const PLATFORM_LABEL: Record<string, string> = {
  AGENT_CHANNEL_PLATFORM_TELEGRAM: "Telegram Poller",
  AGENT_CHANNEL_PLATFORM_DISCORD: "Discord Gateway",
};

function maskToken(token: string | undefined): string {
  if (!token) return "—";
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
    <Card>
      <CardHeader className="flex flex-row items-center justify-between gap-2 pb-2">
        <div className="flex items-center gap-2">
          <Icon className="h-4 w-4 text-muted-foreground" />
          <CardTitle className="text-base">{label}</CardTitle>
          {enabled ? (
            <Badge className="bg-green-500/10 text-green-600">Live</Badge>
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
      <CardContent className="space-y-3">
        <div className="text-xs text-muted-foreground">ID: {channel.name}</div>
        <div>
          <div className="text-xs text-muted-foreground">Bot Token</div>
          <div className="font-mono text-sm">{maskToken(botToken)}</div>
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
        <div className="flex gap-2 pt-2">
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
  const { data, isLoading } = useChannels();
  const channels = data?.channels ?? [];

  return (
    <div className="space-y-6">
      <PageHeader title="Channels" createLabel="Add Channel" createTo="/channels/create" />

      <p className="text-sm text-muted-foreground">
        Manage external entry points and communication bridges for your agents.
      </p>

      {isLoading ? (
        <div className="grid grid-cols-2 gap-4">
          {Array.from({ length: 2 }).map((_, i) => <Skeleton key={i} className="h-64" />)}
        </div>
      ) : channels.length === 0 ? (
        <div className="rounded-lg border border-dashed p-12 text-center text-muted-foreground">
          No channels configured yet.
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {channels.map((c) => (
            <ChannelCard key={c.name} channel={c} />
          ))}
        </div>
      )}
    </div>
  );
}
