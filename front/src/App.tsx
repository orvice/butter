import { lazy, Suspense } from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "next-themes";
import { Toaster } from "@/components/ui/sonner";
import { AuthProvider } from "@/hooks/use-auth";
import { ColorThemeProvider } from "@/hooks/use-color-theme";
import { LayoutDensityProvider } from "@/hooks/use-layout-density";
import { WorkspaceProvider } from "@/hooks/use-workspace";
import DashboardLayout from "@/layouts/dashboard-layout";

const LoginPage = lazy(() => import("@/pages/login"));
const OAuthCallbackPage = lazy(() => import("@/pages/auth-oauth-callback"));
const DashboardPage = lazy(() => import("@/pages/dashboard"));
const IntegrationsPage = lazy(() => import("@/pages/integrations"));
const OperationsPage = lazy(() => import("@/pages/operations"));
const AgentListPage = lazy(() => import("@/pages/agents/list"));
const AgentCreatePage = lazy(() => import("@/pages/agents/create"));
const AgentEditPage = lazy(() => import("@/pages/agents/edit"));
const AgentFilesPage = lazy(() => import("@/pages/agent-files/list"));
const MCPServerListPage = lazy(() => import("@/pages/mcp-servers/list"));
const MCPServerCreatePage = lazy(() => import("@/pages/mcp-servers/create"));
const MCPServerEditPage = lazy(() => import("@/pages/mcp-servers/edit"));
const ModelProviderListPage = lazy(() => import("@/pages/model-providers/list"));
const ModelProviderCreatePage = lazy(() => import("@/pages/model-providers/create"));
const ModelProviderEditPage = lazy(() => import("@/pages/model-providers/edit"));
const NotifyGroupListPage = lazy(() => import("@/pages/notify-groups/list"));
const NotifyGroupCreatePage = lazy(() => import("@/pages/notify-groups/create"));
const NotifyGroupEditPage = lazy(() => import("@/pages/notify-groups/edit"));
const RemoteAgentListPage = lazy(() => import("@/pages/remote-agents/list"));
const RemoteAgentCreatePage = lazy(() => import("@/pages/remote-agents/create"));
const RemoteAgentEditPage = lazy(() => import("@/pages/remote-agents/edit"));
const SessionListPage = lazy(() => import("@/pages/sessions/list"));
const SessionDetailPage = lazy(() => import("@/pages/sessions/detail"));
const ChatPage = lazy(() => import("@/pages/chat"));
const ForumListPage = lazy(() => import("@/pages/forum/list"));
const ForumThreadPage = lazy(() => import("@/pages/forum/thread"));
const CronJobListPage = lazy(() => import("@/pages/cron/list"));
const CronJobCreatePage = lazy(() => import("@/pages/cron/create"));
const CronJobEditPage = lazy(() => import("@/pages/cron/edit"));
const CronExecutionsPage = lazy(() => import("@/pages/cron/executions"));
const DaemonListPage = lazy(() => import("@/pages/daemons/list"));
const ChannelListPage = lazy(() => import("@/pages/channels/list"));
const ChannelCreatePage = lazy(() => import("@/pages/channels/create"));
const ChannelEditPage = lazy(() => import("@/pages/channels/edit"));
const APITokenListPage = lazy(() => import("@/pages/api-tokens/list"));
const UserListPage = lazy(() => import("@/pages/users/list"));
const AdminGlobalMCPServersPage = lazy(() => import("@/pages/admin/global-mcp-servers"));
const ProfilePage = lazy(() => import("@/pages/profile"));
const WorkspacePage = lazy(() => import("@/pages/workspaces"));

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
        <ColorThemeProvider>
          <LayoutDensityProvider>
            <QueryClientProvider client={queryClient}>
              <AuthProvider>
                <WorkspaceProvider>
                  <Suspense fallback={null}>
                    <Routes>
                      <Route path="/login" element={<LoginPage />} />
                      <Route path="/auth/oauth/callback/:provider" element={<OAuthCallbackPage />} />
                      <Route element={<DashboardLayout />}>
                        <Route index element={<DashboardPage />} />
                        <Route path="agents" element={<AgentListPage />} />
                        <Route path="agents/create" element={<AgentCreatePage />} />
                        <Route path="agents/:name/edit" element={<AgentEditPage />} />
                        <Route path="agent-files" element={<AgentFilesPage />} />
                        <Route path="mcp-servers" element={<MCPServerListPage />} />
                        <Route path="mcp-servers/create" element={<MCPServerCreatePage />} />
                        <Route path="mcp-servers/:id/edit" element={<MCPServerEditPage />} />
                        <Route path="model-providers" element={<ModelProviderListPage />} />
                        <Route path="model-providers/create" element={<ModelProviderCreatePage />} />
                        <Route path="model-providers/:name/edit" element={<ModelProviderEditPage />} />
                        <Route path="notify-groups" element={<NotifyGroupListPage />} />
                        <Route path="notify-groups/create" element={<NotifyGroupCreatePage />} />
                        <Route path="notify-groups/:name/edit" element={<NotifyGroupEditPage />} />
                        <Route path="remote-agents" element={<RemoteAgentListPage />} />
                        <Route path="remote-agents/create" element={<RemoteAgentCreatePage />} />
                        <Route path="remote-agents/:id/edit" element={<RemoteAgentEditPage />} />
                        <Route path="integrations" element={<IntegrationsPage />} />
                        <Route path="chat" element={<ChatPage />} />
                        <Route path="forum" element={<ForumListPage />} />
                        <Route path="forum/:id" element={<ForumThreadPage />} />
                        <Route path="sessions" element={<SessionListPage />} />
                        <Route path="sessions/detail" element={<SessionDetailPage />} />
                        <Route path="operations" element={<OperationsPage />} />
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
                        <Route path="admin/users" element={<UserListPage />} />
                        <Route path="admin/global-mcp-servers" element={<AdminGlobalMCPServersPage />} />
                        <Route path="workspaces" element={<WorkspacePage />} />
                      </Route>
                    </Routes>
                  </Suspense>
                  <Toaster />
                </WorkspaceProvider>
              </AuthProvider>
            </QueryClientProvider>
          </LayoutDensityProvider>
        </ColorThemeProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
}
