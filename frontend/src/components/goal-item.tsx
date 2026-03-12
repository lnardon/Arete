import type { Goal } from "@/lib/types"
import { ChecklistItem } from "@/components/checklist-item"

interface GoalItemProps {
  goal: Goal
  onToggle: (id: string) => void
  onDelete: (id: string) => void
}

export function GoalItem({ goal, onToggle, onDelete }: GoalItemProps) {
  return (
    <ChecklistItem
      id={goal.id}
      label={goal.title}
      checked={goal.completed}
      onCheckedChange={() => onToggle(goal.id)}
      onDelete={() => onDelete(goal.id)}
    />
  )
}
