import { useSearch } from '@tanstack/react-router'
import { AuthLayout } from '../auth-layout'
import { UserAuthForm } from './components/user-auth-form'

export function SignIn() {
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })

  return (
    <AuthLayout>
      <section aria-labelledby='sign-in-title'>
        <div className='mb-8'>
          <p className='mb-2 text-sm font-medium text-[#2b658b] dark:text-[#82c4ea]'>
            Workspace access
          </p>
          <h1
            id='sign-in-title'
            className='font-manrope text-2xl leading-tight font-semibold text-foreground'
          >
            Sign in to Butter
          </h1>
          <p className='mt-2 text-sm leading-6 text-muted-foreground'>
            Enter your account details to continue.
          </p>
        </div>

        <UserAuthForm redirectTo={redirect} />
      </section>
    </AuthLayout>
  )
}
