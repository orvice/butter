import { useEffect, useState } from 'react'
import { Link, useNavigate, useSearch } from '@tanstack/react-router'
import { ArrowLeft, ChevronRight, TriangleAlert } from 'lucide-react'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { useCreateAgent } from '@/api/agents'
import { useMCPServers } from '@/api/mcp-servers'
import { useRemoteAgents } from '@/api/remote-agents'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { AgentModelSelect } from './model-select'
import { AgentIconUpload } from './icon-upload'
import { AgentFileMountsField } from './file-mounts-field'
import { AgentRemoteAgentsField } from './remote-agents-field'
import { PiAgentConfigurationCard } from './pi-agent-fields'
import { CursorAgentConfigurationCard } from './cursor-agent-fields'
import { ContextGuardConfigurationCard } from './context-guard-fields'
import {
  buildContextGuardConfig,
  contextGuardFormSchema,
  EMPTY_CONTEXT_GUARD_FORM_VALUES,
  supportsContextGuard,
} from './context-guard-config'
import {
  asPiAgent,
  EMPTY_PI_AGENT_FORM_VALUES,
  piAgentFormSchema,
  type PiAgentFormValues,
  validatePiAgentForm,
} from './pi-config'
import {
  asCursorAgent,
  cursorAgentFormSchema,
  EMPTY_CURSOR_AGENT_FORM_VALUES,
  type CursorAgentFormValues,
  validateCursorAgentForm,
} from './cursor-config'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Page, PageActions, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { enumLabel } from '@/lib/constants'
import { suggestAgentID, validateAgentID } from './agent-id'
import type { AgentFileMountPermission, AgentType } from '@/types/api'

const MOUNT_PERMISSIONS = [
  'AGENT_FILE_MOUNT_PERMISSION_READ',
  'AGENT_FILE_MOUNT_PERMISSION_READ_WRITE',
  'AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE',
] as const satisfies readonly AgentFileMountPermission[]

const agentSchema = z.object({
  name: z.string().min(1, 'Name is required').refine((v) => v !== 'user', "Name cannot be 'user'"),
  agent_id: z.string().superRefine((value, ctx) => {
    const error = validateAgentID(value)
    if (error) ctx.addIssue({ code: 'custom', message: error })
  }),
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
  context_guard: contextGuardFormSchema,
  pi: piAgentFormSchema,
  cursor: cursorAgentFormSchema,
}).superRefine((values, ctx) => {
  if (values.type === 'AGENT_TYPE_PI') {
    validatePiAgentForm(values.pi, ctx)
  }
  if (values.type === 'AGENT_TYPE_CURSOR') {
    validateCursorAgentForm(values.cursor, ctx)
  }
})

type AgentFormValues = z.infer<typeof agentSchema>

function isBoxBackedLeaf(type: string) {
  return type === 'AGENT_TYPE_PI' || type === 'AGENT_TYPE_CURSOR'
}

