import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { ModelProvider } from "@/types/api";

const SVC = "agents.v1.ModelProviderService";

function listModelProviders() {
  return twirpFetch<object, { model_providers: ModelProvider[] }>(SVC, "ListModelProviders", {});
}

function getModelProvider(name: string) {
  return twirpFetch<{ name: string }, { model_provider: ModelProvider }>(SVC, "GetModelProvider", { name });
}

function createModelProvider(model_provider: ModelProvider) {
  return twirpFetch<{ model_provider: ModelProvider }, { model_provider: ModelProvider }>(
    SVC,
    "CreateModelProvider",
    { model_provider },
  );
}

function updateModelProvider(model_provider: ModelProvider) {
  return twirpFetch<{ model_provider: ModelProvider }, { model_provider: ModelProvider }>(
    SVC,
    "UpdateModelProvider",
    { model_provider },
  );
}

function deleteModelProvider(name: string) {
  return twirpFetch<{ name: string }, object>(SVC, "DeleteModelProvider", { name });
}

export function useModelProviders() {
  return useQuery({ queryKey: ["model-providers"], queryFn: listModelProviders });
}

export function useModelProvider(name: string) {
  return useQuery({
    queryKey: ["model-providers", name],
    queryFn: () => getModelProvider(name),
    enabled: !!name,
  });
}

export function useCreateModelProvider() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createModelProvider,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["model-providers"] }),
  });
}

export function useUpdateModelProvider() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateModelProvider,
    onSuccess: (_data, provider) => {
      qc.invalidateQueries({ queryKey: ["model-providers"] });
      qc.invalidateQueries({ queryKey: ["model-providers", provider.name] });
    },
  });
}

export function useDeleteModelProvider() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteModelProvider,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["model-providers"] }),
  });
}
