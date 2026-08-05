import { useEffect, useRef, useState } from 'react'
import { useNavigate, useParams, useSearch } from '@tanstack/react-router'
import { completeOAuthFlow } from '@/api/auth'
import { useAuthStore } from '@/stores/auth-store'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export function OAuthCallback() {
  const { provider } = useParams({
    from: '/(auth)/auth/oauth/callback/$provider',
  })
  const search = useSearch({ from: '/(auth)/auth/oauth/callback/$provider' })
  const navigate = useNavigate()
  const { auth } = useAuthStore()
  const code = search.code ?? ''
  const state = search.state ?? ''
  const callbackError = search.error
    ? `${search.error}: ${search.error_description ?? 'authorization rejected'}`
    : !provider || !code || !state
      ? 'Missing provider, code, or state in callback URL.'
      : ''
  const [error, setError] = useState<string>(callbackError)
  const consumed = useRef(false)

  useEffect(() => {
    if (callbackError) return
    if (consumed.current) return
    consumed.current = true

    completeOAuthFlow(provider, code, state)
      .then((res) => {
        if (!res.token) {
          setError('Login response missing token.')
          return
        }
        auth.applyLoginResponse(res)
        navigate({ to: '/', replace: true })
      })
      .catch((e: unknown) => {
        setError(e instanceof Error ? e.message : 'OAuth login failed.')
      })
  }, [provider, code, state, callbackError, auth, navigate])

  return (
    <div className='flex min-h-svh items-center justify-center bg-background px-4'>
      <Card className='w-full max-w-sm'>
        <CardHeader className='text-center'>
          <CardTitle>Signing you in…</CardTitle>
          <CardDescription>
            Completing {provider || 'OAuth'} login.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {error ? (
            <div className='space-y-3 text-sm'>
              <p className='text-destructive'>{error}</p>
              <Button
                variant='outline'
                className='w-full'
                onClick={() => navigate({ to: '/sign-in', replace: true })}
              >
                Back to sign in
              </Button>
            </div>
          ) : (
            <p className='text-sm text-muted-foreground'>
              Hang tight, you'll be redirected shortly.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
