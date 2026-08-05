import { BrandMark } from '@/components/brand-mark'
import { ThemeSwitch } from '@/components/theme-switch'

type AuthLayoutProps = {
  children: React.ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className='relative min-h-[100dvh] overflow-hidden bg-background'>
      <div className='absolute top-4 right-4 z-10 rounded-full border bg-background/90 shadow-xs backdrop-blur-sm [&>button]:size-11 [&>button]:scale-100'>
        <ThemeSwitch />
      </div>

      <div className='grid min-h-[100dvh] min-w-0 lg:grid-cols-[minmax(0,0.9fr)_minmax(460px,1.1fr)]'>
        <aside className='hidden min-w-0 flex-col justify-between overflow-hidden bg-[#214f6d] p-10 text-white lg:flex xl:p-14 dark:bg-[#16384f]'>
          <div className='flex items-center gap-3'>
            <BrandMark size={42} />
            <span className='font-manrope text-xl font-semibold'>Butter</span>
          </div>

          <div className='max-w-md pb-12'>
            <p className='font-manrope text-4xl leading-tight font-semibold'>
              Welcome back.
            </p>
            <p className='mt-4 max-w-sm text-base leading-7 text-white/70'>
              Sign in to continue to your Butter workspace.
            </p>
          </div>

          <p className='text-sm text-white/55'>Secure workspace access</p>
        </aside>

        <main className='flex min-w-0 items-center justify-center px-5 py-20 sm:px-8 lg:px-12'>
          <div className='w-full max-w-[400px]'>
            <div className='mb-10 flex items-center gap-3 lg:hidden'>
              <BrandMark size={40} />
              <span className='font-manrope text-xl font-semibold'>Butter</span>
            </div>
            {children}
          </div>
        </main>
      </div>
    </div>
  )
}
