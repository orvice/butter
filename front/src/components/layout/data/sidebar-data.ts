import {
  Bot,
  CalendarClock,
  LayoutDashboard,
  ListTodo,
  MessageSquare,
  MessagesSquare,
  PlugZap,
  Sparkles,
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
          title: 'AG-UI Chat',
          url: '/agui-chat',
          icon: PlugZap,
        },
        {
          title: 'Agents',
          url: '/agents',
          icon: Bot,
        },
        {
          title: 'Operations',
          url: '/operations',
          icon: ListTodo,
        },
        {
          title: 'Automations',
          url: '/automations',
          icon: CalendarClock,
        },
        {
          title: 'Forum',
          url: '/forum',
          icon: MessagesSquare,
        },
      ],
    },
  ],
}
