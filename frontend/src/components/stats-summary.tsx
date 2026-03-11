"use client"

import { useMemo } from "react"
import type { Habit, HabitCompletion } from "@/lib/types"
import { formatLocalDate } from "@/lib/date-utils"

interface StatCardProps {
  label: string
  value: string | number
  sublabel?: string
}

function StatCard({ label, value, sublabel }: StatCardProps) {
  return (
    <div className="border border-border rounded-lg bg-card px-5 py-5 flex flex-col gap-2 card-elevated">
      <p className="label-section">
        {label}
      </p>
      <p className="font-display text-3xl font-semibold text-foreground">{value}</p>
      {sublabel && (
        <p className="text-xs text-muted-foreground">
          {sublabel}
        </p>
      )}
    </div>
  )
}

interface StatsSummaryProps {
  habits: Habit[]
  completions: HabitCompletion[]
}

export function StatsSummary({ habits, completions }: StatsSummaryProps) {
  const stats = useMemo(() => {
    const totalCompletions = completions.length
    const totalHabits = habits.length

    const uniqueDates = new Set(completions.map((c) => c.date))
    const activeDays = uniqueDates.size

    const today = new Date()
    today.setHours(12, 0, 0, 0)

    let currentStreak = 0
    if (totalHabits > 0) {
      const cursor = new Date(today)
      while (true) {
        const key = formatLocalDate(cursor)
        const activeOnDay = habits.filter((h) => formatLocalDate(new Date(h.createdAt)) <= key).length
        const dayCompletions = completions.filter((c) => c.date === key).length
        if (activeOnDay > 0 && dayCompletions === activeOnDay) {
          currentStreak++
          cursor.setDate(cursor.getDate() - 1)
        } else {
          break
        }
      }
    }

    let bestStreak = 0
    if (totalHabits > 0) {
      const sortedDates = Array.from(uniqueDates).sort()
      let tempStreak = 0
      for (const dateStr of sortedDates) {
        const activeOnDay = habits.filter((h) => formatLocalDate(new Date(h.createdAt)) <= dateStr).length
        const dayCompletions = completions.filter(
          (c) => c.date === dateStr
        ).length
        if (activeOnDay > 0 && dayCompletions === activeOnDay) {
          tempStreak++
          bestStreak = Math.max(bestStreak, tempStreak)
        } else {
          tempStreak = 0
        }
      }
    }

    let completionRate = 0
    if (totalHabits > 0 && activeDays > 0) {
      const possibleCompletions = activeDays * totalHabits
      completionRate = Math.round((totalCompletions / possibleCompletions) * 100)
    }

    return {
      totalCompletions,
      totalHabits,
      activeDays,
      currentStreak,
      bestStreak,
      completionRate,
    }
  }, [habits, completions])

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
      <StatCard
        label="Active Habits"
        value={stats.totalHabits}
        sublabel="Currently tracking"
      />
      <StatCard
        label="Total Checks"
        value={stats.totalCompletions}
        sublabel={`Across ${stats.activeDays} days`}
      />
      <StatCard
        label="Current Streak"
        value={`${stats.currentStreak}d`}
        sublabel={stats.currentStreak > 0 ? "Keep going" : "Start today"}
      />
      <StatCard
        label="Best Streak"
        value={`${stats.bestStreak}d`}
        sublabel="Personal record"
      />
    </div>
  )
}
