import { RefreshCw } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { GOOGLE_SYNC_REFRESH_MS, useGoogleCalendarStatus, useTriggerGoogleSync } from "@/hooks/use-google-calendar"
import { formatTimeAgo } from "@/lib/date-utils"
import { cn } from "@/lib/utils"

// "Synced 5 min ago" plus a sync button. Renders nothing unless Google
// Calendar is connected.
export function GoogleSyncStatus({ now }: { now: Date }) {
  const { data: status } = useGoogleCalendarStatus({ refetchInterval: GOOGLE_SYNC_REFRESH_MS })
  const sync = useTriggerGoogleSync()

  if (!status?.connected) return null

  return (
    <div className="flex items-center gap-1">
      {status.lastSyncedAt ? (
        <Tooltip>
          <TooltipTrigger render={<span />} className="text-xs text-muted-foreground cursor-default">
            Synced {formatTimeAgo(status.lastSyncedAt, now)}
          </TooltipTrigger>
          <TooltipContent>{new Date(status.lastSyncedAt).toLocaleString()}</TooltipContent>
        </Tooltip>
      ) : (
        <span className="text-xs text-muted-foreground">Not synced yet</span>
      )}
      <Button
        variant="ghost"
        size="icon"
        onClick={() => sync.mutate()}
        disabled={sync.isPending}
        aria-busy={sync.isPending}
        aria-label="Sync Google Calendar"
        className="h-7 w-7 text-muted-foreground hover:text-foreground hover:bg-transparent"
      >
        <RefreshCw className={cn("w-3.5 h-3.5", sync.isPending && "animate-spin")} />
      </Button>
    </div>
  )
}
