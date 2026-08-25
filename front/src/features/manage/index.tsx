import {
  useNavigate,
  useSearch,
  Link,
  type LinkProps,
} from '@tanstack/react-router'
import { fonts } from '@/config/fonts'
import {
  Bell,
  Blocks,
  Bot,
  Box,
  Building2,
  Cable,
  ChevronRight,
  Cpu,
  FileBox,
  GitBranch,
  KeyRound,
  Laptop,
  ListTree,
  Monitor,
  Moon,
  Palette,
  Plug,
  Satellite,
  Send,
  Server,
  Settings2,
  ShieldCheck,
  Sun,
  UserRound,
  Users,
  Workflow,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useFont } from '@/context/font-provider'
import { useLayout } from '@/context/layout-provider'
import { useTheme } from '@/context/theme-provider'
import { useAuth } from '@/hooks/use-auth'
import { Button } from '@/components/ui/button'
import { useSidebar } from '@/components/ui/sidebar'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'

type ManageTab = 'connections' | 'models' | 'workspace' | 'preferences'

type ManageItem = {
  title: string
  description: string
  to: LinkProps['to']
  icon: LucideIcon
}

type ManageSection = {
  title: string
  description: string
  items: ManageItem[]
}

const CONNECTION_SECTIONS: ManageSection[] = [
  {
    title: 'Tools and data',
    description:
      'Connect external capabilities and control how agents access them.',
    items: [
      {
        title: 'Integrations',
        description: 'Monitor MCP tools and remote agent connections.',
        to: '/integrations',
        icon: Blocks,
      },
      {
        title: 'MCP Servers',
        description:
          'Configure workspace-scoped MCP endpoints and credentials.',
        to: '/mcp-servers',
        icon: Server,
      },
      {
        title: 'Agent Files',
        description: 'Manage file spaces mounted into agents.',
        to: '/agent-files',
        icon: FileBox,
      },
      {
        title: 'API Tokens',
        description: 'Issue and revoke programmatic access tokens.',
        to: '/api-tokens',
        icon: KeyRound,
      },
    ],
  },
  {
    title: 'Messaging',
    description:
      'Configure inbound channels and outbound notification targets.',
    items: [
      {
        title: 'Telegram Platform Settings',
        description:
          'Set the public webhook base URL Telegram delivers callbacks to.',
        to: '/admin/telegram',
        icon: ShieldCheck,
      },
      {
        title: 'Telegram Channels',
        description:
          'Register Telegram bots and bind chats, groups, and forum topics.',
        to: '/telegram-channels',
        icon: Send,
      },
      {
        title: 'Telegram Updates',
        description:
          'Inspect processing history and resend replies that never landed.',
        to: '/telegram-updates',
        icon: ListTree,
      },
      {
        title: 'Notify Groups',
        description: 'Choose where automated results and alerts are delivered.',
        to: '/notify-groups',
        icon: Bell,
      },
    ],
  },
]

const MODEL_SECTIONS: ManageSection[] = [
  {
    title: 'Models and execution',
    description:
      'Configure model access and the runtimes that execute agent work.',
    items: [
      {
        title: 'Model Providers',
        description: 'Manage provider credentials and available model aliases.',
        to: '/model-providers',
        icon: Workflow,
      },
      {
        title: 'Remote Agents',
        description: 'Register A2A and daemon-backed remote agents.',
        to: '/remote-agents',
        icon: Satellite,
      },
      {
        title: 'ButterBoxes',
        description: 'Register agent VMs that host pi coding-agent sessions.',
        to: '/butterboxes',
        icon: Box,
      },
      {
        title: 'Daemons',
        description: 'Inspect connected daemon execution environments.',
        to: '/daemons',
        icon: Cpu,
      },
      {
        title: 'Agents',
        description:
          'Create and configure the agents that use these resources.',
        to: '/agents',
        icon: Bot,
      },
      {
        title: 'Sessions',
        description: 'Browse and inspect agent execution history.',
        to: '/sessions',
        icon: ListTree,
      },
    ],
  },
]

const WORKSPACE_SECTION: ManageSection = {
  title: 'Workspace and access',
  description:
    'Manage workspace membership, shared infrastructure, and your profile.',
  items: [
    {
      title: 'Workspaces',
      description: 'Create workspaces and manage their members.',
      to: '/workspaces',
      icon: Building2,
    },
    {
      title: 'Repository binding',
      description: 'Bind this workspace to a Git repository for agent content.',
      to: '/repo-binding',
      icon: GitBranch,
    },
    {
      title: 'Profile',
      description: 'Update your display name and profile image.',
      to: '/profile',
      icon: UserRound,
    },
  ],
}

