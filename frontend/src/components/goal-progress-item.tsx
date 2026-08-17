import { Pencil, Plus, Trash2 } from "lucide-react"
import { Progress } from "@/components/ui/progress"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import type { Goal } from "@/lib/types"

interface GoalProgressItemProps {
  goal: Goal
  onAddProgress: (id: string) => void
  onDelete: (id: string) => void
  onEdit: (goal: Goal) => void
  pending?: boolean
}

export function GoalProgressItem({ goal, onAddProgress, onDelete, onEdit, pending }: GoalProgressItemProps) {
  const target = goal.targetValue ?? 0
  const percent = target > 0 ? Math.min(100, (goal.currentValue / target) * 100) : 0

  return (
    <div
      className={cn(
        "group flex items-center gap-4 px-4 py-3.5 bg-card border border-border rounded-xl transition-all",
        goal.completed && "bg-muted/60"
      )}
    >
      <div className="flex-1 min-w-0 flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <span
            className={cn(
              "text-sm tracking-wide truncate",
              goal.completed && "text-muted-foreground"
            )}
          >
            {goal.title}
          </span>
          <span className="text-xs text-muted-foreground tabular-nums shrink-0">
            {goal.currentValue} / {target}
          </span>
        </div>
        <Progress value={percent} />
      </div>
      <div className="flex items-center gap-0.5 shrink-0">
        <Button
          variant="outline"
          size="icon-sm"
          onClick={() => onAddProgress(goal.id)}
          disabled={pending}
          className="border-foreground/20"
          aria-label="Add progress"
        >
          <Plus className="w-3.5 h-3.5" />
        </Button>
        <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onEdit(goal)}
            className="h-7 w-7 text-muted-foreground hover:text-foreground hover:bg-transparent"
          >
            <Pencil className="w-3.5 h-3.5" />
            <span className="sr-only">Edit</span>
          </Button>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onDelete(goal.id)}
            className="h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-transparent"
          >
            <Trash2 className="w-3.5 h-3.5" />
            <span className="sr-only">Delete</span>
          </Button>
        </div>
      </div>
    </div>
  )
}
