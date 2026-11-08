import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { PlaneAppEdit } from '@/features/plane/edit'

export const Route = createFileRoute('/_authenticated/plane-apps/$id/edit')({
  validateSearch: z.object({
    plane_install: z.string().optional(),
    reason: z.string().optional(),
    installation_id: z.string().optional(),
  }),
  component: PlaneAppEdit,
})
