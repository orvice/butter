import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCronJob, useUpdateCronJob } from "@/api/cron";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import CronJobForm from "./form";
import type { CronJob } from "@/types/api";

export default function CronJobEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useCronJob(name ?? "");
  const updateMutation = useUpdateCronJob();

  function onSubmit(job: CronJob) {
    updateMutation.mutate(job, {
      onSuccess: () => { toast.success("Cron job updated"); navigate("/cron"); },
      onError: (err) => toast.error(err.message),
    });
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/cron">Cron Jobs</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <div className="mb-6 space-y-1">
        <h2 className="text-2xl font-bold">Edit Cron Job</h2>
        <p className="text-sm text-muted-foreground">Review schedule, agent input, and delivery settings before the next automatic run.</p>
      </div>

      <CronJobForm
        mode="edit"
        submitLabel="Save"
        loading={updateMutation.isPending}
        initialValue={data?.cron_job}
        onCancel={() => navigate("/cron")}
        onSubmit={onSubmit}
      />
    </>
  );
}
