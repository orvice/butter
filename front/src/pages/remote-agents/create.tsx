import { useNavigate, useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import { useCreateRemoteAgent } from "@/api/remote-agents";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import RemoteAgentForm from "./form";
import type { RemoteAgent } from "@/types/api";

export default function RemoteAgentCreatePage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const initialDaemonRuntimeId = searchParams.get("daemon_runtime_id") ?? "";
  const initialAcpRuntime = searchParams.get("acp_runtime") === "codex" ? "codex" : "opencode";
  const createMutation = useCreateRemoteAgent();

  function onSubmit(agent: RemoteAgent) {
    createMutation.mutate(agent, {
      onSuccess: () => { toast.success("Remote agent created"); navigate("/remote-agents"); },
      onError: (err) => toast.error(err.message),
    });
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
      <RemoteAgentForm
        mode="create"
        submitLabel="Create"
        loading={createMutation.isPending}
        initialDaemonRuntimeId={initialDaemonRuntimeId}
        initialAcpRuntime={initialAcpRuntime}
        onCancel={() => navigate("/remote-agents")}
        onSubmit={onSubmit}
      />
    </>
  );
}
