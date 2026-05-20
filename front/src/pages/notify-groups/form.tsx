import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useFieldArray, useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import type { NotifyGroup, NotifyTarget, NotifyTargetType } from "@/types/api";

const TARGET_TYPES = [
  "NOTIFY_TARGET_TYPE_TELEGRAM",
  "NOTIFY_TARGET_TYPE_LARK_WEBHOOK",
  "NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK",
] as const;

const TARGET_TYPE_LABELS: Record<(typeof TARGET_TYPES)[number], string> = {
  NOTIFY_TARGET_TYPE_TELEGRAM: "Telegram",
  NOTIFY_TARGET_TYPE_LARK_WEBHOOK: "Lark Webhook",
  NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK: "Discord Webhook",
};

const PARSE_MODE_LABELS: Record<string, string> = {
  NONE: "None",
  Markdown: "Markdown",
  MarkdownV2: "MarkdownV2",
  HTML: "HTML",
};

const targetSchema = z.object({
  name: z.string().optional(),
  enabled: z.boolean(),
  type: z.enum(TARGET_TYPES),
  telegram_bot_token: z.string().optional(),
  telegram_chat_id: z.string().optional(),
  telegram_parse_mode: z.string().optional(),
  telegram_message_thread_id: z.string().regex(/^\d*$/, "Thread ID must be a non-negative integer").optional(),
  lark_webhook_url: z.string().optional(),
  lark_secret: z.string().optional(),
  discord_webhook_url: z.string().optional(),
  discord_username: z.string().optional(),
  discord_avatar_url: z.string().optional(),
  discord_thread_id: z.string().optional(),
  metadata: z.record(z.string(), z.string()).optional(),
}).superRefine((target, ctx) => {
  if (target.type === "NOTIFY_TARGET_TYPE_TELEGRAM") {
    if (!target.telegram_bot_token?.trim()) {
      ctx.addIssue({ code: "custom", path: ["telegram_bot_token"], message: "Bot token is required" });
    }
    if (!target.telegram_chat_id?.trim()) {
      ctx.addIssue({ code: "custom", path: ["telegram_chat_id"], message: "Chat ID is required" });
    }
  }

  if (target.type === "NOTIFY_TARGET_TYPE_LARK_WEBHOOK" && !target.lark_webhook_url?.trim()) {
    ctx.addIssue({ code: "custom", path: ["lark_webhook_url"], message: "Webhook URL is required" });
  }

  if (target.type === "NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK" && !target.discord_webhook_url?.trim()) {
    ctx.addIssue({ code: "custom", path: ["discord_webhook_url"], message: "Webhook URL is required" });
  }
});

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  enabled: z.boolean(),
  targets: z.array(targetSchema),
});

type FormValues = z.infer<typeof schema>;
type TargetFormValue = FormValues["targets"][number];

type NotifyGroupFormProps = {
  initialValue?: NotifyGroup;
  submitting?: boolean;
  submitLabel: string;
  onSubmit: (group: NotifyGroup) => void;
};

