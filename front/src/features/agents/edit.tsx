import { useState, useEffect, useMemo } from 'react'
import { Link, useParams, useNavigate } from '@tanstack/react-router'
import { ArrowLeft, ChevronRight } from 'lucide-react'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import Editor from '@monaco-editor/react'
import { useAgent, useUpdateAgent } from '@/api/agents'
import { useRepoBinding } from '@/api/repo-binding'
import { useMCPServers } from '@/api/mcp-servers'
import { useRemoteAgents } from '@/api/remote-agents'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form'
import { Page, PageActions, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { AGENT_TYPE_LABELS, enumLabel } from '@/lib/constants'
import { useTheme } from '@/context/theme-provider'
import { AgentModelSelect } from './model-select'
import { AgentIconUpload } from './icon-upload'
import { AgentFileMountsField } from './file-mounts-field'
import { AgentRemoteAgentsField } from './remote-agents-field'
import { agentIconUrl } from './icon-utils'
import type { Agent, AgentFileMount, AgentFileMountPermission, AgentType } from '@/types/api'

const MOUNT_PERMISSIONS = [
  'AGENT_FILE_MOUNT_PERMISSION_READ',
  'AGENT_FILE_MOUNT_PERMISSION_READ_WRITE',
  'AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE',
] as const satisfies readonly AgentFileMountPermission[]

const agentSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  type: z.string(),
  enable_a2a: z.boolean(),
  enable_openai_api: z.boolean(),
  enable_agui: z.boolean(),
  model: z.string().optional(),
  instruction: z.string().optional(),
  mcp_server_ids: z.array(z.string()).optional(),
  remote_agent_ids: z.array(z.string()).optional(),
  file_mounts: z.array(z.object({
    space_id: z.string().min(1),
    mount_path: z.string().optional(),
    permission: z.enum(MOUNT_PERMISSIONS).optional(),
  })).optional(),
  icon_url: z.string().optional(),
})

type AgentFormValues = z.infer<typeof agentSchema>

function toAgentFileMountFormValues(mounts?: AgentFileMount[]): AgentFormValues['file_mounts'] {
  return (mounts ?? [])
    .filter((mount) => mount.space_id)
    .map((mount) => ({
      space_id: mount.space_id,
      mount_path: mount.mount_path,
      permission:
        mount.permission === 'AGENT_FILE_MOUNT_PERMISSION_READ_WRITE' ||
        mount.permission === 'AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE'
          ? mount.permission
          : 'AGENT_FILE_MOUNT_PERMISSION_READ',
    }))
}

function mergeAgentIconMetadata(metadata: Agent['metadata'] | undefined, iconUrl: string | undefined) {
  const next = { ...(metadata ?? {}) }
  delete next.avatar_url
  if (iconUrl) next.icon_url = iconUrl
  else delete next.icon_url
  return Object.keys(next).length > 0 ? next : undefined
}

