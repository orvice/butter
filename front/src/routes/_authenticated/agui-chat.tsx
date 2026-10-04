import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { AGUIChatPage } from '@/features/agui-chat'

// thread opens that thread; without it the page is a new-chat draft, and
// agent preselects the agent the draft starts with.
const searchSchema = z.object({
  thread: z.string().optional(),
  agent: z.string().optional(),
})

export const Route = createFileRoute('/_authenticated/agui-chat')({
  component: AGUIChatPage,
  validateSearch: searchSchema,
})
