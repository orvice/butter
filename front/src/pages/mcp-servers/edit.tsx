import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useMCPServer, useUpdateMCPServer } from "@/api/mcp-servers";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import MCPServerForm from "./form";
import type { MCPServer } from "@/types/api";

export default function MCPServerEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useMCPServer(id ?? "");
  const updateMutation = useUpdateMCPServer();

  function onSubmit(server: MCPServer) {
    updateMutation.mutate(
      { ...data?.mcp_server, ...server },
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
      <div className="mb-6 space-y-1">
        <h2 className="text-2xl font-bold">Edit MCP Server</h2>
        <p className="text-sm text-muted-foreground">Review endpoint and authentication changes before saving because agents may use this server immediately.</p>
      </div>

      <MCPServerForm
        mode="edit"
        submitLabel="Save"
        loading={updateMutation.isPending}
        initialValue={data?.mcp_server}
        onCancel={() => navigate("/mcp-servers")}
        onSubmit={onSubmit}
      />
    </>
  );
}
