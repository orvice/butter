import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/cron/$name/executions')({
  beforeLoad: ({ params }) => {
    throw redirect({
      to: '/automations/$name',
      params: { name: params.name },
      replace: true,
    })
  },
})
