import { useState } from "react";
import { Link, Navigate, Outlet, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import type { AuthUser } from "@/api/auth";
import { useWorkspace } from "@/hooks/use-workspace";
import { useOverview } from "@/api/dashboard";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { ThemeControls } from "@/components/theme-controls";
import { BrandMark } from "@/components/brand-mark";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useLayoutDensity } from "@/hooks/use-layout-density";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  Bot,
  Server,
  MessageCircle,
  MessagesSquare,
  LogOut,
  Cpu,
  Cable,
  BrainCircuit,
  KeyRound,
  Users,
  UserCircle,
  Building2,
  CircleCheck,
  CircleAlert,
  Menu,
  Search,
  Bell,
  Database,
  FolderOpen,
  RefreshCw,
  Plus,
  BookOpen,
  LifeBuoy,
  Settings2,
  ShieldCheck,
} from "lucide-react";
import type { ComponentHealth } from "@/types/api";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";

type NavItem = {
  to: string;
  icon: typeof LayoutDashboard;
  label: string;
  activePrefixes?: string[];
  adminOnly?: boolean;
};

const PRIMARY_NAV: NavItem[] = [
  { to: "/", icon: LayoutDashboard, label: "Overview" },
  { to: "/agents", icon: Bot, label: "Agents", activePrefixes: ["/agents"] },
  { to: "/agent-files", icon: FolderOpen, label: "Agent Files", activePrefixes: ["/agent-files"] },
  { to: "/chat", icon: MessageCircle, label: "Chat", activePrefixes: ["/chat"] },
  { to: "/forum", icon: MessagesSquare, label: "Forum", activePrefixes: ["/forum"] },
  {
    to: "/integrations",
    icon: Server,
    label: "Integrations",
    activePrefixes: ["/integrations", "/mcp-servers", "/remote-agents"],
  },
  {
    to: "/daemons",
    icon: Cpu,
    label: "Execution",
    activePrefixes: ["/daemons"],
  },
  { to: "/channels", icon: Cable, label: "Channels", activePrefixes: ["/channels"] },
  {
    to: "/operations",
    icon: Settings2,
    label: "Operations",
    activePrefixes: ["/operations", "/automations", "/cron", "/sessions", "/notify-groups"],
  },
];

const SECONDARY_NAV: NavItem[] = [
  { to: "/workspaces", icon: Building2, label: "Workspaces", activePrefixes: ["/workspaces"] },
  { to: "/model-providers", icon: BrainCircuit, label: "Model Providers", activePrefixes: ["/model-providers"] },
  { to: "/notify-groups", icon: Bell, label: "Notify Groups", activePrefixes: ["/notify-groups"] },
  { to: "/api-tokens", icon: KeyRound, label: "API Tokens", activePrefixes: ["/api-tokens"] },
  { to: "/profile", icon: UserCircle, label: "Profile", activePrefixes: ["/profile"] },
];

const ADMIN_NAV: NavItem[] = [
  { to: "/admin/users", icon: Users, label: "Users", activePrefixes: ["/admin/users"], adminOnly: true },
  {
    to: "/admin/global-mcp-servers",
    icon: ShieldCheck,
    label: "Global MCP",
    activePrefixes: ["/admin/global-mcp-servers"],
    adminOnly: true,
  },
];

type StatusBucket = "healthy" | "degraded" | "down" | "unknown";

function worstStatus(...checks: (ComponentHealth | undefined)[]): StatusBucket {
  let result: StatusBucket = "healthy";
  for (const c of checks) {
    if (!c?.status || c.status === "STATUS_UNSPECIFIED") {
      if (result === "healthy") result = "unknown";
      continue;
    }
    if (c.status === "STATUS_DOWN") {
      return "down";
    }
    if (c.status === "STATUS_DEGRADED") {
      result = "degraded";
    }
  }
  return result;
}

