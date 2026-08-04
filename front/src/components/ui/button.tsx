import { Button as ButtonPrimitive } from "@base-ui/react/button"
import { type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"
import { buttonVariants } from "@/components/ui/button-variants"

function Button({
  className,
  variant = "default",
  size = "default",
  static: isStatic = false,
  ...props
}: ButtonPrimitive.Props & VariantProps<typeof buttonVariants> & { static?: boolean }) {
  return (
    <ButtonPrimitive
      data-slot="button"
      data-static={isStatic || undefined}
      className={cn(
        buttonVariants({ variant, size, className }),
        isStatic && "active:not-aria-[haspopup]:scale-100",
      )}
      {...props}
    />
  )
}

export { Button }