export default function NotifyGroupForm({ initialValue, submitting, submitLabel, onSubmit }: NotifyGroupFormProps) {
  const navigate = useNavigate();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: groupToFormValues(initialValue),
  });
  const { fields, append, remove } = useFieldArray({ control: form.control, name: "targets" });
  const watchedTargets = useWatch({ control: form.control, name: "targets" });

  useEffect(() => {
    form.reset(groupToFormValues(initialValue));
  }, [form, initialValue]);

  function handleSubmit(values: FormValues) {
    onSubmit({
      name: values.name.trim(),
      enabled: values.enabled,
      targets: values.targets.map(formTargetToNotifyTarget),
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
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <CardTitle>Targets</CardTitle>
            <Button type="button" variant="outline" size="sm" onClick={() => append(blankTarget())}>
              <Plus className="mr-1 h-3 w-3" /> Add target
            </Button>
          </CardHeader>
          <CardContent className="space-y-4">
            {fields.length === 0 ? (
              <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
                No notification targets configured.
              </div>
            ) : (
              fields.map((target, index) => {
                const targetType = watchedTargets?.[index]?.type ?? target.type;

                return (
                  <div key={target.id} className="space-y-4 rounded-md border p-4">
                    <div className="grid gap-4 lg:grid-cols-[1fr_220px_auto] lg:items-end">
                      <FormField control={form.control} name={`targets.${index}.name`} render={({ field }) => (
                        <FormItem>
                          <FormLabel>Target Name</FormLabel>
                          <FormControl><Input placeholder="ops-telegram" {...field} /></FormControl>
                          <FormMessage />
                        </FormItem>
                      )} />
                      <FormField control={form.control} name={`targets.${index}.type`} render={({ field }) => (
                        <FormItem>
                          <FormLabel>Type</FormLabel>
                          <Select onValueChange={field.onChange} value={field.value}>
                            <FormControl>
                              <SelectTrigger>
                                <span className="truncate">{TARGET_TYPE_LABELS[field.value]}</span>
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              <SelectItem value="NOTIFY_TARGET_TYPE_TELEGRAM">Telegram</SelectItem>
                              <SelectItem value="NOTIFY_TARGET_TYPE_LARK_WEBHOOK">Lark Webhook</SelectItem>
                              <SelectItem value="NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK">Discord Webhook</SelectItem>
                            </SelectContent>
                          </Select>
                          <FormMessage />
                        </FormItem>
                      )} />
                      <div className="flex items-center justify-between gap-3 lg:justify-end">
                        <FormField control={form.control} name={`targets.${index}.enabled`} render={({ field }) => (
                          <FormItem className="flex items-center gap-3">
                            <FormLabel>Enabled</FormLabel>
                            <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                          </FormItem>
                        )} />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={() => remove(index)}
                          aria-label="Remove target"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>

                    {targetType === "NOTIFY_TARGET_TYPE_TELEGRAM" && (
                      <div className="grid gap-4 md:grid-cols-2">
                        <FormField control={form.control} name={`targets.${index}.telegram_bot_token`} render={({ field }) => (
                          <FormItem>
                            <FormLabel>Bot Token</FormLabel>
                            <FormControl><Input type="password" placeholder="Telegram bot token" {...field} /></FormControl>
                            <FormMessage />
                          </FormItem>
                        )} />
                        <FormField control={form.control} name={`targets.${index}.telegram_chat_id`} render={({ field }) => (
                          <FormItem>
                            <FormLabel>Chat ID</FormLabel>
                            <FormControl><Input placeholder="-1001234567890" {...field} /></FormControl>
                            <FormMessage />
                          </FormItem>
                        )} />
                        <FormField control={form.control} name={`targets.${index}.telegram_parse_mode`} render={({ field }) => (
                          <FormItem>
                            <FormLabel>Parse Mode</FormLabel>
                            <Select
                              onValueChange={(value) => field.onChange(value === "NONE" ? "" : value)}
                              value={field.value || "NONE"}
                            >
                              <FormControl>
                                <SelectTrigger>
                                  <span className="truncate">{PARSE_MODE_LABELS[field.value || "NONE"]}</span>
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent>
                                <SelectItem value="NONE">None</SelectItem>
                                <SelectItem value="Markdown">Markdown</SelectItem>
                                <SelectItem value="MarkdownV2">MarkdownV2</SelectItem>
                                <SelectItem value="HTML">HTML</SelectItem>
                              </SelectContent>
                            </Select>
                            <FormMessage />
                          </FormItem>
                        )} />
                        <FormField control={form.control} name={`targets.${index}.telegram_message_thread_id`} render={({ field }) => (
                          <FormItem>
                            <FormLabel>Message Thread ID</FormLabel>
                            <FormControl><Input inputMode="numeric" placeholder="Optional topic ID" {...field} /></FormControl>
                            <FormMessage />
                          </FormItem>
                        )} />
                      </div>
                    )}

                    {targetType === "NOTIFY_TARGET_TYPE_LARK_WEBHOOK" && (
                      <div className="grid gap-4 md:grid-cols-2">
                        <FormField control={form.control} name={`targets.${index}.lark_webhook_url`} render={({ field }) => (
                          <FormItem>
                            <FormLabel>Webhook URL</FormLabel>
                            <FormControl><Input placeholder="https://open.larksuite.com/..." {...field} /></FormControl>
                            <FormMessage />
                          </FormItem>
                        )} />
                        <FormField control={form.control} name={`targets.${index}.lark_secret`} render={({ field }) => (
                          <FormItem>
                            <FormLabel>Signing Secret</FormLabel>
                            <FormControl><Input type="password" placeholder="Optional secret" {...field} /></FormControl>
                            <FormMessage />
                          </FormItem>
                        )} />
                      </div>
                    )}

                    {targetType === "NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK" && (
                      <div className="grid gap-4 md:grid-cols-2">
                        <FormField control={form.control} name={`targets.${index}.discord_webhook_url`} render={({ field }) => (
                          <FormItem>
                            <FormLabel>Webhook URL</FormLabel>
                            <FormControl><Input placeholder="https://discord.com/api/webhooks/..." {...field} /></FormControl>
                            <FormMessage />
                          </FormItem>
                        )} />
                        <FormField control={form.control} name={`targets.${index}.discord_username`} render={({ field }) => (
                          <FormItem>
                            <FormLabel>Username</FormLabel>
                            <FormControl><Input placeholder="Optional display name" {...field} /></FormControl>
                            <FormMessage />
                          </FormItem>
                        )} />
                        <FormField control={form.control} name={`targets.${index}.discord_avatar_url`} render={({ field }) => (
                          <FormItem>
                            <FormLabel>Avatar URL</FormLabel>
                            <FormControl><Input placeholder="Optional avatar URL" {...field} /></FormControl>
                            <FormMessage />
                          </FormItem>
                        )} />
                        <FormField control={form.control} name={`targets.${index}.discord_thread_id`} render={({ field }) => (
                          <FormItem>
                            <FormLabel>Thread ID</FormLabel>
                            <FormControl><Input placeholder="Optional thread ID" {...field} /></FormControl>
                            <FormMessage />
                          </FormItem>
                        )} />
                      </div>
                    )}
                  </div>
                );
              })
            )}
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

function blankTarget(): TargetFormValue {
  return {
    name: "",
    enabled: true,
    type: "NOTIFY_TARGET_TYPE_TELEGRAM",
    telegram_bot_token: "",
    telegram_chat_id: "",
    telegram_parse_mode: "",
    telegram_message_thread_id: "",
    lark_webhook_url: "",
    lark_secret: "",
    discord_webhook_url: "",
    discord_username: "",
    discord_avatar_url: "",
    discord_thread_id: "",
    metadata: undefined,
  };
}

function groupToFormValues(group?: NotifyGroup): FormValues {
  return {
    name: group?.name ?? "",
    enabled: group?.enabled ?? true,
    targets: (group?.targets ?? []).map(notifyTargetToFormTarget),
  };
}

function notifyTargetToFormTarget(target: NotifyTarget): TargetFormValue {
  return {
    ...blankTarget(),
    name: target.name ?? "",
    enabled: target.enabled ?? true,
    type: isSupportedTargetType(target.type) ? target.type : "NOTIFY_TARGET_TYPE_TELEGRAM",
    telegram_bot_token: target.telegram?.bot_token ?? "",
    telegram_chat_id: target.telegram?.chat_id ?? "",
    telegram_parse_mode: target.telegram?.parse_mode ?? "",
    telegram_message_thread_id: target.telegram?.message_thread_id ? String(target.telegram.message_thread_id) : "",
    lark_webhook_url: target.lark?.webhook_url ?? "",
    lark_secret: target.lark?.secret ?? "",
    discord_webhook_url: target.discord?.webhook_url ?? "",
    discord_username: target.discord?.username ?? "",
    discord_avatar_url: target.discord?.avatar_url ?? "",
    discord_thread_id: target.discord?.thread_id ?? "",
    metadata: target.metadata,
  };
}

function formTargetToNotifyTarget(target: TargetFormValue): NotifyTarget {
  const base = {
    name: target.name?.trim() || undefined,
    enabled: target.enabled,
    type: target.type as NotifyTargetType,
    metadata: target.metadata,
  };

  if (target.type === "NOTIFY_TARGET_TYPE_TELEGRAM") {
    const threadId = Number(target.telegram_message_thread_id || 0);
    return {
      ...base,
      telegram: {
        bot_token: target.telegram_bot_token?.trim() || undefined,
        chat_id: target.telegram_chat_id?.trim() || undefined,
        parse_mode: target.telegram_parse_mode || undefined,
        message_thread_id: threadId > 0 ? threadId : undefined,
      },
    };
  }

  if (target.type === "NOTIFY_TARGET_TYPE_LARK_WEBHOOK") {
    return {
      ...base,
      lark: {
        webhook_url: target.lark_webhook_url?.trim() || undefined,
        secret: target.lark_secret?.trim() || undefined,
      },
    };
  }

  return {
    ...base,
    discord: {
      webhook_url: target.discord_webhook_url?.trim() || undefined,
      username: target.discord_username?.trim() || undefined,
      avatar_url: target.discord_avatar_url?.trim() || undefined,
      thread_id: target.discord_thread_id?.trim() || undefined,
    },
  };
}

function isSupportedTargetType(type: NotifyTarget["type"]): type is TargetFormValue["type"] {
  return TARGET_TYPES.includes(type as TargetFormValue["type"]);
}
