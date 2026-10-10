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

export interface UpNext {
  current: CalendarEvent | null // in-progress timed event ending soonest
  currentExtra: number // other timed events in progress
  next: CalendarEvent | null // first timed event starting after now
  nextExtra: number // others starting at the same instant as `next`
}

// Where the day stands relative to `now`, for the dashboard's "Up next"
// strip. All-day events are left out: they have no start time to count
// down to.
export function getUpNext(events: CalendarEvent[], now: Date): UpNext {
  const timed = events.filter((e) => !e.allDay)
  const inProgress = timed
    .filter((e) => new Date(e.startAt) <= now && now < new Date(e.endAt))
    .sort((a, b) => new Date(a.endAt).getTime() - new Date(b.endAt).getTime())
  const upcoming = timed
    .filter((e) => new Date(e.startAt) > now)
    .sort(
      (a, b) =>
        new Date(a.startAt).getTime() - new Date(b.startAt).getTime() || a.title.localeCompare(b.title)
    )

  const next = upcoming[0] ?? null
  const nextStart = next ? new Date(next.startAt).getTime() : null

  return {
    current: inProgress[0] ?? null,
    currentExtra: Math.max(0, inProgress.length - 1),
    next,
    nextExtra: upcoming.filter((e) => new Date(e.startAt).getTime() === nextStart).length - (next ? 1 : 0),
  }
}

// Timed events that haven't ended yet, including the ones in progress.
export function countEventsLeft(events: CalendarEvent[], now: Date): number {
  return events.filter((e) => !e.allDay && new Date(e.endAt) > now).length
}
