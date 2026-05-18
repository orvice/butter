import { BrowserRouter, Routes, Route } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "next-themes";
import { Toaster } from "@/components/ui/sonner";
import { AuthProvider } from "@/hooks/use-auth";
import { WorkspaceProvider } from "@/hooks/use-workspace";
import DashboardLayout from "@/layouts/dashboard-layout";
import LoginPage from "@/pages/login";
import DashboardPage from "@/pages/dashboard";
import AgentListPage from "@/pages/agents/list";
import AgentCreatePage from "@/pages/agents/create";
import AgentEditPage from "@/pages/agents/edit";
import MCPServerListPage from "@/pages/mcp-servers/list";
import MCPServerCreatePage from "@/pages/mcp-servers/create";
import MCPServerEditPage from "@/pages/mcp-servers/edit";
import ModelProviderListPage from "@/pages/model-providers/list";
import ModelProviderCreatePage from "@/pages/model-providers/create";
import ModelProviderEditPage from "@/pages/model-providers/edit";
import RemoteAgentListPage from "@/pages/remote-agents/list";
import RemoteAgentCreatePage from "@/pages/remote-agents/create";
import RemoteAgentEditPage from "@/pages/remote-agents/edit";
import SessionListPage from "@/pages/sessions/list";
import SessionDetailPage from "@/pages/sessions/detail";
import ChatPage from "@/pages/chat";
import CronJobListPage from "@/pages/cron/list";
import CronJobCreatePage from "@/pages/cron/create";
import CronJobEditPage from "@/pages/cron/edit";
import CronExecutionsPage from "@/pages/cron/executions";
import DaemonListPage from "@/pages/daemons/list";
import ChannelListPage from "@/pages/channels/list";
import ChannelCreatePage from "@/pages/channels/create";
import ChannelEditPage from "@/pages/channels/edit";
import APITokenListPage from "@/pages/api-tokens/list";
import UserListPage from "@/pages/users/list";
import ProfilePage from "@/pages/profile";
import WorkspacePage from "@/pages/workspaces";

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
            <WorkspaceProvider>
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
                  <Route path="model-providers" element={<ModelProviderListPage />} />
                  <Route path="model-providers/create" element={<ModelProviderCreatePage />} />
                  <Route path="model-providers/:name/edit" element={<ModelProviderEditPage />} />
                  <Route path="remote-agents" element={<RemoteAgentListPage />} />
                  <Route path="remote-agents/create" element={<RemoteAgentCreatePage />} />
                  <Route path="remote-agents/:id/edit" element={<RemoteAgentEditPage />} />
                  <Route path="chat" element={<ChatPage />} />
                  <Route path="sessions" element={<SessionListPage />} />
                  <Route path="sessions/detail" element={<SessionDetailPage />} />
                  <Route path="cron" element={<CronJobListPage />} />
                  <Route path="cron/create" element={<CronJobCreatePage />} />
                  <Route path="cron/:name/edit" element={<CronJobEditPage />} />
                  <Route path="cron/:name/executions" element={<CronExecutionsPage />} />
                  <Route path="daemons" element={<DaemonListPage />} />
                  <Route path="channels" element={<ChannelListPage />} />
                  <Route path="channels/create" element={<ChannelCreatePage />} />
                  <Route path="channels/:name/edit" element={<ChannelEditPage />} />
                  <Route path="api-tokens" element={<APITokenListPage />} />
                  <Route path="profile" element={<ProfilePage />} />
                  <Route path="users" element={<UserListPage />} />
                  <Route path="workspaces" element={<WorkspacePage />} />
                </Route>
              </Routes>
              <Toaster />
            </WorkspaceProvider>
          </AuthProvider>
        </QueryClientProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
}
