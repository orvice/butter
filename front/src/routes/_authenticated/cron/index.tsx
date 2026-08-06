import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/cron/')({
  beforeLoad: () => {
    throw redirect({ to: '/automations', replace: true })
  },
})
