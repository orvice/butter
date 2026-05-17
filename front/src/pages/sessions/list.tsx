import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useChannels } from "@/api/channels";
import { useSessions, useDeleteSession } from "@/api/sessions";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { MoreHorizontal, Eye, Trash2, Filter, ChevronLeft, ChevronRight, BarChart3 } from "lucide-react";
import type { SessionInfo } from "@/types/api";

export default function SessionListPage() {
  const [filters, setFilters] = useState({
    app_name: "",
    user_id: "",
    start_time: "",
    end_time: "",
  });
  const [pageTokens, setPageTokens] = useState<string[]>([""]);
  const [pageIndex, setPageIndex] = useState(0);
  const PAGE_SIZE = 20;

  const { data: channels } = useChannels();
  const channelNames = (channels?.channels ?? []).map((c) => c.name);

  const params = {
    app_name: filters.app_name === "__all__" ? "" : filters.app_name || undefined,
    user_id: filters.user_id || undefined,
    start_time: filters.start_time ? new Date(filters.start_time).toISOString() : undefined,
    end_time: filters.end_time ? new Date(filters.end_time).toISOString() : undefined,
    page_size: PAGE_SIZE,
    page_token: pageTokens[pageIndex] || undefined,
  };

  const { data, isLoading } = useSessions(params);
  const deleteMutation = useDeleteSession();
  const navigate = useNavigate();
  const [deleteTarget, setDeleteTarget] = useState<SessionInfo | null>(null);

  const sessions = data?.sessions ?? [];
  const total = data?.total ?? 0;
  const nextToken = data?.next_page_token;

  function applyFilters() {
    setPageTokens([""]);
    setPageIndex(0);
  }

  function nextPage() {
    if (!nextToken) return;
    if (pageIndex === pageTokens.length - 1) {
      setPageTokens([...pageTokens, nextToken]);
    }
    setPageIndex(pageIndex + 1);
  }
  function prevPage() {
    if (pageIndex > 0) setPageIndex(pageIndex - 1);
  }

  const columns: Column<SessionInfo>[] = [
    {
      header: "Session ID",
      cell: (row) => <span className="font-mono text-xs">{row.session_id.slice(0, 24)}…</span>,
    },
    { header: "Channel", accessorKey: "app_name" },
    { header: "User", accessorKey: "user_id" },
    {
      header: "Start Time",
      cell: (row) =>
        row.last_update_time ? (
          <span className="text-xs text-muted-foreground">
            {new Date(row.last_update_time).toLocaleString()}
          </span>
        ) : "—",
    },
    { header: "Turns", cell: (row) => row.turn_count ?? 0 },
    {
      header: "Tracing",
      cell: (row) => (
        <Button
          variant="ghost"
          size="sm"
          className="text-xs"
          onClick={() =>
            navigate(
              `/sessions/detail?app=${encodeURIComponent(row.app_name)}&user=${encodeURIComponent(row.user_id)}&sid=${encodeURIComponent(row.session_id)}`,
            )
          }
        >
          <BarChart3 className="mr-1 h-3 w-3" /> Trace
        </Button>
      ),
    },
    {
      header: "",
      cell: (row) => (
        <div className="flex justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={() =>
                  navigate(
                    `/sessions/detail?app=${encodeURIComponent(row.app_name)}&user=${encodeURIComponent(row.user_id)}&sid=${encodeURIComponent(row.session_id)}`,
                  )
                }
              >
                <Eye className="mr-2 h-4 w-4" /> View
              </DropdownMenuItem>
              <DropdownMenuItem className="text-destructive" onClick={() => setDeleteTarget(row)}>
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
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Session Explorer</h2>
          <p className="text-sm text-muted-foreground">
            View ADK sessions, events, and long-term memory stored in MongoDB.
          </p>
        </div>
      </div>

      {/* Filter bar */}
      <Card className="mb-4">
        <CardContent className="grid grid-cols-1 gap-3 p-4 md:grid-cols-5">
          <div>
            <Label htmlFor="filter-channel" className="text-xs">
              Channel
            </Label>
            <Select
              value={filters.app_name || "__all__"}
              onValueChange={(v) => {
                const next = typeof v === "string" && v !== "__all__" ? v : "";
                setFilters((f) => ({ ...f, app_name: next }));
              }}
            >
              <SelectTrigger id="filter-channel" className="mt-1">
                <SelectValue placeholder="All channels" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">All channels</SelectItem>
                {channelNames.map((n) => (
                  <SelectItem key={n} value={n}>
                    {n}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label htmlFor="filter-user" className="text-xs">
              User
            </Label>
            <Input
              id="filter-user"
              placeholder="User ID"
              value={filters.user_id}
              onChange={(e) => setFilters((f) => ({ ...f, user_id: e.target.value }))}
              className="mt-1"
            />
          </div>
          <div>
            <Label htmlFor="filter-start" className="text-xs">
              From
            </Label>
            <Input
              id="filter-start"
              type="datetime-local"
              value={filters.start_time}
              onChange={(e) => setFilters((f) => ({ ...f, start_time: e.target.value }))}
              className="mt-1"
            />
          </div>
          <div>
            <Label htmlFor="filter-end" className="text-xs">
              To
            </Label>
            <Input
              id="filter-end"
              type="datetime-local"
              value={filters.end_time}
              onChange={(e) => setFilters((f) => ({ ...f, end_time: e.target.value }))}
              className="mt-1"
            />
          </div>
          <div className="flex items-end">
            <Button onClick={applyFilters} className="w-full">
              <Filter className="mr-2 h-4 w-4" /> Apply
            </Button>
          </div>
        </CardContent>
      </Card>

      <DataTable
        columns={columns}
        data={sessions}
        isLoading={isLoading}
        emptyMessage="No sessions match the current filters."
      />

      <div className="mt-3 flex items-center justify-between text-xs text-muted-foreground">
        <span>
          Showing {sessions.length} of {total.toLocaleString()} sessions
        </span>
        <div className="flex items-center gap-2">
          <Button size="icon" variant="outline" onClick={prevPage} disabled={pageIndex === 0}>
            <ChevronLeft className="h-3 w-3" />
          </Button>
          <span>Page {pageIndex + 1}</span>
          <Button size="icon" variant="outline" onClick={nextPage} disabled={!nextToken}>
            <ChevronRight className="h-3 w-3" />
          </Button>
        </div>
      </div>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Session"
        description={`Delete session "${deleteTarget?.session_id}"?`}
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(
              {
                app_name: deleteTarget.app_name,
                user_id: deleteTarget.user_id,
                session_id: deleteTarget.session_id,
              },
              {
                onSuccess: () => {
                  toast.success("Session deleted");
                  setDeleteTarget(null);
                },
                onError: (err) => toast.error(err.message),
              },
            );
          }
        }}
      />
    </>
  );
}
