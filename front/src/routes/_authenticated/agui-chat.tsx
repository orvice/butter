import { createFileRoute, redirect } from '@tanstack/react-router'

// AG-UI Chat is Chat now, at /chat (#409). Its links land there with the
// query they carry, ?thread= and ?agent= included.
export const Route = createFileRoute('/_authenticated/agui-chat')({
  beforeLoad: ({ search }) => {
    throw redirect({ to: '/chat', search, replace: true })
  },
})
