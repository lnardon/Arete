import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-keys'

export function useHabits() {
  return useQuery({
    queryKey: queryKeys.habits.list(),
    queryFn: api.habits.list,
  })
}

export function useCreateHabit() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (name: string) => api.habits.create(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.habits.all })
      toast.success("Habit created")
    },
  })
}

export function useUpdateHabit() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) =>
      api.habits.update(id, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.habits.all })
      toast.success("Habit updated")
    },
  })
}

export function useDeleteHabit() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.habits.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.habits.all })
      toast.success("Habit deleted")
    },
  })
}
