import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-keys'
import type { CalendarEventInput } from '@/lib/types'

export function useCalendarEvents(start: string, end: string) {
  return useQuery({
    queryKey: queryKeys.calendarEvents.list(start, end),
    queryFn: () => api.calendarEvents.list(start, end),
  })
}

export function useCreateCalendarEvent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (event: CalendarEventInput) => api.calendarEvents.create(event),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.calendarEvents.all })
      toast.success('Event created')
    },
  })
}

export function useUpdateCalendarEvent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, event }: { id: string; event: CalendarEventInput }) =>
      api.calendarEvents.update(id, event),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.calendarEvents.all })
      toast.success('Event updated')
    },
  })
}

export function useDeleteCalendarEvent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.calendarEvents.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.calendarEvents.all })
      toast.success('Event deleted')
    },
  })
}