function StatusPill() {
  const { data } = useOverview();
  const status = worstStatus(data?.health?.mongodb, data?.health?.redis, data?.health?.runner);
  const palette = {
    healthy: { cls: "border-white/25 bg-white/15 text-primary-foreground", label: "Healthy", icon: CircleCheck },
    degraded: { cls: "border-orange-200/40 bg-orange-300/20 text-primary-foreground", label: "Degraded", icon: CircleAlert },
    down: { cls: "border-red-200/40 bg-red-400/25 text-primary-foreground", label: "Down", icon: CircleAlert },
    unknown: { cls: "border-white/20 bg-white/10 text-primary-foreground/80", label: "Unknown", icon: CircleAlert },
  }[status];
  const Icon = palette.icon;
  return (
    <div className={`flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium ${palette.cls}`}>
      <Icon className="h-3.5 w-3.5" />
      Status: {palette.label}
    </div>
  );
}

function WorkspaceSwitcher() {
  const { workspaces, selectedWorkspaceId, selectedWorkspace, isLoading, setSelectedWorkspaceId } = useWorkspace();

  return (
    <div className="flex min-w-0 items-center gap-2 rounded-md border border-white/25 bg-white/10 px-2 py-1 text-primary-foreground">
      <Building2 className="h-4 w-4 text-primary-foreground/75" />
      <div className="hidden leading-tight sm:block">
        <div className="text-[10px] uppercase tracking-wide text-primary-foreground/70">Workspace</div>
        <div className="max-w-36 truncate text-xs font-medium">
          {selectedWorkspace?.name || selectedWorkspace?.slug || (isLoading ? "Loading..." : "Not selected")}
        </div>
      </div>
      <Select
        value={selectedWorkspaceId || ""}
        onValueChange={(value) => {
          if (value) setSelectedWorkspaceId(value);
        }}
        disabled={isLoading || workspaces.length === 0}
      >
        <SelectTrigger size="sm" className="w-36 border-white/25 bg-white/10 text-primary-foreground hover:bg-white/15 sm:w-44">
          <SelectValue placeholder={isLoading ? "Loading workspaces" : "Select workspace"} />
        </SelectTrigger>
        <SelectContent align="end">
          {workspaces.map((workspace) => (
            <SelectItem key={workspace.id} value={workspace.id}>
              <span className="flex flex-col items-start leading-tight">
                <span>{workspace.name || workspace.slug || workspace.id}</span>
                {workspace.slug ? <span className="text-[10px] text-muted-foreground">{workspace.slug}</span> : null}
              </span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

function isActiveNav(item: NavItem, pathname: string) {
  if (item.to === "/") return pathname === "/";
  return (item.activePrefixes ?? [item.to]).some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`));
}

function NavList({ items, isAdmin }: { items: NavItem[]; isAdmin: boolean }) {
  const location = useLocation();
  const { isCompact } = useLayoutDensity();
  return (
    <div className={cn(isCompact ? "space-y-0.5" : "space-y-1")}>
      {items
        .filter((item) => !item.adminOnly || isAdmin)
        .map(({ to, icon: Icon, label, activePrefixes, adminOnly }) => {
          const active = isActiveNav({ to, icon: Icon, label, activePrefixes, adminOnly }, location.pathname);
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

function SidebarNav({ isAdmin }: { isAdmin: boolean }) {
  const { isCompact } = useLayoutDensity();
  return (
    <nav className={cn("flex-1 overflow-y-auto px-3", isCompact ? "space-y-3 py-2" : "space-y-5 py-3")}>
      <NavList items={PRIMARY_NAV} isAdmin={isAdmin} />
      <div className={cn("border-t", isCompact ? "pt-3" : "pt-4")}>
        <div className="px-3 pb-2 text-[11px] font-medium text-muted-foreground">
          Settings
        </div>
        <NavList items={SECONDARY_NAV} isAdmin={isAdmin} />
      </div>
      {isAdmin ? (
        <div className={cn("border-t", isCompact ? "pt-3" : "pt-4")}>
          <div className="px-3 pb-2 text-[11px] font-medium text-muted-foreground">
            Admin
          </div>
          <NavList items={ADMIN_NAV} isAdmin={isAdmin} />
        </div>
      ) : null}
      <div className={cn("border-t", isCompact ? "pt-3" : "pt-4")}>
        <a
          className="flex items-center gap-3 rounded-md px-3 py-2 text-sm text-muted-foreground hover:bg-sidebar-accent/70 hover:text-foreground"
          href="https://github.com/orvice/butter"
          target="_blank"
          rel="noreferrer"
        >
          <BookOpen className="h-4 w-4" />
          Documentation
        </a>
        <a
          className="flex items-center gap-3 rounded-md px-3 py-2 text-sm text-muted-foreground hover:bg-sidebar-accent/70 hover:text-foreground"
          href="https://github.com/orvice/butter/issues"
          target="_blank"
          rel="noreferrer"
        >
          <LifeBuoy className="h-4 w-4" />
          Support
        </a>
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

function UserAvatarLink({ user }: { user: AuthUser | null }) {
  if (!user) return null;
  const avatar = user.avatar_url || user.avatarUrl || "";
  const name = user.display_name || user.displayName || user.username;
  const initial = (name || user.username || "?").trim().charAt(0).toUpperCase() || "?";
  return (
    <Link
      to="/profile"
      aria-label={`Profile of ${name}`}
      className="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full border bg-muted text-xs font-medium hover:opacity-80"
    >
      {avatar ? (
        <img src={avatar} alt="" className="h-full w-full object-cover" />
      ) : (
        <span>{initial}</span>
      )}
    </Link>
  );
}

export default function DashboardLayout() {
  const { isAuthenticated, isAdmin, logout, user } = useAuth();
  const { selectedWorkspaceId, workspaces, isLoading: isWorkspaceLoading } = useWorkspace();
  const { isCompact } = useLayoutDensity();
  const navigate = useNavigate();

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
        <div className={cn("px-4", isCompact ? "py-3" : "py-4")}>
          <Button className="w-full" onClick={() => navigate("/agents/create")}>
            <Plus className="mr-2 h-4 w-4" />
            Deploy Agent
          </Button>
        </div>
        <SidebarNav isAdmin={isAdmin} />
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
                <div className="px-4 py-4">
                  <Button className="w-full" onClick={() => navigate("/agents/create")}>
                    <Plus className="mr-2 h-4 w-4" />
                    Deploy Agent
                  </Button>
                </div>
                <SidebarNav isAdmin={isAdmin} />
              </SheetContent>
            </Sheet>
            <Link to="/" className="flex items-center gap-2.5 rounded-md outline-none focus-visible:ring-2 focus-visible:ring-white/40">
              <BrandMark size={30} />
              <span className="text-base font-medium leading-none tracking-tight text-primary-foreground">Butter</span>
            </Link>
          </div>
          <div className="hidden min-w-0 flex-1 items-center gap-3 md:flex">
            <div className="content-search max-w-xl">
              <Search className="h-4 w-4 text-current/70" />
              <Input
                className="h-full border-0 bg-transparent px-2 text-current shadow-none placeholder:text-current/80 focus-visible:border-transparent focus-visible:ring-0 dark:bg-transparent"
                placeholder="Search..."
              />
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <WorkspaceSwitcher />
            <StatusPill />
            <ThemeControls mode="menu" triggerClassName={headerIconBtn} />
            <Button variant="ghost" size="icon" aria-label="Storage status" className={cn(headerIconBtn, "hidden sm:inline-flex")}>
              <Database className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" aria-label="Refresh status" className={cn(headerIconBtn, "hidden sm:inline-flex")}>
              <RefreshCw className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" aria-label="Notifications" className={cn(headerIconBtn, "hidden sm:inline-flex")}>
              <Bell className="h-4 w-4" />
            </Button>
            <UserAvatarLink user={user} />
            <Button variant="ghost" size="icon" onClick={logout} aria-label="Sign out" className={headerIconBtn}>
              <LogOut className="h-4 w-4" />
            </Button>
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
