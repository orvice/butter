import { createFileRoute } from '@tanstack/react-router'
import { AdminGlobalMCPServersPage } from '@/features/admin/global-mcp-servers'

export const Route = createFileRoute('/_authenticated/admin/global-mcp-servers')({
  component: AdminGlobalMCPServersPage,
})
