# Agent Management Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a React-based management dashboard for Butter's agent platform with CRUD for 5 services and a Cron execution monitoring page.

**Architecture:** Single-page app using Vite + React Router. Twirp API calls wrapped in TanStack Query hooks. shadcn/ui components with dark/light theme. All state lives in TanStack Query (server) or React Context (client).

**Tech Stack:** Vite, React 18, TypeScript, React Router 6, TanStack Query 5, shadcn/ui, Tailwind CSS, Recharts, Monaco Editor, next-themes, Sonner, zod, react-hook-form

**Spec:** `docs/superpowers/specs/2026-04-14-agent-dashboard-design.md`

---

## File Map

```
front/
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.app.json
├── tsconfig.node.json
├── vite.config.ts
├── tailwind.config.ts
├── postcss.config.js
├── components.json
├── .env.example
├── public/
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── index.css
│   ├── types/
│   │   └── api.ts
│   ├── api/
│   │   ├── client.ts
│   │   ├── agents.ts
│   │   ├── mcp-servers.ts
│   │   ├── remote-agents.ts
│   │   ├── sessions.ts
│   │   └── cron.ts
│   ├── hooks/
│   │   └── use-auth.ts
│   ├── lib/
│   │   ├── utils.ts
│   │   └── constants.ts
│   ├── components/
│   │   ├── ui/                    # shadcn generated
│   │   ├── data-table.tsx
│   │   ├── delete-dialog.tsx
│   │   └── page-header.tsx
│   ├── layouts/
│   │   └── dashboard-layout.tsx
│   └── pages/
│       ├── login.tsx
│       ├── dashboard.tsx
│       ├── agents/
│       │   ├── list.tsx
│       │   ├── create.tsx
│       │   └── edit.tsx
│       ├── mcp-servers/
│       │   ├── list.tsx
│       │   ├── create.tsx
│       │   └── edit.tsx
│       ├── remote-agents/
│       │   ├── list.tsx
│       │   ├── create.tsx
│       │   └── edit.tsx
│       ├── sessions/
│       │   ├── list.tsx
│       │   └── detail.tsx
│       └── cron/
│           ├── list.tsx
│           ├── create.tsx
│           ├── edit.tsx
│           └── executions.tsx
```

---

### Task 1: Project Scaffolding

**Files:**
- Create: `front/` (entire Vite project)
- Create: `front/.env.example`

- [ ] **Step 1: Scaffold Vite project**

```bash
cd /Users/orvice/workspace/go/orvice/butter
npm create vite@latest front -- --template react-ts
```

- [ ] **Step 2: Install core dependencies**

```bash
cd front
npm install react-router-dom @tanstack/react-query next-themes sonner recharts @monaco-editor/react
npm install -D tailwindcss @tailwindcss/vite
```

- [ ] **Step 3: Configure Tailwind**

Replace `front/src/index.css` with:

```css
@import "tailwindcss";
```

Add Tailwind Vite plugin in `front/vite.config.ts`:

```ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});
```

Update `front/tsconfig.app.json` to add path aliases:

```json
{
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.app.tsbuildinfo",
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "verbatimModuleSyntax": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedSideEffectImports": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"]
}
```

- [ ] **Step 4: Initialize shadcn/ui**

```bash
cd front
npx shadcn@latest init -d
```

This creates `components.json` and sets up the `src/components/ui/` directory. Choose "New York" style, Zinc base color.

- [ ] **Step 5: Add required shadcn components**

```bash
npx shadcn@latest add button card input label select switch textarea table badge dropdown-menu dialog alert-dialog form tabs skeleton separator breadcrumb toast sonner sheet
```

- [ ] **Step 6: Install form dependencies**

```bash
npm install react-hook-form @hookform/resolvers zod
```

- [ ] **Step 7: Create .env.example**

Create `front/.env.example`:

```
VITE_API_BASE_URL=http://localhost:8080
```

- [ ] **Step 8: Verify project builds**

```bash
cd front
cp .env.example .env
npm run dev
```

Open http://localhost:5173 — should show the default Vite+React page.

- [ ] **Step 9: Clean up default files**

Remove `front/src/App.css` and the default Vite content from `front/src/App.tsx`. Replace `front/src/App.tsx` with:

```tsx
function App() {
  return <div className="min-h-screen bg-background text-foreground">Butter Dashboard</div>;
}

export default App;
```

- [ ] **Step 10: Commit**

```bash
git add front/
git commit -m "feat(front): scaffold Vite + React + Tailwind + shadcn/ui project"
```

---

### Task 2: TypeScript Types + Constants

**Files:**
- Create: `front/src/types/api.ts`
- Create: `front/src/lib/constants.ts`
- Create: `front/src/lib/utils.ts` (already created by shadcn, verify)

- [ ] **Step 1: Write API types**

Create `front/src/types/api.ts`:

```ts
// --- Enums ---

export type AgentType =
  | "AGENT_TYPE_UNSPECIFIED"
  | "AGENT_TYPE_LLM"
  | "AGENT_TYPE_LOOP"
  | "AGENT_TYPE_SEQUENTIAL"
  | "AGENT_TYPE_PARALLEL";

export type MCPServerTransport =
  | "MCP_SERVER_TRANSPORT_UNSPECIFIED"
  | "MCP_SERVER_TRANSPORT_STDIO"
  | "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP"
  | "MCP_SERVER_TRANSPORT_SSE";

export type RemoteAgentProtocol =
  | "REMOTE_AGENT_PROTOCOL_UNSPECIFIED"
  | "REMOTE_AGENT_PROTOCOL_A2A";

export type CronDeliveryType =
  | "CRON_DELIVERY_TYPE_UNSPECIFIED"
  | "CRON_DELIVERY_TYPE_LOG"
  | "CRON_DELIVERY_TYPE_WEBHOOK"
  | "CRON_DELIVERY_TYPE_CHANNEL";

export type CronExecutionStatus =
  | "CRON_EXECUTION_STATUS_UNSPECIFIED"
  | "CRON_EXECUTION_STATUS_SUCCESS"
  | "CRON_EXECUTION_STATUS_ERROR";

export type StreamingMode =
  | "STREAMING_MODE_UNSPECIFIED"
  | "STREAMING_MODE_NONE"
  | "STREAMING_MODE_SSE";

export type LLMIncludeContents =
  | "LLM_INCLUDE_CONTENTS_UNSPECIFIED"
  | "LLM_INCLUDE_CONTENTS_DEFAULT"
  | "LLM_INCLUDE_CONTENTS_NONE";

export type ContextGuardStrategy =
  | "CONTEXT_GUARD_STRATEGY_UNSPECIFIED"
  | "CONTEXT_GUARD_STRATEGY_THRESHOLD"
  | "CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW";

// --- Models ---

export interface ContextGuardConfig {
  strategy?: ContextGuardStrategy;
  max_turns?: number;
  max_tokens?: number;
}

export interface AgentRuntime {
  streaming_mode?: StreamingMode;
  save_input_blobs_as_artifacts?: boolean;
}

export interface AgentConfig {
  runtime?: AgentRuntime;
  mcp_servers?: MCPServer[];
  context_guard?: ContextGuardConfig;
  mcp_server_ids?: string[];
  remote_agent_ids?: string[];
  model?: string;
  instruction?: string;
  global_instruction?: string;
  disallow_transfer_to_parent?: boolean;
  disallow_transfer_to_peers?: boolean;
  include_contents?: LLMIncludeContents;
  output_key?: string;
  input_schema_json?: string;
  output_schema_json?: string;
  max_iterations?: number;
}

export interface Agent {
  name: string;
  description?: string;
  sub_agents?: Agent[];
  labels?: Record<string, string>;
  metadata?: Record<string, string>;
  config?: AgentConfig;
  type?: AgentType;
  enable_a2a?: boolean;
}

export interface MCPServer {
  id?: string;
  name: string;
  transport?: MCPServerTransport;
  command?: string;
  args?: string[];
  env?: Record<string, string>;
  url?: string;
  headers?: Record<string, string>;
  tool_filter?: string[];
  metadata?: Record<string, string>;
}

export interface RemoteAgent {
  id: string;
  name: string;
  url: string;
  protocol?: RemoteAgentProtocol;
}

export interface SessionInfo {
  session_id: string;
  app_name: string;
  user_id: string;
  state?: Record<string, unknown>;
  last_update_time?: string;
}

export interface SessionEvent {
  event_id: string;
  invocation_id?: string;
  author?: string;
  branch?: string;
  content_json?: string;
  timestamp?: string;
}

export interface SessionDetail {
  session: SessionInfo;
  events: SessionEvent[];
}

export interface CronDelivery {
  type?: CronDeliveryType;
  webhook_url?: string;
  channel_name?: string;
  chat_id?: string;
}

export interface CronJob {
  name: string;
  schedule: string;
  agent_name: string;
  input?: string;
  timezone?: string;
  enabled?: boolean;
  delivery?: CronDelivery;
  metadata?: Record<string, string>;
}

export interface CronExecution {
  id: string;
  job_name: string;
  agent_name: string;
  status: CronExecutionStatus;
  input?: string;
  output?: string;
  started_at?: string;
  finished_at?: string;
}

// --- Twirp Error ---

export interface TwirpError {
  code: string;
  msg: string;
}
```

- [ ] **Step 2: Write constants**

Create `front/src/lib/constants.ts`:

```ts
export const AGENT_TYPE_LABELS: Record<string, string> = {
  AGENT_TYPE_UNSPECIFIED: "Unspecified",
  AGENT_TYPE_LLM: "LLM",
  AGENT_TYPE_LOOP: "Loop",
  AGENT_TYPE_SEQUENTIAL: "Sequential",
  AGENT_TYPE_PARALLEL: "Parallel",
};

export const MCP_TRANSPORT_LABELS: Record<string, string> = {
  MCP_SERVER_TRANSPORT_UNSPECIFIED: "Unspecified",
  MCP_SERVER_TRANSPORT_STDIO: "Stdio",
  MCP_SERVER_TRANSPORT_STREAMABLE_HTTP: "HTTP",
  MCP_SERVER_TRANSPORT_SSE: "SSE",
};

export const CRON_DELIVERY_LABELS: Record<string, string> = {
  CRON_DELIVERY_TYPE_UNSPECIFIED: "Unspecified",
  CRON_DELIVERY_TYPE_LOG: "Log",
  CRON_DELIVERY_TYPE_WEBHOOK: "Webhook",
  CRON_DELIVERY_TYPE_CHANNEL: "Channel",
};

export const CRON_STATUS_LABELS: Record<string, string> = {
  CRON_EXECUTION_STATUS_UNSPECIFIED: "Unknown",
  CRON_EXECUTION_STATUS_SUCCESS: "Success",
  CRON_EXECUTION_STATUS_ERROR: "Error",
};

export const TOKEN_KEY = "butter_token";
```

