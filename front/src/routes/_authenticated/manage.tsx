import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { ManagePage } from '@/features/manage'

export const Route = createFileRoute('/_authenticated/manage')({
  component: ManagePage,
  validateSearch: z.object({
    tab: z
      .enum(['connections', 'models', 'workspace', 'preferences'])
      .optional(),
  }),
})
