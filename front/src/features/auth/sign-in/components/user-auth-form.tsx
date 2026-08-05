import { useEffect, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import { AlertCircle, Loader2, LogIn } from 'lucide-react'
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
    mode: 'onChange',
    defaultValues: {
      username: '',
      password: '',
    },
  })

  async function onSubmit(data: z.infer<typeof formSchema>) {
    form.clearErrors('root')
    setIsLoading(true)
    try {
      const ok = await auth.login(data.username.trim(), data.password)
      if (!ok) {
        form.setError('root', {
          message: 'Sign-in failed. Check your username and password.',
        })
        return
      }
      const targetPath = redirectTo || '/'
      navigate({ to: targetPath, replace: true })
    } catch {
      form.setError('root', {
        message: 'Connection failed. Check the server and try again.',
      })
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
      toast.error(
        e instanceof Error ? e.message : 'Failed to start OAuth flow.'
      )
      setOauthLoading(null)
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-5', className)}
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
                  autoFocus
                  disabled={isLoading || !!oauthLoading}
                  className='h-11 bg-background px-3.5 shadow-none'
                  {...field}
                  onChange={(event) => {
                    field.onChange(event)
                    form.clearErrors('root')
                  }}
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
                  disabled={isLoading || !!oauthLoading}
                  inputClassName='h-11 bg-background px-3.5 pe-11 shadow-none'
                  {...field}
                  onChange={(event) => {
                    field.onChange(event)
                    form.clearErrors('root')
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        {form.formState.errors.root?.message && (
          <div
            role='alert'
            className='flex items-start gap-2 rounded-md border border-destructive/25 bg-destructive/8 px-3 py-2.5 text-sm leading-5 text-destructive'
          >
            <AlertCircle className='mt-0.5 size-4 shrink-0' />
            <span>{form.formState.errors.root.message}</span>
          </div>
        )}

        <Button
          type='submit'
          className='h-11 w-full bg-[#2b658b] text-white shadow-none hover:bg-[#245a7a] active:translate-y-px dark:bg-[#82c4ea] dark:text-[#112f42] dark:hover:bg-[#99d0ef]'
          disabled={
            isLoading ||
            !!oauthLoading ||
            (!form.formState.isValid && !form.formState.errors.root)
          }
        >
          {isLoading ? <Loader2 className='animate-spin' /> : <LogIn />}
          {isLoading ? 'Signing in' : 'Sign in'}
        </Button>

        {oauthProviders.length > 0 && (
          <>
            <div className='relative my-1'>
              <div className='absolute inset-0 flex items-center'>
                <span className='w-full border-t' />
              </div>
              <div className='relative flex justify-center text-xs uppercase'>
                <span className='bg-background px-3 text-muted-foreground'>
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
                    className='h-11 w-full shadow-none active:translate-y-px'
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
