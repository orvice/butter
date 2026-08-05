import { useMemo } from "react";
import { useModelProviders } from "@/api/model-providers";
import { FormControl } from "@/components/ui/form";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { ModelProvider } from "@/types/api";

type ModelOption = {
  value: string;
  label: string;
  description: string;
};

type AgentModelSelectProps = {
  value?: string;
  onChange: (value: string) => void;
};

function buildModelOptions(providers?: ModelProvider[]): ModelOption[] {
  const seen = new Set<string>();
  const options: ModelOption[] = [];

  for (const provider of providers ?? []) {
    for (const model of provider.models ?? []) {
      const value = (model.alias || model.name || "").trim();
      if (!value || seen.has(value)) continue;

      seen.add(value);
      options.push({
        value,
        label: model.alias || model.name,
        description: model.alias ? `${model.name} · ${provider.name}` : provider.name,
      });
    }
  }

  return options.sort((a, b) => a.label.localeCompare(b.label));
}

export function AgentModelSelect({ value, onChange }: AgentModelSelectProps) {
  const { data, isLoading } = useModelProviders();
  const options = useMemo(() => {
    const configuredOptions = buildModelOptions(data?.model_providers);
    const currentValue = value?.trim();

    if (currentValue && !configuredOptions.some((option) => option.value === currentValue)) {
      return [
        ...configuredOptions,
        {
          value: currentValue,
          label: currentValue,
          description: "Current value (not in configured providers)",
        },
      ];
    }

    return configuredOptions;
  }, [data?.model_providers, value]);

  return (
    <Select onValueChange={(nextValue) => nextValue && onChange(nextValue)} value={value || undefined} disabled={isLoading || options.length === 0}>
      <FormControl>
        <SelectTrigger>
          <SelectValue placeholder={isLoading ? "Loading models..." : "Select a model"} />
        </SelectTrigger>
      </FormControl>
      <SelectContent>
        {options.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            <div className="flex flex-col">
              <span>{option.label}</span>
              <span className="text-xs text-muted-foreground">{option.description}</span>
            </div>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
