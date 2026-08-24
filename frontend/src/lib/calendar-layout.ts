import type { CalendarEvent } from '@/lib/types'
import { addDays, getStartOfDay } from '@/lib/date-utils'

export const HOUR_HEIGHT = 48 // px per hour row, shared by the day and week grids
export const DAY_HEIGHT = HOUR_HEIGHT * 24
const MIN_EVENT_HEIGHT = 20 // px, so short events stay clickable/legible

export interface PositionedEvent {
  event: CalendarEvent
  top: number
  height: number
}

// Clips an event's start/end to the given calendar day and returns its pixel
// top/height within a DAY_HEIGHT-tall column, or null if it doesn't fall on
// this day at all.
export function layoutEventForDay(event: CalendarEvent, day: Date): PositionedEvent | null {
  const dayStart = getStartOfDay(day)
  const dayEnd = addDays(dayStart, 1)

  const eventStart = new Date(event.startAt)
  const eventEnd = new Date(event.endAt)
  if (eventEnd <= dayStart || eventStart >= dayEnd) return null

  const clippedStart = eventStart < dayStart ? dayStart : eventStart
  const clippedEnd = eventEnd > dayEnd ? dayEnd : eventEnd

  const top = ((clippedStart.getTime() - dayStart.getTime()) / (60 * 60 * 1000)) * HOUR_HEIGHT
  const height = Math.max(
    MIN_EVENT_HEIGHT,
    ((clippedEnd.getTime() - clippedStart.getTime()) / (60 * 60 * 1000)) * HOUR_HEIGHT
  )

  return { event, top, height }
}

export function timeRangeLabel(event: CalendarEvent): string {
  const fmt = (d: Date) => d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
  return `${fmt(new Date(event.startAt))} – ${fmt(new Date(event.endAt))}`
}
