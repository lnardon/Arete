"use client"

import { useRef, type MouseEvent } from "react"
import type { CalendarEvent } from "@/lib/types"
import { HOUR_HEIGHT, DAY_HEIGHT, layoutEventForDay } from "@/lib/calendar-layout"
import { formatHourLabel, getStartOfDay } from "@/lib/date-utils"
import { CalendarEventItem } from "@/components/calendar-event-item"
import { Repeat } from "lucide-react"

const HOURS = Array.from({ length: 24 }, (_, i) => i)

interface CalendarDayViewProps {
  date: Date
  events: CalendarEvent[]
  onSlotClick: (start: Date) => void
  onEventClick: (event: CalendarEvent) => void
}

export function CalendarDayView({ date, events, onSlotClick, onEventClick }: CalendarDayViewProps) {
  const gridRef = useRef<HTMLDivElement>(null)

  const allDayEvents = events.filter((e) => e.allDay)
  const timedPositions = events
    .filter((e) => !e.allDay)
    .map((e) => layoutEventForDay(e, date))
    .filter((p): p is NonNullable<typeof p> => p !== null)

  function handleGridClick(e: MouseEvent<HTMLDivElement>) {
    const rect = gridRef.current?.getBoundingClientRect()
    if (!rect) return
    const offsetY = e.clientY - rect.top
    const hour = Math.max(0, Math.min(23, Math.floor(offsetY / HOUR_HEIGHT)))
    const start = getStartOfDay(date)
    start.setHours(hour)
    onSlotClick(start)
  }

  return (
    <div className="flex flex-col">
      {allDayEvents.length > 0 && (
        <div className="flex border-b border-border pb-2 mb-2">
          <div className="w-14 shrink-0" />
          <div className="flex-1 flex flex-col gap-1">
            {allDayEvents.map((event) => (
              <button
                key={event.id}
                type="button"
                onClick={() => onEventClick(event)}
                className="rounded-md px-2 py-1 text-left text-xs text-white flex items-center gap-1 min-w-0"
                style={{ backgroundColor: event.color }}
              >
                {event.recurringEventId && <Repeat className="size-3 shrink-0" aria-label="Repeats" />}
                <span className="truncate">{event.title}</span>
              </button>
            ))}
          </div>
        </div>
      )}
      <div className="flex">
        <div className="w-14 shrink-0 flex flex-col">
          {HOURS.map((hour) => (
            <div
              key={hour}
              style={{ height: HOUR_HEIGHT }}
              className="text-[11px] text-muted-foreground text-right pr-2 -translate-y-2"
            >
              {hour === 0 ? "" : formatHourLabel(hour)}
            </div>
          ))}
        </div>
        <div
          ref={gridRef}
          onClick={handleGridClick}
          className="relative flex-1 border-l border-border cursor-pointer"
          style={{ height: DAY_HEIGHT }}
        >
          {HOURS.map((hour) => (
            <div
              key={hour}
              className="absolute left-0 right-0 border-t border-border/60"
              style={{ top: hour * HOUR_HEIGHT }}
            />
          ))}
          {timedPositions.map(({ event, top, height }) => (
            <CalendarEventItem key={event.id} event={event} top={top} height={height} onClick={onEventClick} />
          ))}
        </div>
      </div>
    </div>
  )
}
