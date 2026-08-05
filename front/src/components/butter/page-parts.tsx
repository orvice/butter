import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

/**
 * Bridges the pre-migration page structure
 * (`<Page><PageHeader/><PageScroll>…</PageScroll></Page>`) onto the
 * shadcn-admin chrome: a sticky Header with global controls, and Main as
 * the scrolling content container.
 */
export function Page({
  children,
  fixed,
}: {
  children: ReactNode
  /** Pin Main to the viewport (full-height flex column, no page scroll). */
  fixed?: boolean
}) {
  return (
    <>
      <Header fixed>
        <Search />
        <div className='ms-auto flex items-center space-x-4'>
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>
      <Main fixed={fixed}>{children}</Main>
    </>
  )
}

export function PageHeader({
  title,
  subtitle,
  actions,
  breadcrumb,
  className,
}: {
  title: ReactNode
  subtitle?: ReactNode
  actions?: ReactNode
  breadcrumb?: ReactNode
  className?: string
}) {
  return (
    <div
      className={cn(
        'mb-6 flex w-full flex-col gap-4 md:flex-row md:items-start md:justify-between',
        className
      )}
    >
      <div className='min-w-0 flex-1'>
        {breadcrumb && (
          <div className='mb-3 text-xs text-muted-foreground'>{breadcrumb}</div>
        )}
        <h1 className='text-2xl font-bold tracking-tight text-balance'>
          {title}
        </h1>
        {subtitle && (
          <div className='mt-1 max-w-3xl text-sm leading-5 text-muted-foreground text-pretty'>
            {subtitle}
          </div>
        )}
      </div>
      {actions && (
        <div className='flex w-full flex-wrap items-center gap-2 sm:w-auto md:max-w-[55%] md:justify-end'>
          {actions}
        </div>
      )}
    </div>
  )
}

/** Content region below the page header. Main handles scrolling. */
export function PageScroll({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  return <div className={cn('w-full', className)}>{children}</div>
}

export function PageActions({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  return (
    <div
      className={cn(
        'sticky bottom-0 z-20 -mx-4 mt-6 flex flex-col gap-2 border-t border-border bg-background/95 px-4 py-4 pb-[max(1rem,env(safe-area-inset-bottom))] backdrop-blur supports-[backdrop-filter]:bg-background/85 [&>*]:w-full sm:mx-0 sm:flex-row sm:justify-end sm:px-0 sm:[&>*]:w-auto',
        className
      )}
    >
      {children}
    </div>
  )
}
