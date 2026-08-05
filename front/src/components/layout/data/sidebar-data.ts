import {
  Blocks,
  Bot,
  CalendarClock,
  FileBox,
  Globe,
  KeyRound,
  LayoutDashboard,
  ListTree,
  MessageSquare,
  MessagesSquare,
  Radio,
  Satellite,
  Server,
  ServerCog,
  Settings,
  ShieldCheck,
  Sparkles,
  Users,
  Workflow,
} from 'lucide-react'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
  user: {
    name: 'Butter',
    email: '',
    avatar: '',
  },
  teams: [
    {
      name: 'Butter',
      logo: Sparkles,
      plan: 'Agent Platform',
    },
  ],
  navGroups: [
    {
      title: 'General',
      items: [
        {
          title: 'Dashboard',
          url: '/',
          icon: LayoutDashboard,
        },
        {
          title: 'Chat',
          url: '/chat',
          icon: MessageSquare,
        },
        {
          title: 'Agents',
          url: '/agents',
          icon: Bot,
        },
        {
          title: 'Automations',
          url: '/automations',
          icon: CalendarClock,
        },
        {
          title: 'Sessions',
          url: '/sessions',
          icon: ListTree,
        },
        {
          title: 'Forum',
          url: '/forum',
          icon: MessagesSquare,
        },
      ],
    },
    {
      title: 'Resources',
      items: [
        {
          title: 'MCP Servers',
          url: '/mcp-servers',
          icon: Server,
        },
        {
          title: 'Model Providers',
          url: '/model-providers',
          icon: Workflow,
        },
        {
          title: 'Remote Agents',
          url: '/remote-agents',
          icon: Satellite,
        },
        {
          title: 'Channels',
          url: '/channels',
          icon: Radio,
        },
        {
          title: 'Notify Groups',
          url: '/notify-groups',
          icon: Globe,
        },
        {
          title: 'Agent Files',
          url: '/agent-files',
          icon: FileBox,
        },
        {
          title: 'Daemons',
          url: '/daemons',
          icon: ServerCog,
        },
        {
          title: 'Integrations',
          url: '/integrations',
          icon: Blocks,
        },
        {
          title: 'API Tokens',
          url: '/api-tokens',
          icon: KeyRound,
        },
      ],
    },
    {
      title: 'Admin',
      items: [
        {
          title: 'Users',
          url: '/users',
          icon: Users,
        },
        {
          title: 'Global MCP Servers',
          url: '/admin/global-mcp-servers',
          icon: ShieldCheck,
        },
        {
          title: 'Workspaces',
          url: '/workspaces',
          icon: Globe,
        },
      ],
    },
    {
      title: 'Other',
      items: [
        {
          title: 'Settings',
          icon: Settings,
          items: [
            {
              title: 'Profile',
              url: '/settings',
            },
            {
              title: 'Account',
              url: '/settings/account',
            },
            {
              title: 'Appearance',
              url: '/settings/appearance',
            },
            {
              title: 'Display',
              url: '/settings/display',
            },
          ],
        },
      ],
    },
  ],
}
