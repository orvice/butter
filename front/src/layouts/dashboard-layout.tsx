import { useMemo, useState, useSyncExternalStore } from "react";
import { Link, Navigate, Outlet, useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { useTheme } from "next-themes";
import { toast } from "sonner";
import { useAuth } from "@/hooks/use-auth";
import type { AuthUser } from "@/api/auth";
import { useWorkspace } from "@/hooks/use-workspace";
import { useSessions } from "@/api/sessions";
import { ButterLogo } from "@/components/butter/logo";
import { AgentAvatar } from "@/components/butter/primitives";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import type { SessionInfo } from "@/types/api";
import {
  Bot,
  Building2,
  Check,
  ChevronsUpDown,
  LayoutDashboard,
  LogOut,
  Menu as MenuIcon,
  PanelLeft,
  PanelLeftClose,
  Search,
  Settings,
  SquarePen,
  User,
  Workflow,
} from "lucide-react";

const CHAT_APP_NAME = "web-chat";
const SIDEBAR_HIDDEN_KEY = "butter.sidebar.hidden";

function sessionAgentName(state: SessionInfo["state"]): string | undefined {
  if (!state) return undefined;
  const v = state["agent_name"];
  return typeof v === "string" && v ? v : undefined;
}

function sessionTitle(session: SessionInfo): string {
  const state = session.state;
  const title = state?.["title"];
  if (typeof title === "string" && title.trim()) return title;
  return sessionAgentName(state) ?? session.session_id.slice(0, 12);
}

type SessionGroupKey = "today" | "week" | "older";

function sessionGroup(session: SessionInfo): SessionGroupKey {
  if (!session.last_update_time) return "older";
  const updated = new Date(session.last_update_time);
  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  if (updated >= startOfToday) return "today";
  if (now.getTime() - updated.getTime() < 7 * 24 * 60 * 60 * 1000) return "week";
  return "older";
}

function NavButton({
  active,
  icon,
  label,
  to,
  onNavigate,
}: {
  active: boolean;
  icon: React.ReactNode;
  label: string;
  to: string;
  onNavigate?: () => void;
}) {
  return (
    <Link
      to={to}
      onClick={onNavigate}
      aria-current={active ? "page" : undefined}
      className={cn(
        "flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-sm font-medium transition-colors",
        active
          ? "bg-sidebar-accent text-sidebar-accent-foreground"
          : "text-sidebar-foreground/80 hover:bg-sidebar-accent/60 hover:text-sidebar-foreground",
      )}
    >
      <span
        className={cn(
          "[&_svg]:size-4",
          active ? "text-sidebar-foreground" : "text-muted-foreground",
        )}
      >
        {icon}
      </span>
      {label}
    </Link>
  );
}

function ConversationGroup({
  title,
  items,
  activeSessionId,
  onNavigate,
}: {
  title: string;
  items: SessionInfo[];
  activeSessionId: string | null;
  onNavigate?: () => void;
}) {
  if (items.length === 0) return null;
  return (
    <div className="mb-1">
      <div className="px-2.5 py-1 text-[0.7rem] font-medium uppercase tracking-wide text-muted-foreground">
        {title}
      </div>
      <ul>
        {items.map((s) => {
          const agent = sessionAgentName(s.state);
          const active = s.session_id === activeSessionId;
          return (
            <li key={s.session_id}>
              <Link
                to={`/chat?session=${encodeURIComponent(s.session_id)}`}
                onClick={onNavigate}
                className={cn(
                  "group flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-sm transition-colors",
                  active
                    ? "bg-sidebar-accent text-sidebar-accent-foreground"
                    : "text-sidebar-foreground/75 hover:bg-sidebar-accent/60 hover:text-sidebar-foreground",
                )}
              >
                {agent && (
                  <AgentAvatar
                    name={agent}
                    size="sm"
                    className="size-4 rounded text-[0.6rem]"
                  />
                )}
                <span className="truncate">{sessionTitle(s)}</span>
              </Link>
            </li>
          );
        })}
      </ul>
    </div>
  );
}

function WorkspaceSwitcher() {
  const { workspaces, selectedWorkspaceId, selectedWorkspace, isLoading, setSelectedWorkspaceId } =
    useWorkspace();
  const name = selectedWorkspace?.name || (isLoading ? "Loading…" : "Select workspace");

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        disabled={isLoading || workspaces.length === 0}
        className="flex w-full items-center gap-2 rounded-md border border-sidebar-border bg-background/50 px-2.5 py-2 text-left text-sm hover:bg-sidebar-accent/60"
      >
        <span className="flex size-6 shrink-0 items-center justify-center rounded bg-primary text-[0.7rem] font-semibold text-primary-foreground">
          {(selectedWorkspace?.name || "?").charAt(0).toUpperCase()}
        </span>
        <span className="min-w-0 flex-1">
          <span className="block truncate font-medium">{name}</span>
          <span className="block truncate text-xs text-muted-foreground">Workspace</span>
        </span>
        <ChevronsUpDown className="size-4 shrink-0 text-muted-foreground" />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" sideOffset={6} className="w-56">
        <DropdownMenuLabel className="text-[0.7rem] font-medium uppercase tracking-wide text-muted-foreground">
          Workspaces
        </DropdownMenuLabel>
        {workspaces.map((w) => (
          <DropdownMenuItem key={w.id} onClick={() => setSelectedWorkspaceId(w.id)}>
            <Building2 />
            <span className="flex flex-1 items-center justify-between gap-2">
              <span className="truncate">{w.name || "Unnamed workspace"}</span>
              {w.id === selectedWorkspaceId && <Check className="size-4 text-success" />}
            </span>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function UserMenu({ user, logout }: { user: AuthUser | null; logout: () => void }) {
  const { theme, setTheme } = useTheme();
  if (!user) return null;
  const avatar = user.avatar_url || user.avatarUrl || "";
  const name = user.display_name || user.displayName || user.username;
  const initial = (name || user.username || "?").trim().charAt(0).toUpperCase() || "?";

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="flex w-full items-center gap-2 rounded-md px-1 py-1 text-left text-sm hover:bg-sidebar-accent/60">
        <span className="flex size-7 shrink-0 items-center justify-center overflow-hidden rounded-full bg-secondary text-xs font-semibold text-secondary-foreground">
          {avatar ? <img src={avatar} alt="" className="h-full w-full object-cover" /> : initial}
        </span>
        <span className="min-w-0 flex-1">
          <span className="block truncate font-medium">{name}</span>
          <span className="block truncate text-xs text-muted-foreground">@{user.username}</span>
        </span>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" sideOffset={6} className="w-56">
        <DropdownMenuItem render={<Link to="/profile" />}>
          <User />
          Profile
        </DropdownMenuItem>
        <DropdownMenuLabel className="pt-1 text-[0.7rem] font-medium uppercase tracking-wide text-muted-foreground">
          Appearance
        </DropdownMenuLabel>
        <div className="flex gap-1 p-1">
          {(["light", "dark", "system"] as const).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setTheme(t)}
              className={cn(
                "flex-1 rounded-md border px-1.5 py-1 text-xs capitalize transition-colors",
                theme === t
                  ? "border-ring bg-accent text-foreground"
                  : "border-border text-muted-foreground hover:bg-accent",
              )}
            >
              {t}
            </button>
          ))}
        </div>
        <DropdownMenuItem render={<Link to="/manage" />}>
          <Settings />
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

function SidebarContent({ onNavigate }: { onNavigate?: () => void }) {
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const { user, logout } = useAuth();
  const [query, setQuery] = useState("");

  const userId = user?.id ?? "";
  const sessionsQuery = useSessions(
    { app_name: CHAT_APP_NAME, user_id: userId || undefined, page_size: 100 },
    { enabled: !!userId },
  );

  const activeSessionId =
    location.pathname === "/chat" ? searchParams.get("session") : null;

  const filtered = useMemo(() => {
    const sessions = sessionsQuery.data?.sessions ?? [];
    const q = query.trim().toLowerCase();
    if (!q) return sessions;
    return sessions.filter((s) => sessionTitle(s).toLowerCase().includes(q));
  }, [sessionsQuery.data, query]);

  const groups = useMemo(
    () => ({
      today: filtered.filter((s) => sessionGroup(s) === "today"),
      week: filtered.filter((s) => sessionGroup(s) === "week"),
      older: filtered.filter((s) => sessionGroup(s) === "older"),
    }),
    [filtered],
  );

  const pathname = location.pathname;

  return (
    <div className="flex h-full flex-col bg-sidebar">
      {/* Brand + collapse */}
      <div className="flex items-center justify-between px-3 pb-2 pt-3">
        <Link to="/" onClick={onNavigate} className="rounded-md outline-none focus-visible:ring-2 focus-visible:ring-ring">
          <ButterLogo />
        </Link>
        <SidebarCollapseButton />
      </div>

      {/* New Chat — most prominent */}
      <div className="px-3 pb-2">
        <Link
          to="/chat?new=1"
          onClick={onNavigate}
          className="flex w-full items-center justify-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90"
        >
          <SquarePen className="size-4" />
          New Chat
        </Link>
      </div>

      {/* Primary nav */}
      <nav className="px-3 pb-2">
        <div className="flex flex-col gap-0.5">
          <NavButton
            to="/"
            active={pathname === "/"}
            icon={<LayoutDashboard />}
            label="Dashboard"
            onNavigate={onNavigate}
          />
          <NavButton
            to="/agents"
            active={pathname.startsWith("/agents")}
            icon={<Bot />}
            label="Agents"
            onNavigate={onNavigate}
          />
          <NavButton
            to="/automations"
            active={pathname.startsWith("/automations")}
            icon={<Workflow />}
            label="Automations"
            onNavigate={onNavigate}
          />
        </div>
      </nav>

      <div className="mx-3 mb-2 h-px bg-sidebar-border" />

      {/* Conversation search */}
      <div className="px-3 pb-2">
        <div className="relative">
          <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search chats"
            className="w-full rounded-md border border-sidebar-border bg-background/60 py-1.5 pl-8 pr-2 text-sm outline-none placeholder:text-muted-foreground focus:border-ring"
          />
        </div>
      </div>

      {/* Conversation history (scrolls) */}
      <div className="scrollbar-thin min-h-0 flex-1 overflow-y-auto px-1.5">
        {filtered.length === 0 ? (
          <p className="px-3 py-6 text-center text-sm text-muted-foreground">
            {sessionsQuery.isLoading ? "Loading chats…" : "No conversations found."}
          </p>
        ) : (
          <>
            <ConversationGroup
              title="Today"
              items={groups.today}
              activeSessionId={activeSessionId}
              onNavigate={onNavigate}
            />
            <ConversationGroup
              title="Previous 7 days"
              items={groups.week}
              activeSessionId={activeSessionId}
              onNavigate={onNavigate}
            />
            <ConversationGroup
              title="Older"
              items={groups.older}
              activeSessionId={activeSessionId}
              onNavigate={onNavigate}
            />
          </>
        )}
      </div>

      {/* Bottom: Manage, workspace switcher, user */}
      <div className="border-t border-sidebar-border p-3">
        <NavButton
          to="/manage"
          active={
            pathname !== "/" &&
            !pathname.startsWith("/chat") &&
            !pathname.startsWith("/agents") &&
            !pathname.startsWith("/automations")
          }
          icon={<Settings />}
          label="Manage"
          onNavigate={onNavigate}
        />
        <div className="mt-2">
          <WorkspaceSwitcher />
        </div>
        <div className="mt-2">
          <UserMenu user={user} logout={logout} />
        </div>
      </div>
    </div>
  );
}

/* Sidebar visibility is shared via a tiny external store so the collapse
   button (inside the sidebar) and the show button (in the content area) stay
   in sync without a context provider. */
const sidebarListeners = new Set<() => void>();
function readSidebarHidden(): boolean {
  return localStorage.getItem(SIDEBAR_HIDDEN_KEY) === "1";
}
function setSidebarHidden(hidden: boolean) {
  localStorage.setItem(SIDEBAR_HIDDEN_KEY, hidden ? "1" : "0");
  sidebarListeners.forEach((fn) => fn());
}
function subscribeSidebar(listener: () => void) {
  sidebarListeners.add(listener);
  return () => sidebarListeners.delete(listener);
}
function useSidebarHidden(): [boolean, (hidden: boolean) => void] {
  const hidden = useSyncExternalStore(subscribeSidebar, readSidebarHidden);
  return [hidden, setSidebarHidden];
}

function SidebarCollapseButton() {
  const [, setHidden] = useSidebarHidden();
  return (
    <button
      type="button"
      onClick={() => setHidden(true)}
      className="hidden rounded-md p-1.5 text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground md:inline-flex"
      aria-label="Hide sidebar"
    >
      <PanelLeftClose className="size-4" />
    </button>
  );
}

function SidebarShowButton() {
  const [hidden, setHidden] = useSidebarHidden();
  if (!hidden) return null;
  return (
    <button
      type="button"
      onClick={() => setHidden(false)}
      aria-label="Show sidebar"
      className="absolute left-3 top-3 z-30 hidden rounded-md border border-border bg-card p-1.5 text-muted-foreground shadow-sm hover:bg-muted md:inline-flex"
    >
      <PanelLeft className="size-4" />
    </button>
  );
}

function MobileTopBar({ onOpenNav }: { onOpenNav: () => void }) {
  const navigate = useNavigate();
  return (
    <header className="flex h-12 shrink-0 items-center justify-between border-b border-border px-3 md:hidden">
      <button
        type="button"
        onClick={onOpenNav}
        aria-label="Open navigation"
        className="rounded-md p-1.5 text-muted-foreground hover:bg-muted"
      >
        <MenuIcon className="size-5" />
      </button>
      <Link to="/">
        <ButterLogo />
      </Link>
      <button
        type="button"
        onClick={() => navigate("/chat?new=1")}
        aria-label="New chat"
        className="rounded-md p-1.5 text-muted-foreground hover:bg-muted"
      >
        <SquarePen className="size-5" />
      </button>
    </header>
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
    <div className="flex h-full items-center justify-center overflow-y-auto p-4">
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

export default function DashboardLayout() {
  const { isAuthenticated } = useAuth();
  const { selectedWorkspaceId, workspaces, isLoading: isWorkspaceLoading } = useWorkspace();
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const [sidebarHidden] = useSidebarHidden();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="flex h-dvh w-full overflow-hidden bg-background text-foreground">
      {/* Desktop sidebar */}
      {!sidebarHidden && (
        <aside className="hidden w-72 shrink-0 border-r border-sidebar-border md:block">
          <SidebarContent />
        </aside>
      )}

      {/* Mobile drawer */}
      {mobileNavOpen && (
        <div className="fixed inset-0 z-50 md:hidden">
          <div
            className="absolute inset-0 bg-black/40"
            onClick={() => setMobileNavOpen(false)}
            aria-hidden
          />
          <div className="absolute inset-y-0 left-0 w-[85%] max-w-xs border-r border-sidebar-border shadow-xl">
            <SidebarContent onNavigate={() => setMobileNavOpen(false)} />
          </div>
        </div>
      )}

      <div className="relative flex min-w-0 flex-1 flex-col">
        <SidebarShowButton />
        <MobileTopBar onOpenNav={() => setMobileNavOpen(true)} />
        <main className="relative min-h-0 flex-1 overflow-hidden">
          {selectedWorkspaceId ? (
            <Outlet />
          ) : isWorkspaceLoading ? (
            <div className="flex h-full items-center justify-center">
              <p className="text-sm text-muted-foreground">Loading workspaces…</p>
            </div>
          ) : workspaces.length === 0 ? (
            <WorkspaceCreateCard />
          ) : (
            <div className="flex h-full items-center justify-center">
              <p className="text-sm text-muted-foreground">Preparing your workspace…</p>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
