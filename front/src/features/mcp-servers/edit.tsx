import { useNavigate, useParams } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useMCPServer, useUpdateMCPServer } from '@/api/mcp-servers'
import type { MCPServer } from '@/types/api'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { Skeleton } from '@/components/ui/skeleton'
import { MCPServerForm } from './form'

export function MCPServerEdit() {
  const { id } = useParams({ from: '/_authenticated/mcp-servers/$id/edit' })
  const navigate = useNavigate()
  const { data, isLoading } = useMCPServer(id ?? '')
  const updateMutation = useUpdateMCPServer()

  function onSubmit(server: MCPServer) {
    updateMutation.mutate(
      { ...data?.mcp_server, ...server },
      {
        onSuccess: () => {
          toast.success('MCP server updated')
          navigate({ to: '/mcp-servers' })
        },
        onError: (err) => toast.error(err.message),
      }
    )
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title='Edit MCP Server'
        subtitle='Review endpoint and authentication changes before saving because agents may use this server immediately.'
      />
      <PageScroll className='max-w-3xl'>
        {isLoading ? (
          <Skeleton className='h-96 w-full' />
        ) : (
          <MCPServerForm
            mode='edit'
            submitLabel='Save'
            loading={updateMutation.isPending}
            initialValue={data?.mcp_server}
            onCancel={() => navigate({ to: '/mcp-servers' })}
            onSubmit={onSubmit}
          />
        )}
      </PageScroll>
    </Page>
  )
}