- [ ] **Step 3: Verify TypeScript compiles**

```bash
cd front
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add front/src/types/ front/src/lib/constants.ts
git commit -m "feat(front): add TypeScript API types and constants"
```

---

### Task 3: API Client

**Files:**
- Create: `front/src/api/client.ts`

- [ ] **Step 1: Write Twirp client**

Create `front/src/api/client.ts`:

```ts
import type { TwirpError } from "@/types/api";
import { TOKEN_KEY } from "@/lib/constants";

const BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

export class ApiError extends Error {
  code: string;
  constructor(code: string, msg: string) {
    super(msg);
    this.code = code;
    this.name = "ApiError";
  }
}

export async function twirpFetch<TReq, TRes>(
  service: string,
  method: string,
  body: TReq,
): Promise<TRes> {
  const token = localStorage.getItem(TOKEN_KEY);
  const url = `${BASE_URL}/api/${service}/${method}`;

  const res = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(body),
  });

  if (res.status === 401) {
    localStorage.removeItem(TOKEN_KEY);
    window.location.href = "/login";
    throw new ApiError("unauthenticated", "Invalid or expired token");
  }

  const data = await res.json();

  if (!res.ok) {
    const err = data as TwirpError;
    throw new ApiError(err.code || "unknown", err.msg || "Unknown error");
  }

  return data as TRes;
}

export async function validateToken(token: string): Promise<boolean> {
  const url = `${BASE_URL}/api/agents.v1.AgentService/ListAgents`;
  const res = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({}),
  });
  return res.ok;
}
```

- [ ] **Step 2: Verify compiles**

```bash
cd front && npx tsc --noEmit
```

- [ ] **Step 3: Commit**

```bash
git add front/src/api/client.ts
git commit -m "feat(front): add Twirp API client with auth handling"
```

---

### Task 4: Auth Hook + Login Page

**Files:**
- Create: `front/src/hooks/use-auth.ts`
- Create: `front/src/pages/login.tsx`

- [ ] **Step 1: Write auth hook**

Create `front/src/hooks/use-auth.ts`:

```tsx
import { createContext, useContext, useState, useCallback, type ReactNode } from "react";
import { TOKEN_KEY } from "@/lib/constants";
import { validateToken } from "@/api/client";

interface AuthContextValue {
  token: string | null;
  isAuthenticated: boolean;
  login: (token: string) => Promise<boolean>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(
    () => localStorage.getItem(TOKEN_KEY),
  );

  const login = useCallback(async (newToken: string): Promise<boolean> => {
    localStorage.setItem(TOKEN_KEY, newToken);
    const valid = await validateToken(newToken);
    if (valid) {
      setToken(newToken);
      return true;
    }
    localStorage.removeItem(TOKEN_KEY);
    return false;
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
  }, []);

  return (
    <AuthContext.Provider value={{ token, isAuthenticated: !!token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be inside AuthProvider");
  return ctx;
}
```

- [ ] **Step 2: Write login page**

Create `front/src/pages/login.tsx`:

```tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function LoginPage() {
  const [tokenInput, setTokenInput] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const ok = await login(tokenInput.trim());
      if (ok) {
        navigate("/");
      } else {
        setError("Invalid token. Please check and try again.");
      }
    } catch {
      setError("Connection failed. Is the server running?");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Butter</CardTitle>
          <CardDescription>Enter your API token to connect</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              type="password"
              placeholder="Paste your API token"
              value={tokenInput}
              onChange={(e) => setTokenInput(e.target.value)}
              disabled={loading}
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full" disabled={loading || !tokenInput.trim()}>
              {loading ? "Connecting..." : "Connect"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
```

- [ ] **Step 3: Verify compiles**

```bash
cd front && npx tsc --noEmit
```

- [ ] **Step 4: Commit**

```bash
git add front/src/hooks/use-auth.ts front/src/pages/login.tsx
git commit -m "feat(front): add auth context and login page"
```

---

### Task 5: Dashboard Layout + Theme + Router

**Files:**
- Create: `front/src/layouts/dashboard-layout.tsx`
- Create: `front/src/components/page-header.tsx`
- Modify: `front/src/App.tsx`
- Modify: `front/src/main.tsx`

- [ ] **Step 1: Write dashboard layout with sidebar**

Create `front/src/layouts/dashboard-layout.tsx`:

```tsx
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
} from "lucide-react";

const NAV_ITEMS = [
  { to: "/", icon: LayoutDashboard, label: "Dashboard" },
  { to: "/agents", icon: Bot, label: "Agents" },
  { to: "/mcp-servers", icon: Server, label: "MCP Servers" },
  { to: "/remote-agents", icon: Globe, label: "Remote Agents" },
  { to: "/sessions", icon: MessageSquare, label: "Sessions" },
  { to: "/cron", icon: Clock, label: "Cron Jobs" },
];

export default function DashboardLayout() {
  const { isAuthenticated, logout } = useAuth();
  const { theme, setTheme } = useTheme();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
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

      {/* Main content */}
      <main className="flex-1 overflow-auto p-6">
        <Outlet />
      </main>
    </div>
  );
}
```

- [ ] **Step 2: Write page header component**

Create `front/src/components/page-header.tsx`:

```tsx
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

interface PageHeaderProps {
  title: string;
  createLabel?: string;
  createTo?: string;
}

export function PageHeader({ title, createLabel, createTo }: PageHeaderProps) {
  const navigate = useNavigate();
  return (
    <div className="mb-6 flex items-center justify-between">
      <h2 className="text-2xl font-bold tracking-tight">{title}</h2>
      {createLabel && createTo && (
        <Button onClick={() => navigate(createTo)}>
          <Plus className="mr-2 h-4 w-4" />
          {createLabel}
        </Button>
      )}
    </div>
  );
}
```

- [ ] **Step 3: Write App.tsx with router**

Replace `front/src/App.tsx`:

```tsx
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "next-themes";
import { Toaster } from "@/components/ui/sonner";
import { AuthProvider } from "@/hooks/use-auth";
import DashboardLayout from "@/layouts/dashboard-layout";
import LoginPage from "@/pages/login";
import DashboardPage from "@/pages/dashboard";
import AgentListPage from "@/pages/agents/list";
import AgentCreatePage from "@/pages/agents/create";
import AgentEditPage from "@/pages/agents/edit";
import MCPServerListPage from "@/pages/mcp-servers/list";
import MCPServerCreatePage from "@/pages/mcp-servers/create";
import MCPServerEditPage from "@/pages/mcp-servers/edit";
import RemoteAgentListPage from "@/pages/remote-agents/list";
import RemoteAgentCreatePage from "@/pages/remote-agents/create";
import RemoteAgentEditPage from "@/pages/remote-agents/edit";
import SessionListPage from "@/pages/sessions/list";
import SessionDetailPage from "@/pages/sessions/detail";
import CronJobListPage from "@/pages/cron/list";
import CronJobCreatePage from "@/pages/cron/create";
import CronJobEditPage from "@/pages/cron/edit";
import CronExecutionsPage from "@/pages/cron/executions";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
});

export default function App() {
  return (
    <BrowserRouter>
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
        <QueryClientProvider client={queryClient}>
          <AuthProvider>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route element={<DashboardLayout />}>
                <Route index element={<DashboardPage />} />
                <Route path="agents" element={<AgentListPage />} />
                <Route path="agents/create" element={<AgentCreatePage />} />
                <Route path="agents/:name/edit" element={<AgentEditPage />} />
                <Route path="mcp-servers" element={<MCPServerListPage />} />
                <Route path="mcp-servers/create" element={<MCPServerCreatePage />} />
                <Route path="mcp-servers/:id/edit" element={<MCPServerEditPage />} />
                <Route path="remote-agents" element={<RemoteAgentListPage />} />
                <Route path="remote-agents/create" element={<RemoteAgentCreatePage />} />
                <Route path="remote-agents/:id/edit" element={<RemoteAgentEditPage />} />
                <Route path="sessions" element={<SessionListPage />} />
                <Route path="sessions/detail" element={<SessionDetailPage />} />
                <Route path="cron" element={<CronJobListPage />} />
                <Route path="cron/create" element={<CronJobCreatePage />} />
                <Route path="cron/:name/edit" element={<CronJobEditPage />} />
                <Route path="cron/:name/executions" element={<CronExecutionsPage />} />
              </Route>
            </Routes>
            <Toaster />
          </AuthProvider>
        </QueryClientProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
}
```

- [ ] **Step 4: Update main.tsx**

Replace `front/src/main.tsx`:

```tsx
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
```

- [ ] **Step 5: Create stub pages for all routes**

Create placeholder pages so the app compiles. Each page is a single exported default component with the page title. Create these files:

`front/src/pages/dashboard.tsx`:
```tsx
export default function DashboardPage() {
  return <div>Dashboard</div>;
}
```

`front/src/pages/agents/list.tsx`:
```tsx
export default function AgentListPage() {
  return <div>Agent List</div>;
}
```

`front/src/pages/agents/create.tsx`:
```tsx
export default function AgentCreatePage() {
  return <div>Create Agent</div>;
}
```

`front/src/pages/agents/edit.tsx`:
```tsx
export default function AgentEditPage() {
  return <div>Edit Agent</div>;
}
```

`front/src/pages/mcp-servers/list.tsx`:
```tsx
export default function MCPServerListPage() {
  return <div>MCP Server List</div>;
}
```

`front/src/pages/mcp-servers/create.tsx`:
```tsx
export default function MCPServerCreatePage() {
  return <div>Create MCP Server</div>;
}
```

`front/src/pages/mcp-servers/edit.tsx`:
```tsx
export default function MCPServerEditPage() {
  return <div>Edit MCP Server</div>;
}
```

`front/src/pages/remote-agents/list.tsx`:
```tsx
export default function RemoteAgentListPage() {
  return <div>Remote Agent List</div>;
}
```

`front/src/pages/remote-agents/create.tsx`:
```tsx
export default function RemoteAgentCreatePage() {
  return <div>Create Remote Agent</div>;
}
```

`front/src/pages/remote-agents/edit.tsx`:
```tsx
export default function RemoteAgentEditPage() {
  return <div>Edit Remote Agent</div>;
}
```

`front/src/pages/sessions/list.tsx`:
```tsx
export default function SessionListPage() {
  return <div>Session List</div>;
}
```

