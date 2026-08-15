import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { AGUIChatPage } from '@/features/agui-chat'

const searchSchema = z.object({
  agent: z.string().optional(),
})

export const Route = createFileRoute('/_authenticated/agui-chat')({
  component: AGUIChatPage,
  validateSearch: searchSchema,
})
