import type { Goal, Habit, HabitCompletion } from '@/lib/types'

export class ApiError extends Error {
  readonly status: number
  constructor(status: number) {
    super('Request failed')
    this.status = status
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    credentials: 'include',
  })
  if (res.status === 401) {
    window.dispatchEvent(new Event('auth:unauthorized'))
    throw new ApiError(res.status)
  }
  if (!res.ok) throw new ApiError(res.status)
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
  goals: {
    list: (periodType: string, periodKey: string) =>
      request<Goal[]>(`/api/v1/goals?period_type=${periodType}&period_key=${periodKey}`),
    create: (title: string, periodType: string, periodKey: string) =>
      request<Goal>('/api/v1/goals', { method: 'POST', body: JSON.stringify({ title, periodType, periodKey }) }),
    update: (id: string, title: string) =>
      request<Goal>(`/api/v1/goals/${id}`, { method: 'PUT', body: JSON.stringify({ title }) }),
    toggle: (id: string) =>
      request<Goal>(`/api/v1/goals/${id}/toggle`, { method: 'PATCH' }),
    delete: (id: string) =>
      request<void>(`/api/v1/goals/${id}`, { method: 'DELETE' }),
  },
  auth: {
    register: (username: string, password: string) =>
      request<{ userId: string; username: string }>('/api/v1/auth/register', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      }),
    login: (username: string, password: string) =>
      request<{ userId: string; username: string }>('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      }),
    logout: () =>
      request<void>('/api/v1/auth/logout', { method: 'POST' }),
    me: () =>
      request<{ userId: string; username: string }>('/api/v1/auth/me'),
  },
}
