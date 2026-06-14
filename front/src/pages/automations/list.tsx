import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  useAutomations,
  useAutomationRuns,
  useDeleteAutomation,
  useRunAutomationNow,
  useUpdateAutomation,
} from "@/api/automations";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import type { Automation, AutomationRun } from "@/types/api";
import { CalendarClock, History, MoreHorizontal, Pencil, Play, Trash2, Workflow } from "lucide-react";

function triggerLabel(automation: Automation) {
  return automation.trigger?.type === "AUTOMATION_TRIGGER_TYPE_SCHEDULE"
    ? automation.trigger.schedule?.schedule || "Schedule"
    : "Manual";
}

function lastRunMap(runs: AutomationRun[]) {
  const map = new Map<string, AutomationRun>();
  for (const run of runs) {
    const prev = map.get(run.automation_name);
    if (!prev || new Date(run.started_at ?? 0) > new Date(prev.started_at ?? 0)) {
      map.set(run.automation_name, run);
    }
  }
  return map;
}

function statusBadge(run?: AutomationRun) {
  if (!run) return <span className="text-xs text-muted-foreground">-</span>;
  const status = run.status.replace("AUTOMATION_RUN_STATUS_", "");
  const cls =
    run.status === "AUTOMATION_RUN_STATUS_SUCCEEDED"
      ? "bg-emerald-500/10 text-emerald-700"
      : run.status === "AUTOMATION_RUN_STATUS_FAILED" || run.status === "AUTOMATION_RUN_STATUS_CANCELLED"
        ? "bg-rose-500/10 text-rose-700"
        : "bg-amber-500/10 text-amber-700";
  return <Badge className={cls}>{status}</Badge>;
}

export default function AutomationListPage() {
  const navigate = useNavigate();
  const { data, isLoading } = useAutomations();
  const { data: runsData } = useAutomationRuns(undefined, 200);
  const updateMutation = useUpdateAutomation();
  const deleteMutation = useDeleteAutomation();
  const runNow = useRunAutomationNow();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const lastByAutomation = lastRunMap(runsData?.runs ?? []);

  function toggleEnabled(automation: Automation) {
    updateMutation.mutate(
      { ...automation, enabled: !automation.enabled },
      {
        onSuccess: () => toast.success(`Automation ${automation.enabled ? "disabled" : "enabled"}`),
        onError: (err) => toast.error(err.message),
      },
    );
  }

  const columns: Column<Automation>[] = [
    {
      header: "Automation",
      cell: (row) => (
        <div>
          <div className="font-medium">{row.name}</div>
          <div className="text-xs text-muted-foreground">{row.steps?.length ?? 0} steps</div>
        </div>
      ),
    },
    {
      header: "Trigger",
      cell: (row) => (
        <Badge variant="outline">
          {row.trigger?.type === "AUTOMATION_TRIGGER_TYPE_SCHEDULE" ? <CalendarClock className="mr-1 h-3 w-3" /> : <Workflow className="mr-1 h-3 w-3" />}
          {triggerLabel(row)}
        </Badge>
      ),
    },
    {
      header: "Conditions",
      cell: (row) => <span className="text-sm">{row.conditions?.length ?? 0}</span>,
    },
    {
      header: "Last Run",
      cell: (row) => statusBadge(lastByAutomation.get(row.name)),
    },
    {
      header: "Enabled",
      cell: (row) => <Switch checked={row.enabled ?? false} onCheckedChange={() => toggleEnabled(row)} />,
    },
    {
      header: "",
      cell: (row) => (
        <div className="flex items-center justify-end gap-1">
          <Button
            variant="ghost"
            size="sm"
            disabled={runNow.isPending || !(row.enabled ?? false)}
            onClick={() =>
              runNow.mutate(
                { name: row.name, trigger_payload_json: "{}" },
                {
                  onSuccess: () => toast.success(`${row.name} executed`),
                  onError: (err) => toast.error(err.message),
                },
              )
            }
          >
            <Play className="mr-1 h-3 w-3" />
            Run now
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => navigate(`/automations/${encodeURIComponent(row.name)}`)}>
                <History className="mr-2 h-4 w-4" /> Runs
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => navigate(`/automations/${encodeURIComponent(row.name)}/edit`)}>
                <Pencil className="mr-2 h-4 w-4" /> Edit
              </DropdownMenuItem>
              <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row.name)}>
                <Trash2 className="mr-2 h-4 w-4" /> Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      ),
    },
  ];

  return (
    <>
      <PageHeader
        title="Automations"
        description="Workspace workflows with triggers, conditions, ordered steps, and durable run history."
        createLabel="Create Automation"
        createTo="/automations/create"
      />
      <DataTable
        columns={columns}
        data={data?.automations}
        isLoading={isLoading}
        emptyMessage="No automations yet"
        emptyDescription="Create a workflow to invoke agents, call webhooks, notify teams, or post into forum threads."
        emptyAction={<Button onClick={() => navigate("/automations/create")}><Workflow className="mr-2 h-4 w-4" />Create Automation</Button>}
      />
      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Automation"
        description={`Delete "${deleteTarget}"? Run history remains available in storage but this definition will be removed.`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          deleteMutation.mutate(deleteTarget, {
            onSuccess: () => {
              toast.success("Automation deleted");
              setDeleteTarget(null);
            },
            onError: (err) => toast.error(err.message),
          });
        }}
      />
    </>
  );
}
