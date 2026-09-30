import { useState } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import type { LinearApp } from '@/gen/agents/v1/linear_pb'
import {
  ListTree,
  MoreVertical,
  Pencil,
  Plus,
  SquareKanban,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
import { useAgents } from '@/api/agents'
import { useDeleteLinearApp, useLinearApps } from '@/api/linear'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { DataTable, type Column } from '@/components/data-table'
import { DeleteDialog } from '@/components/delete-dialog'
import { CredentialStateBadge } from './shared'

export function LinearAppList() {
  const { data, isLoading } = useLinearApps()
  const { data: agentsData } = useAgents()
  const deleteMutation = useDeleteLinearApp()
  const navigate = useNavigate()
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  const agentName = (agentId: string) =>
    agentsData?.agents?.find((agent) => agent.agent_id === agentId)?.name ??
    agentId

  const columns: Column<LinearApp>[] = [
    {
      header: 'Linear App',
      cell: (row) => (
        <div className='flex items-center gap-2'>
          <SquareKanban className='h-4 w-4 text-muted-foreground' />
          <div>
            <div className='font-medium'>{row.displayName}</div>
            <div className='line-clamp-1 max-w-md font-mono text-xs text-muted-foreground'>
              {row.clientId}
            </div>
          </div>
        </div>
      ),
    },
    { header: 'Agent', cell: (row) => agentName(row.agentId) },
    {
      header: 'Inbound',
      cell: (row) =>
        row.inboundEnabled ? (
          <Badge variant='outline'>Enabled</Badge>
        ) : (
          <Badge variant='outline' className='text-muted-foreground'>
            Disabled
          </Badge>
        ),
    },
    {
      header: 'Secrets',
      cell: (row) => <CredentialStateBadge state={row.credentialState} />,
    },
    {
      header: '',
      cell: (row) => (
        <div className='flex justify-end'>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                aria-label={`Actions for ${row.displayName}`}
              >
                <MoreVertical className='h-4 w-4' />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align='end'>
              <DropdownMenuItem
                onClick={() =>
                  navigate({
                    to: '/linear-apps/$id/edit',
                    params: { id: row.id },
                  })
                }
              >
                <Pencil className='mr-2 h-4 w-4' /> Edit
              </DropdownMenuItem>
              <DropdownMenuItem
                className='text-destructive'
                onClick={() => setDeleteTarget(row.id)}
              >
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
        title='Linear Apps'
        subtitle='Let people delegate Linear issues to an Agent, or mention it, and get the answer in Linear.'
        actions={
          <div className='flex gap-2'>
            <Button size='sm' variant='outline' asChild>
              <Link to='/linear-deliveries'>
                <ListTree className='size-4' />
                Deliveries
              </Link>
            </Button>
            <Button size='sm' asChild>
              <Link to='/linear-apps/create'>
                <Plus className='size-4' />
                Register Linear App
              </Link>
            </Button>
          </div>
        }
      />
      <PageScroll>
        <DataTable
          columns={columns}
          data={data}
          isLoading={isLoading}
          emptyMessage='No Linear Apps registered.'
        />
      </PageScroll>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title='Delete Linear App'
        description='Delete this Linear App and its installations? Linear sessions will stop being answered.'
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return
          deleteMutation.mutate(deleteTarget, {
            onSuccess: () => {
              toast.success('Linear App deleted')
              setDeleteTarget(null)
            },
            onError: (err) => toast.error(err.message),
          })
        }}
      />
    </Page>
  )
}
