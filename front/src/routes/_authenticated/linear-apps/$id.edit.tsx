import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { LinearAppEdit } from '@/features/linear/edit'

const searchSchema = z.object({
  // Set by the Linear OAuth callback redirect.
  linear_install: z.string().optional(),
  reason: z.string().optional(),
  installation_id: z.string().optional(),
})

export const Route = createFileRoute('/_authenticated/linear-apps/$id/edit')({
  validateSearch: searchSchema,
  component: LinearAppEdit,
})
