import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import {
  AgentChannelPlatform,
  AgentChannelSchema,
  AgentTriggerSchema,
  AgentTriggerType,
  DiscordChannelConfigSchema,
  TelegramChannelConfigSchema,
  type AgentChannel as PbAgentChannel,
  type AgentTrigger as PbAgentTrigger,
  type DiscordChannelConfig as PbDiscordChannelConfig,
  type TelegramChannelConfig as PbTelegramChannelConfig,
} from "@/gen/agents/v1/agentchannel_pb";
import { ChannelService } from "@/gen/agents/v1/agent_service_pb";
import {
  ChannelStatus_State,
  type ChannelStatus as PbChannelStatus,
} from "@/gen/agents/v1/dashboard_pb";
import type {
  AgentChannel,
  AgentChannelPlatform as LegacyPlatform,
  AgentTrigger,
  AgentTriggerType as LegacyTriggerType,
  ChannelState,
  ChannelStatus,
  DiscordChannelConfig,
  TelegramChannelConfig,
} from "@/types/api";
import { tsToISO } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(ChannelService);

const PLATFORM_NAMES: LegacyPlatform[] = [
  "AGENT_CHANNEL_PLATFORM_UNSPECIFIED",
  "AGENT_CHANNEL_PLATFORM_TELEGRAM",
  "AGENT_CHANNEL_PLATFORM_DISCORD",
];

const TRIGGER_TYPE_NAMES: LegacyTriggerType[] = [
  "AGENT_TRIGGER_TYPE_UNSPECIFIED",
  "AGENT_TRIGGER_TYPE_MESSAGE",
  "AGENT_TRIGGER_TYPE_COMMAND",
  "AGENT_TRIGGER_TYPE_MENTION",
  "AGENT_TRIGGER_TYPE_PRIVATE_CHAT",
];

const CHANNEL_STATE_NAMES: ChannelState[] = [
  "STATE_UNSPECIFIED",
  "STATE_LIVE",
  "STATE_PAUSED",
  "STATE_DISABLED",
  "STATE_ERROR",
];

function platformFromProto(p: AgentChannelPlatform): LegacyPlatform {
  return PLATFORM_NAMES[p] ?? "AGENT_CHANNEL_PLATFORM_UNSPECIFIED";
}

function platformToProto(p: LegacyPlatform | undefined): AgentChannelPlatform {
  const idx = PLATFORM_NAMES.indexOf(p ?? "AGENT_CHANNEL_PLATFORM_UNSPECIFIED");
  return idx < 0 ? AgentChannelPlatform.UNSPECIFIED : (idx as AgentChannelPlatform);
}

function triggerTypeFromProto(t: AgentTriggerType): LegacyTriggerType {
  return TRIGGER_TYPE_NAMES[t] ?? "AGENT_TRIGGER_TYPE_UNSPECIFIED";
}

function triggerTypeToProto(t: LegacyTriggerType | undefined): AgentTriggerType {
  const idx = TRIGGER_TYPE_NAMES.indexOf(t ?? "AGENT_TRIGGER_TYPE_UNSPECIFIED");
  return idx < 0 ? AgentTriggerType.UNSPECIFIED : (idx as AgentTriggerType);
}

function stateFromProto(s: ChannelStatus_State): ChannelState {
  return CHANNEL_STATE_NAMES[s] ?? "STATE_UNSPECIFIED";
}

function triggerFromProto(t: PbAgentTrigger): AgentTrigger {
  return {
    type: triggerTypeFromProto(t.type),
    commands: t.commands,
    prefixes: t.prefixes,
    require_mention: t.requireMention,
  };
}

function triggerToProto(t: AgentTrigger): PbAgentTrigger {
  return create(AgentTriggerSchema, {
    type: triggerTypeToProto(t.type),
    commands: t.commands ?? [],
    prefixes: t.prefixes ?? [],
    requireMention: t.require_mention ?? false,
  });
}

function telegramFromProto(c: PbTelegramChannelConfig | undefined): TelegramChannelConfig | undefined {
  if (!c) return undefined;
  return {
    bot_token: c.botToken,
    allowed_chat_ids: c.allowedChatIds.map((v) => String(v)),
  };
}

function telegramToProto(c: TelegramChannelConfig | undefined): PbTelegramChannelConfig | undefined {
  if (!c) return undefined;
  const ids = c.allowed_chat_ids ?? c.allow_chat_ids ?? [];
  return create(TelegramChannelConfigSchema, {
    botToken: c.bot_token ?? "",
    allowedChatIds: ids.map((s) => BigInt(s)),
  });
}

