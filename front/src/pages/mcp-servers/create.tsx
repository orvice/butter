import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCreateMCPServer } from "@/api/mcp-servers";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
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
    <Page>
      <PageHeader
        className="max-w-3xl"
        title="Create MCP Server"
        subtitle="Connect an HTTP or SSE MCP endpoint, then choose the authentication method it requires."
      />
      <PageScroll className="max-w-3xl">
        <MCPServerForm
          mode="create"
          submitLabel="Create"
          loading={createMutation.isPending}
          onCancel={() => navigate("/mcp-servers")}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  );
}