`front/src/pages/sessions/detail.tsx`:
```tsx
export default function SessionDetailPage() {
  return <div>Session Detail</div>;
}
```

`front/src/pages/cron/list.tsx`:
```tsx
export default function CronJobListPage() {
  return <div>Cron Job List</div>;
}
```

`front/src/pages/cron/create.tsx`:
```tsx
export default function CronJobCreatePage() {
  return <div>Create Cron Job</div>;
}
```

`front/src/pages/cron/edit.tsx`:
```tsx
export default function CronJobEditPage() {
  return <div>Edit Cron Job</div>;
}
```

`front/src/pages/cron/executions.tsx`:
```tsx
export default function CronExecutionsPage() {
  return <div>Cron Executions</div>;
}
```

- [ ] **Step 6: Verify app compiles and runs**

```bash
cd front && npx tsc --noEmit && npm run dev
```

Open http://localhost:5173 — should redirect to `/login`. Enter any token. Should show the sidebar layout with stub pages.

- [ ] **Step 7: Commit**

```bash
git add front/src/
git commit -m "feat(front): add dashboard layout, sidebar, theme toggle, and router with stub pages"
```

---

### Task 6: API Hooks for All Services

**Files:**
- Create: `front/src/api/agents.ts`
- Create: `front/src/api/mcp-servers.ts`
- Create: `front/src/api/remote-agents.ts`
- Create: `front/src/api/sessions.ts`
- Create: `front/src/api/cron.ts`

- [ ] **Step 1: Write agents API hooks**

Create `front/src/api/agents.ts`:

```ts
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { Agent } from "@/types/api";

const SVC = "agents.v1.AgentService";

function listAgents() {
  return twirpFetch<object, { agents: Agent[] }>(SVC, "ListAgents", {});
}

function getAgent(name: string) {
  return twirpFetch<{ name: string }, { agent: Agent }>(SVC, "GetAgent", { name });
}

function createAgent(agent: Agent) {
  return twirpFetch<{ agent: Agent }, { agent: Agent }>(SVC, "CreateAgent", { agent });
}

function updateAgent(agent: Agent) {
  return twirpFetch<{ agent: Agent }, { agent: Agent }>(SVC, "UpdateAgent", { agent });
}

function deleteAgent(name: string) {
  return twirpFetch<{ name: string }, object>(SVC, "DeleteAgent", { name });
}

export function useAgents() {
  return useQuery({ queryKey: ["agents"], queryFn: listAgents });
}

export function useAgent(name: string) {
  return useQuery({ queryKey: ["agents", name], queryFn: () => getAgent(name), enabled: !!name });
}

export function useCreateAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agents"] }),
  });
}

export function useUpdateAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateAgent,
    onSuccess: (_data, agent) => {
      qc.invalidateQueries({ queryKey: ["agents"] });
      qc.invalidateQueries({ queryKey: ["agents", agent.name] });
    },
  });
}

export function useDeleteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agents"] }),
  });
}
```

- [ ] **Step 2: Write MCP servers API hooks**

Create `front/src/api/mcp-servers.ts`:

```ts
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { MCPServer } from "@/types/api";

const SVC = "agents.v1.MCPServerService";

function listMCPServers() {
  return twirpFetch<object, { mcp_servers: MCPServer[] }>(SVC, "ListMCPServers", {});
}

function getMCPServer(id: string) {
  return twirpFetch<{ id: string }, { mcp_server: MCPServer }>(SVC, "GetMCPServer", { id });
}

function createMCPServer(mcp_server: MCPServer) {
  return twirpFetch<{ mcp_server: MCPServer }, { mcp_server: MCPServer }>(SVC, "CreateMCPServer", { mcp_server });
}

function updateMCPServer(mcp_server: MCPServer) {
  return twirpFetch<{ mcp_server: MCPServer }, { mcp_server: MCPServer }>(SVC, "UpdateMCPServer", { mcp_server });
}

function deleteMCPServer(id: string) {
  return twirpFetch<{ id: string }, object>(SVC, "DeleteMCPServer", { id });
}

export function useMCPServers() {
  return useQuery({ queryKey: ["mcp-servers"], queryFn: listMCPServers });
}

export function useMCPServer(id: string) {
  return useQuery({ queryKey: ["mcp-servers", id], queryFn: () => getMCPServer(id), enabled: !!id });
}

export function useCreateMCPServer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createMCPServer,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["mcp-servers"] }),
  });
}

export function useUpdateMCPServer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateMCPServer,
    onSuccess: (_data, server) => {
      qc.invalidateQueries({ queryKey: ["mcp-servers"] });
      qc.invalidateQueries({ queryKey: ["mcp-servers", server.id] });
    },
  });
}

export function useDeleteMCPServer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteMCPServer,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["mcp-servers"] }),
  });
}
```

- [ ] **Step 3: Write remote agents API hooks**

Create `front/src/api/remote-agents.ts`:

```ts
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { RemoteAgent } from "@/types/api";

const SVC = "agents.v1.RemoteAgentService";

function listRemoteAgents() {
  return twirpFetch<object, { remote_agents: RemoteAgent[] }>(SVC, "ListRemoteAgents", {});
}

function getRemoteAgent(id: string) {
  return twirpFetch<{ id: string }, { remote_agent: RemoteAgent }>(SVC, "GetRemoteAgent", { id });
}

function createRemoteAgent(remote_agent: RemoteAgent) {
  return twirpFetch<{ remote_agent: RemoteAgent }, { remote_agent: RemoteAgent }>(SVC, "CreateRemoteAgent", { remote_agent });
}

function updateRemoteAgent(remote_agent: RemoteAgent) {
  return twirpFetch<{ remote_agent: RemoteAgent }, { remote_agent: RemoteAgent }>(SVC, "UpdateRemoteAgent", { remote_agent });
}

function deleteRemoteAgent(id: string) {
  return twirpFetch<{ id: string }, object>(SVC, "DeleteRemoteAgent", { id });
}

export function useRemoteAgents() {
  return useQuery({ queryKey: ["remote-agents"], queryFn: listRemoteAgents });
}

export function useRemoteAgent(id: string) {
  return useQuery({ queryKey: ["remote-agents", id], queryFn: () => getRemoteAgent(id), enabled: !!id });
}

export function useCreateRemoteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createRemoteAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["remote-agents"] }),
  });
}

export function useUpdateRemoteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateRemoteAgent,
    onSuccess: (_data, agent) => {
      qc.invalidateQueries({ queryKey: ["remote-agents"] });
      qc.invalidateQueries({ queryKey: ["remote-agents", agent.id] });
    },
  });
}

export function useDeleteRemoteAgent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteRemoteAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["remote-agents"] }),
  });
}
```

- [ ] **Step 4: Write sessions API hooks**

Create `front/src/api/sessions.ts`:

```ts
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { SessionInfo, SessionDetail } from "@/types/api";

const SVC = "agents.v1.SessionService";

interface ListSessionsParams {
  app_name: string;
  user_id: string;
}

interface GetSessionParams {
  app_name: string;
  user_id: string;
  session_id: string;
  num_recent_events?: number;
}

interface DeleteSessionParams {
  app_name: string;
  user_id: string;
  session_id: string;
}

function listSessions(params: ListSessionsParams) {
  return twirpFetch<ListSessionsParams, { sessions: SessionInfo[] }>(SVC, "ListSessions", params);
}

function getSession(params: GetSessionParams) {
  return twirpFetch<GetSessionParams, { session_detail: SessionDetail }>(SVC, "GetSession", params);
}

function deleteSession(params: DeleteSessionParams) {
  return twirpFetch<DeleteSessionParams, object>(SVC, "DeleteSession", params);
}

export function useSessions(appName: string, userId: string) {
  return useQuery({
    queryKey: ["sessions", { appName, userId }],
    queryFn: () => listSessions({ app_name: appName, user_id: userId }),
    enabled: !!appName && !!userId,
  });
}

export function useSession(appName: string, userId: string, sessionId: string) {
  return useQuery({
    queryKey: ["sessions", { appName, userId, sessionId }],
    queryFn: () => getSession({ app_name: appName, user_id: userId, session_id: sessionId }),
    enabled: !!appName && !!userId && !!sessionId,
  });
}

export function useDeleteSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteSession,
    onSuccess: (_data, params) => {
      qc.invalidateQueries({ queryKey: ["sessions", { appName: params.app_name, userId: params.user_id }] });
    },
  });
}
```

- [ ] **Step 5: Write cron API hooks**

Create `front/src/api/cron.ts`:

```ts
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { twirpFetch } from "./client";
import type { CronJob, CronExecution } from "@/types/api";

const SVC = "agents.v1.CronJobService";

function listCronJobs() {
  return twirpFetch<object, { cron_jobs: CronJob[] }>(SVC, "ListCronJobs", {});
}

function getCronJob(name: string) {
  return twirpFetch<{ name: string }, { cron_job: CronJob }>(SVC, "GetCronJob", { name });
}

function createCronJob(cron_job: CronJob) {
  return twirpFetch<{ cron_job: CronJob }, { cron_job: CronJob }>(SVC, "CreateCronJob", { cron_job });
}

function updateCronJob(cron_job: CronJob) {
  return twirpFetch<{ cron_job: CronJob }, { cron_job: CronJob }>(SVC, "UpdateCronJob", { cron_job });
}

function deleteCronJob(name: string) {
  return twirpFetch<{ name: string }, { cron_job: CronJob }>(SVC, "DeleteCronJob", { name });
}

interface ListExecutionsParams {
  job_name?: string;
  page_size?: number;
  page_token?: string;
}

function listCronExecutions(params: ListExecutionsParams) {
  return twirpFetch<ListExecutionsParams, { executions: CronExecution[]; next_page_token: string }>(
    SVC,
    "ListCronExecutions",
    params,
  );
}

export function useCronJobs() {
  return useQuery({ queryKey: ["cron-jobs"], queryFn: listCronJobs });
}

export function useCronJob(name: string) {
  return useQuery({ queryKey: ["cron-jobs", name], queryFn: () => getCronJob(name), enabled: !!name });
}

export function useCreateCronJob() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createCronJob,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["cron-jobs"] }),
  });
}

export function useUpdateCronJob() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateCronJob,
    onSuccess: (_data, job) => {
      qc.invalidateQueries({ queryKey: ["cron-jobs"] });
      qc.invalidateQueries({ queryKey: ["cron-jobs", job.name] });
    },
  });
}

export function useDeleteCronJob() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteCronJob,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["cron-jobs"] }),
  });
}

export function useCronExecutions(jobName?: string, pageSize?: number, pageToken?: string) {
  return useQuery({
    queryKey: ["cron-executions", { jobName, page: pageToken }],
    queryFn: () => listCronExecutions({ job_name: jobName, page_size: pageSize, page_token: pageToken }),
  });
}

export function useDashboardExecutions() {
  return useQuery({
    queryKey: ["cron-executions", { jobName: "", page: "" }],
    queryFn: () => listCronExecutions({ page_size: 100 }),
    refetchInterval: 60_000,
  });
}
```

