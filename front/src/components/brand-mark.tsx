interface Props {
  className?: string;
  size?: number;
}

export function BrandMark({ className = "", size = 36 }: Props) {
  return (
    <span
      aria-hidden="true"
      className={`inline-flex shrink-0 items-center justify-center rounded-md border border-primary/40 bg-primary text-primary-foreground font-semibold tracking-tight shadow-[inset_0_1px_0_color-mix(in_srgb,white_55%,transparent)] ${className}`}
      style={{ height: size, width: size, fontSize: size * 0.5 }}
    >
      B
    </span>
  );
}
