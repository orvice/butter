import type { ReactNode } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { EmptyState } from "@/components/empty-state";

export interface Column<T> {
  header: string;
  accessorKey?: keyof T;
  cell?: (row: T) => ReactNode;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[] | undefined;
  isLoading: boolean;
  emptyMessage?: string;
  emptyDescription?: string;
  emptyAction?: ReactNode;
}

export function DataTable<T>({ columns, data, isLoading, emptyMessage = "No data", emptyDescription, emptyAction }: DataTableProps<T>) {
  function renderCell(row: T, col: Column<T>) {
    return col.cell
      ? col.cell(row)
      : col.accessorKey
        ? String((row as Record<string, unknown>)[col.accessorKey as string] ?? "")
        : null;
  }

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    );
  }

  if (!data || data.length === 0) {
    return <EmptyState title={emptyMessage} description={emptyDescription} action={emptyAction} />;
  }

  return (
    <div className="rounded-lg border bg-card shadow-[0_1px_3px_rgba(0,0,0,0.05)]">
      <div className="hidden overflow-x-auto md:block">
        <Table className="min-w-full whitespace-nowrap">
          <TableHeader>
            <TableRow>
              {columns.map((col) => (
                <TableHead key={col.header}>{col.header}</TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.map((row, i) => (
              <TableRow key={i}>
                {columns.map((col) => (
                  <TableCell key={col.header}>
                    {renderCell(row, col)}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <div className="divide-y md:hidden">
        {data.map((row, i) => (
          <div key={i} className="space-y-3 p-4">
            {columns.map((col) => (
              <div key={col.header} className="grid gap-1 text-sm">
                <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{col.header}</div>
                <div className="min-w-0 break-words text-card-foreground">{renderCell(row, col)}</div>
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}
