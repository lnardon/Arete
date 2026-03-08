import type { Habit, HabitCompletion } from '@/lib/types'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}: ${path}`)
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  habits: {
    list: () =>
      request<Habit[]>('/api/v1/habits'),
    create: (name: string) =>
      request<Habit>('/api/v1/habits', { method: 'POST', body: JSON.stringify({ name }) }),
    update: (id: string, name: string) =>
      request<Habit>(`/api/v1/habits/${id}`, { method: 'PUT', body: JSON.stringify({ name }) }),
    delete: (id: string) =>
      request<void>(`/api/v1/habits/${id}`, { method: 'DELETE' }),
  },
  completions: {
    forDate: (date: string) =>
      request<HabitCompletion[]>(`/api/v1/completions?date=${date}`),
    forRange: (start: string, end: string) =>
      request<HabitCompletion[]>(`/api/v1/completions/range?start=${start}&end=${end}`),
    toggle: (habitId: string, date: string) =>
      request<void>('/api/v1/completions/toggle', {
        method: 'POST',
        body: JSON.stringify({ habitId, date }),
      }),
  },
}
