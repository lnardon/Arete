import { Checkbox as CheckboxPrimitive } from "@base-ui/react/checkbox"

import { cn } from "@/lib/utils"
import { RiCheckLine } from "@remixicon/react"

function Checkbox({ className, ...props }: CheckboxPrimitive.Root.Props) {
  return (
    <CheckboxPrimitive.Root
      data-slot="checkbox"
      className={cn(
        "border-2 border-foreground/40 bg-transparent data-checked:bg-primary data-checked:text-primary-foreground data-checked:border-primary aria-invalid:border-destructive focus-visible:border-ring focus-visible:ring-ring/50 flex size-5 items-center justify-center rounded-sm border transition-colors group-has-disabled/field:opacity-50 focus-visible:ring-3 peer relative shrink-0 outline-none after:absolute after:-inset-x-3 after:-inset-y-2 disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      {...props}
    >
      <CheckboxPrimitive.Indicator
        data-slot="checkbox-indicator"
        className="[&>svg]:size-3.5 grid place-content-center text-current transition-none"
      >
        <RiCheckLine
        />
      </CheckboxPrimitive.Indicator>
    </CheckboxPrimitive.Root>
  )
}

export { Checkbox }
