"use client"

import { Pencil, Trash2 } from "lucide-react"
import { Checkbox } from "@/components/ui/checkbox"
import { Button } from "@/components/ui/button"
import type { Habit, HabitCompletion } from "@/lib/types"
import { useToggleCompletion } from "@/hooks/use-completions"
import { cn } from "@/lib/utils"

interface HabitItemProps {
  habit: Habit
  date: string
  completions: HabitCompletion[]
  onEdit: (habit: Habit) => void
  onDelete: (habit: Habit) => void
}

export function HabitItem({ habit, date, completions, onEdit, onDelete }: HabitItemProps) {
  const completed = completions.some((c) => c.habitId === habit.id && c.date === date)
  const toggleMutation = useToggleCompletion()

  return (
    <div
      className={cn(
        "group flex items-center gap-4 px-4 py-3.5 bg-card border border-border rounded-xl transition-all",
        completed && "bg-muted/60"
      )}
    >
      <Checkbox
        checked={completed}
        onCheckedChange={() => toggleMutation.mutate({ habitId: habit.id, date })}
        className="h-5 w-5 border-foreground/30 data-[state=checked]:bg-foreground data-[state=checked]:text-background"
      />
      <span
        className={cn(
          "flex-1 text-sm tracking-wide transition-all",
          completed && "line-through text-muted-foreground"
        )}
      >
        {habit.name}
      </span>
      <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => onEdit(habit)}
          className="h-7 w-7 text-muted-foreground hover:text-foreground hover:bg-transparent"
        >
          <Pencil className="w-3.5 h-3.5" />
          <span className="sr-only">Edit habit</span>
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => onDelete(habit)}
          className="h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-transparent"
        >
          <Trash2 className="w-3.5 h-3.5" />
          <span className="sr-only">Delete habit</span>
        </Button>
      </div>
    </div>
  )
}
