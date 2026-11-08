import { createFileRoute } from '@tanstack/react-router'
import { PlaneAppEdit } from '@/features/plane/edit'

export const Route = createFileRoute('/_authenticated/plane-apps/$id/edit')({
  component: PlaneAppEdit,
})
