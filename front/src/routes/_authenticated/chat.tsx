import { z } from 'zod'
import { createFileRoute, redirect } from '@tanstack/react-router'
import { AGUIChatPage } from '@/features/agui-chat'

// thread opens that thread; without it the page is a new-chat draft, and
// agent preselects the agent the draft starts with.
const searchSchema = z.object({
  thread: z.string().optional(),
  agent: z.string().optional(),
})

// The parameters of the Chat that came before AG-UI (#409). Its links, as
// bookmarks or browser history keep them, land on a new-chat draft: its
// web-chat sessions are not carried over.
const RETIRED_PARAMS = ['session', 'new', 'pending_message', 'invocation']

export const Route = createFileRoute('/_authenticated/chat')({
  component: AGUIChatPage,
  validateSearch: searchSchema,
  beforeLoad: ({ search }) => {
    if (!RETIRED_PARAMS.some((param) => param in search)) return
    throw redirect({
      to: '/chat',
      search: { thread: search.thread, agent: search.agent },
      replace: true,
    })
  },
})
