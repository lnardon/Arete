import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-keys'
import type { CalendarEventInput, RecurrenceScope } from '@/lib/types'

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
    mutationFn: ({ id, event, scope }: { id: string; event: CalendarEventInput; scope?: RecurrenceScope }) =>
      api.calendarEvents.update(id, event, scope),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.calendarEvents.all })
      toast.success('Event updated')
    },
  })
}

export function useDeleteCalendarEvent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, scope }: { id: string; scope?: RecurrenceScope }) => api.calendarEvents.delete(id, scope),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.calendarEvents.all })
      toast.success('Event deleted')
    },
  })
}
