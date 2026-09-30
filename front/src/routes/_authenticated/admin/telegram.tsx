import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/admin/telegram')({
  beforeLoad: () => {
    throw redirect({ to: '/admin/endpoints', replace: true })
  },
})
