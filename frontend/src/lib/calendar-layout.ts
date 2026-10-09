import type { CalendarEvent } from '@/lib/types'
import { addDays, getStartOfDay } from '@/lib/date-utils'

export const HOUR_HEIGHT = 48 // px per hour row, shared by the day and week grids
export const DAY_HEIGHT = HOUR_HEIGHT * 24
const MIN_EVENT_HEIGHT = 20 // px, so short events stay clickable/legible
const HOUR_MS = 60 * 60 * 1000

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

  const top = offsetInDay(clippedStart)
  const height = Math.max(
    MIN_EVENT_HEIGHT,
    ((clippedEnd.getTime() - clippedStart.getTime()) / HOUR_MS) * HOUR_HEIGHT
  )

  return { event, top, height }
}

// Pixel offset of an instant within its day's column, measured the same way
// as event tops so the now-line and events never disagree.
export function offsetInDay(time: Date): number {
  return ((time.getTime() - getStartOfDay(time).getTime()) / HOUR_MS) * HOUR_HEIGHT
}

// Whether the part of an event shown on `day` is over. Spanning events are
// judged per day, so yesterday's half of an overnight event is grayed even
// while today's half is still running.
export function isPastOnDay(event: CalendarEvent, day: Date, now: Date): boolean {
  const dayEnd = addDays(getStartOfDay(day), 1)
  const eventEnd = new Date(event.endAt)
  return (eventEnd < dayEnd ? eventEnd : dayEnd) <= now
}

export function timeRangeLabel(event: CalendarEvent): string {
  const fmt = (d: Date) => d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
  return `${fmt(new Date(event.startAt))} – ${fmt(new Date(event.endAt))}`
}
