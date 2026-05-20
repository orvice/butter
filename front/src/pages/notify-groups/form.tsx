import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import type { NotifyGroup, NotifyTarget } from "@/types/api";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  enabled: z.boolean(),
  targets_json: z.string(),
});

type FormValues = z.infer<typeof schema>;

type NotifyGroupFormProps = {
  initialValue?: NotifyGroup;
  submitting?: boolean;
  submitLabel: string;
  onSubmit: (group: NotifyGroup) => void;
};

export default function NotifyGroupForm({ initialValue, submitting, submitLabel, onSubmit }: NotifyGroupFormProps) {
  const navigate = useNavigate();
  const defaultTargets = useMemo(() => JSON.stringify(initialValue?.targets ?? [], null, 2), [initialValue]);
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: initialValue?.name ?? "",
      enabled: initialValue?.enabled ?? true,
      targets_json: defaultTargets,
    },
  });

  function handleSubmit(values: FormValues) {
    let targets: NotifyTarget[];
    try {
      targets = JSON.parse(values.targets_json || "[]") as NotifyTarget[];
      if (!Array.isArray(targets)) {
        form.setError("targets_json", { message: "Targets must be a JSON array" });
        return;
      }
    } catch {
      form.setError("targets_json", { message: "Targets must be valid JSON" });
      return;
    }
    onSubmit({
      name: values.name,
      enabled: values.enabled,
      targets,
      metadata: initialValue?.metadata,
    });
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
        <Card>
          <CardHeader><CardTitle>Notify Group</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="name" render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl><Input placeholder="ops-alerts" disabled={!!initialValue} {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="enabled" render={({ field }) => (
              <FormItem className="flex items-center gap-3">
                <FormLabel>Enabled</FormLabel>
                <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
              </FormItem>
            )} />
            <FormField control={form.control} name="targets_json" render={({ field }) => (
              <FormItem>
                <FormLabel>Targets JSON</FormLabel>
                <FormControl><Textarea rows={14} className="font-mono text-xs" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
          </CardContent>
        </Card>
        <div className="flex gap-3">
          <Button type="button" variant="outline" onClick={() => navigate("/notify-groups")}>Cancel</Button>
          <Button type="submit" disabled={submitting}>{submitting ? "Saving..." : submitLabel}</Button>
        </div>
      </form>
    </Form>
  );
}
