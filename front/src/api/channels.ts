import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { AgentChannel, ChannelStatus } from "@/types/api";

const SVC = "agents.v1.ChannelService";

function listChannels() {
  return twirpFetch<object, { channels?: AgentChannel[] }>(SVC, "ListChannels", {});
}

function getChannel(name: string) {
  return twirpFetch<{ name: string }, { channel: AgentChannel }>(SVC, "GetChannel", { name });
}

function createChannel(channel: AgentChannel) {
  return twirpFetch<{ channel: AgentChannel }, { channel: AgentChannel }>(
    SVC,
    "CreateChannel",
    { channel },
  );
}

function updateChannel(channel: AgentChannel) {
  return twirpFetch<{ channel: AgentChannel }, { channel: AgentChannel }>(
    SVC,
    "UpdateChannel",
    { channel },
  );
}

function deleteChannel(name: string) {
  return twirpFetch<{ name: string }, object>(SVC, "DeleteChannel", { name });
}

function getChannelStatus(name: string) {
  return twirpFetch<{ name: string }, { status: ChannelStatus }>(
    SVC,
    "GetChannelStatus",
    { name },
  );
}

function restartChannel(name: string) {
  return twirpFetch<{ name: string }, { channel?: AgentChannel }>(
    SVC,
    "RestartChannel",
    { name },
  );
}

function pauseChannel(name: string) {
  return twirpFetch<{ name: string }, { channel?: AgentChannel }>(
    SVC,
    "PauseChannel",
    { name },
  );
}

function resumeChannel(name: string) {
  return twirpFetch<{ name: string }, { channel?: AgentChannel }>(
    SVC,
    "ResumeChannel",
    { name },
  );
}

export function useChannels() {
  return useQuery({ queryKey: ["channels"], queryFn: listChannels });
}

export function useChannel(name: string) {
  return useQuery({
    queryKey: ["channels", name],
    queryFn: () => getChannel(name),
    enabled: !!name,
  });
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
