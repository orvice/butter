import { useAuthStore } from '@/stores/auth-store'
import { useLayout } from '@/context/layout-provider'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { sidebarData } from './data/sidebar-data'
import { NavChatHistory } from './nav-chat-history'
import { NavGroup } from './nav-group'
import { NavManage } from './nav-manage'
import { NavUser } from './nav-user'
import { WorkspaceSwitcher } from './workspace-switcher'

export function AppSidebar() {
  const { collapsible, variant } = useLayout()
  const user = useAuthStore((state) => state.auth.user)

  const navUser = {
    name: user?.display_name || user?.displayName || user?.username || 'User',
    email: user?.username ? `@${user.username}` : '',
    avatar: user?.avatar_url || user?.avatarUrl || '',
  }

  const generalGroup = sidebarData.navGroups[0]

  return (
    <Sidebar collapsible={collapsible} variant={variant}>
      <SidebarHeader>
        <WorkspaceSwitcher />
      </SidebarHeader>
      <SidebarContent>
        {generalGroup && <NavGroup {...generalGroup} />}
        <NavChatHistory />
      </SidebarContent>
      <SidebarFooter>
        <NavManage />
        <NavUser user={navUser} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
