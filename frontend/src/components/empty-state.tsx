"use client"

import { Plus } from "lucide-react"
import { Button } from "@/components/ui/button"

interface EmptyStateProps {
  onAdd: () => void
}

export function EmptyState({ onAdd }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-20 px-6 bg-card rounded-2xl border border-dashed border-border">
      <div className="flex items-center justify-center w-14 h-14 rounded-2xl bg-muted mb-6">
        <svg
          viewBox="0 0 24 24"
          fill="none"
          className="w-6 h-6 text-muted-foreground"
          stroke="currentColor"
          strokeWidth="1"
        >
          <path d="M12 2L4 6v6c0 5.25 3.4 10.15 8 11.25C16.6 22.15 20 17.25 20 12V6l-8-4z" />
        </svg>
      </div>
      <h3 className="font-display text-lg font-semibold tracking-tight text-foreground mb-1">
        No Habits Yet
      </h3>
      <p className="text-sm text-muted-foreground mb-6 text-center max-w-xs">
        Begin your journey of discipline by defining your first daily practice
      </p>
      <Button onClick={onAdd} className="rounded-xl">
        <Plus className="w-4 h-4 mr-2" />
        Create First Habit
      </Button>
    </div>
  )
}
