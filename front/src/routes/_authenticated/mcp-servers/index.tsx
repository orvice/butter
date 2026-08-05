import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { MCPServerList } from '@/features/mcp-servers/list'

export const Route = createFileRoute('/_authenticated/mcp-servers/')({
  validateSearch: z.object({
    mcp_oauth: z.string().optional(),
    server_id: z.string().optional(),
  }),
  component: MCPServerList,
})
