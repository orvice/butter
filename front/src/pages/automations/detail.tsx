import { useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import { useAutomation, useAutomationRuns, useAutomationStepRuns, useRunAutomationNow } from "@/api/automations";
import { Badge } from "@/components/ui/badge";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import type { AutomationRun, AutomationStepRun } from "@/types/api";
import { ExternalLink, Play, Workflow } from "lucide-react";

function formatTime(ts?: string) {
  return ts ? new Date(ts).toLocaleString() : "-";
}

function runStatusClass(run: AutomationRun) {
  if (run.status === "AUTOMATION_RUN_STATUS_SUCCEEDED") return "bg-emerald-500/10 text-emerald-700";
  if (run.status === "AUTOMATION_RUN_STATUS_FAILED" || run.status === "AUTOMATION_RUN_STATUS_CANCELLED") return "bg-rose-500/10 text-rose-700";
  return "bg-amber-500/10 text-amber-700";
}

function stepStatusClass(step: AutomationStepRun) {
  if (step.status === "AUTOMATION_STEP_RUN_STATUS_SUCCEEDED") return "bg-emerald-500/10 text-emerald-700";
  if (step.status === "AUTOMATION_STEP_RUN_STATUS_FAILED" || step.status === "AUTOMATION_STEP_RUN_STATUS_CANCELLED") return "bg-rose-500/10 text-rose-700";
  return "bg-amber-500/10 text-amber-700";
}

export default function AutomationDetailPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useAutomation(name ?? "");
  const { data: runsData, isLoading: loadingRuns } = useAutomationRuns(name, 100);
  const [selectedRunId, setSelectedRunId] = useState("");
  const selectedRun = useMemo(() => {
    const runs = runsData?.runs ?? [];
    return runs.find((run) => run.id === selectedRunId) ?? runs[0];
  }, [runsData?.runs, selectedRunId]);
  const { data: stepData, isLoading: loadingSteps } = useAutomationStepRuns(selectedRun?.id ?? "");
  const runNow = useRunAutomationNow();
  const automation = data?.automation;

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/automations">Automations</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold">{name}</h2>
          <p className="text-sm text-muted-foreground">Run history and step-level execution records.</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => navigate(`/automations/${encodeURIComponent(name ?? "")}/edit`)}>
            <ExternalLink className="mr-2 h-4 w-4" />
            Edit
          </Button>
          <Button
            disabled={runNow.isPending || !(automation?.enabled ?? false)}
            onClick={() =>
              runNow.mutate(
                { name: name ?? "", trigger_payload_json: "{}" },
                {
                  onSuccess: () => toast.success(`${name} executed`),
                  onError: (err) => toast.error(err.message),
                },
              )
            }
          >
            <Play className="mr-2 h-4 w-4" />
            Run now
          </Button>
        </div>
      </div>

      <div className="grid gap-6 xl:grid-cols-[360px_minmax(0,1fr)]">
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2"><Workflow className="h-4 w-4" /> Definition</CardTitle>
              <CardDescription>{automation?.enabled ? "Enabled" : "Disabled"}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex justify-between gap-4"><span className="text-muted-foreground">Trigger</span><span>{automation?.trigger?.type?.replace("AUTOMATION_TRIGGER_TYPE_", "")}</span></div>
              <div className="flex justify-between gap-4"><span className="text-muted-foreground">Schedule</span><span>{automation?.trigger?.schedule?.schedule ?? "-"}</span></div>
              <div className="flex justify-between gap-4"><span className="text-muted-foreground">Conditions</span><span>{automation?.conditions?.length ?? 0}</span></div>
              <div className="flex justify-between gap-4"><span className="text-muted-foreground">Steps</span><span>{automation?.steps?.length ?? 0}</span></div>
              <div className="flex justify-between gap-4"><span className="text-muted-foreground">Updated</span><span>{formatTime(automation?.updated_at)}</span></div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Runs</CardTitle>
              <CardDescription>Newest run first.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              {loadingRuns ? <Skeleton className="h-28" /> : null}
              {(runsData?.runs ?? []).length === 0 && !loadingRuns ? <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">No runs yet.</div> : null}
              {(runsData?.runs ?? []).map((run) => (
                <button
                  key={run.id}
                  className={`w-full rounded-md border p-3 text-left transition-colors ${selectedRun?.id === run.id ? "border-primary bg-primary/5" : "hover:bg-muted/50"}`}
                  onClick={() => setSelectedRunId(run.id)}
                >
                  <div className="flex items-center justify-between gap-2">
                    <Badge className={runStatusClass(run)}>{run.status.replace("AUTOMATION_RUN_STATUS_", "")}</Badge>
                    <span className="text-xs text-muted-foreground">{formatTime(run.started_at)}</span>
                  </div>
                  <div className="mt-2 truncate font-mono text-xs text-muted-foreground">{run.id}</div>
                </button>
              ))}
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Step Runs</CardTitle>
            <CardDescription className="font-mono">{selectedRun?.id ?? "No run selected"}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {loadingSteps ? <Skeleton className="h-40" /> : null}
            {(stepData?.step_runs ?? []).length === 0 && !loadingSteps ? <div className="rounded-md border border-dashed p-8 text-center text-sm text-muted-foreground">No step runs for this execution.</div> : null}
            {(stepData?.step_runs ?? []).map((step) => (
              <div key={step.id} className="rounded-md border p-4">
                <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div className="font-medium">{step.order}. {step.step_name}</div>
                    <div className="text-xs text-muted-foreground">{step.step_type.replace("AUTOMATION_STEP_TYPE_", "")}</div>
                  </div>
                  <Badge className={stepStatusClass(step)}>{step.status.replace("AUTOMATION_STEP_RUN_STATUS_", "")}</Badge>
                </div>
                <div className="mt-3 grid gap-3 text-sm md:grid-cols-3">
                  <div><span className="block text-xs text-muted-foreground">Attempts</span>{step.attempt_count ?? 0}</div>
                  <div><span className="block text-xs text-muted-foreground">Duration</span>{step.duration_ms ?? 0}ms</div>
                  <div><span className="block text-xs text-muted-foreground">Truncated</span>{step.truncated ? "Yes" : "No"}</div>
                </div>
                {step.error ? <div className="mt-3 rounded-md bg-rose-500/10 p-3 text-xs text-rose-700">{step.error}</div> : null}
                {step.output_json ? (
                  <pre className="mt-3 max-h-64 overflow-auto rounded-md bg-muted p-3 text-xs leading-5">{step.output_json}</pre>
                ) : null}
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </>
  );
}
