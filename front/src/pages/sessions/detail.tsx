import { useMemo, useState } from "react";
import { useSearchParams, Link } from "react-router-dom";
import { useSession } from "@/api/sessions";
import { parseSessionEventsFull, type FullParsedEvent } from "@/lib/session-events";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import {
  MemoryStick,
  Brain,
  Clock,
  User,
  MessageSquare,
  ExternalLink,
  Maximize2,
  History,
} from "lucide-react";

const RECENT_EVENTS_LIMIT = 50;

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

function EventHeader({ author, timestamp }: { author: string; timestamp?: string }) {
  return (
    <div className="flex items-center gap-2">
      <Badge variant={author === "user" ? "default" : "secondary"}>{author}</Badge>
      <span className="text-xs font-normal text-muted-foreground">
        {timestamp ? new Date(timestamp).toLocaleString() : ""}
      </span>
    </div>
  );
}

function EventContent({ event }: { event: FullParsedEvent }) {
  if (event.textParts.length === 0 && event.functionCalls.length === 0 && event.functionResponses.length === 0) {
    return <p className="text-xs text-muted-foreground">No content.</p>;
  }

  return (
    <div className="space-y-3">
      {event.textParts.map((part, i) => (
        <p key={i} className="whitespace-pre-wrap text-sm">
          {part.text}
        </p>
      ))}
      {event.functionCalls.map((call, i) => (
        <div key={i}>
          <div className="mb-1 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground">
            Tool Call: {call.name}
          </div>
          <pre className="overflow-auto rounded bg-muted p-3 text-xs">
            {JSON.stringify(call.args, null, 2)}
          </pre>
        </div>
      ))}
      {event.functionResponses.map((resp, i) => (
        <div key={i}>
          <div className="mb-1 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground">
            Tool Response: {resp.name}
          </div>
          <pre className="overflow-auto rounded bg-muted p-3 text-xs">
            {JSON.stringify(resp.response, null, 2)}
          </pre>
        </div>
      ))}
    </div>
  );
}

function EventCard({
  event,
  onExpand,
}: {
  event: FullParsedEvent;
  onExpand: () => void;
}) {
  return (
    <Card>
      <CardContent className="p-3">
        <div className="mb-2 flex items-center justify-between">
          <EventHeader author={event.author} timestamp={event.timestamp} />
          <div className="flex items-center gap-2">
            {event.traceUrl ? (
              <a
                href={event.traceUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center text-xs text-primary hover:underline"
              >
                <ExternalLink className="mr-1 h-3 w-3" /> Trace
              </a>
            ) : event.traceId ? (
              <span className="font-mono text-[10px] text-muted-foreground">
                {event.traceId.slice(0, 12)}…
              </span>
            ) : null}
            <Button size="icon-sm" variant="ghost" aria-label="Expand event" onClick={onExpand}>
              <Maximize2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
        <EventContent event={event} />
      </CardContent>
    </Card>
  );
}

function EventLogList({
  events,
  onExpandEvent,
}: {
  events: FullParsedEvent[];
  onExpandEvent: (evt: FullParsedEvent) => void;
}) {
  if (events.length === 0) {
    return <p className="text-sm text-muted-foreground">No events in this session.</p>;
  }
  return (
    <div className="space-y-2">
      {events.map((evt) => (
        <EventCard key={evt.eventId} event={evt} onExpand={() => onExpandEvent(evt)} />
      ))}
    </div>
  );
}

export default function SessionDetailPage() {
  const [searchParams] = useSearchParams();
  // Support both `?sid=` (new) and legacy `?session=`.
  const appName = searchParams.get("app") ?? "";
  const userId = searchParams.get("user") ?? "";
  const sessionId = searchParams.get("sid") ?? searchParams.get("session") ?? "";
  const [showAllEvents, setShowAllEvents] = useState(false);
  const [logExpanded, setLogExpanded] = useState(false);
  const [expandedEvent, setExpandedEvent] = useState<FullParsedEvent | null>(null);
  const { data, isLoading } = useSession(appName, userId, sessionId, showAllEvents ? 0 : RECENT_EVENTS_LIMIT);

  const detail = data?.session_detail;
  const events = useMemo(() => parseSessionEventsFull(detail?.events), [detail?.events]);

  if (isLoading) return <Skeleton className="h-96 w-full" />;
  if (!detail) return <p className="text-muted-foreground">Session not found.</p>;

  const info = detail.session;
  const stateEntries = Object.entries(info.state ?? {});
  // The backend derives duration and turn count from the returned event slice, so in
  // recent-only mode they describe the loaded range rather than the whole session.
  const partial = !showAllEvents && events.length >= RECENT_EVENTS_LIMIT;
  const rangeNote = partial ? `recent ${RECENT_EVENTS_LIMIT}` : undefined;

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
        <CardContent className="grid grid-cols-2 gap-4 p-4 text-sm md:grid-cols-3 lg:grid-cols-5">
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
            <History className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="text-xs text-muted-foreground">Last Update</div>
              <div className="font-medium">
                {info.last_update_time ? new Date(info.last_update_time).toLocaleString() : "-"}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="text-xs text-muted-foreground">
                Duration{rangeNote ? <span className="ml-1 normal-case">({rangeNote})</span> : null}
              </div>
              <div className="font-medium">{fmtDuration(detail.duration)}</div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Brain className="h-4 w-4 text-muted-foreground" />
            <div>
              <div className="text-xs text-muted-foreground">
                Turns{rangeNote ? <span className="ml-1 normal-case">({rangeNote})</span> : null}
              </div>
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
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-lg font-semibold">
          Event Log ({events.length}
          {partial ? "+" : ""})
        </h3>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={() => setShowAllEvents((v) => !v)}>
            {showAllEvents ? `Show recent ${RECENT_EVENTS_LIMIT}` : "Show full history"}
          </Button>
          <Button size="icon-sm" variant="ghost" aria-label="Expand event log" onClick={() => setLogExpanded(true)}>
            <Maximize2 className="h-4 w-4" />
          </Button>
        </div>
      </div>
      {partial ? (
        <p className="mb-3 text-xs text-muted-foreground">
          Showing the most recent {RECENT_EVENTS_LIMIT} events. Duration and turn count reflect this range — use “Show
          full history” for the whole session.
        </p>
      ) : null}
      <EventLogList events={events} onExpandEvent={setExpandedEvent} />

      <Dialog open={logExpanded} onOpenChange={setLogExpanded}>
        <DialogContent className="flex max-h-[90vh] w-full max-w-[95vw] flex-col overflow-hidden sm:max-w-[95vw]">
          <DialogHeader>
            <DialogTitle>
              Event Log ({events.length}
              {partial ? "+" : ""})
            </DialogTitle>
          </DialogHeader>
          <div className="overflow-y-auto pr-1">
            <EventLogList events={events} onExpandEvent={setExpandedEvent} />
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={!!expandedEvent} onOpenChange={(open) => !open && setExpandedEvent(null)}>
        <DialogContent className="flex max-h-[90vh] w-full max-w-[95vw] flex-col overflow-hidden sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>
              {expandedEvent ? (
                <EventHeader author={expandedEvent.author} timestamp={expandedEvent.timestamp} />
              ) : null}
            </DialogTitle>
          </DialogHeader>
          <div className="overflow-y-auto pr-1">
            {expandedEvent ? <EventContent event={expandedEvent} /> : null}
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
