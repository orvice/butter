import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useAgents } from '@/api/agents'
import { useNotifyGroups } from '@/api/notify-groups'
import { useTelegramDestinations } from '@/api/telegram'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { ScheduleBuilder } from '@/components/schedule-builder'
import { PageActions } from '@/components/butter/page-parts'
import type { CronConcurrencyPolicy, CronDeliveryType, CronJob, CronNotifyOn } from '@/types/api'

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  schedule: z.string().min(1, 'Schedule is required'),
  // Opaque agent ref: the immutable agent_id when the agent has one, else the name.
  agent_ref: z.string().min(1, 'Agent is required'),
  input: z.string().optional(),
  timezone: z.string().optional(),
  enabled: z.boolean(),
  delivery_type: z.string(),
  webhook_url: z.string().optional(),
  channel_name: z.string().optional(),
  chat_id: z.string().optional(),
  notify_group_name: z.string().optional(),
  telegram_destination_id: z.string().optional(),
  timeout_seconds: z.number().optional(),
  retry_attempts: z.number().optional(),
  retry_backoff_seconds: z.number().optional(),
  concurrency_policy: z.string(),
  notify_on: z.string(),
  max_output_bytes: z.number().optional(),
})

type FormValues = z.infer<typeof schema>

type AutomationFormProps = {
  mode: 'create' | 'edit'
  initialValue?: CronJob
  loading?: boolean
  submitLabel: string
  onCancel: () => void
  onSubmit: (job: CronJob) => void
}

