import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCreateChannel } from "@/api/channels";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import ChannelForm from "./form";
import type { AgentChannel } from "@/types/api";

export default function ChannelCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateChannel();

  function onSubmit(channel: AgentChannel) {
    createMutation.mutate(channel, {
      onSuccess: () => {
        toast.success("Channel created");
        navigate("/channels");
      },
      onError: (err) => toast.error(err.message),
    });
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/channels">Channels</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Create Channel</h2>

      <ChannelForm
        mode="create"
        submitLabel="Create"
        loading={createMutation.isPending}
        onCancel={() => navigate("/channels")}
        onSubmit={onSubmit}
      />
    </>
  );
}
