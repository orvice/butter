import { cn } from "@/lib/utils";

export function ButterLogo({ className }: { className?: string }) {
  return (
    <span className={cn("inline-flex items-center gap-2", className)}>
      <span
        className="flex size-7 items-center justify-center rounded-md bg-brand text-brand-foreground"
        aria-hidden
      >
        <svg viewBox="0 0 24 24" className="size-4" fill="none">
          <path
            d="M5 8.5 L12 5 L19 8.5 L19 15.5 L12 19 L5 15.5 Z"
            stroke="currentColor"
            strokeWidth="1.6"
            strokeLinejoin="round"
          />
          <path
            d="M12 5 L12 19 M5 8.5 L19 15.5 M19 8.5 L5 15.5"
            stroke="currentColor"
            strokeWidth="1.1"
            strokeLinecap="round"
            opacity="0.5"
          />
        </svg>
      </span>
      <span className="text-[0.95rem] font-semibold tracking-tight text-sidebar-foreground">
        Butter
      </span>
    </span>
  );
}
