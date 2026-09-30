import { createFileRoute } from '@tanstack/react-router'
import { AdminPublicEndpointsPage } from '@/features/admin/public-endpoints'

export const Route = createFileRoute('/_authenticated/admin/endpoints')({
  component: AdminPublicEndpointsPage,
})
