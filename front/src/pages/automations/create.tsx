import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCreateAutomation } from "@/api/automations";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { AutomationForm } from "./form";
import type { Automation } from "@/types/api";

export default function AutomationCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateAutomation();

  function onSubmit(automation: Automation) {
    createMutation.mutate(automation, {
      onSuccess: () => {
        toast.success("Automation created");
        navigate("/automations");
      },
      onError: (err) => toast.error(err.message),
    });
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/automations">Automations</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <div className="mb-6 space-y-1">
        <h2 className="text-2xl font-bold">Create Automation</h2>
        <p className="text-sm text-muted-foreground">Configure a trigger, conditions, and ordered actions.</p>
      </div>
      <AutomationForm isSaving={createMutation.isPending} onSubmit={onSubmit} />
    </>
  );
}
