import { useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import { useChannel, useUpdateChannel } from "@/api/channels";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import ChannelForm from "./form";
import type { AgentChannel } from "@/types/api";

export default function ChannelEditPage() {
  const { name = "" } = useParams();
  const navigate = useNavigate();
  const { data, isLoading } = useChannel(name);
  const updateMutation = useUpdateChannel();

  function onSubmit(channel: AgentChannel) {
    updateMutation.mutate(channel, {
      onSuccess: () => {
        toast.success("Channel updated");
        navigate("/channels");
      },
      onError: (err) => toast.error(err.message),
    });
  }

  if (isLoading) {
    return <Skeleton className="h-96" />;
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/channels">Channels</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit Channel</h2>

      <ChannelForm
        mode="edit"
        initialValue={data?.channel}
        submitLabel="Save"
        loading={updateMutation.isPending}
        onCancel={() => navigate("/channels")}
        onSubmit={onSubmit}
      />
    </>
  );
}
