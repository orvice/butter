import { createFileRoute } from '@tanstack/react-router'
import { WorkStationPage } from '@/features/work-station'

export const Route = createFileRoute('/_authenticated/work-station')({
  validateSearch: (search: Record<string, unknown>) => ({
    kind: search.kind === 'PULL_REQUEST' ? 'PULL_REQUEST' : 'ISSUE',
    status: typeof search.status === 'string' ? search.status : 'open',
    connection: typeof search.connection === 'string' ? search.connection : '',
    scope: typeof search.scope === 'string' ? search.scope : '',
    assignee: typeof search.assignee === 'string' ? search.assignee : '',
    q: typeof search.q === 'string' ? search.q : '',
  }),
  component: WorkStationPage,
})
