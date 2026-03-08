export const queryKeys = {
  habits: {
    all: ['habits'] as const,
    list: () => [...queryKeys.habits.all, 'list'] as const,
  },
  completions: {
    all: ['completions'] as const,
    forDate: (date: string) => [...queryKeys.completions.all, 'date', date] as const,
    forRange: (start: string, end: string) =>
      [...queryKeys.completions.all, 'range', start, end] as const,
  },
}