function discordFromProto(c: PbDiscordChannelConfig | undefined): DiscordChannelConfig | undefined {
  if (!c) return undefined;
  return {
    bot_token: c.botToken,
    allowed_channel_ids: c.allowedChannelIds,
  };
}

function discordToProto(c: DiscordChannelConfig | undefined): PbDiscordChannelConfig | undefined {
  if (!c) return undefined;
  return create(DiscordChannelConfigSchema, {
    botToken: c.bot_token ?? "",
    allowedChannelIds: c.allowed_channel_ids ?? c.allow_channel_ids ?? [],
  });
}

function channelFromProto(ch: PbAgentChannel): AgentChannel {
  return {
    name: ch.name,
    agent_name: ch.agentName,
    platform: platformFromProto(ch.platform),
    enabled: ch.enabled,
    triggers: ch.triggers.map(triggerFromProto),
    telegram: telegramFromProto(ch.telegram),
    discord: discordFromProto(ch.discord),
    model: ch.model,
    metadata: ch.metadata,
  };
}

function channelToProto(ch: AgentChannel): PbAgentChannel {
  return create(AgentChannelSchema, {
    name: ch.name,
    agentName: ch.agent_name,
    platform: platformToProto(ch.platform),
    enabled: ch.enabled ?? false,
    triggers: (ch.triggers ?? []).map(triggerToProto),
    telegram: telegramToProto(ch.telegram),
    discord: discordToProto(ch.discord),
    model: ch.model ?? "",
    metadata: ch.metadata ?? {},
  });
}

function statusFromProto(s: PbChannelStatus): ChannelStatus {
  return {
    name: s.name,
    platform: platformFromProto(s.platform),
    state: stateFromProto(s.state),
    last_poll_at: tsToISO(s.lastPollAt),
    detail: s.detail,
  };
}

async function listChannels(): Promise<{ channels?: AgentChannel[] }> {
  const res = await client.listChannels({});
  return { channels: res.channels.map(channelFromProto) };
}

async function getChannel(name: string): Promise<{ channel: AgentChannel }> {
  const res = await client.getChannel({ name });
  if (!res.channel) throw new Error("not found");
  return { channel: channelFromProto(res.channel) };
}

async function createChannel(channel: AgentChannel): Promise<{ channel: AgentChannel }> {
  const res = await client.createChannel({ channel: channelToProto(channel) });
  if (!res.channel) throw new Error("create returned nothing");
  return { channel: channelFromProto(res.channel) };
}

async function updateChannel(channel: AgentChannel): Promise<{ channel: AgentChannel }> {
  const res = await client.updateChannel({ channel: channelToProto(channel) });
  if (!res.channel) throw new Error("update returned nothing");
  return { channel: channelFromProto(res.channel) };
}

async function deleteChannel(name: string): Promise<void> {
  await client.deleteChannel({ name });
}

async function getChannelStatus(name: string): Promise<{ status: ChannelStatus }> {
  const res = await client.getChannelStatus({ name });
  if (!res.status) throw new Error("status not found");
  return { status: statusFromProto(res.status) };
}

async function restartChannel(name: string): Promise<{ channel?: AgentChannel }> {
  const res = await client.restartChannel({ name });
  return { channel: res.channel ? channelFromProto(res.channel) : undefined };
}

async function pauseChannel(name: string): Promise<{ channel?: AgentChannel }> {
  const res = await client.pauseChannel({ name });
  return { channel: res.channel ? channelFromProto(res.channel) : undefined };
}

async function resumeChannel(name: string): Promise<{ channel?: AgentChannel }> {
  const res = await client.resumeChannel({ name });
  return { channel: res.channel ? channelFromProto(res.channel) : undefined };
}

export function useChannels() {
  return useQuery({ queryKey: ["channels"], queryFn: listChannels });
}

export function useChannel(name: string) {
  return useQuery({ queryKey: ["channels", name], queryFn: () => getChannel(name), enabled: !!name });
}

export function useChannelStatus(name: string) {
  return useQuery({
    queryKey: ["channels", name, "status"],
    queryFn: () => getChannelStatus(name),
    enabled: !!name,
    refetchInterval: 10_000,
  });
}

export function useCreateChannel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createChannel,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["channels"] }),
  });
}

export function useUpdateChannel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateChannel,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["channels"] }),
  });
}

export function useDeleteChannel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteChannel,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["channels"] }),
  });
}

export function useRestartChannel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: restartChannel,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["channels"] }),
  });
}

export function usePauseChannel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: pauseChannel,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["channels"] }),
  });
}

export function useResumeChannel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: resumeChannel,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["channels"] }),
  });
}
