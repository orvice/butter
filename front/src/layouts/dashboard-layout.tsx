import { NavLink, Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
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
  KeyRound,
} from "lucide-react";

const NAV_ITEMS = [
  { to: "/", icon: LayoutDashboard, label: "Overview" },
  { to: "/agents", icon: Bot, label: "Agents" },
  { to: "/mcp-servers", icon: Server, label: "MCP Servers" },
  { to: "/remote-agents", icon: Globe, label: "Remote Agents" },
  { to: "/daemons", icon: Cpu, label: "Daemons" },
  { to: "/channels", icon: Cable, label: "Channels" },
  { to: "/sessions", icon: MessageSquare, label: "Sessions" },
  { to: "/cron", icon: Clock, label: "Cron Jobs" },
  { to: "/api-tokens", icon: KeyRound, label: "API Tokens" },
];

export default function DashboardLayout() {
  const { isAuthenticated, logout } = useAuth();
  const { theme, setTheme } = useTheme();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="flex min-h-screen">
      <aside className="flex w-56 flex-col border-r bg-card">
        <div className="p-4">
          <h1 className="text-lg font-bold">Butter</h1>
        </div>
        <nav className="flex-1 space-y-1 px-2">
          {NAV_ITEMS.map(({ to, icon: Icon, label }) => (
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
        </nav>
        <Separator />
        <div className="flex items-center justify-between p-3">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
          >
            {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
          </Button>
          <Button variant="ghost" size="icon" onClick={logout}>
            <LogOut className="h-4 w-4" />
          </Button>
        </div>
      </aside>
      <main className="flex-1 overflow-auto p-6">
        <Outlet />
      </main>
    </div>
  );
}
