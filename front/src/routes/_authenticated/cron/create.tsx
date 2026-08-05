import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/cron/create')({
  beforeLoad: () => {
    throw redirect({ to: '/automations/create', replace: true })
  },
})
