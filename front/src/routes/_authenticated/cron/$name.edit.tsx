import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/cron/$name/edit')({
  beforeLoad: ({ params }) => {
    throw redirect({
      to: '/automations/$name/edit',
      params: { name: params.name },
      replace: true,
    })
  },
})
