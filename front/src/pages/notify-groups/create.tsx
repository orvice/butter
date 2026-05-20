import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCreateNotifyGroup } from "@/api/notify-groups";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
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
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/notify-groups">Notify Groups</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Create Notify Group</h2>
      <NotifyGroupForm submitLabel="Create" submitting={createMutation.isPending} onSubmit={onSubmit} />
    </>
  );
}