- [ ] **Step 6: Verify compiles**

```bash
cd front && npx tsc --noEmit
```

- [ ] **Step 7: Commit**

```bash
git add front/src/api/
git commit -m "feat(front): add TanStack Query API hooks for all 5 services"
```

---

### Task 7: Shared Components — DataTable + DeleteDialog

**Files:**
- Create: `front/src/components/data-table.tsx`
- Create: `front/src/components/delete-dialog.tsx`

- [ ] **Step 1: Write reusable data table**

Create `front/src/components/data-table.tsx`:

```tsx
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";

export interface Column<T> {
  header: string;
  accessorKey?: keyof T;
  cell?: (row: T) => React.ReactNode;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[] | undefined;
  isLoading: boolean;
  emptyMessage?: string;
}

export function DataTable<T>({ columns, data, isLoading, emptyMessage = "No data" }: DataTableProps<T>) {
  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-lg border border-dashed p-12 text-center">
        <p className="text-muted-foreground">{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            {columns.map((col) => (
              <TableHead key={col.header}>{col.header}</TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((row, i) => (
            <TableRow key={i}>
              {columns.map((col) => (
                <TableCell key={col.header}>
                  {col.cell
                    ? col.cell(row)
                    : col.accessorKey
                      ? String((row as Record<string, unknown>)[col.accessorKey as string] ?? "")
                      : null}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
```

- [ ] **Step 2: Write reusable delete dialog**

Create `front/src/components/delete-dialog.tsx`:

```tsx
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface DeleteDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  onConfirm: () => void;
  loading?: boolean;
}

export function DeleteDialog({ open, onOpenChange, title, description, onConfirm, loading }: DeleteDialogProps) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={loading}>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={onConfirm} disabled={loading} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
            {loading ? "Deleting..." : "Delete"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
```

- [ ] **Step 3: Verify compiles**

```bash
cd front && npx tsc --noEmit
```

- [ ] **Step 4: Commit**

```bash
git add front/src/components/data-table.tsx front/src/components/delete-dialog.tsx
git commit -m "feat(front): add reusable DataTable and DeleteDialog components"
```

---

### Task 8: Agents CRUD Pages

**Files:**
- Modify: `front/src/pages/agents/list.tsx`
- Modify: `front/src/pages/agents/create.tsx`
- Modify: `front/src/pages/agents/edit.tsx`

- [ ] **Step 1: Write agent list page**

Replace `front/src/pages/agents/list.tsx`:

```tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useAgents, useDeleteAgent } from "@/api/agents";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react";
import { AGENT_TYPE_LABELS } from "@/lib/constants";
import type { Agent } from "@/types/api";

export default function AgentListPage() {
  const { data, isLoading } = useAgents();
  const deleteMutation = useDeleteAgent();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const columns: Column<Agent>[] = [
    { header: "Name", accessorKey: "name" },
    {
      header: "Type",
      cell: (row) => <Badge variant="secondary">{AGENT_TYPE_LABELS[row.type ?? "AGENT_TYPE_UNSPECIFIED"]}</Badge>,
    },
    { header: "Description", accessorKey: "description" },
    {
      header: "A2A",
      cell: (row) => row.enable_a2a ? <Badge>Enabled</Badge> : <Badge variant="outline">Disabled</Badge>,
    },
    {
      header: "Actions",
      cell: (row) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/agents/${row.name}/edit`)}>
              <Pencil className="mr-2 h-4 w-4" /> Edit
            </DropdownMenuItem>
            <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row.name)}>
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="Agents" createLabel="Create Agent" createTo="/agents/create" />
      <DataTable columns={columns} data={data?.agents} isLoading={isLoading} emptyMessage="No agents yet. Create your first agent to get started." />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Agent"
        description={`Are you sure you want to delete "${deleteTarget}"? This action cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => { toast.success("Agent deleted"); setDeleteTarget(null); },
              onError: (err) => toast.error(err.message),
            });
          }
        }}
      />
    </>
  );
}
```

- [ ] **Step 2: Write agent create page**

Replace `front/src/pages/agents/create.tsx`:

```tsx
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCreateAgent } from "@/api/agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import type { AgentType } from "@/types/api";

const agentSchema = z.object({
  name: z.string().min(1, "Name is required").refine((v) => v !== "user", "Name cannot be 'user'"),
  description: z.string().optional(),
  type: z.string().default("AGENT_TYPE_LLM"),
  enable_a2a: z.boolean().default(false),
  model: z.string().optional(),
  instruction: z.string().optional(),
});

type AgentFormValues = z.infer<typeof agentSchema>;

export default function AgentCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateAgent();

  const form = useForm<AgentFormValues>({
    resolver: zodResolver(agentSchema),
    defaultValues: { name: "", description: "", type: "AGENT_TYPE_LLM", enable_a2a: false, model: "", instruction: "" },
  });

  function onSubmit(values: AgentFormValues) {
    createMutation.mutate(
      {
        name: values.name,
        description: values.description,
        type: values.type as AgentType,
        enable_a2a: values.enable_a2a,
        config: {
          model: values.model,
          instruction: values.instruction,
        },
      },
      {
        onSuccess: () => { toast.success("Agent created"); navigate("/agents"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/agents">Agents</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Create Agent</h2>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl><Input placeholder="my-agent" {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="description" render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl><Input placeholder="A helpful assistant" {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="type" render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      <SelectItem value="AGENT_TYPE_LLM">LLM</SelectItem>
                      <SelectItem value="AGENT_TYPE_LOOP">Loop</SelectItem>
                      <SelectItem value="AGENT_TYPE_SEQUENTIAL">Sequential</SelectItem>
                      <SelectItem value="AGENT_TYPE_PARALLEL">Parallel</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="enable_a2a" render={({ field }) => (
                <FormItem className="flex items-center gap-3">
                  <FormLabel>Enable A2A</FormLabel>
                  <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                </FormItem>
              )} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle>Model Configuration</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="model" render={({ field }) => (
                <FormItem>
                  <FormLabel>Model</FormLabel>
                  <FormControl><Input placeholder="flash" {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="instruction" render={({ field }) => (
                <FormItem>
                  <FormLabel>Instruction</FormLabel>
                  <FormControl><Textarea placeholder="You are a helpful assistant..." rows={5} {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
            </CardContent>
          </Card>

          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/agents")}>Cancel</Button>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? "Creating..." : "Create Agent"}
            </Button>
          </div>
        </form>
      </Form>
    </>
  );
}
```

- [ ] **Step 3: Write agent edit page with dual mode (Form + JSON)**

Replace `front/src/pages/agents/edit.tsx`:

```tsx
import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import Editor from "@monaco-editor/react";
import { useAgent, useUpdateAgent } from "@/api/agents";
import { useMCPServers } from "@/api/mcp-servers";
import { useRemoteAgents } from "@/api/remote-agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { AGENT_TYPE_LABELS } from "@/lib/constants";
import { useTheme } from "next-themes";
import type { Agent, AgentType } from "@/types/api";

const agentSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  type: z.string().default("AGENT_TYPE_LLM"),
  enable_a2a: z.boolean().default(false),
  model: z.string().optional(),
  instruction: z.string().optional(),
});

type AgentFormValues = z.infer<typeof agentSchema>;

export default function AgentEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { theme } = useTheme();
  const { data, isLoading } = useAgent(name ?? "");
  const { data: mcpData } = useMCPServers();
  const { data: remoteData } = useRemoteAgents();
  const updateMutation = useUpdateAgent();
  const [jsonValue, setJsonValue] = useState("");
  const [activeTab, setActiveTab] = useState("form");

  const form = useForm<AgentFormValues>({
    resolver: zodResolver(agentSchema),
    defaultValues: { name: "", description: "", type: "AGENT_TYPE_LLM", enable_a2a: false, model: "", instruction: "" },
  });

  useEffect(() => {
    if (data?.agent) {
      const a = data.agent;
      form.reset({
        name: a.name,
        description: a.description ?? "",
        type: a.type ?? "AGENT_TYPE_LLM",
        enable_a2a: a.enable_a2a ?? false,
        model: a.config?.model ?? "",
        instruction: a.config?.instruction ?? "",
      });
      setJsonValue(JSON.stringify(a, null, 2));
    }
  }, [data, form]);

  function onFormSubmit(values: AgentFormValues) {
    const agent: Agent = {
      ...data?.agent,
      name: values.name,
      description: values.description,
      type: values.type as AgentType,
      enable_a2a: values.enable_a2a,
      config: {
        ...data?.agent?.config,
        model: values.model,
        instruction: values.instruction,
      },
    };
    submitUpdate(agent);
  }

  function onJsonSubmit() {
    try {
      const agent = JSON.parse(jsonValue) as Agent;
      submitUpdate(agent);
    } catch {
      toast.error("Invalid JSON");
    }
  }

  function submitUpdate(agent: Agent) {
    updateMutation.mutate(agent, {
      onSuccess: () => { toast.success("Agent updated"); navigate("/agents"); },
      onError: (err) => toast.error(err.message),
    });
  }

  function handleTabChange(tab: string) {
    if (tab === "json") {
      const values = form.getValues();
      const agent: Agent = {
        ...data?.agent,
        name: values.name,
        description: values.description,
        type: values.type as AgentType,
        enable_a2a: values.enable_a2a,
        config: { ...data?.agent?.config, model: values.model, instruction: values.instruction },
      };
      setJsonValue(JSON.stringify(agent, null, 2));
    } else if (tab === "form") {
      try {
        const agent = JSON.parse(jsonValue) as Agent;
        form.reset({
          name: agent.name,
          description: agent.description ?? "",
          type: agent.type ?? "AGENT_TYPE_LLM",
          enable_a2a: agent.enable_a2a ?? false,
          model: agent.config?.model ?? "",
          instruction: agent.config?.instruction ?? "",
        });
      } catch { /* keep current form values if JSON is invalid */ }
    }
    setActiveTab(tab);
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/agents">Agents</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit Agent</h2>

      <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList className="mb-4">
          <TabsTrigger value="form">Form</TabsTrigger>
          <TabsTrigger value="json">JSON</TabsTrigger>
        </TabsList>

        <TabsContent value="form">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onFormSubmit)} className="space-y-6">
              <Card>
                <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
                <CardContent className="space-y-4">
                  <FormField control={form.control} name="name" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Name</FormLabel>
                      <FormControl><Input {...field} disabled /></FormControl>
                    </FormItem>
                  )} />
                  <FormField control={form.control} name="description" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Description</FormLabel>
                      <FormControl><Input {...field} /></FormControl>
                    </FormItem>
                  )} />
                  <FormField control={form.control} name="type" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Type</FormLabel>
                      <Select onValueChange={field.onChange} value={field.value}>
                        <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                        <SelectContent>
                          <SelectItem value="AGENT_TYPE_LLM">LLM</SelectItem>
                          <SelectItem value="AGENT_TYPE_LOOP">Loop</SelectItem>
                          <SelectItem value="AGENT_TYPE_SEQUENTIAL">Sequential</SelectItem>
                          <SelectItem value="AGENT_TYPE_PARALLEL">Parallel</SelectItem>
                        </SelectContent>
                      </Select>
                    </FormItem>
                  )} />
                  <FormField control={form.control} name="enable_a2a" render={({ field }) => (
                    <FormItem className="flex items-center gap-3">
                      <FormLabel>Enable A2A</FormLabel>
                      <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                    </FormItem>
                  )} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader><CardTitle>Model Configuration</CardTitle></CardHeader>
                <CardContent className="space-y-4">
                  <FormField control={form.control} name="model" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Model</FormLabel>
                      <FormControl><Input {...field} /></FormControl>
                    </FormItem>
                  )} />
                  <FormField control={form.control} name="instruction" render={({ field }) => (
                    <FormItem>
                      <FormLabel>Instruction</FormLabel>
                      <FormControl><Textarea rows={5} {...field} /></FormControl>
                    </FormItem>
                  )} />
                </CardContent>
              </Card>

              {/* MCP Server IDs - read-only display */}
              {data?.agent?.config?.mcp_server_ids && data.agent.config.mcp_server_ids.length > 0 && (
                <Card>
                  <CardHeader><CardTitle>MCP Servers</CardTitle></CardHeader>
                  <CardContent className="flex flex-wrap gap-2">
                    {data.agent.config.mcp_server_ids.map((id) => {
                      const server = mcpData?.mcp_servers?.find((s) => s.id === id);
                      return <Badge key={id} variant="secondary">{server?.name ?? id}</Badge>;
                    })}
                  </CardContent>
                </Card>
              )}

              {/* Remote Agent IDs - read-only display */}
              {data?.agent?.config?.remote_agent_ids && data.agent.config.remote_agent_ids.length > 0 && (
                <Card>
                  <CardHeader><CardTitle>Remote Agents</CardTitle></CardHeader>
                  <CardContent className="flex flex-wrap gap-2">
                    {data.agent.config.remote_agent_ids.map((id) => {
                      const ra = remoteData?.remote_agents?.find((r) => r.id === id);
                      return <Badge key={id} variant="secondary">{ra?.name ?? id}</Badge>;
                    })}
                  </CardContent>
                </Card>
              )}

              {/* Sub-agents - read-only list */}
              {data?.agent?.sub_agents && data.agent.sub_agents.length > 0 && (
                <Card>
                  <CardHeader><CardTitle>Sub-Agents (read-only, edit in JSON mode)</CardTitle></CardHeader>
                  <CardContent className="space-y-2">
                    {data.agent.sub_agents.map((sa) => (
                      <div key={sa.name} className="flex items-center gap-2">
                        <span className="text-sm font-medium">{sa.name}</span>
                        <Badge variant="outline">{AGENT_TYPE_LABELS[sa.type ?? "AGENT_TYPE_UNSPECIFIED"]}</Badge>
                      </div>
                    ))}
                  </CardContent>
                </Card>
              )}

              <div className="flex gap-3">
                <Button type="button" variant="outline" onClick={() => navigate("/agents")}>Cancel</Button>
                <Button type="submit" disabled={updateMutation.isPending}>
                  {updateMutation.isPending ? "Saving..." : "Save"}
                </Button>
              </div>
            </form>
          </Form>
        </TabsContent>

        <TabsContent value="json">
          <Card>
            <CardContent className="pt-6">
              <Editor
                height="500px"
                language="json"
                theme={theme === "dark" ? "vs-dark" : "light"}
                value={jsonValue}
                onChange={(v) => setJsonValue(v ?? "")}
                options={{ minimap: { enabled: false }, formatOnPaste: true }}
              />
            </CardContent>
          </Card>
          <div className="mt-6 flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/agents")}>Cancel</Button>
            <Button onClick={onJsonSubmit} disabled={updateMutation.isPending}>
              {updateMutation.isPending ? "Saving..." : "Save"}
            </Button>
          </div>
        </TabsContent>
      </Tabs>
    </>
  );
}
```

- [ ] **Step 4: Verify compiles and test in browser**

```bash
cd front && npx tsc --noEmit && npm run dev
```

Navigate to `/agents` — list should render (empty if no backend). Navigate to `/agents/create` — form should render.

- [ ] **Step 5: Commit**

```bash
git add front/src/pages/agents/
git commit -m "feat(front): add Agents CRUD pages with dual Form/JSON edit mode"
```

---

### Task 9: MCP Servers CRUD Pages

**Files:**
- Modify: `front/src/pages/mcp-servers/list.tsx`
- Modify: `front/src/pages/mcp-servers/create.tsx`
- Modify: `front/src/pages/mcp-servers/edit.tsx`

- [ ] **Step 1: Write MCP server list page**

Replace `front/src/pages/mcp-servers/list.tsx`:

```tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useMCPServers, useDeleteMCPServer } from "@/api/mcp-servers";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react";
import { MCP_TRANSPORT_LABELS } from "@/lib/constants";
import type { MCPServer } from "@/types/api";

export default function MCPServerListPage() {
  const { data, isLoading } = useMCPServers();
  const deleteMutation = useDeleteMCPServer();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const columns: Column<MCPServer>[] = [
    { header: "ID", accessorKey: "id" },
    { header: "Name", accessorKey: "name" },
    {
      header: "Transport",
      cell: (row) => <Badge variant="secondary">{MCP_TRANSPORT_LABELS[row.transport ?? "MCP_SERVER_TRANSPORT_UNSPECIFIED"]}</Badge>,
    },
    {
      header: "URL / Command",
      cell: (row) => <span className="max-w-xs truncate text-sm text-muted-foreground">{row.url || row.command || "-"}</span>,
    },
    {
      header: "Actions",
      cell: (row) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/mcp-servers/${row.id}/edit`)}>
              <Pencil className="mr-2 h-4 w-4" /> Edit
            </DropdownMenuItem>
            <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row.id ?? null)}>
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="MCP Servers" createLabel="Create MCP Server" createTo="/mcp-servers/create" />
      <DataTable columns={columns} data={data?.mcp_servers} isLoading={isLoading} emptyMessage="No MCP servers yet. Create your first server to get started." />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete MCP Server"
        description={`Are you sure you want to delete this MCP server? This action cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => { toast.success("MCP server deleted"); setDeleteTarget(null); },
              onError: (err) => toast.error(err.message),
            });
          }
        }}
      />
    </>
  );
}
```

- [ ] **Step 2: Write MCP server create page**

Replace `front/src/pages/mcp-servers/create.tsx`:

```tsx
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCreateMCPServer } from "@/api/mcp-servers";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import type { MCPServerTransport } from "@/types/api";

const schema = z.object({
  id: z.string().min(1, "ID is required"),
  name: z.string().min(1, "Name is required"),
  transport: z.string().default("MCP_SERVER_TRANSPORT_STDIO"),
  command: z.string().optional(),
  url: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

export default function MCPServerCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateMCPServer();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { id: "", name: "", transport: "MCP_SERVER_TRANSPORT_STDIO", command: "", url: "" },
  });

  const transport = form.watch("transport");

  function onSubmit(values: FormValues) {
    createMutation.mutate(
      {
        id: values.id,
        name: values.name,
        transport: values.transport as MCPServerTransport,
        command: values.command,
        url: values.url,
      },
      {
        onSuccess: () => { toast.success("MCP server created"); navigate("/mcp-servers"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/mcp-servers">MCP Servers</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Create MCP Server</h2>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="id" render={({ field }) => (
                <FormItem><FormLabel>ID</FormLabel><FormControl><Input placeholder="my-mcp-server" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input placeholder="My MCP Server" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="transport" render={({ field }) => (
                <FormItem>
                  <FormLabel>Transport</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      <SelectItem value="MCP_SERVER_TRANSPORT_STDIO">Stdio</SelectItem>
                      <SelectItem value="MCP_SERVER_TRANSPORT_STREAMABLE_HTTP">HTTP</SelectItem>
                      <SelectItem value="MCP_SERVER_TRANSPORT_SSE">SSE</SelectItem>
                    </SelectContent>
                  </Select>
                </FormItem>
              )} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle>Connection</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              {transport === "MCP_SERVER_TRANSPORT_STDIO" ? (
                <FormField control={form.control} name="command" render={({ field }) => (
                  <FormItem><FormLabel>Command</FormLabel><FormControl><Input placeholder="npx @modelcontextprotocol/server" {...field} /></FormControl></FormItem>
                )} />
              ) : (
                <FormField control={form.control} name="url" render={({ field }) => (
                  <FormItem><FormLabel>URL</FormLabel><FormControl><Input placeholder="https://..." {...field} /></FormControl></FormItem>
                )} />
              )}
            </CardContent>
          </Card>

          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/mcp-servers")}>Cancel</Button>
            <Button type="submit" disabled={createMutation.isPending}>{createMutation.isPending ? "Creating..." : "Create"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
```

- [ ] **Step 3: Write MCP server edit page**

Replace `front/src/pages/mcp-servers/edit.tsx`:

```tsx
import { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useMCPServer, useUpdateMCPServer } from "@/api/mcp-servers";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import type { MCPServerTransport } from "@/types/api";

const schema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  transport: z.string(),
  command: z.string().optional(),
  url: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

export default function MCPServerEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useMCPServer(id ?? "");
  const updateMutation = useUpdateMCPServer();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { id: "", name: "", transport: "MCP_SERVER_TRANSPORT_STDIO", command: "", url: "" },
  });

  const transport = form.watch("transport");

  useEffect(() => {
    if (data?.mcp_server) {
      const s = data.mcp_server;
      form.reset({ id: s.id ?? "", name: s.name, transport: s.transport ?? "MCP_SERVER_TRANSPORT_STDIO", command: s.command ?? "", url: s.url ?? "" });
    }
  }, [data, form]);

  function onSubmit(values: FormValues) {
    updateMutation.mutate(
      { ...data?.mcp_server, id: values.id, name: values.name, transport: values.transport as MCPServerTransport, command: values.command, url: values.url },
      {
        onSuccess: () => { toast.success("MCP server updated"); navigate("/mcp-servers"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/mcp-servers">MCP Servers</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{id}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit MCP Server</h2>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="id" render={({ field }) => (
                <FormItem><FormLabel>ID</FormLabel><FormControl><Input {...field} disabled /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="transport" render={({ field }) => (
                <FormItem>
                  <FormLabel>Transport</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      <SelectItem value="MCP_SERVER_TRANSPORT_STDIO">Stdio</SelectItem>
                      <SelectItem value="MCP_SERVER_TRANSPORT_STREAMABLE_HTTP">HTTP</SelectItem>
                      <SelectItem value="MCP_SERVER_TRANSPORT_SSE">SSE</SelectItem>
                    </SelectContent>
                  </Select>
                </FormItem>
              )} />
            </CardContent>
          </Card>
          <Card>
            <CardHeader><CardTitle>Connection</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              {transport === "MCP_SERVER_TRANSPORT_STDIO" ? (
                <FormField control={form.control} name="command" render={({ field }) => (
                  <FormItem><FormLabel>Command</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                )} />
              ) : (
                <FormField control={form.control} name="url" render={({ field }) => (
                  <FormItem><FormLabel>URL</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                )} />
              )}
            </CardContent>
          </Card>
          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/mcp-servers")}>Cancel</Button>
            <Button type="submit" disabled={updateMutation.isPending}>{updateMutation.isPending ? "Saving..." : "Save"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
```

- [ ] **Step 4: Verify compiles**

```bash
cd front && npx tsc --noEmit
```

- [ ] **Step 5: Commit**

```bash
git add front/src/pages/mcp-servers/
git commit -m "feat(front): add MCP Servers CRUD pages"
```

---

### Task 10: Remote Agents CRUD Pages

**Files:**
- Modify: `front/src/pages/remote-agents/list.tsx`
- Modify: `front/src/pages/remote-agents/create.tsx`
- Modify: `front/src/pages/remote-agents/edit.tsx`

- [ ] **Step 1: Write remote agents list page**

Replace `front/src/pages/remote-agents/list.tsx`:

```tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useRemoteAgents, useDeleteRemoteAgent } from "@/api/remote-agents";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react";
import type { RemoteAgent } from "@/types/api";

export default function RemoteAgentListPage() {
  const { data, isLoading } = useRemoteAgents();
  const deleteMutation = useDeleteRemoteAgent();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const columns: Column<RemoteAgent>[] = [
    { header: "ID", accessorKey: "id" },
    { header: "Name", accessorKey: "name" },
    { header: "URL", accessorKey: "url" },
    {
      header: "Protocol",
      cell: (row) => <Badge variant="secondary">{row.protocol === "REMOTE_AGENT_PROTOCOL_A2A" ? "A2A" : "Unknown"}</Badge>,
    },
    {
      header: "Actions",
      cell: (row) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/remote-agents/${row.id}/edit`)}>
              <Pencil className="mr-2 h-4 w-4" /> Edit
            </DropdownMenuItem>
            <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row.id)}>
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="Remote Agents" createLabel="Create Remote Agent" createTo="/remote-agents/create" />
      <DataTable columns={columns} data={data?.remote_agents} isLoading={isLoading} emptyMessage="No remote agents yet." />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Remote Agent"
        description="Are you sure? This action cannot be undone."
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => { toast.success("Remote agent deleted"); setDeleteTarget(null); },
              onError: (err) => toast.error(err.message),
            });
          }
        }}
      />
    </>
  );
}
```

- [ ] **Step 2: Write remote agent create page**

Replace `front/src/pages/remote-agents/create.tsx`:

```tsx
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCreateRemoteAgent } from "@/api/remote-agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";

