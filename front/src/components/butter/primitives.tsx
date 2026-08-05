import { cn } from "@/lib/utils";
import {
  CircleCheck,
  CircleDashed,
  CircleDot,
  CircleSlash,
  Loader2,
  TriangleAlert,
} from "lucide-react";
import type { ReactNode } from "react";

/* ----------------------------- Status badge ----------------------------- */

export type RunStatus =
  | "running"
  | "success"
  | "failed"
  | "waiting"
  | "disabled"
  | "never";

const statusConfig: Record<
  RunStatus,
  { label: string; className: string; icon: ReactNode }
> = {
  running: {
    label: "Running",
    className: "bg-running-muted text-running-foreground",
    icon: <Loader2 className="size-3 animate-spin" />,
  },
  success: {
    label: "Success",
    className: "bg-success-muted text-success-foreground",
    icon: <CircleCheck className="size-3" />,
  },
  failed: {
    label: "Failed",
    className: "bg-danger-muted text-danger-foreground",
    icon: <TriangleAlert className="size-3" />,
  },
  waiting: {
    label: "Waiting for input",
    className: "bg-warning-muted text-warning-foreground",
    icon: <CircleDot className="size-3" />,
  },
  disabled: {
    label: "Disabled",
    className: "bg-muted text-muted-foreground",
    icon: <CircleSlash className="size-3" />,
  },
  never: {
    label: "Never run",
    className: "bg-muted text-muted-foreground",
    icon: <CircleDashed className="size-3" />,
  },
};

export function StatusBadge({
  status,
  label,
  className,
}: {
  status: RunStatus;
  label?: string;
  className?: string;
}) {
  const cfg = statusConfig[status];
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 text-xs font-medium",
        cfg.className,
        className,
      )}
    >
      {cfg.icon}
      <span>{label ?? cfg.label}</span>
    </span>
  );
}

export function StatusDot({ status }: { status: RunStatus }) {
  const color: Record<RunStatus, string> = {
    running: "bg-running",
    success: "bg-success",
    failed: "bg-danger",
    waiting: "bg-warning",
    disabled: "bg-muted-foreground/50",
    never: "bg-muted-foreground/40",
  };
  return (
    <span className="inline-flex size-2 items-center justify-center">
      <span
        className={cn(
          "size-2 rounded-full",
          color[status],
          status === "running" && "animate-blink",
        )}
      />
    </span>
  );
}

/* ----------------------------- Agent avatar ------------------------------ */

const avatarTints = [
  "bg-running-muted text-running-foreground",
  "bg-success-muted text-success-foreground",
  "bg-warning-muted text-warning-foreground",
  "bg-[oklch(0.94_0.06_75)] text-[oklch(0.45_0.1_65)] dark:bg-[oklch(0.34_0.06_75)] dark:text-[oklch(0.82_0.12_78)]",
  "bg-danger-muted text-danger-foreground",
];

function tintFor(name: string) {
  let hash = 0;
  for (let i = 0; i < name.length; i += 1) {
    hash = (hash * 31 + name.charCodeAt(i)) | 0;
  }
  return avatarTints[Math.abs(hash) % avatarTints.length];
}

export function AgentAvatar({
  name,
  iconUrl,
  size = "md",
  className,
}: {
  name: string;
  iconUrl?: string;
  size?: "sm" | "md" | "lg";
  className?: string;
}) {
  const sizes = {
    sm: "size-6 text-xs rounded-md",
    md: "size-8 text-sm rounded-md",
    lg: "size-10 text-base rounded-lg",
  };
  if (iconUrl) {
    return (
      <img
        src={iconUrl}
        alt=""
        aria-hidden
        className={cn(
          "shrink-0 object-cover outline outline-1 -outline-offset-1 outline-black/10 dark:outline-white/10",
          sizes[size],
          className,
        )}
      />
    );
  }
  return (
    <span
      className={cn(
        "inline-flex shrink-0 items-center justify-center font-semibold uppercase",
        sizes[size],
        tintFor(name),
        className,
      )}
      aria-hidden
    >
      {name.trim().charAt(0) || "?"}
    </span>
  );
}

/* ------------------------------ Section label ---------------------------- */

export function SectionLabel({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "px-2 text-[0.7rem] font-medium uppercase tracking-wide text-muted-foreground",
        className,
      )}
    >
      {children}
    </div>
  );
}
