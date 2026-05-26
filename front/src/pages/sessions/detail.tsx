import { useSearchParams, Link } from "react-router-dom";
import { useSession } from "@/api/sessions";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { MemoryStick, Brain, Clock, User, MessageSquare, ExternalLink } from "lucide-react";

function fmtDuration(d?: string): string {
  if (!d) return "-";
  // Protobuf Duration over JSON is something like "12.345s"
  const m = /^(-?\d+(?:\.\d+)?)s$/.exec(d);
  if (!m) return d;
  const secs = parseFloat(m[1]);
  if (secs < 60) return `${secs.toFixed(1)}s`;
  const mins = Math.floor(secs / 60);
  const rem = Math.floor(secs % 60);
  return `${mins}m ${rem}s`;
}

export default function SessionDetailPage() {
  const [searchParams] = useSearchParams();
  // Support both `?sid=` (new) and legacy `?session=`.
  const appName = searchParams.get("app") ?? "";
  const userId = searchParams.get("user") ?? "";
  const sessionId = searchParams.get("sid") ?? searchParams.get("session") ?? "";
  const { data, isLoading } = useSession(appName, userId, sessionId);

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  const detail = data?.session_detail;
  if (!detail) return <p className="text-muted-foreground">Session not found.</p>;

  const events = detail.events ?? [];
  const info = detail.session;
  const stateEntries = Object.entries(info.state ?? {});

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem>
            <Link to="/sessions" className="text-sm hover:underline">
              Sessions
            </Link>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem className="font-mono text-xs">{sessionId}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <div className="mb-6">
        <h2 className="text-2xl font-bold tracking-tight">Session Detail</h2>
        <p className="font-mono text-xs text-muted-foreground">{info.session_id}</p>
      </div>

      {/* Header info */}
      <Card className="mb-6">
        <CardContent className="grid grid-cols-2 gap-4 p-4 text-sm md:grid-cols-4">
          <div className="flex items-center gap-2">
            <User className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="text-xs text-muted-foreground">User</div>
              <div className="font-medium">{info.user_id}</div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <MessageSquare className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="text-xs text-muted-foreground">Channel</div>
              <div className="font-medium">{info.app_name}</div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="text-xs text-muted-foreground">Duration</div>
              <div className="font-medium">{fmtDuration(detail.duration)}</div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Brain className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="text-xs text-muted-foreground">Turns</div>
              <div className="font-medium">{info.turn_count ?? events.length}</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Memory context */}
      <Card className="mb-6">
        <CardHeader className="pb-2">
          <CardTitle className="flex items-center gap-2 text-base">
            <MemoryStick className="h-4 w-4" /> Memory Context
          </CardTitle>
        </CardHeader>
        <CardContent>
          {stateEntries.length === 0 ? (
            <p className="text-sm text-muted-foreground">No session state recorded.</p>
          ) : (
            <pre className="overflow-x-auto rounded bg-muted p-3 text-xs">
              {JSON.stringify(info.state, null, 2)}
            </pre>
          )}
        </CardContent>
      </Card>

      {/* Event log */}
      <h3 className="mb-3 text-lg font-semibold">Event Log ({events.length})</h3>
      <div className="space-y-2">
        {events.map((evt) => (
          <Card key={evt.event_id}>
            <CardContent className="p-3">
              <div className="mb-2 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Badge variant={evt.author === "user" ? "default" : "secondary"}>
                    {evt.author ?? "unknown"}
                  </Badge>
                  <span className="text-xs text-muted-foreground">
                    {evt.timestamp ? new Date(evt.timestamp).toLocaleString() : ""}
                  </span>
                </div>
                {evt.trace_url ? (
                  <a
                    href={evt.trace_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center text-xs text-primary hover:underline"
                  >
                    <ExternalLink className="mr-1 h-3 w-3" /> Trace
                  </a>
                ) : evt.trace_id ? (
                  <span className="font-mono text-[10px] text-muted-foreground">
                    {evt.trace_id.slice(0, 12)}…
                  </span>
                ) : null}
              </div>
              <pre className="max-h-40 overflow-auto rounded bg-muted p-3 text-xs">
                {evt.content_json ?? ""}
              </pre>
            </CardContent>
          </Card>
        ))}
        {events.length === 0 && (
          <p className="text-sm text-muted-foreground">No events in this session.</p>
        )}
      </div>
    </>
  );
}
