import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

export function PageHeader({
  title,
  subtitle,
  actions,
}: {
  title: ReactNode;
  subtitle?: ReactNode;
  actions?: ReactNode;
}) {
  return (
    <div className="flex flex-col gap-3 border-b border-border px-4 py-4 md:flex-row md:items-center md:justify-between md:px-6">
      <div className="min-w-0">
        <h1 className="text-lg font-semibold tracking-tight">{title}</h1>
        {subtitle && (
          <div className="mt-0.5 text-sm text-muted-foreground">{subtitle}</div>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
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
      <div className={cn("mx-auto w-full max-w-6xl px-4 py-5 md:px-6", className)}>
        {children}
      </div>
    </div>
  );
}

export function Page({ children }: { children: ReactNode }) {
  return <div className="flex h-full flex-col">{children}</div>;
}
