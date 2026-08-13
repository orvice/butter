import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import {
  DiscordNotifyTargetSchema,
  LarkNotifyTargetSchema,
  NotifyGroupSchema,
  NotifyTargetSchema,
  NotifyTargetType,
  TelegramNotifyTargetSchema,
  type NotifyGroup as PbNotifyGroup,
  type NotifyTarget as PbNotifyTarget,
} from "@/gen/agents/v1/agent_pb";
import { NotifyGroupService } from "@/gen/agents/v1/agent_service_pb";
import type { NotifyGroup, NotifyTarget, NotifyTargetType as LegacyType } from "@/types/api";
import { bigintToNumber } from "./_proto-bridge";
import { makeClient } from "./transport";

const client = makeClient(NotifyGroupService);

const NOTIFY_TARGET_TYPE_NAMES: LegacyType[] = [
  "NOTIFY_TARGET_TYPE_UNSPECIFIED",
  "NOTIFY_TARGET_TYPE_TELEGRAM",
  "NOTIFY_TARGET_TYPE_LARK_WEBHOOK",
  "NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK",
];

function typeFromProto(t: NotifyTargetType): LegacyType {
  return NOTIFY_TARGET_TYPE_NAMES[t] ?? "NOTIFY_TARGET_TYPE_UNSPECIFIED";
}

function typeToProto(t: LegacyType | undefined): NotifyTargetType {
  const idx = NOTIFY_TARGET_TYPE_NAMES.indexOf(t ?? "NOTIFY_TARGET_TYPE_UNSPECIFIED");
  return idx < 0 ? NotifyTargetType.UNSPECIFIED : (idx as NotifyTargetType);
}

function targetFromProto(t: PbNotifyTarget): NotifyTarget {
  return {
    name: t.name,
    enabled: t.enabled,
    type: typeFromProto(t.type),
    telegram: t.telegram
      ? {
          destination_id: t.telegram.destinationId,
          // Legacy fields are surfaced read-only so an old record still
          // renders; the form never sends them back (issue #264).
          bot_token: t.telegram.botToken,
          chat_id: t.telegram.chatId,
          parse_mode: t.telegram.parseMode,
          message_thread_id: bigintToNumber(t.telegram.messageThreadId),
        }
      : undefined,
    lark: t.lark
      ? { webhook_url: t.lark.webhookUrl, secret: t.lark.secret }
      : undefined,
    discord: t.discord
      ? {
          webhook_url: t.discord.webhookUrl,
          username: t.discord.username,
          avatar_url: t.discord.avatarUrl,
          thread_id: t.discord.threadId,
        }
      : undefined,
    metadata: t.metadata,
  };
}

function targetToProto(t: NotifyTarget): PbNotifyTarget {
  return create(NotifyTargetSchema, {
    name: t.name ?? "",
    enabled: t.enabled ?? false,
    type: typeToProto(t.type),
    telegram: t.telegram
      ? create(TelegramNotifyTargetSchema, {
          // Only the destination reference is written. The raw address fields
          // are rejected by the API, so sending them would turn every save
          // into a validation error (issue #264).
          destinationId: t.telegram.destination_id ?? "",
        })
      : undefined,
    lark: t.lark
      ? create(LarkNotifyTargetSchema, {
          webhookUrl: t.lark.webhook_url ?? "",
          secret: t.lark.secret ?? "",
        })
      : undefined,
    discord: t.discord
      ? create(DiscordNotifyTargetSchema, {
          webhookUrl: t.discord.webhook_url ?? "",
          username: t.discord.username ?? "",
          avatarUrl: t.discord.avatar_url ?? "",
          threadId: t.discord.thread_id ?? "",
        })
      : undefined,
    metadata: t.metadata ?? {},
  });
}

function fromProto(g: PbNotifyGroup): NotifyGroup {
  return {
    name: g.name,
    enabled: g.enabled,
    targets: g.targets.map(targetFromProto),
    metadata: g.metadata,
  };
}

function toProto(g: NotifyGroup): PbNotifyGroup {
  return create(NotifyGroupSchema, {
    name: g.name,
    enabled: g.enabled ?? false,
    targets: (g.targets ?? []).map(targetToProto),
    metadata: g.metadata ?? {},
  });
}

async function listNotifyGroups(): Promise<{ notify_groups: NotifyGroup[] }> {
  const res = await client.listNotifyGroups({});
  return { notify_groups: res.notifyGroups.map(fromProto) };
}

async function getNotifyGroup(name: string): Promise<{ notify_group: NotifyGroup }> {
  const res = await client.getNotifyGroup({ name });
  if (!res.notifyGroup) throw new Error("not found");
  return { notify_group: fromProto(res.notifyGroup) };
}

async function createNotifyGroup(group: NotifyGroup): Promise<{ notify_group: NotifyGroup }> {
  const res = await client.createNotifyGroup({ notifyGroup: toProto(group) });
  if (!res.notifyGroup) throw new Error("create returned no group");
  return { notify_group: fromProto(res.notifyGroup) };
}

async function updateNotifyGroup(group: NotifyGroup): Promise<{ notify_group: NotifyGroup }> {
  const res = await client.updateNotifyGroup({ notifyGroup: toProto(group) });
  if (!res.notifyGroup) throw new Error("update returned no group");
  return { notify_group: fromProto(res.notifyGroup) };
}

async function deleteNotifyGroup(name: string): Promise<void> {
  await client.deleteNotifyGroup({ name });
}

export function useNotifyGroups() {
  return useQuery({ queryKey: ["notify-groups"], queryFn: listNotifyGroups });
}

export function useNotifyGroup(name: string) {
  return useQuery({ queryKey: ["notify-groups", name], queryFn: () => getNotifyGroup(name), enabled: !!name });
}

export function useCreateNotifyGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createNotifyGroup,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notify-groups"] }),
  });
}

export function useUpdateNotifyGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateNotifyGroup,
    onSuccess: (_data, group) => {
      qc.invalidateQueries({ queryKey: ["notify-groups"] });
      qc.invalidateQueries({ queryKey: ["notify-groups", group.name] });
    },
  });
}

export function useDeleteNotifyGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteNotifyGroup,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notify-groups"] }),
  });
}
