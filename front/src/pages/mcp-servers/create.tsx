import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCreateMCPServer } from "@/api/mcp-servers";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import MCPServerForm from "./form";
import type { MCPServer } from "@/types/api";

export default function MCPServerCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateMCPServer();

  function onSubmit(server: MCPServer) {
    createMutation.mutate(server, {
      onSuccess: () => { toast.success("MCP server created"); navigate("/mcp-servers"); },
      onError: (err) => toast.error(err.message),
    });
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
      <div className="mb-6 space-y-1">
        <h2 className="text-2xl font-bold">Create MCP Server</h2>
        <p className="text-sm text-muted-foreground">Connect an HTTP or SSE MCP endpoint, then choose the authentication method it requires.</p>
      </div>

      <MCPServerForm
        mode="create"
        submitLabel="Create"
        loading={createMutation.isPending}
        onCancel={() => navigate("/mcp-servers")}
        onSubmit={onSubmit}
      />
    </>
  );
}
