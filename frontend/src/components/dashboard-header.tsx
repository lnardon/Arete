import type { ReactNode } from "react"

export interface DaySummary {
  eventsLeft: number
  hasTimedEvents: boolean
  habitsDone: number
  habitsTotal: number
  goalsDone: number
  goalsTotal: number
}

function summarySegments(summary: DaySummary): string[] {
  const segments: string[] = []

  if (summary.eventsLeft > 0) {
    segments.push(`${summary.eventsLeft} ${summary.eventsLeft === 1 ? "event" : "events"} left`)
  } else {
    segments.push(summary.hasTimedEvents ? "No events left" : "No events today")
  }
  if (summary.habitsTotal > 0) {
    segments.push(`${summary.habitsDone}/${summary.habitsTotal} habits`)
  }
  if (summary.goalsTotal > 0) {
    segments.push(`${summary.goalsDone}/${summary.goalsTotal} goals this month`)
  }

  return segments
}

interface DashboardHeaderProps {
  now: Date
  // null while the data behind it is loading, so the counts never flash 0/0.
  summary: DaySummary | null
  chip?: ReactNode
}

export function DashboardHeader({ now, summary, chip }: DashboardHeaderProps) {
  const date = now.toLocaleDateString("en-US", { weekday: "long", month: "long", day: "numeric" })
  const segments = [date, ...(summary ? summarySegments(summary) : [])]

  return (
    <div className="mb-2 flex flex-wrap items-start justify-between gap-4">
      <div className="min-w-0">
        <h2 className="font-display text-2xl font-semibold tracking-wide text-foreground text-balance">
          Dashboard
        </h2>
        <p className="text-sm text-muted-foreground mt-1">
          {segments.map((segment, i) => (
            <span key={i}>
              {i > 0 && <span className="text-muted-foreground/50">{" · "}</span>}
              {segment}
            </span>
          ))}
        </p>
      </div>
      {chip}
    </div>
  )
}