const schema = z.object({
  id: z.string().min(1, "ID is required"),
  name: z.string().min(1, "Name is required"),
  url: z.string().min(1, "URL is required").url("Must be a valid URL"),
});

type FormValues = z.infer<typeof schema>;

export default function RemoteAgentCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateRemoteAgent();
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { id: "", name: "", url: "" } });

  function onSubmit(values: FormValues) {
    createMutation.mutate(
      { id: values.id, name: values.name, url: values.url, protocol: "REMOTE_AGENT_PROTOCOL_A2A" },
      {
        onSuccess: () => { toast.success("Remote agent created"); navigate("/remote-agents"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/remote-agents">Remote Agents</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Create Remote Agent</h2>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Remote Agent</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="id" render={({ field }) => (
                <FormItem><FormLabel>ID</FormLabel><FormControl><Input placeholder="remote-assistant" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input placeholder="Remote Assistant" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="url" render={({ field }) => (
                <FormItem><FormLabel>URL</FormLabel><FormControl><Input placeholder="https://agent.example.com/a2a" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
            </CardContent>
          </Card>
          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/remote-agents")}>Cancel</Button>
            <Button type="submit" disabled={createMutation.isPending}>{createMutation.isPending ? "Creating..." : "Create"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
```

- [ ] **Step 3: Write remote agent edit page**

Replace `front/src/pages/remote-agents/edit.tsx`:

```tsx
import { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useRemoteAgent, useUpdateRemoteAgent } from "@/api/remote-agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";

const schema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  url: z.string().min(1).url(),
});

type FormValues = z.infer<typeof schema>;

export default function RemoteAgentEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useRemoteAgent(id ?? "");
  const updateMutation = useUpdateRemoteAgent();
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { id: "", name: "", url: "" } });

  useEffect(() => {
    if (data?.remote_agent) {
      const ra = data.remote_agent;
      form.reset({ id: ra.id, name: ra.name, url: ra.url });
    }
  }, [data, form]);

  function onSubmit(values: FormValues) {
    updateMutation.mutate(
      { id: values.id, name: values.name, url: values.url, protocol: "REMOTE_AGENT_PROTOCOL_A2A" },
      {
        onSuccess: () => { toast.success("Remote agent updated"); navigate("/remote-agents"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/remote-agents">Remote Agents</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{id}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit Remote Agent</h2>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Remote Agent</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="id" render={({ field }) => (
                <FormItem><FormLabel>ID</FormLabel><FormControl><Input {...field} disabled /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="url" render={({ field }) => (
                <FormItem><FormLabel>URL</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
              )} />
            </CardContent>
          </Card>
          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/remote-agents")}>Cancel</Button>
            <Button type="submit" disabled={updateMutation.isPending}>{updateMutation.isPending ? "Saving..." : "Save"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
```

- [ ] **Step 4: Verify compiles and commit**

```bash
cd front && npx tsc --noEmit
git add front/src/pages/remote-agents/
git commit -m "feat(front): add Remote Agents CRUD pages"
```

---

### Task 11: Sessions Pages

**Files:**
- Modify: `front/src/pages/sessions/list.tsx`
- Modify: `front/src/pages/sessions/detail.tsx`

- [ ] **Step 1: Write session list page**

Replace `front/src/pages/sessions/list.tsx`:

```tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useSessions, useDeleteSession } from "@/api/sessions";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { MoreHorizontal, Eye, Trash2 } from "lucide-react";
import type { SessionInfo } from "@/types/api";

export default function SessionListPage() {
  const [appName, setAppName] = useState("");
  const [userId, setUserId] = useState("");
  const { data, isLoading } = useSessions(appName, userId);
  const deleteMutation = useDeleteSession();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<SessionInfo | null>(null);

  const columns: Column<SessionInfo>[] = [
    { header: "Session ID", accessorKey: "session_id" },
    { header: "App Name", accessorKey: "app_name" },
    { header: "User ID", accessorKey: "user_id" },
    {
      header: "Last Update",
      cell: (row) => row.last_update_time ? new Date(row.last_update_time).toLocaleString() : "-",
    },
    {
      header: "Actions",
      cell: (row) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/sessions/detail?app=${row.app_name}&user=${row.user_id}&session=${row.session_id}`)}>
              <Eye className="mr-2 h-4 w-4" /> View
            </DropdownMenuItem>
            <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row)}>
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="Sessions" />
      <div className="mb-4 flex gap-3">
        <Input placeholder="App Name" value={appName} onChange={(e) => setAppName(e.target.value)} className="max-w-xs" />
        <Input placeholder="User ID" value={userId} onChange={(e) => setUserId(e.target.value)} className="max-w-xs" />
      </div>
      {appName && userId ? (
        <DataTable columns={columns} data={data?.sessions} isLoading={isLoading} emptyMessage="No sessions found." />
      ) : (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed p-12 text-center">
          <p className="text-muted-foreground">Enter an App Name and User ID to search sessions.</p>
        </div>
      )}
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Session"
        description={`Delete session "${deleteTarget?.session_id}"? This cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(
              { app_name: deleteTarget.app_name, user_id: deleteTarget.user_id, session_id: deleteTarget.session_id },
              {
                onSuccess: () => { toast.success("Session deleted"); setDeleteTarget(null); },
                onError: (err) => toast.error(err.message),
              },
            );
          }
        }}
      />
    </>
  );
}
```

- [ ] **Step 2: Write session detail page**

Replace `front/src/pages/sessions/detail.tsx`:

```tsx
import { useSearchParams } from "react-router-dom";
import { useSession } from "@/api/sessions";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";

export default function SessionDetailPage() {
  const [searchParams] = useSearchParams();
  const appName = searchParams.get("app") ?? "";
  const userId = searchParams.get("user") ?? "";
  const sessionId = searchParams.get("session") ?? "";
  const { data, isLoading } = useSession(appName, userId, sessionId);

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  const detail = data?.session_detail;
  if (!detail) return <p className="text-muted-foreground">Session not found.</p>;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/sessions">Sessions</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{sessionId}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Session Detail</h2>

      <Card className="mb-6">
        <CardHeader><CardTitle>Info</CardTitle></CardHeader>
        <CardContent className="grid grid-cols-2 gap-4 text-sm">
          <div><span className="text-muted-foreground">Session ID:</span> {detail.session.session_id}</div>
          <div><span className="text-muted-foreground">App:</span> {detail.session.app_name}</div>
          <div><span className="text-muted-foreground">User:</span> {detail.session.user_id}</div>
          <div><span className="text-muted-foreground">Last Update:</span> {detail.session.last_update_time ? new Date(detail.session.last_update_time).toLocaleString() : "-"}</div>
        </CardContent>
      </Card>

      <h3 className="mb-4 text-lg font-semibold">Events ({detail.events?.length ?? 0})</h3>
      <div className="space-y-3">
        {(detail.events ?? []).map((evt) => (
          <Card key={evt.event_id}>
            <CardContent className="pt-4">
              <div className="mb-2 flex items-center gap-2">
                <Badge variant={evt.author === "user" ? "default" : "secondary"}>{evt.author ?? "unknown"}</Badge>
                <span className="text-xs text-muted-foreground">{evt.timestamp ? new Date(evt.timestamp).toLocaleString() : ""}</span>
              </div>
              <pre className="max-h-40 overflow-auto rounded bg-muted p-3 text-xs">{evt.content_json ?? ""}</pre>
            </CardContent>
          </Card>
        ))}
        {(!detail.events || detail.events.length === 0) && (
          <p className="text-muted-foreground">No events in this session.</p>
        )}
      </div>
    </>
  );
}
```

- [ ] **Step 3: Verify compiles and commit**

```bash
cd front && npx tsc --noEmit
git add front/src/pages/sessions/
git commit -m "feat(front): add Sessions list and detail pages"
```

---

### Task 12: Cron Jobs CRUD + Executions Pages

**Files:**
- Modify: `front/src/pages/cron/list.tsx`
- Modify: `front/src/pages/cron/create.tsx`
- Modify: `front/src/pages/cron/edit.tsx`
- Modify: `front/src/pages/cron/executions.tsx`

- [ ] **Step 1: Write cron job list page**

Replace `front/src/pages/cron/list.tsx`:

```tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCronJobs, useDeleteCronJob, useUpdateCronJob } from "@/api/cron";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { MoreHorizontal, Pencil, Trash2, History } from "lucide-react";
import type { CronJob } from "@/types/api";

export default function CronJobListPage() {
  const { data, isLoading } = useCronJobs();
  const deleteMutation = useDeleteCronJob();
  const updateMutation = useUpdateCronJob();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  function toggleEnabled(job: CronJob) {
    updateMutation.mutate(
      { ...job, enabled: !job.enabled },
      {
        onSuccess: () => toast.success(`Job ${job.enabled ? "disabled" : "enabled"}`),
        onError: (err) => toast.error(err.message),
      },
    );
  }

  const columns: Column<CronJob>[] = [
    { header: "Name", accessorKey: "name" },
    { header: "Schedule", cell: (row) => <code className="text-xs">{row.schedule}</code> },
    { header: "Agent", accessorKey: "agent_name" },
    { header: "Timezone", cell: (row) => row.timezone || "UTC" },
    {
      header: "Enabled",
      cell: (row) => <Switch checked={row.enabled ?? false} onCheckedChange={() => toggleEnabled(row)} />,
    },
    {
      header: "Actions",
      cell: (row) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => navigate(`/cron/${row.name}/edit`)}>
              <Pencil className="mr-2 h-4 w-4" /> Edit
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => navigate(`/cron/${row.name}/executions`)}>
              <History className="mr-2 h-4 w-4" /> Executions
            </DropdownMenuItem>
            <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row.name)}>
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="Cron Jobs" createLabel="Create Cron Job" createTo="/cron/create" />
      <DataTable columns={columns} data={data?.cron_jobs} isLoading={isLoading} emptyMessage="No cron jobs yet." />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Cron Job"
        description={`Delete "${deleteTarget}"? This cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget, {
              onSuccess: () => { toast.success("Cron job deleted"); setDeleteTarget(null); },
              onError: (err) => toast.error(err.message),
            });
          }
        }}
      />
    </>
  );
}
```

- [ ] **Step 2: Write cron job create page**

Replace `front/src/pages/cron/create.tsx`:

```tsx
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCreateCronJob } from "@/api/cron";
import { useAgents } from "@/api/agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import type { CronDeliveryType } from "@/types/api";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  schedule: z.string().min(1, "Schedule is required"),
  agent_name: z.string().min(1, "Agent is required"),
  input: z.string().optional(),
  timezone: z.string().optional(),
  enabled: z.boolean().default(true),
  delivery_type: z.string().default("CRON_DELIVERY_TYPE_LOG"),
  webhook_url: z.string().optional(),
  channel_name: z.string().optional(),
  chat_id: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

export default function CronJobCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateCronJob();
  const { data: agentsData } = useAgents();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", schedule: "", agent_name: "", input: "", timezone: "UTC", enabled: true, delivery_type: "CRON_DELIVERY_TYPE_LOG" },
  });

  const deliveryType = form.watch("delivery_type");

  function onSubmit(values: FormValues) {
    createMutation.mutate(
      {
        name: values.name,
        schedule: values.schedule,
        agent_name: values.agent_name,
        input: values.input,
        timezone: values.timezone,
        enabled: values.enabled,
        delivery: {
          type: values.delivery_type as CronDeliveryType,
          webhook_url: values.webhook_url,
          channel_name: values.channel_name,
          chat_id: values.chat_id,
        },
      },
      {
        onSuccess: () => { toast.success("Cron job created"); navigate("/cron"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/cron">Cron Jobs</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Create Cron Job</h2>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input placeholder="daily-summary" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="schedule" render={({ field }) => (
                <FormItem><FormLabel>Schedule</FormLabel><FormControl><Input placeholder="0 9 * * *" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="agent_name" render={({ field }) => (
                <FormItem>
                  <FormLabel>Agent</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl><SelectTrigger><SelectValue placeholder="Select agent" /></SelectTrigger></FormControl>
                    <SelectContent>
                      {(agentsData?.agents ?? []).map((a) => (
                        <SelectItem key={a.name} value={a.name}>{a.name}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="input" render={({ field }) => (
                <FormItem><FormLabel>Input Message</FormLabel><FormControl><Textarea placeholder="Generate a daily summary" rows={3} {...field} /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="timezone" render={({ field }) => (
                <FormItem><FormLabel>Timezone</FormLabel><FormControl><Input placeholder="UTC" {...field} /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="enabled" render={({ field }) => (
                <FormItem className="flex items-center gap-3">
                  <FormLabel>Enabled</FormLabel>
                  <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                </FormItem>
              )} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle>Delivery</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="delivery_type" render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      <SelectItem value="CRON_DELIVERY_TYPE_LOG">Log</SelectItem>
                      <SelectItem value="CRON_DELIVERY_TYPE_WEBHOOK">Webhook</SelectItem>
                      <SelectItem value="CRON_DELIVERY_TYPE_CHANNEL">Channel</SelectItem>
                    </SelectContent>
                  </Select>
                </FormItem>
              )} />
              {deliveryType === "CRON_DELIVERY_TYPE_WEBHOOK" && (
                <FormField control={form.control} name="webhook_url" render={({ field }) => (
                  <FormItem><FormLabel>Webhook URL</FormLabel><FormControl><Input placeholder="https://..." {...field} /></FormControl></FormItem>
                )} />
              )}
              {deliveryType === "CRON_DELIVERY_TYPE_CHANNEL" && (
                <>
                  <FormField control={form.control} name="channel_name" render={({ field }) => (
                    <FormItem><FormLabel>Channel Name</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                  )} />
                  <FormField control={form.control} name="chat_id" render={({ field }) => (
                    <FormItem><FormLabel>Chat ID</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                  )} />
                </>
              )}
            </CardContent>
          </Card>

          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/cron")}>Cancel</Button>
            <Button type="submit" disabled={createMutation.isPending}>{createMutation.isPending ? "Creating..." : "Create"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
```

- [ ] **Step 3: Write cron job edit page**

Replace `front/src/pages/cron/edit.tsx`:

```tsx
import { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCronJob, useUpdateCronJob } from "@/api/cron";
import { useAgents } from "@/api/agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import type { CronDeliveryType } from "@/types/api";

const schema = z.object({
  name: z.string().min(1),
  schedule: z.string().min(1),
  agent_name: z.string().min(1),
  input: z.string().optional(),
  timezone: z.string().optional(),
  enabled: z.boolean().default(true),
  delivery_type: z.string().default("CRON_DELIVERY_TYPE_LOG"),
  webhook_url: z.string().optional(),
  channel_name: z.string().optional(),
  chat_id: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

export default function CronJobEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useCronJob(name ?? "");
  const { data: agentsData } = useAgents();
  const updateMutation = useUpdateCronJob();
  const form = useForm<FormValues>({ resolver: zodResolver(schema) });
  const deliveryType = form.watch("delivery_type");

  useEffect(() => {
    if (data?.cron_job) {
      const j = data.cron_job;
      form.reset({
        name: j.name,
        schedule: j.schedule,
        agent_name: j.agent_name,
        input: j.input ?? "",
        timezone: j.timezone ?? "UTC",
        enabled: j.enabled ?? true,
        delivery_type: j.delivery?.type ?? "CRON_DELIVERY_TYPE_LOG",
        webhook_url: j.delivery?.webhook_url ?? "",
        channel_name: j.delivery?.channel_name ?? "",
        chat_id: j.delivery?.chat_id ?? "",
      });
    }
  }, [data, form]);

  function onSubmit(values: FormValues) {
    updateMutation.mutate(
      {
        name: values.name,
        schedule: values.schedule,
        agent_name: values.agent_name,
        input: values.input,
        timezone: values.timezone,
        enabled: values.enabled,
        delivery: { type: values.delivery_type as CronDeliveryType, webhook_url: values.webhook_url, channel_name: values.channel_name, chat_id: values.chat_id },
      },
      {
        onSuccess: () => { toast.success("Cron job updated"); navigate("/cron"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/cron">Cron Jobs</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit Cron Job</h2>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Basic Info</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input {...field} disabled /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="schedule" render={({ field }) => (
                <FormItem><FormLabel>Schedule</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="agent_name" render={({ field }) => (
                <FormItem>
                  <FormLabel>Agent</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      {(agentsData?.agents ?? []).map((a) => (
                        <SelectItem key={a.name} value={a.name}>{a.name}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormItem>
              )} />
              <FormField control={form.control} name="input" render={({ field }) => (
                <FormItem><FormLabel>Input Message</FormLabel><FormControl><Textarea rows={3} {...field} /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="timezone" render={({ field }) => (
                <FormItem><FormLabel>Timezone</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
              )} />
              <FormField control={form.control} name="enabled" render={({ field }) => (
                <FormItem className="flex items-center gap-3">
                  <FormLabel>Enabled</FormLabel>
                  <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                </FormItem>
              )} />
            </CardContent>
          </Card>
          <Card>
            <CardHeader><CardTitle>Delivery</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="delivery_type" render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                    <SelectContent>
                      <SelectItem value="CRON_DELIVERY_TYPE_LOG">Log</SelectItem>
                      <SelectItem value="CRON_DELIVERY_TYPE_WEBHOOK">Webhook</SelectItem>
                      <SelectItem value="CRON_DELIVERY_TYPE_CHANNEL">Channel</SelectItem>
                    </SelectContent>
                  </Select>
                </FormItem>
              )} />
              {deliveryType === "CRON_DELIVERY_TYPE_WEBHOOK" && (
                <FormField control={form.control} name="webhook_url" render={({ field }) => (
                  <FormItem><FormLabel>Webhook URL</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                )} />
              )}
              {deliveryType === "CRON_DELIVERY_TYPE_CHANNEL" && (
                <>
                  <FormField control={form.control} name="channel_name" render={({ field }) => (
                    <FormItem><FormLabel>Channel Name</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                  )} />
                  <FormField control={form.control} name="chat_id" render={({ field }) => (
                    <FormItem><FormLabel>Chat ID</FormLabel><FormControl><Input {...field} /></FormControl></FormItem>
                  )} />
                </>
              )}
            </CardContent>
          </Card>
          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/cron")}>Cancel</Button>
            <Button type="submit" disabled={updateMutation.isPending}>{updateMutation.isPending ? "Saving..." : "Save"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
```

- [ ] **Step 4: Write cron executions page**

Replace `front/src/pages/cron/executions.tsx`:

```tsx
import { useParams } from "react-router-dom";
import { useCronExecutions } from "@/api/cron";
import { DataTable, type Column } from "@/components/data-table";
import { Badge } from "@/components/ui/badge";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import type { CronExecution } from "@/types/api";

function formatDuration(start?: string, end?: string): string {
  if (!start || !end) return "-";
  const ms = new Date(end).getTime() - new Date(start).getTime();
  return `${(ms / 1000).toFixed(1)}s`;
}

export default function CronExecutionsPage() {
  const { name } = useParams<{ name: string }>();
  const { data, isLoading } = useCronExecutions(name);

  const columns: Column<CronExecution>[] = [
    { header: "ID", cell: (row) => <span className="max-w-[100px] truncate text-xs">{row.id}</span> },
    { header: "Agent", accessorKey: "agent_name" },
    {
      header: "Status",
      cell: (row) =>
        row.status === "CRON_EXECUTION_STATUS_SUCCESS"
          ? <Badge className="bg-green-500/10 text-green-500">Success</Badge>
          : <Badge variant="destructive">Error</Badge>,
    },
    { header: "Duration", cell: (row) => formatDuration(row.started_at, row.finished_at) },
    { header: "Started", cell: (row) => row.started_at ? new Date(row.started_at).toLocaleString() : "-" },
    {
      header: "Output",
      cell: (row) => <span className="block max-w-xs truncate text-xs text-muted-foreground">{row.output ?? "-"}</span>,
    },
  ];

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/cron">Cron Jobs</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Executions</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Executions: {name}</h2>
      <DataTable columns={columns} data={data?.executions} isLoading={isLoading} emptyMessage="No executions yet." />
    </>
  );
}
```

- [ ] **Step 5: Verify compiles and commit**

```bash
cd front && npx tsc --noEmit
git add front/src/pages/cron/
git commit -m "feat(front): add Cron Jobs CRUD and executions pages"
```

---

### Task 13: Dashboard Page (Monitoring)

**Files:**
- Modify: `front/src/pages/dashboard.tsx`

- [ ] **Step 1: Write dashboard page with stats, charts, and recent executions**

Replace `front/src/pages/dashboard.tsx`:

```tsx
import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useCronJobs, useDashboardExecutions } from "@/api/cron";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts";
import type { CronExecution } from "@/types/api";

function formatDuration(start?: string, end?: string): string {
  if (!start || !end) return "-";
  const ms = new Date(end).getTime() - new Date(start).getTime();
  return `${(ms / 1000).toFixed(1)}s`;
}

function computeStats(executions: CronExecution[], activeJobs: number) {
  const total = executions.length;
  const success = executions.filter((e) => e.status === "CRON_EXECUTION_STATUS_SUCCESS").length;
  const errors = total - success;
  const rate = total > 0 ? ((success / total) * 100).toFixed(1) : "0";
  const durations = executions
    .map((e) => {
      if (!e.started_at || !e.finished_at) return 0;
      return (new Date(e.finished_at).getTime() - new Date(e.started_at).getTime()) / 1000;
    })
    .filter((d) => d > 0);
  const avgDuration = durations.length > 0 ? (durations.reduce((a, b) => a + b, 0) / durations.length).toFixed(1) : "0";

  return { total, success, errors, rate, activeJobs, avgDuration };
}

function buildTimelineData(executions: CronExecution[]) {
  const now = new Date();
  const buckets: Record<string, { hour: string; success: number; error: number }> = {};

  for (let i = 23; i >= 0; i--) {
    const d = new Date(now.getTime() - i * 3600_000);
    const key = `${d.getHours().toString().padStart(2, "0")}:00`;
    buckets[key] = { hour: key, success: 0, error: 0 };
  }

  for (const e of executions) {
    if (!e.started_at) continue;
    const d = new Date(e.started_at);
    const key = `${d.getHours().toString().padStart(2, "0")}:00`;
    if (buckets[key]) {
      if (e.status === "CRON_EXECUTION_STATUS_SUCCESS") buckets[key].success++;
      else buckets[key].error++;
    }
  }

  return Object.values(buckets);
}

const COLORS = { success: "#4ade80", error: "#ef4444" };

export default function DashboardPage() {
  const { data: jobsData, isLoading: jobsLoading } = useCronJobs();
  const { data: execData, isLoading: execLoading } = useDashboardExecutions();
  const navigate = useNavigate();

  const executions = execData?.executions ?? [];
  const activeJobs = (jobsData?.cron_jobs ?? []).filter((j) => j.enabled).length;

  const stats = useMemo(() => computeStats(executions, activeJobs), [executions, activeJobs]);
  const timelineData = useMemo(() => buildTimelineData(executions), [executions]);
  const pieData = useMemo(
    () => [
      { name: "Success", value: stats.success },
      { name: "Error", value: stats.errors },
    ],
    [stats],
  );

  const isLoading = jobsLoading || execLoading;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h2 className="text-2xl font-bold">Dashboard</h2>
        <div className="grid grid-cols-4 gap-4">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-28" />)}</div>
        <Skeleton className="h-72" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Dashboard</h2>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground">Total Executions</CardTitle></CardHeader>
          <CardContent><div className="text-3xl font-bold">{stats.total}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground">Success Rate</CardTitle></CardHeader>
          <CardContent><div className="text-3xl font-bold text-green-500">{stats.rate}%</div><p className="text-xs text-muted-foreground">{stats.success} passed / {stats.errors} failed</p></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground">Active Cron Jobs</CardTitle></CardHeader>
          <CardContent><div className="text-3xl font-bold">{stats.activeJobs}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground">Avg Duration</CardTitle></CardHeader>
          <CardContent><div className="text-3xl font-bold">{stats.avgDuration}s</div></CardContent>
        </Card>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-3 gap-4">
        {/* Timeline */}
        <Card className="col-span-2">
          <CardHeader><CardTitle>Execution Timeline</CardTitle></CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={timelineData}>
                <XAxis dataKey="hour" tick={{ fontSize: 11 }} />
                <YAxis tick={{ fontSize: 11 }} />
                <Tooltip />
                <Legend />
                <Bar dataKey="success" stackId="a" fill={COLORS.success} name="Success" />
                <Bar dataKey="error" stackId="a" fill={COLORS.error} name="Error" />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* Donut */}
        <Card>
          <CardHeader><CardTitle>Status Breakdown</CardTitle></CardHeader>
          <CardContent className="flex items-center justify-center">
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie data={pieData} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, value }) => `${name}: ${value}`}>
                  <Cell fill={COLORS.success} />
                  <Cell fill={COLORS.error} />
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* Recent Executions */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Recent Executions</CardTitle>
          <button className="text-sm text-primary hover:underline" onClick={() => navigate("/cron")}>View all &rarr;</button>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {executions.slice(0, 10).map((e) => (
              <div key={e.id} className="flex items-center gap-4 rounded-md border px-4 py-2 text-sm">
                <span className="w-36 font-medium truncate">{e.job_name}</span>
                <span className="w-24 text-muted-foreground">{e.agent_name}</span>
                {e.status === "CRON_EXECUTION_STATUS_SUCCESS"
                  ? <Badge className="bg-green-500/10 text-green-500">Success</Badge>
                  : <Badge variant="destructive">Error</Badge>}
                <span className="w-16 text-muted-foreground">{formatDuration(e.started_at, e.finished_at)}</span>
                <span className="w-40 text-xs text-muted-foreground">{e.started_at ? new Date(e.started_at).toLocaleString() : "-"}</span>
                <span className="flex-1 truncate text-xs text-muted-foreground">{e.output ?? ""}</span>
              </div>
            ))}
            {executions.length === 0 && <p className="text-center text-muted-foreground py-8">No executions yet.</p>}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
```

- [ ] **Step 2: Verify compiles**

```bash
cd front && npx tsc --noEmit
```

- [ ] **Step 3: Run dev server and verify dashboard renders**

```bash
cd front && npm run dev
```

Open http://localhost:5173 — Dashboard should show stats cards, charts (empty), and recent executions. Navigate through all sidebar routes to confirm no crashes.

- [ ] **Step 4: Commit**

```bash
git add front/src/pages/dashboard.tsx
git commit -m "feat(front): add Dashboard page with stats, timeline chart, and recent executions"
```

---

### Task 14: Final Integration Verification

- [ ] **Step 1: Full TypeScript check**

```bash
cd front && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 2: Production build**

```bash
cd front && npm run build
```

Expected: successful build to `front/dist/`.

- [ ] **Step 3: Manual smoke test**

Start dev server (`npm run dev`). Walk through:
1. Login page renders at `/login`
2. Sidebar navigation works for all 6 routes
3. Theme toggle switches between dark and light
4. All list pages render (empty state)
5. All create forms render and validate
6. Dashboard page renders with charts
7. Logout clears token and redirects to login

- [ ] **Step 4: Final commit**

```bash
git add -A front/
git commit -m "feat(front): complete agent management dashboard v1"
```
