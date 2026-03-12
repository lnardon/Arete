"use client"

import type { Habit, HabitCompletion } from "@/lib/types"
import { useToggleCompletion } from "@/hooks/use-completions"
import { ChecklistItem } from "@/components/checklist-item"

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
    <ChecklistItem
      id={habit.id}
      label={habit.name}
      checked={completed}
      onCheckedChange={() => toggleMutation.mutate({ habitId: habit.id, date })}
      onEdit={() => onEdit(habit)}
      onDelete={() => onDelete(habit)}
    />
  )
}
