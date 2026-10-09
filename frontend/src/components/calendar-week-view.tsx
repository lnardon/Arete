"use client"

import { useRef, type MouseEvent } from "react"
import type { CalendarEvent } from "@/lib/types"
import { HOUR_HEIGHT, DAY_HEIGHT, isPastOnDay, layoutEventForDay } from "@/lib/calendar-layout"
import { formatHourLabel, formatWeekdayLabel, addDays, getStartOfDay, isSameDay } from "@/lib/date-utils"
import { CalendarAllDayChip, CalendarEventItem } from "@/components/calendar-event-item"
import { CalendarNowIndicator } from "@/components/calendar-now-indicator"

const HOURS = Array.from({ length: 24 }, (_, i) => i)

interface CalendarWeekViewProps {
  weekStart: Date
  events: CalendarEvent[]
  now: Date
  scrollToNowSignal: number
  onSlotClick: (start: Date) => void
  onEventClick: (event: CalendarEvent) => void
}

export function CalendarWeekView({
  weekStart,
  events,
  now,
  scrollToNowSignal,
  onSlotClick,
  onEventClick,
}: CalendarWeekViewProps) {
  const days = Array.from({ length: 7 }, (_, i) => addDays(weekStart, i))
  const allDayEvents = events.filter((e) => e.allDay)

  return (
    <div className="flex flex-col">
      <div className="flex border-b border-border pb-2 mb-2">
        <div className="w-14 shrink-0" />
        {days.map((day) => (
          <div key={day.toISOString()} className="flex-1 text-center">
            <p className="text-[11px] text-muted-foreground">{formatWeekdayLabel(day)}</p>
            <p className={`text-sm font-medium ${isSameDay(day, now) ? "text-primary" : "text-foreground"}`}>
              {day.getDate()}
            </p>
          </div>
        ))}
      </div>

      {allDayEvents.length > 0 && (
        <div className="flex border-b border-border pb-2 mb-2">
          <div className="w-14 shrink-0" />
          {days.map((day) => (
            <div key={day.toISOString()} className="flex-1 flex flex-col gap-1 px-0.5">
              {allDayEvents
                .filter((e) => layoutEventForDay(e, day) !== null)
                .map((event) => (
                  <CalendarAllDayChip
                    key={event.id}
                    event={event}
                    past={isPastOnDay(event, day, now)}
                    onClick={onEventClick}
                    className="px-1.5 py-0.5 text-[11px]"
                  />
                ))}
            </div>
          ))}
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
        <div className="flex flex-1">
          {days.map((day) => (
            <DayColumn
              key={day.toISOString()}
              day={day}
              events={events}
              now={now}
              scrollToNowSignal={scrollToNowSignal}
              onSlotClick={onSlotClick}
              onEventClick={onEventClick}
            />
          ))}
        </div>
      </div>
    </div>
  )
}

interface DayColumnProps {
  day: Date
  events: CalendarEvent[]
  now: Date
  scrollToNowSignal: number
  onSlotClick: (start: Date) => void
  onEventClick: (event: CalendarEvent) => void
}

function DayColumn({ day, events, now, scrollToNowSignal, onSlotClick, onEventClick }: DayColumnProps) {
  const gridRef = useRef<HTMLDivElement>(null)
  const positions = events
    .filter((e) => !e.allDay)
    .map((e) => layoutEventForDay(e, day))
    .filter((p): p is NonNullable<typeof p> => p !== null)

  function handleClick(e: MouseEvent<HTMLDivElement>) {
    const rect = gridRef.current?.getBoundingClientRect()
    if (!rect) return
    const offsetY = e.clientY - rect.top
    const hour = Math.max(0, Math.min(23, Math.floor(offsetY / HOUR_HEIGHT)))
    const start = getStartOfDay(day)
    start.setHours(hour)
    onSlotClick(start)
  }

  return (
    <div
      ref={gridRef}
      onClick={handleClick}
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
      {positions.map(({ event, top, height }) => (
        <CalendarEventItem
          key={event.id}
          event={event}
          top={top}
          height={height}
          past={isPastOnDay(event, day, now)}
          onClick={onEventClick}
        />
      ))}
      {isSameDay(day, now) && <CalendarNowIndicator now={now} scrollSignal={scrollToNowSignal} />}
    </div>
  )
}
