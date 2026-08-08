import { createFileRoute } from '@tanstack/react-router'
import { RepoBindingPage } from '@/features/repo-binding'

export const Route = createFileRoute('/_authenticated/repo-binding')({
  component: RepoBindingPage,
})
