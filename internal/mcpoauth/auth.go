package mcpoauth

import agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"

func AuthType(srv *agentsv1.MCPServer) agentsv1.MCPServerAuthType {
	authType := srv.GetAuth().GetType()
	if authType != agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_UNSPECIFIED {
		return authType
	}
	if len(srv.GetHeaders()) > 0 {
		return agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_STATIC_HEADERS
	}
	return agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_NONE
}
