export interface Habit {
  id: string
  name: string
  createdAt: string
}

export interface HabitCompletion {
  habitId: string
  date: string // YYYY-MM-DD
}
