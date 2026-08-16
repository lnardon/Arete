import { useMutation, useQuery, useQueryClient, type UseQueryOptions } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-keys'
import type { WhatsAppStatus } from '@/lib/types'

export function useWhatsAppStatus(options?: {
  refetchInterval?: UseQueryOptions<WhatsAppStatus>['refetchInterval']
}) {
  return useQuery({
    queryKey: queryKeys.whatsapp.status(),
    queryFn: () => api.whatsapp.status(),
    refetchInterval: options?.refetchInterval,
  })
}

export function useGenerateLinkCode() {
  return useMutation({
    mutationFn: () => api.whatsapp.getLinkCode(),
    onError: () => {
      toast.error('Failed to generate a linking code')
    },
  })
}

export function useUnlinkWhatsApp() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.whatsapp.unlink(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.whatsapp.all })
      toast.success('WhatsApp disconnected')
    },
    onError: () => {
      toast.error('Failed to disconnect WhatsApp')
    },
  })
}
