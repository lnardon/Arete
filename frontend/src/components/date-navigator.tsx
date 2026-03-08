"use client"

import { ChevronLeft, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"

interface DateNavigatorProps {
  currentDate: Date
  onPrevious: () => void
  onNext: () => void
  onToday: () => void
}

function formatDisplayDate(date: Date): string {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const compare = new Date(date)
  compare.setHours(0, 0, 0, 0)

  const diffTime = compare.getTime() - today.getTime()
  const diffDays = Math.round(diffTime / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return "Today"
  if (diffDays === -1) return "Yesterday"
  if (diffDays === 1) return "Tomorrow"

  return date.toLocaleDateString("en-US", {
    weekday: "long",
    month: "long",
    day: "numeric",
  })
}

function formatSubDate(date: Date): string {
  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  })
}

function isToday(date: Date): boolean {
  const today = new Date()
  return (
    date.getDate() === today.getDate() &&
    date.getMonth() === today.getMonth() &&
    date.getFullYear() === today.getFullYear()
  )
}

export function DateNavigator({
  currentDate,
  onPrevious,
  onNext,
  onToday,
}: DateNavigatorProps) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <h2 className="text-2xl font-display font-semibold tracking-tight text-foreground">
          {formatDisplayDate(currentDate)}
        </h2>
        <p className="text-xs text-muted-foreground mt-0.5">
          {formatSubDate(currentDate)}
        </p>
      </div>
      <div className="flex items-center gap-1">
        {!isToday(currentDate) && (
          <Button
            variant="outline"
            size="sm"
            onClick={onToday}
            className="h-8 px-3 text-sm rounded-lg border-border hover:bg-muted mr-1"
          >
            Today
          </Button>
        )}
        <Button
          variant="outline"
          size="icon"
          onClick={onPrevious}
          className="h-8 w-8 rounded-lg border-border hover:bg-muted"
        >
          <ChevronLeft className="w-4 h-4" />
          <span className="sr-only">Previous day</span>
        </Button>
        <Button
          variant="outline"
          size="icon"
          onClick={onNext}
          className="h-8 w-8 rounded-lg border-border hover:bg-muted"
        >
          <ChevronRight className="w-4 h-4" />
          <span className="sr-only">Next day</span>
        </Button>
      </div>
    </div>
  )
}
