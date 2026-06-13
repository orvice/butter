import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCreateCronJob } from "@/api/cron";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import CronJobForm from "./form";
import type { CronJob } from "@/types/api";

export default function CronJobCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateCronJob();

  function onSubmit(job: CronJob) {
    createMutation.mutate(job, {
      onSuccess: () => { toast.success("Cron job created"); navigate("/cron"); },
      onError: (err) => toast.error(err.message),
    });
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/cron">Cron Jobs</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <div className="mb-6 space-y-1">
        <h2 className="text-2xl font-bold">Create Cron Job</h2>
        <p className="text-sm text-muted-foreground">Schedule an agent run, define the prompt, and choose where the result is delivered.</p>
      </div>

      <CronJobForm
        mode="create"
        submitLabel="Create"
        loading={createMutation.isPending}
        onCancel={() => navigate("/cron")}
        onSubmit={onSubmit}
      />
    </>
  );
}