const ADMIN_ITEMS: ManageItem[] = [
  {
    title: 'Users',
    description: 'Manage dashboard accounts and global access.',
    to: '/admin/users',
    icon: Users,
  },
  {
    title: 'Global MCP Servers',
    description: 'Manage MCP servers shared across every workspace.',
    to: '/admin/global-mcp-servers',
    icon: ShieldCheck,
  },
  {
    title: 'Git hosts',
    description:
      'Manage the platform allowlist of Git hosts for repository bindings.',
    to: '/admin/git-hosts',
    icon: GitBranch,
  },
]

const PREFERENCE_LINKS: ManageItem[] = [
  {
    title: 'Profile',
    description: 'Update your name and avatar.',
    to: '/profile',
    icon: UserRound,
  },
  {
    title: 'Advanced appearance',
    description: 'Open the detailed font and theme form.',
    to: '/settings/appearance',
    icon: Palette,
  },
  {
    title: 'Display preferences',
    description: 'Control which optional items are visible in the app.',
    to: '/settings/display',
    icon: Monitor,
  },
]

const MANAGE_TABS: Array<{
  value: ManageTab
  label: string
  mobileLabel: string
  icon: LucideIcon
}> = [
  {
    value: 'connections',
    label: 'Connections',
    mobileLabel: 'Connect',
    icon: Plug,
  },
  {
    value: 'models',
    label: 'Models and runtime',
    mobileLabel: 'Models',
    icon: Cpu,
  },
  {
    value: 'workspace',
    label: 'Workspace',
    mobileLabel: 'Workspace',
    icon: Building2,
  },
  {
    value: 'preferences',
    label: 'Preferences',
    mobileLabel: 'Prefs',
    icon: Monitor,
  },
]

export function ManagePage() {
  const { isAdmin } = useAuth()
  const { tab = 'connections' } = useSearch({
    from: '/_authenticated/manage',
  })
  const navigate = useNavigate({ from: '/manage' })

  const workspaceSections = isAdmin
    ? [
        WORKSPACE_SECTION,
        {
          title: 'Administration',
          description: 'Global controls available only to administrators.',
          items: ADMIN_ITEMS,
        },
      ]
    : [WORKSPACE_SECTION]

  return (
    <Page>
      <PageHeader
        title='Manage'
        subtitle='Workspace connections, models, access, and application preferences.'
        className='mb-3'
      />
      <PageScroll>
        <Tabs
          value={tab}
          onValueChange={(value) =>
            void navigate({
              search: { tab: value as ManageTab },
              replace: true,
            })
          }
          className='gap-0'
        >
          <TabsList className='grid h-auto w-full grid-cols-4 rounded-none border-b bg-transparent p-0 sm:flex sm:justify-start sm:overflow-x-auto'>
            {MANAGE_TABS.map((item) => (
              <TabsTrigger
                key={item.value}
                value={item.value}
                className='h-11 min-w-0 rounded-none border-0 border-b-2 border-transparent px-1 text-xs text-muted-foreground shadow-none data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:text-foreground data-[state=active]:shadow-none sm:flex-none sm:px-3 sm:text-sm dark:data-[state=active]:border-primary dark:data-[state=active]:bg-transparent'
              >
                <item.icon className='hidden sm:block' />
                <span className='truncate sm:hidden'>{item.mobileLabel}</span>
                <span className='hidden sm:inline'>{item.label}</span>
              </TabsTrigger>
            ))}
          </TabsList>

          <TabsContent value='connections' className='mt-0 py-6'>
            <ManageDirectory sections={CONNECTION_SECTIONS} />
          </TabsContent>
          <TabsContent value='models' className='mt-0 py-6'>
            <ManageDirectory sections={MODEL_SECTIONS} />
          </TabsContent>
          <TabsContent value='workspace' className='mt-0 py-6'>
            <ManageDirectory sections={workspaceSections} />
          </TabsContent>
          <TabsContent value='preferences' className='mt-0 py-6'>
            <PreferencesPanel />
          </TabsContent>
        </Tabs>
      </PageScroll>
    </Page>
  )
}

function ManageDirectory({ sections }: { sections: ManageSection[] }) {
  return (
    <div className='grid max-w-5xl gap-6 xl:grid-cols-2'>
      {sections.map((section) => (
        <section key={section.title}>
          <div className='mb-3'>
            <h2 className='text-sm font-semibold'>{section.title}</h2>
            <p className='mt-0.5 text-sm text-muted-foreground'>
              {section.description}
            </p>
          </div>
          <div className='overflow-hidden rounded-md border bg-card'>
            {section.items.map((item) => (
              <ManageRow key={item.title} item={item} />
            ))}
          </div>
        </section>
      ))}
    </div>
  )
}

