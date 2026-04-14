import { useSearchParams } from "react-router-dom";
import { useSession } from "@/api/sessions";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";

export default function SessionDetailPage() {
  const [searchParams] = useSearchParams();
  const appName = searchParams.get("app") ?? "";
  const userId = searchParams.get("user") ?? "";
  const sessionId = searchParams.get("session") ?? "";
  const { data, isLoading } = useSession(appName, userId, sessionId);

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  const detail = data?.session_detail;
  if (!detail) return <p className="text-muted-foreground">Session not found.</p>;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/sessions">Sessions</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{sessionId}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Session Detail</h2>

      <Card className="mb-6">
        <CardHeader><CardTitle>Info</CardTitle></CardHeader>
        <CardContent className="grid grid-cols-2 gap-4 text-sm">
          <div><span className="text-muted-foreground">Session ID:</span> {detail.session.session_id}</div>
          <div><span className="text-muted-foreground">App:</span> {detail.session.app_name}</div>
          <div><span className="text-muted-foreground">User:</span> {detail.session.user_id}</div>
          <div><span className="text-muted-foreground">Last Update:</span> {detail.session.last_update_time ? new Date(detail.session.last_update_time).toLocaleString() : "-"}</div>
        </CardContent>
      </Card>

      <h3 className="mb-4 text-lg font-semibold">Events ({detail.events?.length ?? 0})</h3>
      <div className="space-y-3">
        {(detail.events ?? []).map((evt) => (
          <Card key={evt.event_id}>
            <CardContent className="pt-4">
              <div className="mb-2 flex items-center gap-2">
                <Badge variant={evt.author === "user" ? "default" : "secondary"}>{evt.author ?? "unknown"}</Badge>
                <span className="text-xs text-muted-foreground">{evt.timestamp ? new Date(evt.timestamp).toLocaleString() : ""}</span>
              </div>
              <pre className="max-h-40 overflow-auto rounded bg-muted p-3 text-xs">{evt.content_json ?? ""}</pre>
            </CardContent>
          </Card>
        ))}
        {(!detail.events || detail.events.length === 0) && (
          <p className="text-muted-foreground">No events in this session.</p>
        )}
      </div>
    </>
  );
}
