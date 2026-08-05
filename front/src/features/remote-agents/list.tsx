import { useState } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useDeleteRemoteAgent, useRemoteAgents, useRemoteAgentStatus } from '@/api/remote-agents'
import { DataTable, type Column } from '@/components/data-table'
import { DeleteDialog } from '@/components/delete-dialog'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { StatusBadge, type RunStatus } from '@/components/butter/primitives'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  MoreVertical,
  Pencil,
  Trash2,
  Link as LinkIcon,
  Cpu,
  ShieldCheck,
  Plus,
} from 'lucide-react'
import type { RemoteAgent, RemoteAgentState } from '@/types/api'
import { enumLabel } from '@/lib/constants'

const STATE_BADGE: Record<RemoteAgentState, { status: RunStatus; label: string }> = {
  STATE_UNSPECIFIED: { status: 'never', label: 'Unknown' },
  STATE_CONFIGURED: { status: 'disabled', label: 'Configured' },
  STATE_ACTIVE: { status: 'running', label: 'Active' },
  STATE_IDLE: { status: 'success', label: 'Idle' },
  STATE_UNREACHABLE: { status: 'failed', label: 'Unreachable' },
  STATE_ERROR: { status: 'failed', label: 'Error' },
}

function RemoteAgentStatusBadge({ id }: { id: string }) {
  const { data, isLoading } = useRemoteAgentStatus(id)
  if (isLoading || !data) return <Badge variant='outline' className='text-xs'>…</Badge>
  const state = (data.status.state ?? 'STATE_UNSPECIFIED') as RemoteAgentState
  const badge = STATE_BADGE[state]
  return <StatusBadge status={badge.status} label={badge.label} />
}

export function RemoteAgentList() {
  const { data, isLoading } = useRemoteAgents()
  const deleteMutation = useDeleteRemoteAgent()
  const navigate = useNavigate()
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  const columns: Column<RemoteAgent>[] = [
    {
      header: 'Remote Agent',
      cell: (row) => {
        const isDaemon = row.protocol === 'REMOTE_AGENT_PROTOCOL_DAEMON'
        const Icon = isDaemon ? Cpu : LinkIcon
        return (
          <div className='flex items-center gap-2'>
            <Icon className='h-4 w-4 text-muted-foreground' />
            <div>
              <div className='font-medium'>{row.name}</div>
              <div className='text-xs text-muted-foreground line-clamp-1 max-w-md'>
                {isDaemon
                  ? `${row.daemon_runtime_id ?? '-'} / ${row.acp_runtime ?? '-'}`
                  : row.url}
              </div>
            </div>
          </div>
        )
      },
    },
    {
      header: 'Protocol',
      cell: (row) => (
        <Badge variant='outline' className='font-mono text-[10px]'>
          {enumLabel(row.protocol, 'Unknown')}
        </Badge>
      ),
    },
    {
      header: 'Status',
      cell: (row) => <RemoteAgentStatusBadge id={row.id} />,
    },
    {
      header: 'Verified',
      cell: (row) =>
        row.protocol === 'REMOTE_AGENT_PROTOCOL_DAEMON' ? (
          <ShieldCheck className='h-4 w-4 text-success' />
        ) : (
          <span className='text-xs text-muted-foreground'>-</span>
        ),
    },
    {
      header: '',
      cell: (row) => (
        <div className='flex justify-end'>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant='ghost' size='icon'>
                <MoreVertical className='h-4 w-4' />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align='end'>
              <DropdownMenuItem
                onClick={() =>
                  navigate({ to: '/remote-agents/$id/edit', params: { id: row.id } })
                }
              >
                <Pencil className='mr-2 h-4 w-4' /> Edit
              </DropdownMenuItem>
              <DropdownMenuItem className='text-destructive' onClick={() => setDeleteTarget(row.id)}>
                <Trash2 className='mr-2 h-4 w-4' /> Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      ),
    },
  ]

  return (
    <Page>
      <PageHeader
        title='Remote Agents'
        subtitle='External orchestrators and autonomous daemon instances.'
        actions={
          <Button size='sm' asChild>
            <Link to='/remote-agents/create'>
              <Plus className='size-4' />
              Register Agent
            </Link>
          </Button>
        }
      />
      <PageScroll>
        <DataTable
          columns={columns}
          data={data?.remote_agents}
          isLoading={isLoading}
          emptyMessage='No remote agents registered.'
        />
      </PageScroll>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title='Delete Remote Agent'
        description='Delete this remote agent? This action cannot be undone.'
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => {
                toast.success('Remote agent deleted')
                setDeleteTarget(null)
              },
              onError: (err) => toast.error(err.message),
            })
          }
        }}
      />
    </Page>
  )
}
