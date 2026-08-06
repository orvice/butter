import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { SessionDetailPage } from '@/features/sessions/detail'

const searchSchema = z.object({
  app: z.string().optional(),
  user: z.string().optional(),
  sid: z.string().optional(),
  // Legacy links used `?session=` instead of `?sid=`.
  session: z.string().optional(),
})

export const Route = createFileRoute('/_authenticated/sessions/detail')({
  validateSearch: searchSchema,
  component: SessionDetailPage,
})
