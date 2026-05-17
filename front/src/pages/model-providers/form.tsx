import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import type { ModelConfig, ModelProvider } from "@/types/api";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  type: z.string().min(1, "Type is required"),
  api_key: z.string().optional(),
  base_url: z.string().optional(),
  models_text: z.string().min(1, "At least one model is required"),
});

type FormValues = z.infer<typeof schema>;

type ModelProviderFormProps = {
  mode: "create" | "edit";
  initialValue?: ModelProvider;
  loading?: boolean;
  submitLabel: string;
  onCancel: () => void;
  onSubmit: (provider: ModelProvider) => void;
};

function modelsToText(models?: ModelConfig[]) {
  return (models ?? [])
    .map((m) => (m.alias ? `${m.name}:${m.alias}` : m.name))
    .join("\n");
}

function parseModels(text: string): ModelConfig[] {
  return text
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const [name, ...aliasParts] = line.split(":");
      const alias = aliasParts.join(":").trim();
      return { name: name.trim(), ...(alias ? { alias } : {}) };
    });
}

export default function ModelProviderForm({
  mode,
  initialValue,
  loading,
  submitLabel,
  onCancel,
  onSubmit,
}: ModelProviderFormProps) {
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      type: "openai",
      api_key: "",
      base_url: "",
      models_text: "",
    },
  });

  useEffect(() => {
    if (initialValue) {
      form.reset({
        name: initialValue.name ?? "",
        type: initialValue.type || "openai",
        api_key: initialValue.api_key ?? "",
        base_url: initialValue.base_url ?? "",
        models_text: modelsToText(initialValue.models),
      });
    }
  }, [form, initialValue]);

  function appendModelExample(model: string, alias: string) {
    const current = form.getValues("models_text").trim();
    const next = `${model}:${alias}`;
    form.setValue("models_text", current ? `${current}\n${next}` : next, { shouldDirty: true, shouldValidate: true });
  }

  function handleSubmit(values: FormValues) {
    onSubmit({
      name: values.name,
      type: values.type,
      api_key: values.api_key || undefined,
      base_url: values.base_url || undefined,
      models: parseModels(values.models_text),
    });
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
        <Card>
          <CardHeader><CardTitle>Provider</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="name" render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl><Input placeholder="openai" {...field} disabled={mode === "edit"} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="type" render={({ field }) => (
              <FormItem>
                <FormLabel>Type</FormLabel>
                <FormControl><Input placeholder="openai or gemini" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="api_key" render={({ field }) => (
              <FormItem>
                <FormLabel>API Key</FormLabel>
                <FormControl><Input type="password" placeholder="sk-... or env-resolved value" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="base_url" render={({ field }) => (
              <FormItem>
                <FormLabel>Base URL</FormLabel>
                <FormControl><Input placeholder="https://api.openai.com/v1" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Models</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="models_text" render={({ field }) => (
              <FormItem>
                <FormLabel>Models</FormLabel>
                <FormControl>
                  <Textarea
                    rows={8}
                    placeholder={"gpt-4o:4o\ngpt-4.1:4.1\ngemini-2.5-pro:pro"}
                    className="font-mono"
                    {...field}
                  />
                </FormControl>
                <p className="text-xs text-muted-foreground">
                  One model per line. Use <span className="font-mono">model-name:alias</span>; alias is optional.
                </p>
                <FormMessage />
              </FormItem>
            )} />
            <div className="flex flex-wrap gap-2">
              <Button type="button" variant="outline" size="sm" onClick={() => appendModelExample("gpt-4o", "4o")}>
                <Plus className="mr-2 h-3.5 w-3.5" /> gpt-4o
              </Button>
              <Button type="button" variant="outline" size="sm" onClick={() => appendModelExample("gemini-2.5-pro", "pro")}>
                <Plus className="mr-2 h-3.5 w-3.5" /> gemini pro
              </Button>
              <Button type="button" variant="outline" size="sm" onClick={() => form.setValue("models_text", "", { shouldDirty: true, shouldValidate: true })}>
                <Trash2 className="mr-2 h-3.5 w-3.5" /> Clear
              </Button>
            </div>
          </CardContent>
        </Card>

        <div className="flex gap-3">
          <Button type="button" variant="outline" onClick={onCancel}>Cancel</Button>
          <Button type="submit" disabled={loading}>{loading ? "Saving..." : submitLabel}</Button>
        </div>
      </form>
    </Form>
  );
}
