export const queryKeys = {
  habits: {
    all: ['habits'] as const,
    list: () => [...queryKeys.habits.all, 'list'] as const,
  },
  goals: {
    all: ['goals'] as const,
    list: (periodType: string, periodKey: string) =>
      [...queryKeys.goals.all, periodType, periodKey] as const,
  },
  journal: {
    all: ['journal'] as const,
    list: (limit?: number) => [...queryKeys.journal.all, 'list', limit ?? null] as const,
    entry: (date: string) => [...queryKeys.journal.all, 'entry', date] as const,
  },
  completions: {
    all: ['completions'] as const,
    forDate: (date: string) => [...queryKeys.completions.all, 'date', date] as const,
    forRange: (start: string, end: string) =>
      [...queryKeys.completions.all, 'range', start, end] as const,
  },
  pomodoro: {
    all: ['pomodoro'] as const,
    projects: () => [...queryKeys.pomodoro.all, 'projects'] as const,
    active: () => [...queryKeys.pomodoro.all, 'active'] as const,
    entries: (start: string, end: string, projectId?: string) =>
      [...queryKeys.pomodoro.all, 'entries', start, end, projectId ?? null] as const,
  },
  whatsapp: {
    all: ['whatsapp'] as const,
    status: () => [...queryKeys.whatsapp.all, 'status'] as const,
  },
}
