import { Link, useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import { ArrowLeft, ChevronRight } from "lucide-react";
import { useCronJob, useUpdateCronJob } from "@/api/cron";
import { Page, PageScroll } from "@/components/butter/page-parts";
import { Skeleton } from "@/components/ui/skeleton";
import { AutomationForm } from "./form";
import type { CronJob } from "@/types/api";

export default function AutomationEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useCronJob(name ?? "");
  const updateMutation = useUpdateCronJob();

  const job = data?.cron_job;

  function handleSubmit(next: CronJob) {
    updateMutation.mutate(next, {
      onSuccess: () => {
        toast.success("Automation updated");
        navigate(`/automations/${encodeURIComponent(next.name)}`);
      },
      onError: (err) => toast.error(err.message),
    });
  }

  if (isLoading) {
    return (
      <div className="p-6">
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  if (!job) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-sm text-muted-foreground">Automation not found.</p>
      </div>
    );
  }

  return (
    <Page>
      <div className="border-b border-border px-4 py-4 md:px-6">
        <Link
          to={`/automations/${encodeURIComponent(job.name)}`}
          className="mb-3 inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="size-3.5" />
          Automations
          <ChevronRight className="size-3" />
          {job.name}
          <ChevronRight className="size-3" />
          <span className="text-foreground">Edit</span>
        </Link>
        <h1 className="text-lg font-semibold tracking-tight">Edit {job.name}</h1>
        <p className="mt-0.5 text-sm text-muted-foreground">
          Update the schedule, agent, reliability, and delivery settings.
        </p>
      </div>
      <PageScroll className="max-w-3xl">
        <AutomationForm
          mode="edit"
          initialValue={job}
          loading={updateMutation.isPending}
          submitLabel="Save Changes"
          onCancel={() => navigate(`/automations/${encodeURIComponent(job.name)}`)}
          onSubmit={handleSubmit}
        />
      </PageScroll>
    </Page>
  );
}