export function AutomationForm({
  mode,
  initialValue,
  loading,
  submitLabel,
  onCancel,
  onSubmit,
}: AutomationFormProps) {
  const { data: agentsData } = useAgents()
  const { data: notifyGroupsData } = useNotifyGroups()
  const agents = agentsData?.agents ?? []

  // Normalizes a stored ref (agent_id or legacy agent_name) to the matching
  // agent's stable ref so the select finds its option once agents load.
  function resolveAgentRef(ref: string): string | undefined {
    const match = agents.find((a) => a.agent_id === ref || a.name === ref)
    return match ? match.agent_id || match.name : undefined
  }

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: formValuesFromJob(initialValue),
  })
  const deliveryType = useWatch({ control: form.control, name: 'delivery_type' })
  // Telegram cron delivery references a destination; the chat, forum topic,
  // and bot credential all live there (issue #264).
  const { data: telegramDestinations = [] } = useTelegramDestinations()
  const isEdit = mode === 'edit'

  function handleSubmit(values: FormValues) {
    const selectedAgent = agents.find(
      (a) => (a.agent_id || a.name) === resolveAgentRef(values.agent_ref),
    )
    onSubmit({
      name: values.name,
      schedule: values.schedule,
      // Submit the immutable agent_id when the agent has one; agent_name is
      // kept alongside as the human-readable label (the backend normalizes
      // and backfills both fields).
      agent_name: selectedAgent?.name ?? values.agent_ref,
      agent_id: selectedAgent?.agent_id || undefined,
      input: values.input,
      timezone: values.timezone,
      enabled: values.enabled,
      delivery: {
        type: values.delivery_type as CronDeliveryType,
        webhook_url: values.webhook_url,
        channel_name: values.channel_name,
        chat_id: values.chat_id,
        notify_group_name: values.notify_group_name,
        telegram_destination_id: values.telegram_destination_id,
      },
      timeout_seconds: values.timeout_seconds || undefined,
      retry: values.retry_attempts
        ? { max_attempts: values.retry_attempts, backoff_seconds: values.retry_backoff_seconds || undefined }
        : undefined,
      concurrency_policy: values.concurrency_policy as CronConcurrencyPolicy,
      notify_on: values.notify_on as CronNotifyOn,
      max_output_bytes: values.max_output_bytes || undefined,
    })
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className='space-y-6'>
        <Card>
          <CardHeader>
            <CardTitle>Basic Info</CardTitle>
            <CardDescription>Pick the schedule, target agent, input, and timezone for automatic runs.</CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>
            <FormField control={form.control} name='name' render={({ field }) => (
              <FormItem><FormLabel>Name</FormLabel><FormControl><Input placeholder='daily-summary' {...field} disabled={isEdit} /></FormControl><FormMessage /></FormItem>
            )} />
            <FormField control={form.control} name='schedule' render={({ field }) => (
              <FormItem>
                <FormLabel>Schedule</FormLabel>
                <FormControl>
                  <ScheduleBuilder value={field.value} onChange={field.onChange} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name='agent_ref' render={({ field }) => (
              <FormItem>
                <FormLabel>Agent</FormLabel>
                <Select
                  onValueChange={(value) => {
                    if (value) field.onChange(value)
                  }}
                  value={resolveAgentRef(field.value)}
                >
                  <FormControl><SelectTrigger><SelectValue placeholder='Select agent' /></SelectTrigger></FormControl>
                  <SelectContent>
                    {agents.map((a) => (
                      <SelectItem key={a.agent_id || a.name} value={a.agent_id || a.name}>{a.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name='input' render={({ field }) => (
              <FormItem><FormLabel>Input Message</FormLabel><FormControl><Textarea placeholder='Generate a daily summary' rows={3} {...field} /></FormControl></FormItem>
            )} />
            <FormField control={form.control} name='timezone' render={({ field }) => (
              <FormItem><FormLabel>Timezone</FormLabel><FormControl><Input placeholder='UTC' {...field} /></FormControl></FormItem>
            )} />
            <FormField control={form.control} name='enabled' render={({ field }) => (
              <FormItem className='flex items-center gap-3'>
                <FormLabel>Enabled</FormLabel>
                <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
              </FormItem>
            )} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Reliability</CardTitle>
            <CardDescription>Control timeouts, retries, overlap handling, and result notifications.</CardDescription>
          </CardHeader>
          <CardContent className='grid gap-4 md:grid-cols-2'>
            <FormField control={form.control} name='timeout_seconds' render={({ field }) => (
              <FormItem><FormLabel>Timeout Seconds</FormLabel><FormControl><Input type='number' min={0} value={field.value ?? 0} onChange={(event) => field.onChange(event.currentTarget.valueAsNumber || 0)} /></FormControl></FormItem>
            )} />
            <FormField control={form.control} name='retry_attempts' render={({ field }) => (
              <FormItem><FormLabel>Retry Attempts</FormLabel><FormControl><Input type='number' min={0} value={field.value ?? 0} onChange={(event) => field.onChange(event.currentTarget.valueAsNumber || 0)} /></FormControl></FormItem>
            )} />
            <FormField control={form.control} name='retry_backoff_seconds' render={({ field }) => (
              <FormItem><FormLabel>Retry Backoff Seconds</FormLabel><FormControl><Input type='number' min={0} value={field.value ?? 0} onChange={(event) => field.onChange(event.currentTarget.valueAsNumber || 0)} /></FormControl></FormItem>
            )} />
            <FormField control={form.control} name='max_output_bytes' render={({ field }) => (
              <FormItem><FormLabel>Max Output Bytes</FormLabel><FormControl><Input type='number' min={0} value={field.value ?? 0} onChange={(event) => field.onChange(event.currentTarget.valueAsNumber || 0)} /></FormControl></FormItem>
            )} />
            <FormField control={form.control} name='concurrency_policy' render={({ field }) => (
              <FormItem>
                <FormLabel>Concurrency</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                  <SelectContent>
                    <SelectItem value='CRON_CONCURRENCY_POLICY_SKIP'>Skip</SelectItem>
                    <SelectItem value='CRON_CONCURRENCY_POLICY_QUEUE'>Queue</SelectItem>
                    <SelectItem value='CRON_CONCURRENCY_POLICY_REPLACE'>Replace</SelectItem>
                    <SelectItem value='CRON_CONCURRENCY_POLICY_ALLOW'>Allow</SelectItem>
                  </SelectContent>
                </Select>
              </FormItem>
            )} />
            <FormField control={form.control} name='notify_on' render={({ field }) => (
              <FormItem>
                <FormLabel>Notify On</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                  <SelectContent>
                    <SelectItem value='CRON_NOTIFY_ON_ALWAYS'>Always</SelectItem>
                    <SelectItem value='CRON_NOTIFY_ON_FAILURE'>Failure</SelectItem>
                    <SelectItem value='CRON_NOTIFY_ON_SUCCESS'>Success</SelectItem>
                  </SelectContent>
                </Select>
              </FormItem>
            )} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Delivery</CardTitle>
            <CardDescription>Choose whether results stay in logs or are sent to a webhook, channel, or notify group.</CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>
            <FormField control={form.control} name='delivery_type' render={({ field }) => (
              <FormItem>
                <FormLabel>Type</FormLabel>
                <Select onValueChange={field.onChange} value={field.value}>
                  <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                  <SelectContent>
                    <SelectItem value='CRON_DELIVERY_TYPE_LOG'>Log</SelectItem>
                    <SelectItem value='CRON_DELIVERY_TYPE_WEBHOOK'>Webhook</SelectItem>
                    <SelectItem value='CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION'>Telegram Destination</SelectItem>
                    <SelectItem value='CRON_DELIVERY_TYPE_NOTIFY_GROUP'>Notify Group</SelectItem>
                  </SelectContent>
                </Select>
              </FormItem>
            )} />
            {deliveryType === 'CRON_DELIVERY_TYPE_WEBHOOK' && (
              <FormField control={form.control} name='webhook_url' render={({ field }) => (
                <FormItem><FormLabel>Webhook URL</FormLabel><FormControl><Input placeholder='https://...' {...field} /></FormControl></FormItem>
              )} />
            )}
            {deliveryType === 'CRON_DELIVERY_TYPE_CHANNEL' && (
              // Retained read-only for jobs saved before issue #264. Saving
              // one now is rejected by the API, so the form says so instead
              // of letting the operator discover it on submit.
              <p className='text-sm text-destructive'>
                Channel delivery is retired. Choose Telegram Destination instead; this job
                cannot be saved until you do.
              </p>
            )}
            {deliveryType === 'CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION' && (
              <FormField control={form.control} name='telegram_destination_id' render={({ field }) => (
                <FormItem>
                  <FormLabel>Telegram destination</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value || undefined}>
                    <FormControl>
                      <SelectTrigger aria-label='Telegram destination'>
                        <SelectValue placeholder='Select a destination' />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {telegramDestinations.map((destination) => (
                        <SelectItem key={destination.id} value={destination.id}>
                          {destination.name || destination.key}
                          {destination.messageThreadId ? ` \u00b7 topic ${destination.messageThreadId}` : ''}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormItem>
              )} />
            )}
            {deliveryType === 'CRON_DELIVERY_TYPE_NOTIFY_GROUP' && (
              <FormField control={form.control} name='notify_group_name' render={({ field }) => (
                <FormItem>
                  <FormLabel>Notify Group</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl><SelectTrigger><SelectValue placeholder='Select notify group' /></SelectTrigger></FormControl>
                    <SelectContent>
                      {(notifyGroupsData?.notify_groups ?? []).map((group) => (
                        <SelectItem key={group.name} value={group.name}>{group.name}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormItem>
              )} />
            )}
          </CardContent>
        </Card>

        <PageActions>
          <Button type='button' variant='outline' onClick={onCancel}>Cancel</Button>
          <Button type='submit' disabled={loading}>{loading ? 'Saving...' : submitLabel}</Button>
        </PageActions>
      </form>
    </Form>
  )
}

function formValuesFromJob(job?: CronJob): FormValues {
  return {
    name: job?.name ?? '',
    schedule: job?.schedule ?? '',
    agent_ref: job?.agent_id || job?.agent_name || '',
    input: job?.input ?? '',
    timezone: job?.timezone ?? 'UTC',
    enabled: job?.enabled ?? true,
    delivery_type: normalizeDeliveryType(job?.delivery?.type),
    webhook_url: job?.delivery?.webhook_url ?? '',
    channel_name: job?.delivery?.channel_name ?? '',
    chat_id: job?.delivery?.chat_id ?? '',
    notify_group_name: job?.delivery?.notify_group_name ?? '',
    telegram_destination_id: job?.delivery?.telegram_destination_id ?? '',
    timeout_seconds: job?.timeout_seconds ?? 0,
    retry_attempts: job?.retry?.max_attempts ?? 0,
    retry_backoff_seconds: job?.retry?.backoff_seconds ?? 0,
    concurrency_policy: normalizeConcurrencyPolicy(job?.concurrency_policy),
    notify_on: normalizeNotifyOn(job?.notify_on),
    max_output_bytes: job?.max_output_bytes ?? 4096,
  }
}

function normalizeDeliveryType(type?: CronDeliveryType): CronDeliveryType {
  return type && type !== 'CRON_DELIVERY_TYPE_UNSPECIFIED'
    ? type
    : 'CRON_DELIVERY_TYPE_LOG'
}

function normalizeConcurrencyPolicy(
  policy?: CronConcurrencyPolicy
): CronConcurrencyPolicy {
  return policy && policy !== 'CRON_CONCURRENCY_POLICY_UNSPECIFIED'
    ? policy
    : 'CRON_CONCURRENCY_POLICY_SKIP'
}

function normalizeNotifyOn(notifyOn?: CronNotifyOn): CronNotifyOn {
  return notifyOn && notifyOn !== 'CRON_NOTIFY_ON_UNSPECIFIED'
    ? notifyOn
    : 'CRON_NOTIFY_ON_ALWAYS'
}
