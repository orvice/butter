import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { ArrowLeft, ChevronRight } from "lucide-react";
import { useCreateCronJob } from "@/api/cron";
import { Page, PageScroll } from "@/components/butter/page-parts";
import { AutomationForm } from "./form";
import type { CronJob } from "@/types/api";

export default function AutomationCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateCronJob();

  function handleSubmit(job: CronJob) {
    createMutation.mutate(job, {
      onSuccess: () => {
        toast.success("Automation created");
        navigate("/automations");
      },
      onError: (err) => toast.error(err.message),
    });
  }

  return (
    <Page>
      <div className="border-b border-border px-4 py-4 md:px-6">
        <Link
          to="/automations"
          className="mb-3 inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="size-3.5" />
          Automations
          <ChevronRight className="size-3" />
          <span className="text-foreground">Create</span>
        </Link>
        <h1 className="text-lg font-semibold tracking-tight">Create Automation</h1>
        <p className="mt-0.5 text-sm text-muted-foreground">
          Run an agent on a schedule and deliver the result to a webhook, channel, or notify group.
        </p>
      </div>
      <PageScroll className="max-w-3xl">
        <AutomationForm
          mode="create"
          loading={createMutation.isPending}
          submitLabel="Create Automation"
          onCancel={() => navigate("/automations")}
          onSubmit={handleSubmit}
        />
      </PageScroll>
    </Page>
  );
}
