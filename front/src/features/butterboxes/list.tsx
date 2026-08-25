import { useState } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import type { ButterBox } from '@/gen/agents/v1/butterbox_pb'
import {
  Box,
  ExternalLink,
  KeyRound,
  MoreVertical,
  Pencil,
  Plus,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
import { useButterBoxes, useDeleteButterBox } from '@/api/butterboxes'
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
import { BoxSessionsInline, BoxStatusBadge } from './status-cell'

export function ButterBoxList() {
  const { data, isLoading } = useButterBoxes()
  const deleteMutation = useDeleteButterBox()
  const navigate = useNavigate()
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  const columns: Column<ButterBox>[] = [
    {
      header: 'ButterBox',
      cell: (row) => (
        <div className='flex items-center gap-2'>
          <Box className='h-4 w-4 text-muted-foreground' />
          <div>
            <div className='font-medium'>{row.name}</div>
            <div className='line-clamp-1 max-w-md text-xs text-muted-foreground'>
              {row.baseUrl}
            </div>
          </div>
        </div>
      ),
    },
    {
      header: 'Enabled',
      cell: (row) =>
        row.enabled ? (
          <Badge variant='outline'>Enabled</Badge>
        ) : (
          <Badge variant='outline' className='text-muted-foreground'>
            Disabled
          </Badge>
        ),
    },
    {
      header: 'Token',
      cell: (row) =>
        row.credentialSet ? (
          <KeyRound className='h-4 w-4 text-success' />
        ) : (
          <span className='text-xs text-muted-foreground'>-</span>
        ),
    },
    {
      header: 'Status',
      cell: (row) => <BoxStatusBadge id={row.id} />,
    },
    {
      header: 'Sessions',
      cell: (row) => <BoxSessionsInline id={row.id} />,
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
              <DropdownMenuItem asChild>
                <a href={row.baseUrl} target='_blank' rel='noreferrer'>
                  <ExternalLink className='mr-2 h-4 w-4' /> pi-web
                </a>
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() =>
                  navigate({
                    to: '/butterboxes/$id/edit',
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
        title='ButterBoxes'
        subtitle='Agent VMs running butter-box that host pi coding-agent sessions.'
        actions={
          <Button size='sm' asChild>
            <Link to='/butterboxes/create'>
              <Plus className='size-4' />
              Register Box
            </Link>
          </Button>
        }
      />
      <PageScroll>
        <DataTable
          columns={columns}
          data={data}
          isLoading={isLoading}
          emptyMessage='No ButterBoxes registered.'
        />
      </PageScroll>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title='Delete ButterBox'
        description='Delete this ButterBox? This action cannot be undone.'
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => {
                toast.success('ButterBox deleted')
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
