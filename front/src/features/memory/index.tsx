import { useEffect, useState } from 'react'
import { z } from 'zod'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import type { WorkspaceMemoryConfig } from '@/gen/agents/v1/workspace_memory_pb'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import {
  Brain,
  CircleCheck,
  Info,
  PlugZap,
  Trash2,
  TriangleAlert,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  useDeleteWorkspaceMemoryConfig,
  usePutWorkspaceMemoryConfig,
  useTestWorkspaceMemoryConnection,
  useWorkspaceMemoryConfig,
} from '@/api/workspace-memory'
import { useCanManageWorkspace } from '@/hooks/use-workspace-role'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { DeleteDialog } from '@/components/delete-dialog'
import { apiKeyForSave } from './api-key'

function isHttpUrl(value: string): boolean {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

const schema = z.object({
  baseUrl: z
    .string()
    .trim()
    .min(1, 'Base URL is required')
    .refine(isHttpUrl, 'Must be an absolute http(s) URL'),
  enabled: z.boolean(),
  apiKey: z.string(),
  clearKey: z.boolean(),
})

type FormValues = z.infer<typeof schema>

const EMPTY_VALUES: FormValues = {
  baseUrl: '',
  enabled: true,
  apiKey: '',
  clearKey: false,
}

function valuesFromConfig(
  config: WorkspaceMemoryConfig | null | undefined
): FormValues {
  if (!config) return { ...EMPTY_VALUES }
  return {
    baseUrl: config.baseUrl,
    enabled: config.enabled,
    apiKey: '',
    clearKey: false,
  }
}

export function MemorySettingsPage() {
  const { data: config, isLoading } = useWorkspaceMemoryConfig()
  const { canManage, isLoading: isLoadingRole } = useCanManageWorkspace()
  const putMutation = usePutWorkspaceMemoryConfig()
  const deleteMutation = useDeleteWorkspaceMemoryConfig()
  const testMutation = useTestWorkspaceMemoryConnection()
  const [warning, setWarning] = useState('')
  const [deleteOpen, setDeleteOpen] = useState(false)

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { ...EMPTY_VALUES },
  })

  useEffect(() => {
    form.reset(valuesFromConfig(config))
  }, [config, form])

  const clearKey = useWatch({ control: form.control, name: 'clearKey' })
  const readOnly = !canManage
  const hasKey = !!config?.credentialSet

  function onSubmit(values: FormValues) {
    setWarning('')
    testMutation.reset()
    putMutation.mutate(
      {
        baseUrl: values.baseUrl,
        enabled: values.enabled,
        apiKey: apiKeyForSave(values),
      },
      {
        onSuccess: (res) => {
          if (res.warning) {
            setWarning(res.warning)
            toast.warning('Saved, but the mem0 server did not answer')
          } else {
            toast.success('Memory settings saved')
          }
        },
        onError: (err) => toast.error(err.message),
      }
    )
  }

  function onDelete() {
    deleteMutation.mutate(undefined, {
      onSuccess: () => {
        toast.success('Memory connection removed')
        setDeleteOpen(false)
        setWarning('')
        testMutation.reset()
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title='Memory'
        subtitle='Connect this workspace to a self-hosted mem0 server. Agents with memory enabled recall relevant memories before each turn and save what they learn.'
        actions={config ? <ConnectionBadges config={config} /> : undefined}
      />
      <PageScroll className='max-w-3xl'>
        {isLoading || isLoadingRole ? (
          <Skeleton className='h-96 w-full' />
        ) : (
          <div className='space-y-6'>
            <Alert>
              <Info className='size-4' />
              <AlertTitle>Workspace Memory is shared</AlertTitle>
              <AlertDescription>
                Every member, entry point (dashboard, API, Telegram, cron…), and
                memory-enabled agent in this workspace reads and writes the same
                memory pool. Anyone who can talk to a memory-enabled agent can
                influence what other agents recall.
              </AlertDescription>
            </Alert>

            {readOnly && (
              <p className='text-sm text-muted-foreground'>
                Only workspace owners and admins can change these settings.
              </p>
            )}

            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(onSubmit)}
                className='space-y-6'
              >
                <Card>
                  <CardHeader>
                    <CardTitle className='flex items-center gap-2'>
                      <Brain className='size-4' />
                      mem0 connection
                    </CardTitle>
                    <CardDescription>
                      butter talks to the mem0 OSS REST API. The extraction
                      model and embedder are configured on the mem0 server.
                    </CardDescription>
                  </CardHeader>
                  <CardContent className='space-y-4'>
                    <FormField
                      control={form.control}
                      name='baseUrl'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Base URL</FormLabel>
                          <FormControl>
                            <Input
                              placeholder='https://mem0.example.com'
                              disabled={readOnly}
                              {...field}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name='enabled'
                      render={({ field }) => (
                        <FormItem className='flex flex-row items-center justify-between rounded-md border p-3'>
                          <div className='space-y-0.5'>
                            <FormLabel>Enabled</FormLabel>
                            <FormDescription>
                              When off, memory-enabled agents run without
                              memory. Saving an enabled connection checks the
                              server first.
                            </FormDescription>
                          </div>
                          <FormControl>
                            <Switch
                              checked={field.value}
                              onCheckedChange={field.onChange}
                              disabled={readOnly}
                              aria-label='Enabled'
                            />
                          </FormControl>
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name='apiKey'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>
                            {hasKey ? 'Replace API key' : 'API key'}
                          </FormLabel>
                          <FormControl>
                            <Input
                              type='password'
                              autoComplete='new-password'
                              disabled={readOnly || clearKey}
                              {...field}
                            />
                          </FormControl>
                          <FormDescription>
                            Write-only: encrypted at rest and never displayed
                            again.{' '}
                            {hasKey
                              ? `A key is set${
                                  config?.credentialUpdatedAt
                                    ? ` (updated ${timestampDate(config.credentialUpdatedAt).toLocaleString()})`
                                    : ''
                                }; leave empty to keep it.`
                              : 'Leave empty only if the server runs with authentication disabled.'}
                          </FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    {hasKey && !readOnly && (
                      <FormField
                        control={form.control}
                        name='clearKey'
                        render={({ field }) => (
                          <FormItem className='flex flex-row items-center gap-2 space-y-0'>
                            <FormControl>
                              <Checkbox
                                checked={field.value}
                                onCheckedChange={(checked) =>
                                  field.onChange(checked === true)
                                }
                              />
                            </FormControl>
                            <FormLabel className='font-normal'>
                              Clear the stored API key on save
                            </FormLabel>
                          </FormItem>
                        )}
                      />
                    )}
                  </CardContent>
                </Card>

                {warning && (
                  <Alert variant='destructive'>
                    <TriangleAlert className='size-4' />
                    <AlertTitle>
                      Saved, but the connection check failed
                    </AlertTitle>
                    <AlertDescription className='break-all'>
                      {warning}
                    </AlertDescription>
                  </Alert>
                )}

                <ConnectionTestResult
                  result={testMutation.data}
                  error={testMutation.error}
                />

                <div className='flex flex-wrap items-center justify-between gap-2'>
                  <div className='flex gap-2'>
                    <Button
                      type='button'
                      variant='outline'
                      disabled={!config || testMutation.isPending}
                      onClick={() => testMutation.mutate()}
                    >
                      <PlugZap className='size-4' />
                      {testMutation.isPending
                        ? 'Testing...'
                        : 'Test connection'}
                    </Button>
                    {config && !readOnly && (
                      <Button
                        type='button'
                        variant='outline'
                        onClick={() => setDeleteOpen(true)}
                      >
                        <Trash2 className='size-4' />
                        Remove
                      </Button>
                    )}
                  </div>
                  {!readOnly && (
                    <Button type='submit' disabled={putMutation.isPending}>
                      {putMutation.isPending ? 'Saving...' : 'Save'}
                    </Button>
                  )}
                </div>
              </form>
            </Form>
          </div>
        )}
      </PageScroll>
      <DeleteDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title='Remove memory connection?'
        description='The mem0 URL and API key are removed from this workspace. Agents with memory enabled keep running, without memory. Memories already stored on the mem0 server are not deleted.'
        confirmLabel='Remove'
        loadingLabel='Removing...'
        loading={deleteMutation.isPending}
        onConfirm={onDelete}
      />
    </Page>
  )
}

function ConnectionBadges({ config }: { config: WorkspaceMemoryConfig }) {
  return (
    <div className='flex gap-2'>
      <Badge variant={config.enabled ? 'default' : 'secondary'}>
        {config.enabled ? 'Enabled' : 'Disabled'}
      </Badge>
      <Badge variant='outline'>
        {config.credentialSet ? 'API key set' : 'No API key'}
      </Badge>
    </div>
  )
}

function ConnectionTestResult({
  result,
  error,
}: {
  result?: { ok: boolean; error: string }
  error: Error | null
}) {
  if (error) {
    return (
      <Alert variant='destructive'>
        <TriangleAlert className='size-4' />
        <AlertTitle>Connection test failed</AlertTitle>
        <AlertDescription className='break-all'>
          {error.message}
        </AlertDescription>
      </Alert>
    )
  }
  if (!result) return null
  if (result.ok) {
    return (
      <Alert>
        <CircleCheck className='size-4' />
        <AlertTitle>Connected</AlertTitle>
        <AlertDescription>
          The mem0 server answered and accepted the API key.
        </AlertDescription>
      </Alert>
    )
  }
  return (
    <Alert variant='destructive'>
      <TriangleAlert className='size-4' />
      <AlertTitle>Connection test failed</AlertTitle>
      <AlertDescription className='break-all'>{result.error}</AlertDescription>
    </Alert>
  )
}
