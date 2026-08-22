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
