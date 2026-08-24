"use client"

import { ChevronLeft, ChevronRight, Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { addDays, getStartOfWeek } from "@/lib/date-utils"

export type CalendarViewMode = "day" | "week"

interface CalendarToolbarProps {
  selectedDate: Date
  viewMode: CalendarViewMode
  onViewModeChange: (mode: CalendarViewMode) => void
  onPrevious: () => void
  onNext: () => void
  onToday: () => void
  onCreate: () => void
}

function formatRangeLabel(date: Date, viewMode: CalendarViewMode): string {
  if (viewMode === "day") {
    return date.toLocaleDateString("en-US", { weekday: "long", month: "long", day: "numeric", year: "numeric" })
  }

  const start = getStartOfWeek(date)
  const end = addDays(start, 6)
  const sameMonth = start.getMonth() === end.getMonth()
  const startLabel = start.toLocaleDateString("en-US", { month: "short", day: "numeric" })
  const endLabel = end.toLocaleDateString("en-US", {
    month: sameMonth ? undefined : "short",
    day: "numeric",
    year: "numeric",
  })
  return `${startLabel} – ${endLabel}`
}

export function CalendarToolbar({
  selectedDate,
  viewMode,
  onViewModeChange,
  onPrevious,
  onNext,
  onToday,
  onCreate,
}: CalendarToolbarProps) {
  return (
    <div className="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h2 className="text-2xl font-display font-semibold tracking-tight text-foreground">
          {formatRangeLabel(selectedDate, viewMode)}
        </h2>
      </div>
      <div className="flex items-center gap-2">
        <div className="flex items-center gap-1 rounded-lg border border-border p-0.5">
          <button
            type="button"
            onClick={() => onViewModeChange("day")}
            className={cn(
              "px-3 py-1 text-sm rounded-md transition-colors",
              viewMode === "day"
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            Day
          </button>
          <button
            type="button"
            onClick={() => onViewModeChange("week")}
            className={cn(
              "px-3 py-1 text-sm rounded-md transition-colors",
              viewMode === "week"
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            Week
          </button>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={onToday}
          className="h-8 px-3 text-sm rounded-lg border-border hover:bg-muted"
        >
          Today
        </Button>
        <Button
          variant="outline"
          size="icon"
          onClick={onPrevious}
          className="h-8 w-8 rounded-lg border-border hover:bg-muted"
        >
          <ChevronLeft className="w-4 h-4" />
          <span className="sr-only">Previous</span>
        </Button>
        <Button
          variant="outline"
          size="icon"
          onClick={onNext}
          className="h-8 w-8 rounded-lg border-border hover:bg-muted"
        >
          <ChevronRight className="w-4 h-4" />
          <span className="sr-only">Next</span>
        </Button>
        <Button
          size="sm"
          onClick={onCreate}
          className="h-8 gap-1.5 px-3 text-sm rounded-lg"
        >
          <Plus className="w-4 h-4" />
          New Event
        </Button>
      </div>
    </div>
  )
}
