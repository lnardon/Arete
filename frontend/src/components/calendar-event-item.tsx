"use client"

import { Repeat } from "lucide-react"
import type { CalendarEvent } from "@/lib/types"
import { timeRangeLabel } from "@/lib/calendar-layout"

interface CalendarEventItemProps {
  event: CalendarEvent
  top: number
  height: number
  onClick: (event: CalendarEvent) => void
}

export function CalendarEventItem({ event, top, height, onClick }: CalendarEventItemProps) {
  const compact = height < 36

  return (
    <button
      type="button"
      onClick={(e) => {
        e.stopPropagation()
        onClick(event)
      }}
      className="absolute left-1 right-1 rounded-md px-2 py-1 text-left text-xs text-white shadow-sm overflow-hidden transition-transform hover:scale-[1.01] hover:z-10"
      style={{ top, height, backgroundColor: event.color }}
    >
      <span className="font-medium truncate flex items-center gap-1">
        {event.recurringEventId && <Repeat className="size-3 shrink-0" aria-label="Repeats" />}
        <span className="truncate">{event.title}</span>
      </span>
      {!compact && <span className="block opacity-90 truncate">{timeRangeLabel(event)}</span>}
    </button>
  )
}
