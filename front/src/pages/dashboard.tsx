import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useCronJobs, useDashboardExecutions } from "@/api/cron";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts";
import type { CronExecution } from "@/types/api";

function formatDuration(start?: string, end?: string): string {
  if (!start || !end) return "-";
  const ms = new Date(end).getTime() - new Date(start).getTime();
  return `${(ms / 1000).toFixed(1)}s`;
}

function computeStats(executions: CronExecution[], activeJobs: number) {
  const total = executions.length;
  const success = executions.filter((e) => e.status === "CRON_EXECUTION_STATUS_SUCCESS").length;
  const errors = total - success;
  const rate = total > 0 ? ((success / total) * 100).toFixed(1) : "0";
  const durations = executions
    .map((e) => {
      if (!e.started_at || !e.finished_at) return 0;
      return (new Date(e.finished_at).getTime() - new Date(e.started_at).getTime()) / 1000;
    })
    .filter((d) => d > 0);
  const avgDuration = durations.length > 0 ? (durations.reduce((a, b) => a + b, 0) / durations.length).toFixed(1) : "0";

  return { total, success, errors, rate, activeJobs, avgDuration };
}

function buildTimelineData(executions: CronExecution[]) {
  const now = new Date();
  const buckets: Record<string, { hour: string; success: number; error: number }> = {};

  for (let i = 23; i >= 0; i--) {
    const d = new Date(now.getTime() - i * 3600_000);
    const key = `${d.getHours().toString().padStart(2, "0")}:00`;
    buckets[key] = { hour: key, success: 0, error: 0 };
  }

  for (const e of executions) {
    if (!e.started_at) continue;
    const d = new Date(e.started_at);
    const key = `${d.getHours().toString().padStart(2, "0")}:00`;
    if (buckets[key]) {
      if (e.status === "CRON_EXECUTION_STATUS_SUCCESS") buckets[key].success++;
      else buckets[key].error++;
    }
  }

  return Object.values(buckets);
}

const COLORS = { success: "#4ade80", error: "#ef4444" };

export default function DashboardPage() {
  const { data: jobsData, isLoading: jobsLoading } = useCronJobs();
  const { data: execData, isLoading: execLoading } = useDashboardExecutions();
  const navigate = useNavigate();

  const executions = execData?.executions ?? [];
  const activeJobs = (jobsData?.cron_jobs ?? []).filter((j) => j.enabled).length;

  const stats = useMemo(() => computeStats(executions, activeJobs), [executions, activeJobs]);
  const timelineData = useMemo(() => buildTimelineData(executions), [executions]);
  const pieData = useMemo(
    () => [
      { name: "Success", value: stats.success },
      { name: "Error", value: stats.errors },
    ],
    [stats],
  );

  const isLoading = jobsLoading || execLoading;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h2 className="text-2xl font-bold">Dashboard</h2>
        <div className="grid grid-cols-4 gap-4">{Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-28" />)}</div>
        <Skeleton className="h-72" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Dashboard</h2>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground">Total Executions</CardTitle></CardHeader>
          <CardContent><div className="text-3xl font-bold">{stats.total}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground">Success Rate</CardTitle></CardHeader>
          <CardContent><div className="text-3xl font-bold text-green-500">{stats.rate}%</div><p className="text-xs text-muted-foreground">{stats.success} passed / {stats.errors} failed</p></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground">Active Cron Jobs</CardTitle></CardHeader>
          <CardContent><div className="text-3xl font-bold">{stats.activeJobs}</div></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-sm text-muted-foreground">Avg Duration</CardTitle></CardHeader>
          <CardContent><div className="text-3xl font-bold">{stats.avgDuration}s</div></CardContent>
        </Card>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-3 gap-4">
        {/* Timeline */}
        <Card className="col-span-2">
          <CardHeader><CardTitle>Execution Timeline</CardTitle></CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={timelineData}>
                <XAxis dataKey="hour" tick={{ fontSize: 11 }} />
                <YAxis tick={{ fontSize: 11 }} />
                <Tooltip />
                <Legend />
                <Bar dataKey="success" stackId="a" fill={COLORS.success} name="Success" />
                <Bar dataKey="error" stackId="a" fill={COLORS.error} name="Error" />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* Donut */}
        <Card>
          <CardHeader><CardTitle>Status Breakdown</CardTitle></CardHeader>
          <CardContent className="flex items-center justify-center">
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie data={pieData} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, value }: { name: string; value: number }) => `${name}: ${value}`}>
                  <Cell fill={COLORS.success} />
                  <Cell fill={COLORS.error} />
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* Recent Executions */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Recent Executions</CardTitle>
          <button className="text-sm text-primary hover:underline" onClick={() => navigate("/cron")}>View all &rarr;</button>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {executions.slice(0, 10).map((e) => (
              <div key={e.id} className="flex items-center gap-4 rounded-md border px-4 py-2 text-sm">
                <span className="w-36 font-medium truncate">{e.job_name}</span>
                <span className="w-24 text-muted-foreground">{e.agent_name}</span>
                {e.status === "CRON_EXECUTION_STATUS_SUCCESS"
                  ? <Badge className="bg-green-500/10 text-green-500">Success</Badge>
                  : <Badge variant="destructive">Error</Badge>}
                <span className="w-16 text-muted-foreground">{formatDuration(e.started_at, e.finished_at)}</span>
                <span className="w-40 text-xs text-muted-foreground">{e.started_at ? new Date(e.started_at).toLocaleString() : "-"}</span>
                <span className="flex-1 truncate text-xs text-muted-foreground">{e.output ?? ""}</span>
              </div>
            ))}
            {executions.length === 0 && <p className="text-center text-muted-foreground py-8">No executions yet.</p>}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
