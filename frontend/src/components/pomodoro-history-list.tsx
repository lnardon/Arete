"use client"

import { useMemo } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useEntries, useProjects } from "@/hooks/use-pomodoro"
import { formatLocalDate } from "@/lib/date-utils"
import type { PomodoroEntry } from "@/lib/types"

const HISTORY_DAYS = 14

function formatDuration(startedAt: string, endedAt: string): string {
  const minutes = Math.round((new Date(endedAt).getTime() - new Date(startedAt).getTime()) / 60_000)
  if (minutes < 60) return `${minutes}m`
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`
}

function rangeStart(days: number): string {
  const d = new Date()
  d.setDate(d.getDate() - days)
  return formatLocalDate(d)
}

export function PomodoroHistoryList() {
  const start = rangeStart(HISTORY_DAYS - 1)
  const end = formatLocalDate(new Date())
  const { data: entries = [] } = useEntries(start, end)
  const { data: projects = [] } = useProjects()

  const grouped = useMemo(() => {
    const completed = entries.filter((e) => e.endedAt !== null)
    const groups = new Map<string, PomodoroEntry[]>()
    for (const entry of completed) {
      const list = groups.get(entry.localDate) ?? []
      list.push(entry)
      groups.set(entry.localDate, list)
    }
    return Array.from(groups.entries()).sort((a, b) => (a[0] < b[0] ? 1 : -1))
  }, [entries])

  function projectFor(id: string | null) {
    return projects.find((p) => p.id === id)
  }

  return (
    <Card className="border-foreground/20">
      <CardHeader>
        <CardTitle className="font-display text-lg font-semibold tracking-tight">
          History
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4 max-h-96 overflow-y-auto">
        {grouped.length === 0 ? (
          <p className="text-sm text-muted-foreground py-6 text-center">
            No sessions logged in the last two weeks
          </p>
        ) : (
          grouped.map(([date, dayEntries]) => (
            <div key={date}>
              <p className="label-section mb-2">{date}</p>
              <div className="flex flex-col gap-1.5">
                {dayEntries.map((entry) => {
                  const project = projectFor(entry.projectId)
                  return (
                    <div key={entry.id} className="flex items-center gap-2.5 text-sm">
                      <span
                        className="w-2 h-2 rounded-full shrink-0"
                        style={{ backgroundColor: project?.color ?? "var(--border)" }}
                      />
                      <span className="flex-1 truncate text-foreground/90">
                        {project?.name ?? "Uncategorized"}
                      </span>
                      <span className="text-muted-foreground text-xs tabular-nums">
                        {new Date(entry.startedAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
                        {"–"}
                        {new Date(entry.endedAt!).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
                      </span>
                      <span className="text-muted-foreground text-xs w-12 text-right tabular-nums">
                        {formatDuration(entry.startedAt, entry.endedAt!)}
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
          ))
        )}
      </CardContent>
    </Card>
  )
}
