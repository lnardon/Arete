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

export interface Goal {
  id: string
  userId: string
  title: string
  periodType: GoalPeriodType
  periodKey: string
  completed: boolean
  createdAt: string
}

export type WhatsAppStatus =
  | { linked: false }
  | { linked: true; phoneNumberMasked: string; linkedAt: string }

export interface WhatsAppLinkCode {
  code: string
  expiresAt: string
}
