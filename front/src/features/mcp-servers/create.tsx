import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useCreateMCPServer } from '@/api/mcp-servers'
import type { MCPServer } from '@/types/api'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { MCPServerForm } from './form'

export function MCPServerCreate() {
  const navigate = useNavigate()
  const createMutation = useCreateMCPServer()

  function onSubmit(server: MCPServer) {
    createMutation.mutate(server, {
      onSuccess: () => {
        toast.success('MCP server created')
        navigate({ to: '/mcp-servers' })
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title='Create MCP Server'
        subtitle='Connect an HTTP or SSE MCP endpoint, then choose the authentication method it requires.'
      />
      <PageScroll className='max-w-3xl'>
        <MCPServerForm
          mode='create'
          submitLabel='Create'
          loading={createMutation.isPending}
          onCancel={() => navigate({ to: '/mcp-servers' })}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  )
}
