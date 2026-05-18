import { useEffect } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useAgents } from "@/api/agents";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import type { AgentChannel, AgentChannelPlatform, AgentTriggerType } from "@/types/api";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  agent_name: z.string().min(1, "Agent is required"),
  platform: z.enum(["AGENT_CHANNEL_PLATFORM_TELEGRAM", "AGENT_CHANNEL_PLATFORM_DISCORD"]),
  enabled: z.boolean(),
  model: z.string().optional(),
  trigger_type: z.enum([
    "AGENT_TRIGGER_TYPE_MESSAGE",
    "AGENT_TRIGGER_TYPE_COMMAND",
    "AGENT_TRIGGER_TYPE_PRIVATE_CHAT",
  ]),
  commands_text: z.string().optional(),
  prefixes_text: z.string().optional(),
  require_mention: z.boolean(),
  bot_token: z.string().optional(),
  allow_chat_ids_text: z.string().optional(),
  allow_channel_ids_text: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

type ChannelFormProps = {
  mode: "create" | "edit";
  initialValue?: AgentChannel;
  loading?: boolean;
  submitLabel: string;
  onCancel: () => void;
  onSubmit: (channel: AgentChannel) => void;
};

function linesToList(text?: string): string[] {
  return (text ?? "")
    .split(/\r?\n|,/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function listToLines(items?: string[]): string {
  return (items ?? []).join("\n");
}

function firstTrigger(channel?: AgentChannel) {
  return channel?.triggers?.[0];
}

export default function ChannelForm({
  mode,
  initialValue,
  loading,
  submitLabel,
  onCancel,
  onSubmit,
}: ChannelFormProps) {
  const { data: agentsData } = useAgents();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      agent_name: "",
      platform: "AGENT_CHANNEL_PLATFORM_TELEGRAM",
      enabled: true,
      model: "",
      trigger_type: "AGENT_TRIGGER_TYPE_MESSAGE",
      commands_text: "",
      prefixes_text: "",
      require_mention: false,
      bot_token: "",
      allow_chat_ids_text: "",
      allow_channel_ids_text: "",
    },
  });

  const platform = useWatch({ control: form.control, name: "platform" });
  const triggerType = useWatch({ control: form.control, name: "trigger_type" });

  useEffect(() => {
    if (!initialValue) return;

    const trigger = firstTrigger(initialValue);
    form.reset({
      name: initialValue.name ?? "",
      agent_name: initialValue.agent_name ?? "",
      platform: (initialValue.platform === "AGENT_CHANNEL_PLATFORM_DISCORD"
        ? "AGENT_CHANNEL_PLATFORM_DISCORD"
        : "AGENT_CHANNEL_PLATFORM_TELEGRAM"),
      enabled: initialValue.enabled ?? true,
      model: initialValue.model ?? "",
      trigger_type: (trigger?.type === "AGENT_TRIGGER_TYPE_COMMAND"
        || trigger?.type === "AGENT_TRIGGER_TYPE_PRIVATE_CHAT"
        || trigger?.type === "AGENT_TRIGGER_TYPE_MESSAGE"
        ? trigger.type
        : "AGENT_TRIGGER_TYPE_MESSAGE"),
      commands_text: listToLines(trigger?.commands),
      prefixes_text: listToLines(trigger?.prefixes),
      require_mention: trigger?.require_mention ?? false,
      bot_token: initialValue.telegram?.bot_token ?? initialValue.discord?.bot_token ?? "",
      allow_chat_ids_text: listToLines(initialValue.telegram?.allow_chat_ids),
      allow_channel_ids_text: listToLines(initialValue.discord?.allow_channel_ids),
    });
  }, [form, initialValue]);

  function handleSubmit(values: FormValues) {
    const trigger = {
      type: values.trigger_type as AgentTriggerType,
      commands: values.trigger_type === "AGENT_TRIGGER_TYPE_COMMAND" ? linesToList(values.commands_text) : [],
      prefixes: [],
      require_mention: false,
    };

    const channel: AgentChannel = {
      name: values.name.trim(),
      agent_name: values.agent_name,
      platform: values.platform as AgentChannelPlatform,
      enabled: values.enabled,
      model: values.model?.trim() || undefined,
      triggers: [trigger],
      ...(values.platform === "AGENT_CHANNEL_PLATFORM_TELEGRAM"
        ? {
            telegram: {
              bot_token: values.bot_token?.trim() || undefined,
              allow_chat_ids: linesToList(values.allow_chat_ids_text),
            },
          }
        : {
            discord: {
              bot_token: values.bot_token?.trim() || undefined,
              allow_channel_ids: linesToList(values.allow_channel_ids_text),
            },
          }),
    };

    onSubmit(channel);
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
        <Card>
          <CardHeader><CardTitle>Channel</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="name" render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl><Input placeholder="telegram-main" {...field} disabled={mode === "edit"} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="agent_name" render={({ field }) => (
              <FormItem>
                <FormLabel>Agent</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl><SelectTrigger><SelectValue placeholder="Select agent" /></SelectTrigger></FormControl>
                  <SelectContent>
                    {(agentsData?.agents ?? []).map((agent) => (
                      <SelectItem key={agent.name} value={agent.name}>{agent.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="platform" render={({ field }) => (
              <FormItem>
                <FormLabel>Platform</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                  <SelectContent>
                    <SelectItem value="AGENT_CHANNEL_PLATFORM_TELEGRAM">Telegram</SelectItem>
                    <SelectItem value="AGENT_CHANNEL_PLATFORM_DISCORD">Discord</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="model" render={({ field }) => (
              <FormItem>
                <FormLabel>Model Override</FormLabel>
                <FormControl><Input placeholder="Optional model alias/name" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="enabled" render={({ field }) => (
              <FormItem className="flex items-center gap-3">
                <FormLabel>Enabled</FormLabel>
                <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
              </FormItem>
            )} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Platform Settings</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="bot_token" render={({ field }) => (
              <FormItem>
                <FormLabel>Bot Token</FormLabel>
                <FormControl><Input type="password" placeholder="Bot token" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            {platform === "AGENT_CHANNEL_PLATFORM_TELEGRAM" ? (
              <FormField control={form.control} name="allow_chat_ids_text" render={({ field }) => (
                <FormItem>
                  <FormLabel>Allowed Chat IDs</FormLabel>
                  <FormControl><Textarea rows={4} placeholder={"One chat ID per line\nLeave empty to allow all"} {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
            ) : (
              <FormField control={form.control} name="allow_channel_ids_text" render={({ field }) => (
                <FormItem>
                  <FormLabel>Allowed Channel IDs</FormLabel>
                  <FormControl><Textarea rows={4} placeholder={"One channel ID per line\nLeave empty to allow all"} {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Triggers</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <FormField control={form.control} name="trigger_type" render={({ field }) => (
              <FormItem>
                <FormLabel>Type</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                  <SelectContent>
                    <SelectItem value="AGENT_TRIGGER_TYPE_MESSAGE">All messages</SelectItem>
                    <SelectItem value="AGENT_TRIGGER_TYPE_COMMAND">Commands</SelectItem>
                    <SelectItem value="AGENT_TRIGGER_TYPE_PRIVATE_CHAT">Private chat only</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )} />
            {triggerType === "AGENT_TRIGGER_TYPE_COMMAND" && (
              <FormField control={form.control} name="commands_text" render={({ field }) => (
                <FormItem>
                  <FormLabel>Commands</FormLabel>
                  <FormControl><Textarea rows={4} placeholder={"/start\n/ask"} {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
            )}

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
