import { createGrpcWebTransport } from "@connectrpc/connect-web";
import {
  Code,
  ConnectError,
  createClient,
  type Client,
  type Interceptor,
  type Transport,
} from "@connectrpc/connect";
import type { DescService } from "@bufbuild/protobuf";

import { ApiError, BASE_URL } from "./client";
import { TOKEN_KEY, WORKSPACE_KEY } from "@/lib/constants";

// authInterceptor reads the bearer token and active workspace from
// localStorage on every call so a token rotation in another tab (or a
// workspace switch) takes effect without a full reload. The same headers
// are CORS-allowed by NewGRPCWebHandler on the server.
const authInterceptor: Interceptor = (next) => async (req) => {
  const token = localStorage.getItem(TOKEN_KEY);
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  if (token) req.header.set("Authorization", `Bearer ${token}`);
  if (workspaceId) req.header.set("X-Workspace-ID", workspaceId);

  try {
    return await next(req);
  } catch (err) {
    if (err instanceof ConnectError && err.code === Code.Unauthenticated) {
      localStorage.removeItem(TOKEN_KEY);
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
      throw new ApiError("unauthenticated", "Invalid or expired token");
    }
    if (err instanceof ConnectError) {
      throw new ApiError(connectCodeToString(err.code), err.rawMessage || err.message);
    }
    throw err;
  }
};

function connectCodeToString(code: Code): string {
  return Code[code] ?? "unknown";
}

// transport is configured once and shared across all generated clients.
// The Go server is a grpc-go server wrapped for grpc-web, so use binary
// protobuf framing instead of JSON grpc-web, which would require a server
// JSON codec.
export const transport: Transport = createGrpcWebTransport({
  baseUrl: BASE_URL || "/",
  useBinaryFormat: true,
  interceptors: [authInterceptor],
});

// makeClient is a thin wrapper over createClient that pins the shared
// transport. Use it per-service: `const client = makeClient(AgentService);`
export function makeClient<T extends DescService>(service: T): Client<T> {
  return createClient(service, transport);
}
