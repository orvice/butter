import { createFileRoute } from '@tanstack/react-router'
import { MCPServerEdit } from '@/features/mcp-servers/edit'

export const Route = createFileRoute('/_authenticated/mcp-servers/$id/edit')({
  component: MCPServerEdit,
})
