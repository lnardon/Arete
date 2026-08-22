import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-keys'

export function useActiveTimer() {
  return useQuery({
    queryKey: queryKeys.pomodoro.active(),
    queryFn: api.pomodoro.active,
    refetchInterval: 30_000,
  })
}

export function useProjects() {
  return useQuery({
    queryKey: queryKeys.pomodoro.projects(),
    queryFn: api.pomodoro.projects.list,
  })
}

export function useCreateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ name, color }: { name: string; color: string }) =>
      api.pomodoro.projects.create(name, color),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pomodoro.projects() })
      toast.success("Project created")
    },
  })
}

export function useUpdateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, name, color }: { id: string; name: string; color: string }) =>
      api.pomodoro.projects.update(id, name, color),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pomodoro.projects() })
      toast.success("Project updated")
    },
  })
}

export function useDeleteProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.pomodoro.projects.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pomodoro.all })
      toast.success("Project deleted")
    },
  })
}

export function useStartTimer() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ projectId, plannedMinutes, localDate }: { projectId: string | null; plannedMinutes: number; localDate: string }) =>
      api.pomodoro.start(projectId, plannedMinutes, localDate),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pomodoro.active() })
      toast.success("Timer started")
    },
    onError: () => {
      toast.error("A timer is already running")
    },
  })
}

export function useStopTimer() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.pomodoro.stop,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pomodoro.active() })
      queryClient.invalidateQueries({ queryKey: queryKeys.pomodoro.all })
      toast.success("Timer stopped")
    },
  })
}

export function useEntries(start: string, end: string, projectId?: string) {
  return useQuery({
    queryKey: queryKeys.pomodoro.entries(start, end, projectId),
    queryFn: () => api.pomodoro.entries.list(start, end, projectId),
  })
}

export function useDeleteEntry() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.pomodoro.entries.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pomodoro.all })
      toast.success("Entry deleted")
    },
  })
}
