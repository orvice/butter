import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/operations')({
  beforeLoad: () => {
    throw redirect({ to: '/sessions', replace: true })
  },
})
