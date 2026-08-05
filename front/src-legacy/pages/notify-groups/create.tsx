import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCreateNotifyGroup } from "@/api/notify-groups";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
import type { NotifyGroup } from "@/types/api";
import NotifyGroupForm from "./form";

export default function NotifyGroupCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateNotifyGroup();

  function onSubmit(group: NotifyGroup) {
    createMutation.mutate(group, {
      onSuccess: () => {
        toast.success("Notify group created");
        navigate("/notify-groups");
      },
      onError: (err) => toast.error(err.message),
    });
  }

  return (
    <Page>
      <PageHeader
        className="max-w-3xl"
        title="Create Notify Group"
        subtitle="Configure outbound notification targets for cron jobs."
      />
      <PageScroll className="max-w-3xl">
        <NotifyGroupForm submitLabel="Create" submitting={createMutation.isPending} onSubmit={onSubmit} />
      </PageScroll>
    </Page>
  );
}
