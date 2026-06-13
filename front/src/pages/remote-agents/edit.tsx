import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useRemoteAgent, useUpdateRemoteAgent } from "@/api/remote-agents";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
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

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/remote-agents">Remote Agents</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{id}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit Remote Agent</h2>
      <RemoteAgentForm
        mode="edit"
        submitLabel="Save"
        loading={updateMutation.isPending}
        initialValue={data?.remote_agent}
        onCancel={() => navigate("/remote-agents")}
        onSubmit={onSubmit}
      />
    </>
  );
}
