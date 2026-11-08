import { createFileRoute } from '@tanstack/react-router'
import { PlaneAppCreate } from '@/features/plane/create'

export const Route = createFileRoute('/_authenticated/plane-apps/create')({
  component: PlaneAppCreate,
})
