"use client"

import { useMemo } from "react"
import type { Habit, HabitCompletion } from "@/lib/types"
import { formatLocalDate } from "@/lib/date-utils"
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
} from "recharts"

const DAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]

export function WeeklyChart({ habits, completions }: { habits: Habit[], completions: HabitCompletion[] }) {
  const data = useMemo(() => {
    const today = new Date()
    today.setHours(12, 0, 0, 0)
    const dayOfWeek = today.getDay()
    const mondayOffset = dayOfWeek === 0 ? -6 : 1 - dayOfWeek
    const monday = new Date(today)
    monday.setDate(monday.getDate() + mondayOffset)

    return DAY_LABELS.map((label, idx) => {
      const d = new Date(monday)
      d.setDate(d.getDate() + idx)
      const key = formatLocalDate(d)
      const completed = completions.filter((c) => c.date === key).length
      const activeCount = habits.filter((h) => formatLocalDate(new Date(h.createdAt)) <= key).length
      const missed = Math.max(0, activeCount - completed)
      return { day: label, completed, missed }
    })
  }, [habits, completions])

  const maxValue = Math.max(
    1,
    ...data.map((d) => d.completed + d.missed)
  )

  return (
    <div className="border border-border rounded-lg bg-card px-5 py-6 height-fit">
      <p className="label-section mb-6">
        This Week
      </p>
      {habits.length === 0 ? (
        <p className="text-sm text-muted-foreground py-8 text-center">
          Add habits to see weekly data
        </p>
      ) : (
        <ResponsiveContainer width="100%" height={180}>
          <BarChart
            data={data}
            barCategoryGap="35%"
            margin={{ top: 4, right: 0, left: -20, bottom: 0 }}
          >
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="var(--border)"
              vertical={false}
            />
            <XAxis
              dataKey="day"
              tick={{
                fontSize: 11,
                fill: "var(--muted-foreground)",
              }}
              tickLine={false}
              axisLine={{ stroke: "var(--border)" }}
            />
            <YAxis
              domain={[0, maxValue || 1]}
              width={36}
              tick={{
                fontSize: 11,
                fill: "var(--muted-foreground)",
              }}
              tickLine={false}
              axisLine={false}
              allowDecimals={false}
              tickCount={Math.min(maxValue + 1, 5)}
            />
            <Tooltip
              cursor={{ fill: "var(--muted)" }}
              contentStyle={{
                backgroundColor: "var(--popover)",
                border: "1px solid var(--border)",
                borderRadius: 8,
                fontSize: 12,
              }}
              labelStyle={{
                fontFamily: "var(--font-sans)",
                fontWeight: 600,
                marginBottom: 4,
              }}
            />
            <Bar
              dataKey="completed"
              fill="var(--foreground)"
              name="Completed"
              stackId="a"
              radius={[0, 0, 0, 0]}
            />
            <Bar
              dataKey="missed"
              fill="var(--border)"
              name="Missed"
              stackId="a"
              radius={[3, 3, 0, 0]}
            />
          </BarChart>
        </ResponsiveContainer>
      )}
    </div>
  )
}
