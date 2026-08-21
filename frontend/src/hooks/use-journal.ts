import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api, ApiError } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-keys'

export function useJournalEntries(limit?: number) {
  return useQuery({
    queryKey: queryKeys.journal.list(limit),
    queryFn: () => api.journal.list(limit),
  })
}

export function useJournalEntry(date: string) {
  return useQuery({
    queryKey: queryKeys.journal.entry(date),
    queryFn: async () => {
      try {
        return await api.journal.get(date)
      } catch (err) {
        if (err instanceof ApiError && err.status === 404) return null
        throw err
      }
    },
  })
}

export function useUpsertJournalEntry() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ date, mood, content }: { date: string; mood: number; content: string }) =>
      api.journal.upsert(date, mood, content),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.journal.all })
      toast.success('Entry saved')
    },
  })
}

export function useDeleteJournalEntry() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (date: string) => api.journal.delete(date),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.journal.all })
      toast.success('Entry deleted')
    },
  })
}
