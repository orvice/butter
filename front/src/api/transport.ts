import { Code, ConnectError, createClient, type Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import type { DescService } from "@bufbuild/protobuf";

import { TOKEN_KEY, WORKSPACE_KEY } from "@/lib/constants";

export const BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

// authInterceptor injects the Authorization and X-Workspace-ID headers on
// every Connect RPC. It also rewrites 401 / Unauthenticated errors into a
// hard redirect to /login so existing React Query callers don't have to
// special-case auth failures.
const authInterceptor: Interceptor = (next) => async (req) => {
  const token = localStorage.getItem(TOKEN_KEY);
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  if (token) req.header.set("Authorization", `Bearer ${token}`);
  if (workspaceId) req.header.set("X-Workspace-ID", workspaceId);

  try {
    return await next(req);
  } catch (err) {
    if (err instanceof ConnectError && err.code === Code.Unauthenticated) {
      // Don't trash the session when a deliberate "Me" check probes auth on
      // the login page — login.tsx and validateToken() are responsible for
      // their own redirects.
      if (!isProbeRequest(req.url)) {
        localStorage.removeItem(TOKEN_KEY);
        if (typeof window !== "undefined" && window.location.pathname !== "/login") {
          window.location.href = "/login";
        }
      }
    }
    throw err;
  }
};

const probeUrls = new Set<string>([
  "/agents.v1.AuthService/Me",
  "/agents.v1.AuthService/Login",
  "/agents.v1.AuthService/ListOAuthProviders",
  "/agents.v1.AuthService/BeginOAuthFlow",
  "/agents.v1.AuthService/CompleteOAuthFlow",
]);

function isProbeRequest(url: string): boolean {
  for (const path of probeUrls) {
    if (url.endsWith(path)) return true;
  }
  return false;
}

export const transport = createConnectTransport({
  baseUrl: `${BASE_URL}/api`,
  interceptors: [authInterceptor],
});

// makeClient is a thin wrapper to keep import sites short.
export function makeClient<T extends DescService>(service: T) {
  return createClient(service, transport);
}

// Re-export so callers don't have to import @connectrpc/connect themselves.
export { ConnectError, Code };
