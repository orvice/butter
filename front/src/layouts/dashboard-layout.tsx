import { useState } from "react";
import { Link, Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import type { AuthUser } from "@/api/auth";
import { useWorkspace } from "@/hooks/use-workspace";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { BrandMark } from "@/components/brand-mark";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useLayoutDensity } from "@/hooks/use-layout-density";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  MessageCircle,
  LogOut,
  Building2,
  Menu,
  Settings2,
  UserCircle,
} from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

type NavItem = {
  to: string;
  icon: typeof LayoutDashboard;
  label: string;
  activePrefixes?: string[];
};

const PRIMARY_NAV: NavItem[] = [
  { to: "/", icon: LayoutDashboard, label: "Dashboard" },
  { to: "/chat", icon: MessageCircle, label: "Chat", activePrefixes: ["/chat"] },
];

const MANAGE_NAV: NavItem[] = [
  {
    to: "/manage",
    icon: Settings2,
    label: "Manage",
  },
];

function WorkspaceSwitcher() {
  const { workspaces, selectedWorkspaceId, selectedWorkspace, isLoading, setSelectedWorkspaceId } = useWorkspace();

  return (
    <Select
      value={selectedWorkspaceId || ""}
      onValueChange={(value) => {
        if (value) setSelectedWorkspaceId(value);
      }}
      disabled={isLoading || workspaces.length === 0}
    >
      <SelectTrigger size="sm" className="w-40 border-white/25 bg-white/10 text-primary-foreground hover:bg-white/15 sm:w-48">
        <SelectValue placeholder={isLoading ? "Loading workspaces" : "Select workspace"}>
          <span className="flex min-w-0 items-center gap-2">
            <Building2 className="h-4 w-4 shrink-0 text-primary-foreground/75" />
            <span className="truncate">{selectedWorkspace?.name || "Select workspace"}</span>
          </span>
        </SelectValue>
      </SelectTrigger>
      <SelectContent align="end">
        {workspaces.map((workspace) => (
          <SelectItem key={workspace.id} value={workspace.id}>
            {workspace.name || "Unnamed workspace"}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

function isActiveNav(item: NavItem, pathname: string) {
  if (item.to === "/") return pathname === "/";
  if (item.to === "/manage") return pathname !== "/" && pathname !== "/chat" && !pathname.startsWith("/chat/");
  return (item.activePrefixes ?? [item.to]).some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`));
}

function NavList({ items }: { items: NavItem[] }) {
  const location = useLocation();
  const { isCompact } = useLayoutDensity();
  return (
    <div className={cn(isCompact ? "space-y-0.5" : "space-y-1")}>
      {items.map(({ to, icon: Icon, label, activePrefixes }) => {
        const active = isActiveNav({ to, icon: Icon, label, activePrefixes }, location.pathname);
        return (
          <Link
            key={to}
            to={to}
            className={`flex items-center gap-3 rounded-md px-3 text-sm font-medium transition-colors duration-200 ${isCompact ? "py-1.5" : "py-2.5"} ${
              active
                ? "bg-primary/10 text-primary"
                : "text-muted-foreground hover:bg-sidebar-accent/80 hover:text-foreground"
            }`}
          >
            <Icon className={cn("shrink-0 stroke-[1.7]", isCompact ? "h-4 w-4" : "h-5 w-5")} />
            <span>{label}</span>
          </Link>
        );
      })}
    </div>
  );
}

function SidebarNav() {
  const { isCompact } = useLayoutDensity();
  return (
    <nav className={cn("flex flex-1 flex-col overflow-y-auto px-3", isCompact ? "py-2" : "py-3")}>
      <NavList items={PRIMARY_NAV} />
      <div className={cn("mt-auto border-t", isCompact ? "pt-3" : "pt-4")}>
        <NavList items={MANAGE_NAV} />
      </div>
    </nav>
  );
}

function WorkspaceCreateCard() {
  const { createWorkspace, isCreating } = useWorkspace();
  const [name, setName] = useState("Default");
  const [slug, setSlug] = useState("default");
  const [description, setDescription] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmedName = name.trim();
    const trimmedSlug = slug.trim();
    if (!trimmedName) {
      toast.error("Workspace name is required");
      return;
    }
    if (!trimmedSlug) {
      toast.error("Workspace slug is required");
      return;
    }

    try {
      await createWorkspace({
        name: trimmedName,
        slug: trimmedSlug,
        description: description.trim(),
      });
      toast.success("Workspace created");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create workspace");
    }
  }

  return (
    <div className="flex min-h-[calc(100vh-8rem)] items-center justify-center">
      <Card className="w-full max-w-xl">
        <CardHeader>
          <CardTitle>Create your first workspace</CardTitle>
          <p className="text-sm text-muted-foreground">
            Workspaces scope agents, channels, cron jobs, model providers, and API tokens. Create one to enter the dashboard.
          </p>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="workspace-name">Name</label>
              <Input
                id="workspace-name"
                value={name}
                onChange={(e) => {
                  const next = e.target.value;
                  setName(next);
                  setSlug(next.toLowerCase().trim().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, ""));
                }}
                placeholder="Default"
                disabled={isCreating}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="workspace-slug">Slug</label>
              <Input
                id="workspace-slug"
                value={slug}
                onChange={(e) => setSlug(e.target.value.toLowerCase().trim().replace(/[^a-z0-9-]+/g, "-"))}
                placeholder="default"
                disabled={isCreating}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium" htmlFor="workspace-description">Description</label>
              <Textarea
                id="workspace-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional description"
                disabled={isCreating}
              />
            </div>
            <Button type="submit" disabled={isCreating}>
              {isCreating ? "Creating..." : "Create workspace"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

function UserMenu({ user, logout }: { user: AuthUser | null; logout: () => void }) {
  if (!user) return null;
  const avatar = user.avatar_url || user.avatarUrl || "";
  const name = user.display_name || user.displayName || user.username;
  const initial = (name || user.username || "?").trim().charAt(0).toUpperCase() || "?";
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          aria-label={`Open menu for ${name}`}
          className="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full border bg-muted text-xs font-medium text-foreground transition-opacity hover:opacity-80"
        >
          {avatar ? (
            <img src={avatar} alt="" className="h-full w-full object-cover" />
          ) : (
            <span>{initial}</span>
          )}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" sideOffset={8} className="w-52">
        <DropdownMenuLabel className="px-2 py-1.5">
          <span className="block truncate text-sm font-medium text-foreground">{name}</span>
          <span className="block truncate text-xs font-normal">@{user.username}</span>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem render={<Link to="/profile" />}>
          <UserCircle />
          Profile
        </DropdownMenuItem>
        <DropdownMenuItem render={<Link to="/manage" />}>
          <Settings2 />
          Manage
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem variant="destructive" onClick={logout}>
          <LogOut />
          Sign out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default function DashboardLayout() {
  const { isAuthenticated, logout, user } = useAuth();
  const { selectedWorkspaceId, workspaces, isLoading: isWorkspaceLoading } = useWorkspace();
  const { isCompact } = useLayoutDensity();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  const headerIconBtn =
    "material-header-action";

  return (
    <div className="flex min-h-[100dvh] overflow-hidden bg-background">
      <aside className={cn("hidden shrink-0 flex-col border-r border-sidebar-border bg-sidebar shadow-sidebar md:flex", isCompact ? "w-64" : "w-72")}>
        <div className={cn("flex items-center gap-2 border-b border-sidebar-border bg-primary px-5 text-primary-foreground", isCompact ? "h-16" : "h-[4.5rem]")}>
          <Link to="/" className="flex items-center gap-2.5 rounded-md outline-none focus-visible:ring-2 focus-visible:ring-white/40">
            <BrandMark size={30} />
            <span className="text-base font-medium leading-none tracking-tight">Butter</span>
          </Link>
        </div>
        <SidebarNav />
      </aside>

      <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
        <header className={cn("z-20 flex shrink-0 flex-wrap items-center justify-between gap-2 bg-primary px-3 py-2 text-primary-foreground shadow-header", isCompact ? "min-h-16 sm:px-5" : "min-h-[4.5rem] sm:px-6")}>
          <div className="flex items-center gap-2 md:hidden">
            <Sheet>
              <SheetTrigger render={<Button variant="ghost" size="icon" aria-label="Open navigation" className={headerIconBtn} />}>
                <Menu className="h-4 w-4" />
              </SheetTrigger>
              <SheetContent side="left" className="w-72 gap-0 p-0" showCloseButton={false}>
                <SheetHeader className="h-[4.5rem] justify-center border-b bg-primary text-primary-foreground">
                  <SheetTitle className="text-primary-foreground">
                    <Link to="/" className="flex items-center gap-2.5 rounded-md outline-none focus-visible:ring-2 focus-visible:ring-white/40">
                      <BrandMark size={30} />
                      <span className="text-base font-medium leading-none tracking-tight">Butter</span>
                    </Link>
                  </SheetTitle>
                </SheetHeader>
                <SidebarNav />
              </SheetContent>
            </Sheet>
            <Link to="/" className="flex items-center gap-2.5 rounded-md outline-none focus-visible:ring-2 focus-visible:ring-white/40">
              <BrandMark size={30} />
              <span className="text-base font-medium leading-none tracking-tight text-primary-foreground">Butter</span>
            </Link>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <WorkspaceSwitcher />
            <UserMenu user={user} logout={logout} />
          </div>
        </header>
        <main className={cn("flex-1 overflow-auto p-4", isCompact ? "sm:p-4" : "sm:p-6")}>
          {selectedWorkspaceId ? (
            <Outlet />
          ) : isWorkspaceLoading ? (
            <Card>
              <CardHeader>
                <CardTitle>Loading workspaces</CardTitle>
              </CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                Loading available workspaces...
              </CardContent>
            </Card>
          ) : workspaces.length === 0 ? (
            <WorkspaceCreateCard />
          ) : (
            <Card>
              <CardHeader>
                <CardTitle>Selecting workspace</CardTitle>
              </CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                Preparing your workspace...
              </CardContent>
            </Card>
          )}
        </main>
      </div>
    </div>
  );
}
