"use client"

import { useMemo } from "react"
import type { Habit, HabitCompletion } from "@/lib/types"

export function HabitBreakdown({ habits, completions }: { habits: Habit[], completions: HabitCompletion[] }) {
  const breakdown = useMemo(() => {
    const today = new Date()
    today.setHours(12, 0, 0, 0)
    const thirtyDaysAgo = new Date(today)
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 29)

    return habits.map((habit) => {
      const habitCompletions = completions.filter((c) => {
        if (c.habitId !== habit.id) return false
        return c.date >= thirtyDaysAgo.toISOString().split("T")[0]
      })
      const count = habitCompletions.length
      const rate = Math.round((count / 30) * 100)
      return { ...habit, count, rate }
    })
  }, [habits, completions])

  if (habits.length === 0) {
    return null
  }

  return (
    <div className="border border-border rounded-lg bg-card px-5 py-6">
      <p className="label-section mb-5">
        30-Day Completion Rate
      </p>
      <div className="flex flex-col gap-4">
        {breakdown.map((item) => (
          <div key={item.id} className="flex flex-col gap-2">
            <div className="flex items-baseline justify-between">
              <p className="text-sm tracking-wide text-foreground">
                {item.name}
              </p>
              <p className="font-display text-sm font-semibold text-foreground">
                {item.rate}
                <span className="text-muted-foreground text-xs">%</span>
              </p>
            </div>
            <div className="h-1.5 rounded-full bg-muted overflow-hidden">
              <div
                className="h-full rounded-full bg-foreground transition-all duration-500 ease-out"
                style={{ width: `${item.rate}%` }}
              />
            </div>
            <p className="text-[10px] text-muted-foreground">
              {item.count} of 30 days
            </p>
          </div>
        ))}
      </div>
    </div>
  )
}
