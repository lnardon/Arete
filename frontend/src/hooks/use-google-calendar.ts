import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-keys'

export function useGoogleCalendarStatus() {
  return useQuery({
    queryKey: queryKeys.googleCalendar.status(),
    queryFn: () => api.google.status(),
  })
}

export function useDisconnectGoogleCalendar() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.google.disconnect(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.googleCalendar.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.calendarEvents.all })
      toast.success('Google Calendar disconnected')
    },
    onError: () => {
      toast.error('Failed to disconnect Google Calendar')
    },
  })
}

export function useTriggerGoogleSync() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.google.sync(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.googleCalendar.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.calendarEvents.all })
      toast.success('Synced with Google Calendar')
    },
    onError: () => {
      toast.error('Sync failed — try again in a moment')
    },
  })
}
