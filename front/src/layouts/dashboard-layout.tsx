import { NavLink, Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { useTheme } from "next-themes";
import { useOverview } from "@/api/dashboard";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
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

export default function DashboardLayout() {
  const { isAuthenticated, logout } = useAuth();
  const { theme, setTheme } = useTheme();

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
          <StatusPill />
          <Badge variant="outline" className="text-xs">Production</Badge>
        </header>
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
