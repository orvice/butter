import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { create } from "@bufbuild/protobuf";
import { ModelProviderService } from "@/gen/agents/v1/agent_service_pb";
import {
  ModelProviderSchema,
  ModelConfigSchema,
  type ModelProvider as PbModelProvider,
} from "@/gen/agents/v1/agent_pb";
import type { ModelProvider } from "@/types/api";
import { makeClient } from "./transport";

const client = makeClient(ModelProviderService);

function toProto(p: ModelProvider): PbModelProvider {
  return create(ModelProviderSchema, {
    name: p.name,
    type: p.type,
    apiKey: p.api_key ?? "",
    baseUrl: p.base_url ?? "",
    models: (p.models ?? []).map((m) =>
      create(ModelConfigSchema, { name: m.name, alias: m.alias ?? "" }),
    ),
  });
}

function fromProto(p: PbModelProvider): ModelProvider {
  return {
    name: p.name,
    type: p.type,
    api_key: p.apiKey,
    base_url: p.baseUrl,
    models: p.models,
  };
}

async function listModelProviders(): Promise<{ model_providers: ModelProvider[] }> {
  const res = await client.listModelProviders({});
  return { model_providers: res.modelProviders.map(fromProto) };
}

async function getModelProvider(name: string): Promise<{ model_provider: ModelProvider }> {
  const res = await client.getModelProvider({ name });
  if (!res.modelProvider) throw new Error("not found");
  return { model_provider: fromProto(res.modelProvider) };
}

async function createModelProvider(model_provider: ModelProvider): Promise<{ model_provider: ModelProvider }> {
  const res = await client.createModelProvider({ modelProvider: toProto(model_provider) });
  if (!res.modelProvider) throw new Error("create returned no provider");
  return { model_provider: fromProto(res.modelProvider) };
}

async function updateModelProvider(model_provider: ModelProvider): Promise<{ model_provider: ModelProvider }> {
  const res = await client.updateModelProvider({ modelProvider: toProto(model_provider) });
  if (!res.modelProvider) throw new Error("update returned no provider");
  return { model_provider: fromProto(res.modelProvider) };
}

async function deleteModelProvider(name: string): Promise<void> {
  await client.deleteModelProvider({ name });
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
