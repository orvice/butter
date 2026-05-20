import { useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import { useNotifyGroup, useUpdateNotifyGroup } from "@/api/notify-groups";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import type { NotifyGroup } from "@/types/api";
import NotifyGroupForm from "./form";

export default function NotifyGroupEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useNotifyGroup(name ?? "");
  const updateMutation = useUpdateNotifyGroup();

  function onSubmit(group: NotifyGroup) {
    updateMutation.mutate(group, {
      onSuccess: () => {
        toast.success("Notify group updated");
        navigate("/notify-groups");
      },
      onError: (err) => toast.error(err.message),
    });
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/notify-groups">Notify Groups</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit Notify Group</h2>
      <NotifyGroupForm
        initialValue={data?.notify_group}
        submitLabel="Save"
        submitting={updateMutation.isPending}
        onSubmit={onSubmit}
      />
    </>
  );
}
