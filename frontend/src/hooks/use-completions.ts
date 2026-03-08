import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-keys'
import type { HabitCompletion } from '@/lib/types'

export function useCompletionsForDate(date: string) {
  return useQuery({
    queryKey: queryKeys.completions.forDate(date),
    queryFn: () => api.completions.forDate(date),
  })
}

export function useCompletionsForRange(start: string, end: string) {
  return useQuery({
    queryKey: queryKeys.completions.forRange(start, end),
    queryFn: () => api.completions.forRange(start, end),
  })
}

export function useToggleCompletion() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ habitId, date }: { habitId: string; date: string }) =>
      api.completions.toggle(habitId, date),
    onMutate: async ({ habitId, date }) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.completions.forDate(date) })
      const prev = queryClient.getQueryData<HabitCompletion[]>(
        queryKeys.completions.forDate(date)
      )
      const old = prev ?? []
      const isCompleting = !old.some((c) => c.habitId === habitId && c.date === date)
      if (isCompleting) toast.success("Habit completed!")
      queryClient.setQueryData<HabitCompletion[]>(
        queryKeys.completions.forDate(date),
        () => {
          return isCompleting
            ? [...old, { habitId, date }]
            : old.filter((c) => !(c.habitId === habitId && c.date === date))
        }
      )
      return { prev }
    },
    onError: (_err, { date }, ctx) => {
      queryClient.setQueryData(queryKeys.completions.forDate(date), ctx?.prev)
    },
    onSettled: (_data, _err, { date }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.completions.forDate(date) })
    },
  })
}
