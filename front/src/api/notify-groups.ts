import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { NotifyGroup } from "@/types/api";

const SVC = "agents.v1.NotifyGroupService";

function listNotifyGroups() {
  return twirpFetch<object, { notify_groups: NotifyGroup[] }>(SVC, "ListNotifyGroups", {});
}

function getNotifyGroup(name: string) {
  return twirpFetch<{ name: string }, { notify_group: NotifyGroup }>(SVC, "GetNotifyGroup", { name });
}

function createNotifyGroup(notify_group: NotifyGroup) {
  return twirpFetch<{ notify_group: NotifyGroup }, { notify_group: NotifyGroup }>(
    SVC,
    "CreateNotifyGroup",
    { notify_group },
  );
}

function updateNotifyGroup(notify_group: NotifyGroup) {
  return twirpFetch<{ notify_group: NotifyGroup }, { notify_group: NotifyGroup }>(
    SVC,
    "UpdateNotifyGroup",
    { notify_group },
  );
}

function deleteNotifyGroup(name: string) {
  return twirpFetch<{ name: string }, object>(SVC, "DeleteNotifyGroup", { name });
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