export function AgentEdit() {
  // The $name param is the agent's immutable agent_id.
  const { name: agentRef } = useParams({ from: '/_authenticated/agents/$name/edit' })
  const navigate = useNavigate()
  const { resolvedTheme } = useTheme()
  const { data, isLoading } = useAgent(agentRef ?? '')
  const { data: mcpData } = useMCPServers()
  const { data: remoteData, isLoading: isLoadingRemoteAgents } = useRemoteAgents()
  const { data: repoBindingData, isLoading: isLoadingRepoBinding, isError: isRepoBindingError } = useRepoBinding()
  const updateMutation = useUpdateAgent()
  const [operation, setOperation] = useState<{ request: string; id: string } | null>(null)
  const initialJsonValue = useMemo(() => (data?.agent ? JSON.stringify(data.agent, null, 2) : ''), [data])
  const [jsonValue, setJsonValue] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState('form')

  const form = useForm<AgentFormValues>({
    resolver: zodResolver(agentSchema),
    defaultValues: {
      name: '',
      description: '',
      type: 'AGENT_TYPE_LLM',
      enable_a2a: false,
      enable_openai_api: false,
      enable_agui: false,
      model: '',
      instruction: '',
      mcp_server_ids: [],
      remote_agent_ids: [],
      file_mounts: [],
      icon_url: '',
    },
  })
  const agentName = useWatch({ control: form.control, name: 'name' })
  const iconUrl = useWatch({ control: form.control, name: 'icon_url' })

  useEffect(() => {
    if (data?.agent) {
      const a = data.agent
      form.reset({
        name: a.name,
        description: a.description ?? '',
        type: a.type ?? 'AGENT_TYPE_LLM',
        enable_a2a: a.enable_a2a ?? false,
        enable_openai_api: a.enable_openai_api ?? false,
        enable_agui: a.enable_agui ?? false,
        model: a.config?.model ?? '',
        instruction: a.config?.instruction ?? '',
        mcp_server_ids: a.config?.mcp_server_ids ?? [],
        remote_agent_ids: a.config?.remote_agent_ids ?? [],
        file_mounts: toAgentFileMountFormValues(a.config?.file_mounts),
        icon_url: agentIconUrl(a),
      })
    }
  }, [data, form])

  function onFormSubmit(values: AgentFormValues) {
    const agent: Agent = {
      ...data?.agent,
      name: values.name,
      description: values.description,
      type: values.type as AgentType,
      enable_a2a: values.enable_a2a,
      enable_openai_api: values.enable_openai_api,
      enable_agui: values.enable_agui,
      metadata: mergeAgentIconMetadata(data?.agent?.metadata, values.icon_url),
      config: {
        ...data?.agent?.config,
        model: values.model,
        instruction: values.instruction,
        mcp_server_ids: values.mcp_server_ids ?? [],
        remote_agent_ids: values.remote_agent_ids ?? [],
        file_mounts: values.file_mounts ?? [],
      },
    }
    submitUpdate(agent)
  }

  function onJsonSubmit() {
    try {
      const agent = JSON.parse(jsonValue ?? initialJsonValue) as Agent
      submitUpdate(agent)
    } catch {
      toast.error('Invalid JSON')
    }
  }

  function submitUpdate(agent: Agent) {
    if (!data?.agent) return
    if (isRepoBindingError) {
      toast.error('Unable to determine repository binding status')
      return
    }
    const params = {
      agent,
      previous_agent: data.agent,
      repository_bound: !!repoBindingData?.binding,
      base_commit_sha: repoBindingData?.binding?.activeCommitSha ?? '',
    }
    const request = JSON.stringify(params)
    let requestOperation = operation
    if (!requestOperation || requestOperation.request !== request) {
      requestOperation = { request, id: crypto.randomUUID() }
      setOperation(requestOperation)
    }
    updateMutation.mutate(
      { ...params, operation_id: requestOperation.id },
      {
        onSuccess: () => { toast.success('Agent updated'); navigate({ to: '/agents' }) },
        onError: (err) => toast.error(err.message),
      },
    )
  }

  function handleTabChange(tab: string) {
    if (tab === 'json') {
      const values = form.getValues()
      const agent: Agent = {
        ...data?.agent,
        name: values.name,
        description: values.description,
        type: values.type as AgentType,
        enable_a2a: values.enable_a2a,
        enable_openai_api: values.enable_openai_api,
        enable_agui: values.enable_agui,
        metadata: mergeAgentIconMetadata(data?.agent?.metadata, values.icon_url),
        config: {
          ...data?.agent?.config,
          model: values.model,
          instruction: values.instruction,
          mcp_server_ids: values.mcp_server_ids ?? [],
          remote_agent_ids: values.remote_agent_ids ?? [],
          file_mounts: values.file_mounts ?? [],
        },
      }
      setJsonValue(JSON.stringify(agent, null, 2))
    } else if (tab === 'form') {
      try {
        const agent = JSON.parse(jsonValue ?? initialJsonValue) as Agent
        form.reset({
          name: agent.name,
          description: agent.description ?? '',
          type: agent.type ?? 'AGENT_TYPE_LLM',
          enable_a2a: agent.enable_a2a ?? false,
          enable_openai_api: agent.enable_openai_api ?? false,
          enable_agui: agent.enable_agui ?? false,
          model: agent.config?.model ?? '',
          instruction: agent.config?.instruction ?? '',
          mcp_server_ids: agent.config?.mcp_server_ids ?? [],
          remote_agent_ids: agent.config?.remote_agent_ids ?? [],
          file_mounts: toAgentFileMountFormValues(agent.config?.file_mounts),
          icon_url: agentIconUrl(agent),
        })
      } catch { /* keep current form values if JSON is invalid */ }
    }
    setActiveTab(tab)
  }

  if (isLoading || isLoadingRepoBinding) {
    return (
      <div className='p-6'>
        <Skeleton className='h-96 w-full' />
      </div>
    )
  }

  return (
    <Page>
      <PageHeader
        className='max-w-4xl'
        title='Edit Agent'
        subtitle='Use the guided form for common settings or JSON mode for advanced agent topology.'
        breadcrumb={
          <Link to='/agents' className='inline-flex min-w-0 items-center gap-1.5 hover:text-foreground'>
            <ArrowLeft className='size-3.5 shrink-0' />
            <span>Agents</span>
            <ChevronRight className='size-3 shrink-0' />
            <span className='truncate text-foreground'>{data?.agent?.name ?? agentRef}</span>
          </Link>
        }
      />

      <PageScroll className='max-w-4xl'>
        <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList className='mb-4'>
          <TabsTrigger value='form'>Form</TabsTrigger>
          <TabsTrigger value='json'>JSON</TabsTrigger>
        </TabsList>

        <TabsContent value='form'>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onFormSubmit)} className='space-y-6'>
              <Card>
                <CardHeader>
                  <CardTitle>Basic Info</CardTitle>
                  <CardDescription>Update the visible description and orchestration type.</CardDescription>
                </CardHeader>
                <CardContent className='space-y-4'>
                  <FormField control={form.control} name='name' render={({ field }) => (
                    <FormItem>
                      <FormLabel>Name</FormLabel>
                      <FormControl><Input {...field} disabled /></FormControl>
                    </FormItem>
                  )} />
                  {data?.agent?.agent_id && (
                    <div className='space-y-2'>
                      <Label htmlFor='agent-id-readonly'>Agent ID</Label>
                      <Input id='agent-id-readonly' value={data.agent.agent_id} disabled className='font-mono' />
                      <p className='text-xs text-muted-foreground'>
                        The Agent ID is immutable and can never be changed.
                      </p>
                    </div>
                  )}
                  <FormField control={form.control} name='description' render={({ field }) => (
                    <FormItem>
                      <FormLabel>Description</FormLabel>
                      <FormControl><Input {...field} /></FormControl>
                    </FormItem>
                  )} />
                  <FormField control={form.control} name='type' render={({ field }) => (
                    <FormItem>
                      <FormLabel>Type</FormLabel>
                      <Select onValueChange={field.onChange} value={field.value}>
                        <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                        <SelectContent>
                          <SelectItem value='AGENT_TYPE_LLM'>LLM</SelectItem>
                          <SelectItem value='AGENT_TYPE_LOOP'>Loop</SelectItem>
                          <SelectItem value='AGENT_TYPE_SEQUENTIAL'>Sequential</SelectItem>
                          <SelectItem value='AGENT_TYPE_PARALLEL'>Parallel</SelectItem>
                        </SelectContent>
                      </Select>
                    </FormItem>
                  )} />
                  <div className='space-y-3 border-t pt-4'>
                    <div>
                      <h3 className='text-sm font-medium'>External Access</h3>
                      <p className='text-xs text-muted-foreground'>Choose which external protocols may invoke this agent.</p>
                    </div>
                    <FormField control={form.control} name='enable_a2a' render={({ field }) => (
                      <FormItem className='flex items-center justify-between gap-4 rounded-md border px-3 py-3'>
                        <div className='space-y-0.5'>
                          <FormLabel>A2A</FormLabel>
                          <p className='text-xs text-muted-foreground'>Expose this agent through the A2A endpoint.</p>
                        </div>
                        <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                      </FormItem>
                    )} />
                    <FormField control={form.control} name='enable_openai_api' render={({ field }) => (
                      <FormItem className='flex items-center justify-between gap-4 rounded-md border px-3 py-3'>
                        <div className='space-y-0.5'>
                          <FormLabel>OpenAI API</FormLabel>
                          <p className='text-xs text-muted-foreground'>Expose this agent as a model through the OpenAI-compatible API.</p>
                        </div>
                        <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                      </FormItem>
                    )} />
                    <FormField control={form.control} name='enable_agui' render={({ field }) => (
                      <FormItem className='flex items-center justify-between gap-4 rounded-md border px-3 py-3'>
                        <div className='space-y-0.5'>
                          <FormLabel>AG-UI</FormLabel>
                          <p className='text-xs text-muted-foreground'>Expose this agent to AG-UI frontends over the streaming protocol endpoint.</p>
                        </div>
                        <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                      </FormItem>
                    )} />
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Icon</CardTitle>
                  <CardDescription>Keep the agent visually identifiable in chat, sessions, and lists.</CardDescription>
                </CardHeader>
                <CardContent>
                  <AgentIconUpload
                    agentName={agentName}
                    value={iconUrl}
                    onChange={(url) => form.setValue('icon_url', url, { shouldDirty: true })}
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Model Configuration</CardTitle>
                  <CardDescription>Adjust the model alias and system instruction used for LLM responses.</CardDescription>
                </CardHeader>
                <CardContent className='space-y-4'>
                  <FormField control={form.control} name='model' render={({ field }) => (
                    <FormItem>
                      <FormLabel>Model</FormLabel>
                      <AgentModelSelect value={field.value} onChange={field.onChange} />
                      <p className='text-xs text-muted-foreground'>
                        Models are loaded from configured model providers. Agents use the model alias when available.
                      </p>
                    </FormItem>
                  )} />
                  <FormField control={form.control} name='instruction' render={({ field }) => (
                    <FormItem>
                      <FormLabel>Instruction</FormLabel>
                      <FormControl><Textarea rows={5} {...field} /></FormControl>
                    </FormItem>
                  )} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>MCP Servers</CardTitle>
                  <CardDescription>Shared tool servers available to this agent. Inline servers remain in JSON mode.</CardDescription>
                </CardHeader>
                <CardContent className='space-y-3'>
                  <p className='text-sm text-muted-foreground'>
                    Select shared MCP servers this agent can use. Inline MCP servers can still be managed in JSON mode.
                  </p>
                  <FormField control={form.control} name='mcp_server_ids' render={({ field }) => {
                    const selected = field.value ?? []
                    const servers = mcpData?.mcp_servers ?? []
                    const toggle = (id: string) => {
                      field.onChange(
                        selected.includes(id)
                          ? selected.filter((selectedId) => selectedId !== id)
                          : [...selected, id],
                      )
                    }

                    if (servers.length === 0) {
                      return <p className='text-sm text-muted-foreground'>No shared MCP servers configured yet.</p>
                    }

                    return (
                      <FormItem>
                        <div className='grid gap-2 md:grid-cols-2'>
                          {servers.map((server) => {
                            const id = server.id ?? ''
                            const isSelected = selected.includes(id)
                            return (
                              <button
                                key={id || server.name}
                                type='button'
                                disabled={!id}
                                onClick={() => id && toggle(id)}
                                className={`rounded-md border p-3 text-left transition-colors ${
                                  isSelected ? 'border-primary bg-primary/10' : 'hover:bg-muted'
                                } ${!id ? 'cursor-not-allowed opacity-60' : ''}`}
                              >
                                <div className='flex items-center justify-between gap-2'>
                                  <span className='font-medium'>{server.name}</span>
                                  {isSelected && <Badge>Selected</Badge>}
                                </div>
                                <div className='mt-1 text-xs text-muted-foreground'>
                                  {enumLabel(server.transport, 'Transport unspecified')}{server.url ? ` · ${server.url}` : ''}
                                </div>
                              </button>
                            )
                          })}
                        </div>
                      </FormItem>
                    )
                  }} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Remote Agents</CardTitle>
                  <CardDescription>Daemon and A2A agents this agent can delegate work to.</CardDescription>
                </CardHeader>
                <CardContent>
                  <FormField control={form.control} name='remote_agent_ids' render={({ field }) => (
                    <FormItem>
                      <AgentRemoteAgentsField
                        value={field.value}
                        onChange={field.onChange}
                        remoteAgents={remoteData?.remote_agents}
                        isLoading={isLoadingRemoteAgents}
                      />
                    </FormItem>
                  )} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Agent Files</CardTitle>
                  <CardDescription>Review file-space mounts and permissions before saving.</CardDescription>
                </CardHeader>
                <CardContent className='space-y-3'>
                  <p className='text-sm text-muted-foreground'>
                    Mount workspace file spaces into this agent's built-in agent_files tools.
                  </p>
                  <FormField control={form.control} name='file_mounts' render={({ field }) => (
                    <FormItem>
                      <AgentFileMountsField value={field.value} onChange={field.onChange} />
                    </FormItem>
                  )} />
                </CardContent>
              </Card>

              {/* Sub-agents - read-only list */}
              {data?.agent?.sub_agents && data.agent.sub_agents.length > 0 && (
                <Card>
                  <CardHeader><CardTitle>Sub-Agents (read-only, edit in JSON mode)</CardTitle></CardHeader>
                  <CardContent className='space-y-2'>
                    {data.agent.sub_agents.map((sa) => (
                      <div key={sa.name} className='flex items-center gap-2'>
                        <span className='text-sm font-medium'>{sa.name}</span>
                        <Badge variant='outline'>{AGENT_TYPE_LABELS[sa.type ?? 'AGENT_TYPE_UNSPECIFIED']}</Badge>
                      </div>
                    ))}
                  </CardContent>
                </Card>
              )}

              <PageActions>
                <Button variant='outline' asChild>
                  <Link to='/agents'>Cancel</Link>
                </Button>
                <Button type='submit' disabled={updateMutation.isPending}>
                  {updateMutation.isPending ? 'Saving...' : 'Save'}
                </Button>
              </PageActions>
            </form>
          </Form>
        </TabsContent>

        <TabsContent value='json'>
          <Card>
            <CardContent className='pt-6'>
              <Editor
                height='500px'
                language='json'
                theme={resolvedTheme === 'dark' ? 'vs-dark' : 'light'}
                value={jsonValue ?? initialJsonValue}
                onChange={(v) => setJsonValue(v ?? '')}
                options={{ minimap: { enabled: false }, formatOnPaste: true }}
              />
            </CardContent>
          </Card>
          <PageActions>
            <Button variant='outline' asChild>
              <Link to='/agents'>Cancel</Link>
            </Button>
            <Button onClick={onJsonSubmit} disabled={updateMutation.isPending}>
              {updateMutation.isPending ? 'Saving...' : 'Save'}
            </Button>
          </PageActions>
        </TabsContent>
        </Tabs>
      </PageScroll>
    </Page>
  )
}
