"use client"

import { useMemo, useRef, useEffect, useState } from "react"
import type { Habit, HabitCompletion } from "@/lib/types"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"

function formatDateKey(d: Date): string {
  return d.toISOString().split("T")[0]
}

function formatDisplayDate(dateStr: string): string {
  const d = new Date(dateStr + "T12:00:00")
  return d.toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    year: "numeric",
  })
}

const MONTHS = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
]


interface DayData {
  date: string
  completed: number
  total: number
  ratio: number
}

function getIntensityClass(ratio: number, total: number): string {
  if (total === 0) return "bg-muted"
  if (ratio === 0) return "bg-muted"
  if (ratio <= 0.25) return "bg-foreground/15"
  if (ratio <= 0.5) return "bg-foreground/30"
  if (ratio <= 0.75) return "bg-foreground/55"
  return "bg-foreground/90"
}

export function HabitHeatmap({ habits, completions }: { habits: Habit[], completions: HabitCompletion[] }) {
  const [_, setHoveredDay] = useState<DayData | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollLeft = scrollRef.current.scrollWidth
    }
  }, [])

  const { weeks, monthLabels } = useMemo(() => {
    const today = new Date()
    today.setHours(12, 0, 0, 0)

    const startDate = new Date(today)
    startDate.setDate(startDate.getDate() - 364)

    const dayOfWeek = startDate.getDay()
    startDate.setDate(startDate.getDate() - dayOfWeek)

    const completionsByDate = new Map<string, number>()
    for (const c of completions) {
      completionsByDate.set(c.date, (completionsByDate.get(c.date) || 0) + 1)
    }

    const weeksArr: DayData[][] = []
    const monthLabelArr: { label: string; weekIndex: number }[] = []
    let currentWeek: DayData[] = []
    let lastMonth = -1
    let weekIndex = 0

    const cursor = new Date(startDate)
    while (cursor <= today) {
      const key = formatDateKey(cursor)
      const completed = completionsByDate.get(key) || 0
      const total = habits.filter((h) => h.createdAt.split("T")[0] <= key).length
      currentWeek.push({
        date: key,
        completed,
        total,
        ratio: total > 0 ? completed / total : 0,
      })

      if (cursor.getMonth() !== lastMonth) {
        lastMonth = cursor.getMonth()
        monthLabelArr.push({ label: MONTHS[lastMonth], weekIndex })
      }

      if (cursor.getDay() === 6 || cursor.getTime() === today.getTime()) {
        weeksArr.push(currentWeek)
        currentWeek = []
        weekIndex++
      }

      cursor.setDate(cursor.getDate() + 1)
    }

    if (currentWeek.length > 0) {
      weeksArr.push(currentWeek)
    }

    return { weeks: weeksArr, monthLabels: monthLabelArr }
  }, [habits, completions])

  return (
    <div className="border border-border rounded-lg bg-card px-5 py-6">
      <div className="flex items-baseline justify-between mb-6">
        <p className="label-section">
          365-Day Activity
        </p>
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-muted-foreground">
            Less
          </span>
          <div className="flex gap-0.5">
            <div className="w-2.5 h-2.5 bg-muted" />
            <div className="w-2.5 h-2.5 bg-foreground/15" />
            <div className="w-2.5 h-2.5 bg-foreground/30" />
            <div className="w-2.5 h-2.5 bg-foreground/55" />
            <div className="w-2.5 h-2.5 bg-foreground/90" />
          </div>
          <span className="text-[10px] text-muted-foreground">
            More
          </span>
        </div>
      </div>

      <div ref={scrollRef} className="overflow-x-auto heatmap-scroll pb-4">
        <div className="min-w-[720px]">
          <div className="flex gap-0.5 mb-1 ml-8">
            {monthLabels.map((m, i) => {
              const nextWeek = monthLabels[i + 1]?.weekIndex ?? weeks.length
              const span = nextWeek - m.weekIndex
              return (
                <div
                  key={`${m.label}-${m.weekIndex}`}
                  className="text-[10px] tracking-wider uppercase text-muted-foreground"
                  style={{
                    width: `${span * 13}px`,
                    flexShrink: 0,
                  }}
                >
                  {span >= 3 ? m.label : ""}
                </div>
              )
            })}
          </div>

          <div className="flex gap-0">
            <div className="flex flex-col gap-0.5 mr-1.5 pt-0">
              {[0, 1, 2, 3, 4, 5, 6].map((dayIdx) => (
                <div
                  key={dayIdx}
                  className="h-[11px] flex items-center justify-end"
                >
                  <span className="text-[9px] tracking-wider uppercase text-muted-foreground w-6 text-right">
                    {dayIdx === 0 ? "Sun" : dayIdx === 3 ? "Wed" : dayIdx === 6 ? "Sat" : ""}
                  </span>
                </div>
              ))}
            </div>

            <TooltipProvider delay={1000}>
              <div className="flex gap-0.5">
                {weeks.map((week, wIdx) => (
                  <div key={wIdx} className="flex flex-col gap-0.5">
                    {week.map((day) => (
                      <Tooltip key={day.date}>
                        <TooltipTrigger>
                          <div
                            className={`w-[11px] h-[11px] ${getIntensityClass(day.ratio, day.total)} transition-all duration-150 hover:ring-1 hover:ring-foreground/40 cursor-default`}
                            onMouseEnter={() => setHoveredDay(day)}
                            onMouseLeave={() => setHoveredDay(null)}
                          />
                        </TooltipTrigger>
                        <TooltipContent
                          side="top"
                          className="border border-border bg-popover text-popover-foreground p-2"
                        >
                          <p className="text-xs font-display font-medium">
                            {formatDisplayDate(day.date)}
                          </p>
                          <p className="text-[11px] text-muted-foreground mt-0.5">
                            {day.total === 0
                              ? "No habits tracked"
                              : `${day.completed} completed, ${day.total - day.completed} remaining`}
                          </p>
                        </TooltipContent>
                      </Tooltip>
                    ))}
                    {Array.from({ length: 7 - week.length }).map((_, pIdx) => (
                      <div key={`pad-${pIdx}`} className="w-[11px] h-[11px]" />
                    ))}
                  </div>
                ))}
              </div>
            </TooltipProvider>
          </div>
        </div>
      </div>
    </div>
  )
}
