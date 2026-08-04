import { useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import { useChannel, useUpdateChannel } from "@/api/channels";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
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

  return (
    <Page>
      <PageHeader className="max-w-3xl" title="Edit Channel" subtitle={name} />
      <PageScroll className="max-w-3xl">
        {isLoading ? (
          <Skeleton className="h-96" />
        ) : (
          <ChannelForm
            mode="edit"
            initialValue={data?.channel}
            submitLabel="Save"
            loading={updateMutation.isPending}
            onCancel={() => navigate("/channels")}
            onSubmit={onSubmit}
          />
        )}
      </PageScroll>
    </Page>
  );
}
