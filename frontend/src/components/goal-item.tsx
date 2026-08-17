import type { Goal } from "@/lib/types"
import { ChecklistItem } from "@/components/checklist-item"
import { GoalProgressItem } from "@/components/goal-progress-item"

interface GoalItemProps {
  goal: Goal
  onToggle: (id: string) => void
  onAddProgress: (id: string) => void
  onDelete: (id: string) => void
  onEdit: (goal: Goal) => void
  progressPending?: boolean
}

export function GoalItem({ goal, onToggle, onAddProgress, onDelete, onEdit, progressPending }: GoalItemProps) {
  if (goal.goalType === "numeric") {
    return (
      <GoalProgressItem
        goal={goal}
        onAddProgress={onAddProgress}
        onDelete={onDelete}
        onEdit={onEdit}
        pending={progressPending}
      />
    )
  }

  return (
    <ChecklistItem
      id={goal.id}
      label={goal.title}
      checked={goal.completed}
      onCheckedChange={() => onToggle(goal.id)}
      onEdit={() => onEdit(goal)}
      onDelete={() => onDelete(goal.id)}
    />
  )
}
