import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useMCPServer, useUpdateMCPServer } from "@/api/mcp-servers";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
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

  return (
    <Page>
      <PageHeader
        title="Edit MCP Server"
        subtitle="Review endpoint and authentication changes before saving because agents may use this server immediately."
      />
      <PageScroll className="max-w-3xl">
        {isLoading ? (
          <Skeleton className="h-96 w-full" />
        ) : (
          <MCPServerForm
            mode="edit"
            submitLabel="Save"
            loading={updateMutation.isPending}
            initialValue={data?.mcp_server}
            onCancel={() => navigate("/mcp-servers")}
            onSubmit={onSubmit}
          />
        )}
      </PageScroll>
    </Page>
  );
}
