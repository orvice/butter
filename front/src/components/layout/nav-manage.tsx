import { Link, useLocation } from '@tanstack/react-router'
import { Settings } from 'lucide-react'
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'

export function NavManage() {
  const pathname = useLocation({ select: (location) => location.pathname })
  const { setOpenMobile } = useSidebar()

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          asChild
          isActive={pathname === '/manage'}
          tooltip='Manage'
          className='h-10'
        >
          <Link to='/manage' onClick={() => setOpenMobile(false)}>
            <Settings />
            <span>Manage</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