export function AgentCreate() {
  const navigate = useNavigate()
  const { remote_agent_id } = useSearch({ from: '/_authenticated/agents/create' })
  const initialRemoteAgentId = remote_agent_id ?? ''
  const createMutation = useCreateAgent()
  const { data: mcpData, isLoading: isLoadingMCPServers } = useMCPServers()
  const { data: remoteData, isLoading: isLoadingRemoteAgents } = useRemoteAgents()
  const [operation, setOperation] = useState<{ request: string; id: string } | null>(null)

  const form = useForm<AgentFormValues>({
    resolver: zodResolver(agentSchema),
    defaultValues: {
      name: '',
      agent_id: '',
      description: '',
      type: 'AGENT_TYPE_LLM',
      enable_a2a: false,
      enable_openai_api: false,
      enable_agui: false,
      model: '',
      instruction: '',
      mcp_server_ids: [],
      remote_agent_ids: initialRemoteAgentId ? [initialRemoteAgentId] : [],
      file_mounts: [],
      icon_url: '',
      context_guard: { ...EMPTY_CONTEXT_GUARD_FORM_VALUES },
      pi: { ...EMPTY_PI_AGENT_FORM_VALUES },
      cursor: { ...EMPTY_CURSOR_AGENT_FORM_VALUES },
    },
  })
  const agentName = useWatch({ control: form.control, name: 'name' })
  const iconUrl = useWatch({ control: form.control, name: 'icon_url' })
  const agentType = useWatch({ control: form.control, name: 'type' })
  const contextGuardValues = useWatch({ control: form.control, name: 'context_guard' })
  const piValues = useWatch({ control: form.control, name: 'pi' })
  const cursorValues = useWatch({ control: form.control, name: 'cursor' })

  useEffect(() => {
    if (supportsContextGuard(agentType)) return
    if (form.getValues('context_guard').mode === 'off') return
    form.setValue('context_guard', { ...EMPTY_CONTEXT_GUARD_FORM_VALUES }, {
      shouldDirty: true,
      shouldValidate: true,
    })
  }, [agentType, form])

  // Suggest the slug from the name as the user types, until the user edits
  // the Agent ID field themselves.
  const [agentIdTouched, setAgentIdTouched] = useState(false)
  useEffect(() => {
    if (agentIdTouched) return
    form.setValue('agent_id', suggestAgentID(agentName ?? ''))
  }, [agentIdTouched, agentName, form])

  function onSubmit(values: AgentFormValues) {
    const isBoxBacked = isBoxBackedLeaf(values.type)
    const baseAgent = {
      name: values.name,
      agent_id: values.agent_id,
      description: values.description,
      type: values.type as AgentType,
      enable_a2a: values.enable_a2a,
      enable_openai_api: values.enable_openai_api,
      enable_agui: values.enable_agui,
      metadata: values.icon_url ? { icon_url: values.icon_url } : undefined,
      config: {
        model: values.model,
        instruction: values.instruction,
        mcp_server_ids: values.mcp_server_ids ?? [],
        remote_agent_ids: values.remote_agent_ids ?? [],
        file_mounts: values.file_mounts ?? [],
        context_guard: supportsContextGuard(values.type)
          ? buildContextGuardConfig(values.context_guard)
          : undefined,
      },
    }
    const agent = values.type === 'AGENT_TYPE_PI'
      ? asPiAgent(baseAgent, values.pi)
      : values.type === 'AGENT_TYPE_CURSOR'
        ? asCursorAgent(baseAgent, values.cursor)
        : baseAgent
    const initialContent = {
      description: values.description ?? '',
      prompt: isBoxBacked ? '' : (values.instruction ?? ''),
      global_prompt: '',
    }
    const request = JSON.stringify({ agent, initialContent })
    let requestOperation = operation
    if (!requestOperation || requestOperation.request !== request) {
      requestOperation = { request, id: crypto.randomUUID() }
      setOperation(requestOperation)
    }

    createMutation.mutate(
      {
        agent,
        initial_content: initialContent,
        operation_id: requestOperation.id,
      },
      {
        onSuccess: () => { toast.success('Agent created'); navigate({ to: '/agents' }) },
        onError: (err) => toast.error(err.message),
      },
    )
  }

  return (
    <Page>
      <PageHeader
        className='max-w-4xl'
        title='Create Agent'
        subtitle='Start with identity and model settings, then optionally connect tools and file spaces.'
        breadcrumb={
          <Link to='/agents' className='inline-flex items-center gap-1.5 hover:text-foreground'>
            <ArrowLeft className='size-3.5' />
            Agents
            <ChevronRight className='size-3' />
            <span className='text-foreground'>Create</span>
          </Link>
        }
      />

      <PageScroll className='max-w-4xl'>
        <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
          <Card>
            <CardHeader>
              <CardTitle>Basic Info</CardTitle>
              <CardDescription>Name and describe how this agent appears across the workspace.</CardDescription>
            </CardHeader>
            <CardContent className='space-y-4'>
              <FormField control={form.control} name='name' render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl><Input placeholder='My Agent' {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name='agent_id' render={({ field }) => (
                <FormItem>
                  <FormLabel>Agent ID</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='my-agent'
                      autoComplete='off'
                      spellCheck={false}
                      className='font-mono'
                      {...field}
                      onChange={(e) => {
                        setAgentIdTouched(true)
                        field.onChange(e)
                      }}
                    />
                  </FormControl>
                  <p className='text-xs text-muted-foreground'>
                    1–64 lowercase letters, digits, or hyphens. Must be unique in this workspace.
                  </p>
                  <FormMessage />
                </FormItem>
              )} />
              <div className='flex items-start gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 p-3 text-xs leading-relaxed text-amber-700 dark:text-amber-400'>
                <TriangleAlert className='mt-0.5 size-4 shrink-0' />
                <span>
                  The Agent ID is <strong>immutable</strong> — once assigned it can never be changed or
                  reused. Choose carefully before creating.
                </span>
              </div>
              <FormField control={form.control} name='description' render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl><Input placeholder='A helpful assistant' {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name='type' render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      <SelectItem value='AGENT_TYPE_LLM'>LLM</SelectItem>
                      <SelectItem value='AGENT_TYPE_LOOP'>Loop</SelectItem>
                      <SelectItem value='AGENT_TYPE_SEQUENTIAL'>Sequential</SelectItem>
                      <SelectItem value='AGENT_TYPE_PARALLEL'>Parallel</SelectItem>
                      <SelectItem value='AGENT_TYPE_PI'>Pi</SelectItem>
                      <SelectItem value='AGENT_TYPE_CURSOR'>Cursor</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
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
              <CardDescription>Upload or link an avatar so the agent is easier to recognize in lists and chat.</CardDescription>
            </CardHeader>
            <CardContent>
              <AgentIconUpload
                agentName={agentName}
                value={iconUrl}
                onChange={(url) => form.setValue('icon_url', url, { shouldDirty: true })}
              />
            </CardContent>
          </Card>

          {agentType === 'AGENT_TYPE_PI' ? (
            <PiAgentConfigurationCard
              value={(piValues ?? EMPTY_PI_AGENT_FORM_VALUES) as PiAgentFormValues}
              onChange={(field, value) => form.setValue(
                'pi',
                { ...form.getValues('pi'), [field]: value },
                { shouldDirty: true, shouldValidate: true },
              )}
              errors={{
                butterboxId: form.formState.errors.pi?.butterboxId?.message,
                maxRunSeconds: form.formState.errors.pi?.maxRunSeconds?.message,
              }}
            />
          ) : agentType === 'AGENT_TYPE_CURSOR' ? (
            <CursorAgentConfigurationCard
              value={(cursorValues ?? EMPTY_CURSOR_AGENT_FORM_VALUES) as CursorAgentFormValues}
              onChange={(field, value) => form.setValue(
                'cursor',
                { ...form.getValues('cursor'), [field]: value },
                { shouldDirty: true, shouldValidate: true },
              )}
              errors={{
                butterboxId: form.formState.errors.cursor?.butterboxId?.message,
                maxRunSeconds: form.formState.errors.cursor?.maxRunSeconds?.message,
                mode: form.formState.errors.cursor?.mode?.message,
              }}
            />
          ) : (
            <>
          <Card>
            <CardHeader>
              <CardTitle>Model Configuration</CardTitle>
              <CardDescription>Pick the model and instruction the agent will use for LLM responses.</CardDescription>
            </CardHeader>
            <CardContent className='space-y-4'>
              <FormField control={form.control} name='model' render={({ field }) => (
                <FormItem>
                  <FormLabel>Model</FormLabel>
                  <AgentModelSelect value={field.value} onChange={field.onChange} />
                  <p className='text-xs text-muted-foreground'>
                    Models are loaded from configured model providers. Agents use the model alias when available.
                  </p>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name='instruction' render={({ field }) => (
                <FormItem>
                  <FormLabel>Instruction</FormLabel>
                  <FormControl><Textarea placeholder='You are a helpful assistant...' rows={5} {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
            </CardContent>
          </Card>

          {supportsContextGuard(agentType) && (
            <ContextGuardConfigurationCard
              value={contextGuardValues ?? EMPTY_CONTEXT_GUARD_FORM_VALUES}
              onChange={(value) => form.setValue('context_guard', value, {
                shouldDirty: true,
                shouldValidate: true,
              })}
              errors={{
                mode: form.formState.errors.context_guard?.mode?.message,
                maxTokens: form.formState.errors.context_guard?.maxTokens?.message,
                maxTurns: form.formState.errors.context_guard?.maxTurns?.message,
              }}
            />
          )}

          <Card>
            <CardHeader>
              <CardTitle>MCP Servers</CardTitle>
              <CardDescription>Optional tools the agent can call during runs.</CardDescription>
            </CardHeader>
            <CardContent className='space-y-3'>
              <p className='text-sm text-muted-foreground'>
                Select shared MCP servers this agent can use. Leave empty to disable shared MCP tools.
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

                if (isLoadingMCPServers) {
                  return <p className='text-sm text-muted-foreground'>Loading MCP servers...</p>
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
                  <FormMessage />
                </FormItem>
              )} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Agent Files</CardTitle>
              <CardDescription>Mount file spaces only when the agent needs persistent workspace files.</CardDescription>
            </CardHeader>
            <CardContent className='space-y-3'>
              <p className='text-sm text-muted-foreground'>
                Mount workspace file spaces into this agent's built-in agent_files tools.
              </p>
              <FormField control={form.control} name='file_mounts' render={({ field }) => (
                <FormItem>
                  <AgentFileMountsField value={field.value} onChange={field.onChange} />
                  <FormMessage />
                </FormItem>
              )} />
            </CardContent>
          </Card>

            </>
          )}

          <PageActions>
            <Button variant='outline' asChild>
              <Link to='/agents'>Cancel</Link>
            </Button>
            <Button type='submit' disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Creating...' : 'Create Agent'}
            </Button>
          </PageActions>
        </form>
        </Form>
      </PageScroll>
    </Page>
  )
}
