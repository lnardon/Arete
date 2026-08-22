"use client"

import { useMemo } from "react"
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
} from "recharts"
import { useEntries, useProjects } from "@/hooks/use-pomodoro"
import { formatLocalDate } from "@/lib/date-utils"

const DAYS = 14
const UNCATEGORIZED_KEY = "uncategorized"

type ChartRow = { date: string; label: string; [seriesKey: string]: string | number }

function formatMinutesTooltip(value: number | undefined) {
  return [`${value ?? 0}m`, ""]
}

export function PomodoroChart() {
  const end = formatLocalDate(new Date())
  const start = useMemo(() => {
    const d = new Date()
    d.setDate(d.getDate() - (DAYS - 1))
    return formatLocalDate(d)
  }, [])

  const { data: entries = [] } = useEntries(start, end)
  const { data: projects = [] } = useProjects()

  const data = useMemo(() => {
    const days: ChartRow[] = []
    const cursor = new Date()
    cursor.setDate(cursor.getDate() - (DAYS - 1))
    for (let i = 0; i < DAYS; i++) {
      const key = formatLocalDate(cursor)
      const row: ChartRow = {
        date: key,
        label: cursor.toLocaleDateString([], { month: "short", day: "numeric" }),
      }
      for (const project of projects) row[project.id] = 0
      row[UNCATEGORIZED_KEY] = 0
      days.push(row)
      cursor.setDate(cursor.getDate() + 1)
    }

    const byDate = new Map(days.map((d) => [d.date, d]))
    for (const entry of entries) {
      if (!entry.endedAt) continue
      const row = byDate.get(entry.localDate)
      if (!row) continue
      const minutes = Math.round((new Date(entry.endedAt).getTime() - new Date(entry.startedAt).getTime()) / 60_000)
      const key = entry.projectId ?? UNCATEGORIZED_KEY
      row[key] = (row[key] as number) + minutes
    }

    return days
  }, [entries, projects])

  const hasData = entries.some((e) => e.endedAt)

  return (
    <div className="border border-border rounded-lg bg-card px-5 py-6">
      <p className="label-section mb-6">Focus Time (14 days)</p>
      {!hasData ? (
        <p className="text-sm text-muted-foreground py-8 text-center">
          Complete a timer to see your focus trend
        </p>
      ) : (
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={data} barCategoryGap="20%" margin={{ top: 4, right: 0, left: -20, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
            <XAxis
              dataKey="label"
              tick={{ fontSize: 10, fill: "var(--muted-foreground)" }}
              tickLine={false}
              axisLine={{ stroke: "var(--border)" }}
              interval={1}
            />
            <YAxis
              width={36}
              tick={{ fontSize: 11, fill: "var(--muted-foreground)" }}
              tickLine={false}
              axisLine={false}
              allowDecimals={false}
            />
            <Tooltip
              cursor={{ fill: "var(--muted)" }}
              contentStyle={{
                backgroundColor: "var(--popover)",
                border: "1px solid var(--border)",
                borderRadius: 8,
                fontSize: 12,
              }}
              formatter={formatMinutesTooltip}
            />
            {projects.map((project) => (
              <Bar key={project.id} dataKey={project.id} name={project.name} stackId="a" fill={project.color} />
            ))}
            <Bar dataKey={UNCATEGORIZED_KEY} name="Uncategorized" stackId="a" fill="var(--border)" radius={[3, 3, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      )}
    </div>
  )
}
