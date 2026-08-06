import {
  Bot,
  CalendarClock,
  LayoutDashboard,
  MessageSquare,
  MessagesSquare,
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
          title: 'Forum',
          url: '/forum',
          icon: MessagesSquare,
        },
      ],
    },
  ],
}
