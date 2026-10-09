"use client"

import { Repeat } from "lucide-react"
import type { CalendarEvent } from "@/lib/types"
import { timeRangeLabel } from "@/lib/calendar-layout"
import { cn } from "@/lib/utils"

// Past events drop their color so the upcoming ones stand out. The fill stays
// solid because overlapping events stack exactly on top of each other.
const PAST_CLASSES = "bg-muted text-muted-foreground border border-border"

interface CalendarEventItemProps {
  event: CalendarEvent
  top: number
  height: number
  past: boolean
  onClick: (event: CalendarEvent) => void
}

export function CalendarEventItem({ event, top, height, past, onClick }: CalendarEventItemProps) {
  const compact = height < 36

  return (
    <button
      type="button"
      onClick={(e) => {
        e.stopPropagation()
        onClick(event)
      }}
      className={cn(
        "absolute left-1 right-1 rounded-md px-2 py-1 text-left text-xs overflow-hidden transition-transform hover:scale-[1.01] hover:z-10",
        past ? PAST_CLASSES : "text-white shadow-sm"
      )}
      style={{ top, height, backgroundColor: past ? undefined : event.color }}
    >
      <span className="font-medium truncate flex items-center gap-1">
        {event.recurringEventId && <Repeat className="size-3 shrink-0" aria-label="Repeats" />}
        <span className="truncate">{event.title}</span>
      </span>
      {!compact && <span className="block opacity-90 truncate">{timeRangeLabel(event)}</span>}
    </button>
  )
}

interface CalendarAllDayChipProps {
  event: CalendarEvent
  past: boolean
  className?: string
  onClick: (event: CalendarEvent) => void
}

export function CalendarAllDayChip({ event, past, className, onClick }: CalendarAllDayChipProps) {
  return (
    <button
      type="button"
      onClick={() => onClick(event)}
      className={cn(
        "rounded-md text-left flex items-center gap-1 min-w-0",
        past ? PAST_CLASSES : "text-white",
        className
      )}
      style={{ backgroundColor: past ? undefined : event.color }}
    >
      {event.recurringEventId && <Repeat className="size-3 shrink-0" aria-label="Repeats" />}
      <span className="truncate">{event.title}</span>
    </button>
  )
}
