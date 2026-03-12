import { Pencil, Trash2 } from "lucide-react"
import { Checkbox } from "@/components/ui/checkbox"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

interface ChecklistItemProps {
  id: string
  label: string
  checked: boolean
  onCheckedChange: () => void
  onEdit?: () => void
  onDelete: () => void
}

export function ChecklistItem({
  id,
  label,
  checked,
  onCheckedChange,
  onEdit,
  onDelete,
}: ChecklistItemProps) {
  return (
    <div
      className={cn(
        "group flex items-center gap-4 px-4 py-3.5 bg-card border border-border rounded-xl transition-all",
        checked && "bg-muted/60"
      )}
    >
      <Checkbox
        id={id}
        checked={checked}
        onCheckedChange={onCheckedChange}
        className="h-5 w-5 shrink-0 border-foreground/30 data-[state=checked]:bg-foreground data-[state=checked]:text-background"
      />
      <label
        htmlFor={id}
        className={cn(
          "flex-1 text-sm tracking-wide cursor-pointer select-none transition-all",
          checked && "line-through text-muted-foreground"
        )}
      >
        {label}
      </label>
      <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
        {onEdit && (
          <Button
            variant="ghost"
            size="icon"
            onClick={onEdit}
            className="h-7 w-7 text-muted-foreground hover:text-foreground hover:bg-transparent"
          >
            <Pencil className="w-3.5 h-3.5" />
            <span className="sr-only">Edit</span>
          </Button>
        )}
        <Button
          variant="ghost"
          size="icon"
          onClick={onDelete}
          className="h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-transparent"
        >
          <Trash2 className="w-3.5 h-3.5" />
          <span className="sr-only">Delete</span>
        </Button>
      </div>
    </div>
  )
}
