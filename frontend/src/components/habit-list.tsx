"use client"

import { useState } from "react"
import { Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { DateNavigator } from "@/components/date-navigator"
import { HabitItem } from "@/components/habit-item"
import { HabitDialog } from "@/components/habit-dialog"
import { DeleteDialog } from "@/components/delete-dialog"
import { DailyProgress } from "@/components/daily-progress"
import { EmptyState } from "@/components/empty-state"
import { useHabits, useCreateHabit, useUpdateHabit, useDeleteHabit } from "@/hooks/use-habits"
import { useCompletionsForDate } from "@/hooks/use-completions"
import type { Habit } from "@/lib/types"

function formatDate(date: Date): string {
  return date.toISOString().split("T")[0]
}

export function HabitList() {
  const [currentDate, setCurrentDate] = useState(new Date())
  const [habitDialogOpen, setHabitDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingHabit, setEditingHabit] = useState<Habit | null>(null)
  const [deletingHabit, setDeletingHabit] = useState<Habit | null>(null)

  const dateStr = formatDate(currentDate)
  const { data: habits = [], isLoading: habitsLoading } = useHabits()
  const { data: completions = [] } = useCompletionsForDate(dateStr)

  const createHabit = useCreateHabit()
  const updateHabit = useUpdateHabit()
  const deleteHabit = useDeleteHabit()

  const activeHabits = habits.filter((h) => h.createdAt.split("T")[0] <= dateStr)
  const activeIds = new Set(activeHabits.map((h) => h.id))
  const completed = completions.filter((c) => c.date === dateStr && activeIds.has(c.habitId)).length
  const total = activeHabits.length

  function handlePrevious() {
    setCurrentDate((d) => {
      const prev = new Date(d)
      prev.setDate(prev.getDate() - 1)
      return prev
    })
  }

  function handleNext() {
    setCurrentDate((d) => {
      const next = new Date(d)
      next.setDate(next.getDate() + 1)
      return next
    })
  }

  function handleToday() {
    setCurrentDate(new Date())
  }

  function handleOpenCreate() {
    setEditingHabit(null)
    setHabitDialogOpen(true)
  }

  function handleOpenEdit(habit: Habit) {
    setEditingHabit(habit)
    setHabitDialogOpen(true)
  }

  function handleOpenDelete(habit: Habit) {
    setDeletingHabit(habit)
    setDeleteDialogOpen(true)
  }

  function handleSave(name: string) {
    if (editingHabit) {
      updateHabit.mutate({ id: editingHabit.id, name })
    } else {
      createHabit.mutate(name)
    }
  }

  function handleConfirmDelete() {
    if (deletingHabit) {
      deleteHabit.mutate(deletingHabit.id)
    }
  }

  if (habitsLoading) {
    return (
      <main className="flex-1 overflow-y-auto bg-background">
        <div className="max-w-2xl mx-auto px-5 py-8 md:px-8 md:py-12">
          <div className="flex flex-col gap-2">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-14 rounded-xl bg-muted animate-pulse" />
            ))}
          </div>
        </div>
      </main>
    )
  }

  return (
    <main className="flex-1 overflow-y-auto bg-background">
      <div className="max-w-2xl mx-auto px-5 py-8 md:px-8 md:py-12">
        <DateNavigator
          currentDate={currentDate}
          onPrevious={handlePrevious}
          onNext={handleNext}
          onToday={handleToday}
        />

        <div className="my-6 h-px bg-border" />

        {activeHabits.length > 0 && (
          <div className="mb-6">
            <DailyProgress completed={completed} total={total} />
          </div>
        )}

        {habits.length === 0 ? (
          <EmptyState onAdd={handleOpenCreate} />
        ) : activeHabits.length === 0 ? (
          <p className="text-sm text-muted-foreground py-8 text-center">
            No habits on this date
          </p>
        ) : (
          <>
            <div className="flex items-center justify-between mb-4">
              <p className="label-section">
                Habits
              </p>
              <Button
                variant="outline"
                size="sm"
                onClick={handleOpenCreate}
                className="h-8 text-sm gap-1.5 px-3 rounded-lg border-border hover:bg-muted"
              >
                <Plus className="w-3.5 h-3.5 mr-1.5" />
                Add Habit
              </Button>
            </div>
            <div className="flex flex-col gap-2">
              {activeHabits.map((habit) => (
                <HabitItem
                  key={habit.id}
                  habit={habit}
                  date={dateStr}
                  completions={completions}
                  onEdit={handleOpenEdit}
                  onDelete={handleOpenDelete}
                />
              ))}
            </div>
          </>
        )}

        <HabitDialog
          open={habitDialogOpen}
          onOpenChange={setHabitDialogOpen}
          habit={editingHabit}
          onSave={handleSave}
        />

        <DeleteDialog
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          habit={deletingHabit}
          onConfirm={handleConfirmDelete}
        />
      </div>
    </main>
  )
}
