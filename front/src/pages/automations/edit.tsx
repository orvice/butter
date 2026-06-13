import { useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import { useAutomation, useUpdateAutomation } from "@/api/automations";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import { AutomationForm } from "./form";
import type { Automation } from "@/types/api";

export default function AutomationEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useAutomation(name ?? "");
  const updateMutation = useUpdateAutomation();

  function onSubmit(automation: Automation) {
    updateMutation.mutate(automation, {
      onSuccess: () => {
        toast.success("Automation updated");
        navigate(`/automations/${encodeURIComponent(automation.name)}`);
      },
      onError: (err) => toast.error(err.message),
    });
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/automations">Automations</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Edit</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <div className="mb-6 space-y-1">
        <h2 className="text-2xl font-bold">Edit Automation</h2>
        <p className="text-sm text-muted-foreground">Update trigger policy, conditions, and ordered actions.</p>
      </div>
      <AutomationForm automation={data?.automation} isSaving={updateMutation.isPending} onSubmit={onSubmit} />
    </>
  );
}
