import { useAuthStore, useIsAdmin } from '@/stores/auth-store'
import { useLayout } from '@/context/layout-provider'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { sidebarData } from './data/sidebar-data'
import { NavGroup } from './nav-group'
import { NavUser } from './nav-user'
import { WorkspaceSwitcher } from './workspace-switcher'

export function AppSidebar() {
  const { collapsible, variant } = useLayout()
  const user = useAuthStore((state) => state.auth.user)
  const isAdmin = useIsAdmin()

  const navUser = {
    name: user?.display_name || user?.displayName || user?.username || 'User',
    email: user?.username ? `@${user.username}` : '',
    avatar: user?.avatar_url || user?.avatarUrl || '',
  }

  const navGroups = sidebarData.navGroups.filter(
    (group) => group.title !== 'Admin' || isAdmin
  )

  return (
    <Sidebar collapsible={collapsible} variant={variant}>
      <SidebarHeader>
        <WorkspaceSwitcher />
      </SidebarHeader>
      <SidebarContent>
        {navGroups.map((props) => (
          <NavGroup key={props.title} {...props} />
        ))}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={navUser} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
