import { useNavigate, useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import { useCreateRemoteAgent } from "@/api/remote-agents";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
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
    <Page>
      <PageHeader
        className="max-w-3xl"
        title="Create Remote Agent"
        subtitle="Register an external orchestrator or autonomous daemon instance."
      />
      <PageScroll className="max-w-3xl">
        <RemoteAgentForm
          mode="create"
          submitLabel="Create"
          loading={createMutation.isPending}
          initialDaemonRuntimeId={initialDaemonRuntimeId}
          initialAcpRuntime={initialAcpRuntime}
          onCancel={() => navigate("/remote-agents")}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  );
}
