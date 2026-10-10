import { Link } from "@tanstack/react-router"
import { ArrowRight } from "lucide-react"
import { getUpNext, timeRangeLabel } from "@/lib/calendar-layout"
import { formatDurationShort } from "@/lib/date-utils"
import type { CalendarEvent } from "@/lib/types"

interface UpNextRowProps {
  label: string
  event: CalendarEvent
  extra: string | null
  relative: string
}

function UpNextRow({ label, event, extra, relative }: UpNextRowProps) {
  return (
    <div className="flex gap-3 px-3.5 py-2.5">
      <span className="w-1 shrink-0 self-stretch rounded-full" style={{ backgroundColor: event.color }} />
      <div className="flex-1 min-w-0">
        <div className="flex items-baseline justify-between gap-2">
          <p className="label-section">{label}</p>
          <p className="text-xs text-muted-foreground tabular-nums shrink-0">{relative}</p>
        </div>
        <p className="mt-0.5 flex items-baseline gap-1.5 min-w-0">
          <span className="text-sm font-medium text-foreground truncate">{event.title}</span>
          {extra && <span className="text-xs text-muted-foreground shrink-0">{extra}</span>}
        </p>
        <p className="text-xs text-muted-foreground">{timeRangeLabel(event)}</p>
      </div>
    </div>
  )
}

interface DashboardUpNextProps {
  events: CalendarEvent[]
  now: Date
  // undefined while the status is loading, so the connect hint doesn't flash.
  googleConnected: boolean | undefined
}

export function DashboardUpNext({ events, now, googleConnected }: DashboardUpNextProps) {
  const { current, currentExtra, next, nextExtra } = getUpNext(events, now)

  if (!current && !next) {
    const hadTimedEvents = events.some((e) => !e.allDay)
    return (
      <div className="flex flex-col gap-1 rounded-lg border border-dashed border-border px-3.5 py-3">
        <p className="text-sm text-muted-foreground">
          {hadTimedEvents ? "You're clear for the rest of the day" : "Nothing scheduled today"}
        </p>
        {!hadTimedEvents && googleConnected === false && (
          <Link
            to="/settings"
            className="text-xs text-muted-foreground hover:text-foreground inline-flex items-center gap-1 transition-colors"
          >
            Connect Google Calendar
            <ArrowRight className="w-3 h-3" />
          </Link>
        )}
      </div>
    )
  }

  return (
    <Link
      to="/calendar"
      className="flex flex-col divide-y divide-border rounded-lg border border-border bg-background hover:bg-muted/40 transition-colors"
    >
      {current && (
        <UpNextRow
          label="Now"
          event={current}
          extra={currentExtra > 0 ? `+${currentExtra} more` : null}
          relative={`ends in ${formatDurationShort(new Date(current.endAt).getTime() - now.getTime())}`}
        />
      )}
      {next && (
        <UpNextRow
          label="Up next"
          event={next}
          extra={nextExtra > 0 ? `+${nextExtra} at the same time` : null}
          relative={`in ${formatDurationShort(new Date(next.startAt).getTime() - now.getTime())}`}
        />
      )}
    </Link>
  )
}
