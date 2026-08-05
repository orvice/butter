import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

export function PageHeader({
  title,
  subtitle,
  actions,
  breadcrumb,
  className,
}: {
  title: ReactNode;
  subtitle?: ReactNode;
  actions?: ReactNode;
  breadcrumb?: ReactNode;
  className?: string;
}) {
  return (
    <div className="shrink-0 border-b border-border px-4 py-4 sm:py-5 md:px-6 lg:px-8">
      <div className={cn("mx-auto flex w-full max-w-7xl flex-col gap-4 md:flex-row md:items-start md:justify-between", className)}>
        <div className="min-w-0 flex-1">
          {breadcrumb && <div className="mb-3 text-xs text-muted-foreground">{breadcrumb}</div>}
          <h1 className="text-xl font-semibold tracking-tight text-balance">{title}</h1>
          {subtitle && (
            <div className="mt-1 max-w-3xl text-sm leading-5 text-muted-foreground text-pretty">{subtitle}</div>
          )}
        </div>
        {actions && (
          <div className="flex w-full flex-wrap items-center gap-2 sm:w-auto md:max-w-[55%] md:justify-end">
            {actions}
          </div>
        )}
      </div>
    </div>
  );
}

export function PageScroll({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div className="scrollbar-thin min-h-0 flex-1 overflow-y-auto">
      <div
        className={cn(
          "mx-auto w-full max-w-7xl px-4 py-5 pb-[max(1.25rem,env(safe-area-inset-bottom))] md:px-6 md:py-6 lg:px-8",
          className,
        )}
      >
        {children}
      </div>
    </div>
  );
}

export function Page({ children }: { children: ReactNode }) {
  return <div className="flex h-full min-w-0 flex-col">{children}</div>;
}

export function PageActions({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "sticky bottom-0 z-20 -mx-4 mt-6 flex flex-col gap-2 border-t border-border bg-background/95 px-4 py-4 pb-[max(1rem,env(safe-area-inset-bottom))] backdrop-blur supports-[backdrop-filter]:bg-background/85 [&>*]:w-full sm:mx-0 sm:flex-row sm:justify-end sm:px-0 sm:[&>*]:w-auto",
        className,
      )}
    >
      {children}
    </div>
  );
}
