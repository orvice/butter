import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { ChatPage } from '@/features/chat'

const searchSchema = z.object({
  new: z.number().optional(),
  session: z.string().optional(),
  agent: z.string().optional(),
  pending_message: z.string().optional(),
  invocation: z.string().optional(),
  aui: z.number().optional(),
})

export const Route = createFileRoute('/_authenticated/chat')({
  component: ChatPage,
  validateSearch: searchSchema,
})
