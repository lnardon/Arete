export interface Habit {
  id: string
  name: string
  createdAt: string
}

export interface HabitCompletion {
  habitId: string
  date: string // YYYY-MM-DD
}

export type GoalPeriodType = 'month' | 'quarter' | 'semester' | 'year'
export type GoalType = 'binary' | 'numeric'

export interface Goal {
  id: string
  userId: string
  title: string
  periodType: GoalPeriodType
  periodKey: string
  goalType: GoalType
  targetValue: number | null
  currentValue: number
  completed: boolean
  createdAt: string
}

export interface JournalEntry {
  id: string
  userId: string
  entryDate: string // YYYY-MM-DD
  mood: number // 1-5
  content: string
  createdAt: string
  updatedAt: string
}

export const MOOD_EMOJI: Record<number, string> = {
  1: '😞',
  2: '😕',
  3: '😐',
  4: '🙂',
  5: '😄',
}

export const MOOD_LABEL: Record<number, string> = {
  1: 'Rough',
  2: 'Meh',
  3: 'Okay',
  4: 'Good',
  5: 'Great',
}

export interface PomodoroProject {
  id: string
  name: string
  color: string
  createdAt: string
}

export interface PomodoroEntry {
  id: string
  projectId: string | null
  plannedMinutes: number
  startedAt: string
  endedAt: string | null
  localDate: string // YYYY-MM-DD
  createdAt: string
}

export type ActiveTimer =
  | { active: false }
  | { active: true; entry: PomodoroEntry }

export type WhatsAppStatus =
  | { linked: false }
  | { linked: true; phoneNumberMasked: string; linkedAt: string }

export interface WhatsAppLinkCode {
  code: string
  expiresAt: string
}

// One entry of the calendar: a single event, or one occurrence of a
// recurring series. An occurrence's id is an instance ID
// (`<seriesId>_<stamp>`), stable even after the occurrence is edited.
export interface CalendarEvent {
  id: string
  title: string
  description: string | null
  location: string | null
  startAt: string // ISO 8601
  endAt: string // ISO 8601
  allDay: boolean
  timezone: string
  // RFC 5545 lines of the series, as Google stores them, e.g.
  // ["RRULE:FREQ=WEEKLY;BYDAY=FR"]. null for single events.
  recurrence: string[] | null
  recurringEventId: string | null
  originalStartAt: string | null // ISO 8601
  color: string
  createdAt: string
  updatedAt: string
}

export interface CalendarEventInput {
  title: string
  description: string | null
  location: string | null
  startAt: string // ISO 8601
  endAt: string // ISO 8601
  allDay: boolean
  timezone: string
  recurrence: string[] | null
  color: string
}

// Which occurrences of a recurring event an edit or delete applies to,
// matching Google Calendar's "This event / This and following / All events".
export type RecurrenceScope = 'this' | 'following' | 'all'

export type GoogleCalendarStatus =
  | { connected: false }
  | { connected: true; email: string | null; lastSyncedAt: string | null }
