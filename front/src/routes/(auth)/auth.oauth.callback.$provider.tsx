import { z } from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { OAuthCallback } from '@/features/auth/oauth-callback'

const searchSchema = z.object({
  code: z.string().optional(),
  state: z.string().optional(),
  error: z.string().optional(),
  error_description: z.string().optional(),
})

export const Route = createFileRoute('/(auth)/auth/oauth/callback/$provider')({
  component: OAuthCallback,
  validateSearch: searchSchema,
})
