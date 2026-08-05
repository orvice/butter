import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useRemoteAgent, useUpdateRemoteAgent } from "@/api/remote-agents";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
import { Skeleton } from "@/components/ui/skeleton";
import RemoteAgentForm from "./form";
import type { RemoteAgent } from "@/types/api";

export default function RemoteAgentEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useRemoteAgent(id ?? "");
  const updateMutation = useUpdateRemoteAgent();

  function onSubmit(agent: RemoteAgent) {
    updateMutation.mutate(agent, {
      onSuccess: () => { toast.success("Remote agent updated"); navigate("/remote-agents"); },
      onError: (err) => toast.error(err.message),
    });
  }

  if (isLoading) {
    return (
      <div className="p-6">
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  return (
    <Page>
      <PageHeader className="max-w-3xl" title="Edit Remote Agent" subtitle={id} />
      <PageScroll className="max-w-3xl">
        <RemoteAgentForm
          mode="edit"
          submitLabel="Save"
          loading={updateMutation.isPending}
          initialValue={data?.remote_agent}
          onCancel={() => navigate("/remote-agents")}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  );
}
