import { useState } from "react";
import { NavLink, Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { useWorkspace } from "@/hooks/use-workspace";
import { useTheme } from "next-themes";
import { useOverview } from "@/api/dashboard";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  LayoutDashboard,
  Bot,
  Server,
  Globe,
  MessageSquare,
  Clock,
  Sun,
  Moon,
  LogOut,
  Cpu,
  Cable,
  BrainCircuit,
  KeyRound,
  Users,
  Building2,
  CircleCheck,
  CircleAlert,
} from "lucide-react";
import type { ComponentHealth } from "@/types/api";

const NAV_GROUPS: { label: string; items: { to: string; icon: typeof LayoutDashboard; label: string }[] }[] = [
  {
    label: "Dashboard",
    items: [{ to: "/", icon: LayoutDashboard, label: "Overview" }],
  },
  {
    label: "Orchestration",
    items: [
      { to: "/agents", icon: Bot, label: "Agents" },
      { to: "/cron", icon: Clock, label: "Cron Jobs" },
    ],
  },
  {
    label: "Integrations",
    items: [
      { to: "/mcp-servers", icon: Server, label: "MCP Servers" },
      { to: "/remote-agents", icon: Globe, label: "Remote Agents" },
    ],
  },
  {
    label: "Execution",
    items: [
      { to: "/daemons", icon: Cpu, label: "Daemons" },
      { to: "/sessions", icon: MessageSquare, label: "Sessions" },
    ],
  },
  {
    label: "Settings",
    items: [
      { to: "/channels", icon: Cable, label: "Channels" },
      { to: "/model-providers", icon: BrainCircuit, label: "Model Providers" },
      { to: "/api-tokens", icon: KeyRound, label: "API Tokens" },
      { to: "/users", icon: Users, label: "Users" },
    ],
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
    healthy: { cls: "bg-green-500/10 text-green-600 border-green-500/20", label: "Healthy", icon: CircleCheck },
    degraded: { cls: "bg-amber-500/10 text-amber-600 border-amber-500/20", label: "Degraded", icon: CircleAlert },
    down: { cls: "bg-red-500/10 text-red-600 border-red-500/20", label: "Down", icon: CircleAlert },
    unknown: { cls: "bg-muted text-muted-foreground border-border", label: "Unknown", icon: CircleAlert },
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
    <div className="flex items-center gap-2 rounded-lg border bg-background px-2 py-1">
      <Building2 className="h-4 w-4 text-muted-foreground" />
      <div className="hidden leading-tight sm:block">
        <div className="text-[10px] uppercase tracking-wide text-muted-foreground">Workspace</div>
        <div className="max-w-36 truncate text-xs font-medium">
          {selectedWorkspace?.name || selectedWorkspace?.slug || (isLoading ? "Loading..." : "Not selected")}
        </div>
      </div>
      <Select
        value={selectedWorkspaceId || undefined}
        onValueChange={(value) => {
          if (value) setSelectedWorkspaceId(value);
        }}
        disabled={isLoading || workspaces.length === 0}
      >
        <SelectTrigger size="sm" className="w-44">
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

export default function DashboardLayout() {
  const { isAuthenticated, logout } = useAuth();
  const { theme, setTheme } = useTheme();
  const { selectedWorkspaceId, workspaces, isLoading: isWorkspaceLoading } = useWorkspace();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="flex min-h-screen bg-background">
      {/* Sidebar */}
      <aside className="flex w-60 flex-col border-r bg-card">
        <div className="flex items-center gap-2 p-4">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary text-primary-foreground font-bold">
            B
          </div>
          <div>
            <div className="text-sm font-bold leading-tight">Butter</div>
            <div className="text-[10px] uppercase tracking-wide text-muted-foreground">Orchestration</div>
          </div>
        </div>
        <Separator />
        <nav className="flex-1 space-y-4 overflow-y-auto px-2 py-3">
          {NAV_GROUPS.map((group) => (
            <div key={group.label}>
              <div className="px-3 pb-1 text-[10px] uppercase tracking-wider text-muted-foreground">
                {group.label}
              </div>
              <div className="space-y-0.5">
                {group.items.map(({ to, icon: Icon, label }) => (
                  <NavLink
                    key={to}
                    to={to}
                    end={to === "/"}
                    className={({ isActive }) =>
                      `flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors ${
                        isActive
                          ? "bg-accent text-accent-foreground font-medium"
                          : "text-muted-foreground hover:bg-accent/50 hover:text-foreground"
                      }`
                    }
                  >
                    <Icon className="h-4 w-4" />
                    {label}
                  </NavLink>
                ))}
              </div>
            </div>
          ))}
        </nav>
        <Separator />
        <div className="flex items-center justify-between p-3">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            aria-label="Toggle theme"
          >
            {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
          </Button>
          <Button variant="ghost" size="icon" onClick={logout} aria-label="Sign out">
            <LogOut className="h-4 w-4" />
          </Button>
        </div>
      </aside>

      {/* Main */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Topbar */}
        <header className="flex h-14 items-center justify-end gap-3 border-b bg-card/40 px-6">
          <WorkspaceSwitcher />
          <StatusPill />
          <Badge variant="outline" className="text-xs">Production</Badge>
        </header>
        <main className="flex-1 overflow-auto p-6">
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
