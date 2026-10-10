import { useRef } from "react"
import { useNavigate } from "@tanstack/react-router"
import { CalendarDayView } from "@/components/calendar-day-view"
import { DashboardCard } from "@/components/dashboard-card"
import { DashboardUpNext } from "@/components/dashboard-up-next"
import { GoogleSyncStatus } from "@/components/google-sync-status"
import { GOOGLE_SYNC_REFRESH_MS, useGoogleCalendarStatus } from "@/hooks/use-google-calendar"
import { cn } from "@/lib/utils"
import type { CalendarEvent } from "@/lib/types"

interface DashboardScheduleCardProps {
  dayStart: Date
  now: Date
  events: CalendarEvent[]
  loading: boolean
  className?: string
}

// Today's events, read-only: the grid scrolls inside the card and anything
// clicked opens the Calendar page, where events are edited.
export function DashboardScheduleCard({ dayStart, now, events, loading, className }: DashboardScheduleCardProps) {
  const navigate = useNavigate()
  const scrollRef = useRef<HTMLDivElement>(null)
  const { data: googleStatus } = useGoogleCalendarStatus({ refetchInterval: GOOGLE_SYNC_REFRESH_MS })

  return (
    <DashboardCard
      title="Today's Schedule"
      href="/calendar"
      linkLabel="Open calendar"
      actions={<GoogleSyncStatus now={now} />}
      className={cn("min-h-0", className)}
    >
      {loading ? (
        <div className="flex-1 rounded-xl bg-muted animate-pulse" />
      ) : (
        <>
          <DashboardUpNext events={events} now={now} googleConnected={googleStatus?.connected} />
          <div ref={scrollRef} className="flex-1 min-h-0 overflow-y-auto pt-2">
            <CalendarDayView
              date={dayStart}
              events={events}
              now={now}
              scrollToNowSignal={0}
              onEventClick={() => navigate({ to: "/calendar" })}
              scrollContainerRef={scrollRef}
            />
          </div>
        </>
      )}
    </DashboardCard>
  )
}
