import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-keys'

export function useGoals(periodType: string, periodKey: string) {
  return useQuery({
    queryKey: queryKeys.goals.list(periodType, periodKey),
    queryFn: () => api.goals.list(periodType, periodKey),
  })
}

export function useCreateGoal() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ title, periodType, periodKey }: { title: string; periodType: string; periodKey: string }) =>
      api.goals.create(title, periodType, periodKey),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals.all })
      toast.success('Goal created')
    },
  })
}

export function useToggleGoal() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id }: { id: string }) => api.goals.toggle(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals.list(data.periodType, data.periodKey) })
      toast.success(data.completed ? 'Goal completed!' : 'Goal uncompleted')
    },
  })
}

export function useUpdateGoal() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, title }: { id: string; title: string }) => api.goals.update(id, title),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals.all })
      toast.success('Goal updated')
    },
  })
}

export function useDeleteGoal() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.goals.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.goals.all })
      toast.success('Goal deleted')
    },
  })
}
