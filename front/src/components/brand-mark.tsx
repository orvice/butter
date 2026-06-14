import { cn } from "@/lib/utils";

interface Props {
  className?: string;
  size?: number;
}

export function BrandMark({ className = "", size = 36 }: Props) {
  return (
    <span
      aria-hidden="true"
      className={cn(
        "inline-flex shrink-0 items-center justify-center rounded-full border border-white/20 bg-[#2b658b] shadow-[inset_0_1px_0_color-mix(in_srgb,white_35%,transparent)]",
        className,
      )}
      style={{ height: size, width: size, fontSize: size * 0.5 }}
    >
      <img src="/md-theme/logo_b.svg" alt="" className="h-full w-full" />
    </span>
  );
}
