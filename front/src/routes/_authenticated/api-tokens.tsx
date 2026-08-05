import { createFileRoute } from '@tanstack/react-router'
import { APITokens } from '@/features/api-tokens'

export const Route = createFileRoute('/_authenticated/api-tokens')({
  component: APITokens,
})
