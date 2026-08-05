import { useEffect, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import { Loader2, LogIn } from 'lucide-react'
import { toast } from 'sonner'
import {
  beginOAuthFlow,
  listOAuthProviders,
  type OAuthProviderInfo,
} from '@/api/auth'
import { useAuthStore } from '@/stores/auth-store'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { OAuthProviderIcon } from '@/components/oauth-provider-icon'
import { PasswordInput } from '@/components/password-input'

const formSchema = z.object({
  username: z.string().min(1, 'Please enter your username.'),
  password: z.string().min(1, 'Please enter your password.'),
})

interface UserAuthFormProps extends React.HTMLAttributes<HTMLFormElement> {
  redirectTo?: string
}

export function UserAuthForm({
  className,
  redirectTo,
  ...props
}: UserAuthFormProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [oauthProviders, setOauthProviders] = useState<OAuthProviderInfo[]>([])
  const [oauthLoading, setOauthLoading] = useState<string | null>(null)
  const navigate = useNavigate()
  const { auth } = useAuthStore()

  useEffect(() => {
    let cancelled = false
    listOAuthProviders()
      .then((res) => {
        if (!cancelled) setOauthProviders(res.providers ?? [])
      })
      .catch(() => {
        if (!cancelled) setOauthProviders([])
      })
    return () => {
      cancelled = true
    }
  }, [])

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  })

  async function onSubmit(data: z.infer<typeof formSchema>) {
    setIsLoading(true)
    try {
      const ok = await auth.login(data.username.trim(), data.password)
      if (!ok) {
        toast.error('Invalid username or password.')
        return
      }
      const targetPath = redirectTo || '/'
      navigate({ to: targetPath, replace: true })
    } catch {
      toast.error('Connection failed. Is the server running?')
    } finally {
      setIsLoading(false)
    }
  }

  async function handleOAuth(providerName: string) {
    setOauthLoading(providerName)
    try {
      const redirectUri = `${window.location.origin}/auth/oauth/callback/${providerName}`
      const res = await beginOAuthFlow(providerName, redirectUri)
      const url = res.authorize_url ?? res.authorizeUrl
      if (!url) {
        toast.error('Provider did not return an authorize URL.')
        setOauthLoading(null)
        return
      }
      window.location.assign(url)
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : 'Failed to start OAuth flow.')
      setOauthLoading(null)
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-3', className)}
        {...props}
      >
        <FormField
          control={form.control}
          name='username'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Username</FormLabel>
              <FormControl>
                <Input
                  placeholder='username'
                  autoComplete='username'
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='password'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder='********'
                  autoComplete='current-password'
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button className='mt-2' disabled={isLoading}>
          {isLoading ? <Loader2 className='animate-spin' /> : <LogIn />}
          Sign in
        </Button>

        {oauthProviders.length > 0 && (
          <>
            <div className='relative my-2'>
              <div className='absolute inset-0 flex items-center'>
                <span className='w-full border-t' />
              </div>
              <div className='relative flex justify-center text-xs uppercase'>
                <span className='bg-background px-2 text-muted-foreground'>
                  Or continue with
                </span>
              </div>
            </div>

            <div className='grid gap-2'>
              {oauthProviders.map((p) => {
                const label = p.display_name ?? p.displayName ?? p.name
                return (
                  <Button
                    key={p.name}
                    variant='outline'
                    type='button'
                    disabled={!!oauthLoading || isLoading}
                    onClick={() => handleOAuth(p.name)}
                  >
                    <OAuthProviderIcon
                      provider={p.name}
                      className='h-4 w-4 shrink-0'
                    />
                    {oauthLoading === p.name
                      ? `Redirecting to ${label}…`
                      : `Sign in with ${label}`}
                  </Button>
                )
              })}
            </div>
          </>
        )}
      </form>
    </Form>
  )
}