function ManageRow({ item }: { item: ManageItem }) {
  return (
    <Link
      to={item.to}
      className='group flex min-h-17 items-center gap-3 border-b px-4 py-3 transition-colors last:border-b-0 hover:bg-muted/55 focus-visible:bg-muted/55 focus-visible:outline-none'
    >
      <span className='flex size-9 shrink-0 items-center justify-center rounded-md border bg-muted/45 text-muted-foreground group-hover:text-foreground'>
        <item.icon className='size-4' />
      </span>
      <span className='min-w-0 flex-1'>
        <span className='block text-sm font-medium text-foreground'>
          {item.title}
        </span>
        <span className='mt-0.5 block text-xs leading-5 text-muted-foreground'>
          {item.description}
        </span>
      </span>
      <ChevronRight className='size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5' />
    </Link>
  )
}

function PreferencesPanel() {
  const { theme, setTheme } = useTheme()
  const { font, setFont } = useFont()
  const { collapsible, setCollapsible } = useLayout()
  const { isMobile, open, setOpen } = useSidebar()

  const sidebarMode = open
    ? 'expanded'
    : collapsible === 'offcanvas'
      ? 'hidden'
      : 'compact'

  const setSidebarMode = (mode: 'expanded' | 'compact' | 'hidden') => {
    if (mode === 'expanded') {
      setCollapsible('icon')
      setOpen(true)
      return
    }
    setCollapsible(mode === 'hidden' ? 'offcanvas' : 'icon')
    setOpen(false)
  }

  return (
    <div className='max-w-5xl space-y-7'>
      <PreferenceSection
        title='Appearance'
        description='Choose how Butter looks on this device.'
      >
        <SegmentedControl
          value={theme}
          options={[
            { value: 'light', label: 'Light', icon: Sun },
            { value: 'dark', label: 'Dark', icon: Moon },
            { value: 'system', label: 'System', icon: Laptop },
          ]}
          onChange={(value) => setTheme(value as typeof theme)}
        />
      </PreferenceSection>

      <PreferenceSection
        title='Typography'
        description='Select the interface font used throughout the dashboard.'
      >
        <SegmentedControl
          value={font}
          options={fonts.map((value) => ({
            value,
            label: value === 'system' ? 'System' : capitalize(value),
          }))}
          onChange={(value) => setFont(value as typeof font)}
        />
      </PreferenceSection>

      {!isMobile && (
        <PreferenceSection
          title='Navigation'
          description='Choose how much space the desktop sidebar uses.'
        >
          <SegmentedControl
            value={sidebarMode}
            options={[
              { value: 'expanded', label: 'Expanded', icon: Cable },
              { value: 'compact', label: 'Compact', icon: Settings2 },
              { value: 'hidden', label: 'Hidden', icon: Monitor },
            ]}
            onChange={(value) =>
              setSidebarMode(value as 'expanded' | 'compact' | 'hidden')
            }
          />
        </PreferenceSection>
      )}

      <section>
        <div className='mb-3'>
          <h2 className='text-sm font-semibold'>Personal settings</h2>
          <p className='mt-0.5 text-sm text-muted-foreground'>
            Open detailed profile and display configuration.
          </p>
        </div>
        <div className='overflow-hidden rounded-md border bg-card'>
          {PREFERENCE_LINKS.map((item) => (
            <ManageRow key={item.title} item={item} />
          ))}
        </div>
      </section>
    </div>
  )
}

function PreferenceSection({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: React.ReactNode
}) {
  return (
    <section>
      <h2 className='text-sm font-semibold'>{title}</h2>
      <p className='mt-0.5 text-xs text-muted-foreground'>{description}</p>
      <div className='mt-3'>{children}</div>
    </section>
  )
}

function SegmentedControl({
  value,
  options,
  onChange,
}: {
  value: string
  options: Array<{ value: string; label: string; icon?: LucideIcon }>
  onChange: (value: string) => void
}) {
  return (
    <div className='inline-flex max-w-full overflow-x-auto rounded-md border bg-background p-1'>
      {options.map((option) => (
        <Button
          key={option.value}
          type='button'
          size='sm'
          variant='ghost'
          aria-pressed={value === option.value}
          onClick={() => onChange(option.value)}
          className={cn(
            'h-8 shrink-0 gap-1.5 px-3 font-normal text-muted-foreground shadow-none',
            value === option.value &&
              'bg-muted text-foreground hover:bg-muted hover:text-foreground'
          )}
        >
          {option.icon && <option.icon className='size-4' />}
          {option.label}
        </Button>
      ))}
    </div>
  )
}

function capitalize(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1)
}
