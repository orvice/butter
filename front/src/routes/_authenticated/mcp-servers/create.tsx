import { createFileRoute } from '@tanstack/react-router'
import { MCPServerCreate } from '@/features/mcp-servers/create'

export const Route = createFileRoute('/_authenticated/mcp-servers/create')({
  component: MCPServerCreate,
})
