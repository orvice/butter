import { useState } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import type { PlaneApp } from '@/gen/agents/v1/plane_pb'
import { MoreVertical, Pencil, Plane, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useAgents } from '@/api/agents'
import { useDeletePlaneApp, usePlaneApps } from '@/api/plane'
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
import { PlaneCredentialBadge } from './shared'

export function PlaneAppList() {
  const { data, isLoading } = usePlaneApps()
  const { data: agentsData } = useAgents()
  const deleteMutation = useDeletePlaneApp()
  const navigate = useNavigate()
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  const agentName = (agentId: string) =>
    agentsData?.agents?.find((agent) => agent.agent_id === agentId)?.name ??
    agentId

  const columns: Column<PlaneApp>[] = [
    {
      header: 'Plane App',
      cell: (row) => (
        <div className='flex items-center gap-2'>
          <Plane className='h-4 w-4 text-muted-foreground' />
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
      header: 'Provider',
      cell: (row) => (
        <Badge variant='outline' className='font-mono'>
          {row.providerProfile}
        </Badge>
      ),
    },
    {
      header: 'Inbound',
      cell: () => (
        <Badge variant='outline' className='text-muted-foreground'>
          Disabled
        </Badge>
      ),
    },
    {
      header: 'Secrets',
      cell: (row) => <PlaneCredentialBadge state={row.credentialState} />,
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
                    to: '/plane-apps/$id/edit',
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
        title='Plane Apps'
        subtitle='Register a Plane app and bind it to an Agent. This version registers the resource only; prompt delivery ships with the execution path.'
        actions={
          <Button size='sm' asChild>
            <Link to='/plane-apps/create'>
              <Plus className='size-4' />
              Register Plane App
            </Link>
          </Button>
        }
      />
      <PageScroll>
        <DataTable
          columns={columns}
          data={data}
          isLoading={isLoading}
          emptyMessage='No Plane Apps registered.'
        />
      </PageScroll>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title='Delete Plane App'
        description='Delete this Plane App and its stored credentials? This cannot be undone.'
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return
          deleteMutation.mutate(deleteTarget, {
            onSuccess: () => {
              toast.success('Plane App deleted')
              setDeleteTarget(null)
            },
            onError: (err) => toast.error(err.message),
          })
        }}
      />
    </Page>
  )
}
