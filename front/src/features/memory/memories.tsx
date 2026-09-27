import { useState, type FormEvent } from 'react'
import type { WorkspaceMemory } from '@/gen/agents/v1/workspace_memory_pb'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import { Search, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { useAgents } from '@/api/agents'
import {
  useDeleteWorkspaceMemory,
  useSearchWorkspaceMemories,
  useWorkspaceMemories,
  type MemoryScopeSelection,
} from '@/api/workspace-memory'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { DeleteDialog } from '@/components/delete-dialog'

type ScopeTab = 'workspace' | 'agent'

export function MemoriesSection({ canManage }: { canManage: boolean }) {
  const [tab, setTab] = useState<ScopeTab>('workspace')
  const [agentId, setAgentId] = useState('')
  const [draft, setDraft] = useState('')
  const [query, setQuery] = useState('')
  const [pendingDelete, setPendingDelete] = useState<WorkspaceMemory | null>(
    null
  )

  const { data: agentsData } = useAgents({}, { enabled: tab === 'agent' })
  const agents = (agentsData?.agents ?? []).filter((a) => a.agent_id)

  const selection: MemoryScopeSelection =
    tab === 'agent' ? { scope: 'agent', agentId } : { scope: 'workspace' }
  const searching = query.trim() !== ''
  const list = useWorkspaceMemories(selection, !searching)
  const search = useSearchWorkspaceMemories(selection, query)
  const deleteMutation = useDeleteWorkspaceMemory()

  const active = searching ? search : list
  const memories: WorkspaceMemory[] =
    (searching ? search.data : list.data?.memories) ?? []
  const needsAgent = tab === 'agent' && !agentId

  function onSearch(e: FormEvent) {
    e.preventDefault()
    setQuery(draft)
  }

  function clearSearch() {
    setDraft('')
    setQuery('')
  }

  function confirmDelete() {
    if (!pendingDelete) return
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        toast.success('Memory deleted')
        setPendingDelete(null)
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Memories</CardTitle>
        <CardDescription>
          What the mem0 server holds for this workspace. Memories are only ever
          added, so outdated facts and near-duplicates accumulate
          {canManage
            ? ' — delete the ones that are wrong.'
            : '; workspace owners and admins can delete them.'}
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <div className='flex flex-col gap-3 sm:flex-row sm:items-center'>
          <Tabs value={tab} onValueChange={(v) => setTab(v as ScopeTab)}>
            <TabsList>
              <TabsTrigger value='workspace'>Workspace</TabsTrigger>
              <TabsTrigger value='agent'>Agent</TabsTrigger>
            </TabsList>
          </Tabs>
          {tab === 'agent' && (
            <Select value={agentId} onValueChange={setAgentId}>
              <SelectTrigger className='w-full sm:w-64' aria-label='Agent'>
                <SelectValue placeholder='Choose an agent' />
              </SelectTrigger>
              <SelectContent>
                {agents.map((a) => (
                  <SelectItem key={a.agent_id} value={a.agent_id!}>
                    {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
          <form onSubmit={onSearch} className='flex flex-1 gap-2'>
            <Input
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              placeholder='Search memories'
              aria-label='Search memories'
            />
            <Button
              type='submit'
              variant='outline'
              size='icon'
              className='size-9 shrink-0'
              aria-label='Search'
            >
              <Search className='size-4' />
            </Button>
            {searching && (
              <Button
                type='button'
                variant='ghost'
                size='icon'
                className='size-9 shrink-0'
                aria-label='Clear search'
                onClick={clearSearch}
              >
                <X className='size-4' />
              </Button>
            )}
          </form>
        </div>

        {needsAgent ? (
          <p className='text-sm text-muted-foreground'>
            Choose an agent to see its Agent Memory.
          </p>
        ) : active.isLoading ? (
          <Skeleton className='h-40 w-full' />
        ) : active.error ? (
          <p className='text-sm text-destructive' role='alert'>
            {active.error.message}
          </p>
        ) : memories.length === 0 ? (
          <p className='text-sm text-muted-foreground'>
            {searching ? 'No memories match this search.' : 'No memories yet.'}
          </p>
        ) : (
          <MemoryTable
            memories={memories}
            canManage={canManage}
            onDelete={setPendingDelete}
          />
        )}

        {!searching && list.data?.truncated && (
          <p className='text-xs text-muted-foreground'>
            Showing the newest {list.data.limit} memories. Search to find older
            ones.
          </p>
        )}
      </CardContent>
      <DeleteDialog
        open={pendingDelete !== null}
        onOpenChange={(open) => !open && setPendingDelete(null)}
        title='Delete this memory?'
        description={`"${pendingDelete?.memory ?? ''}" is removed from the mem0 server and will no longer be recalled.`}
        confirmLabel='Delete'
        loadingLabel='Deleting...'
        loading={deleteMutation.isPending}
        onConfirm={confirmDelete}
      />
    </Card>
  )
}

function MemoryTable({
  memories,
  canManage,
  onDelete,
}: {
  memories: WorkspaceMemory[]
  canManage: boolean
  onDelete: (m: WorkspaceMemory) => void
}) {
  // The server only returns principals to owners/admins.
  const showPrincipal = memories.some((m) => m.principal)
  return (
    <div className='overflow-x-auto rounded-md border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Memory</TableHead>
            <TableHead className='w-28'>Date</TableHead>
            <TableHead className='w-40'>Source</TableHead>
            {showPrincipal && <TableHead className='w-32'>From</TableHead>}
            {canManage && (
              <TableHead className='w-12'>
                <span className='sr-only'>Actions</span>
              </TableHead>
            )}
          </TableRow>
        </TableHeader>
        <TableBody>
          {memories.map((m) => (
            <TableRow key={m.id}>
              <TableCell className='whitespace-normal'>{m.memory}</TableCell>
              <TableCell className='text-xs text-muted-foreground'>
                {m.createdAt
                  ? timestampDate(m.createdAt).toLocaleDateString()
                  : '-'}
              </TableCell>
              <TableCell>
                <div className='flex flex-wrap gap-1'>
                  {m.agentId && <Badge variant='secondary'>{m.agentId}</Badge>}
                  {m.channel && <Badge variant='outline'>{m.channel}</Badge>}
                </div>
              </TableCell>
              {showPrincipal && (
                <TableCell className='font-mono text-xs text-muted-foreground'>
                  {m.principal || '-'}
                </TableCell>
              )}
              {canManage && (
                <TableCell>
                  <Button
                    variant='ghost'
                    size='icon'
                    className='size-8'
                    aria-label={`Delete memory: ${m.memory}`}
                    onClick={() => onDelete(m)}
                  >
                    <Trash2 className='size-4' />
                  </Button>
                </TableCell>
              )}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
