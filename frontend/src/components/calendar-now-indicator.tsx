"use client"

import { useEffect, useRef } from "react"
import { offsetInDay } from "@/lib/calendar-layout"

interface CalendarNowIndicatorProps {
  now: Date
  // Bumped by the "Today" button so the view re-centers on the line even when
  // it's already showing today (and the indicator doesn't remount).
  scrollSignal: number
}

export function CalendarNowIndicator({ now, scrollSignal }: CalendarNowIndicatorProps) {
  const ref = useRef<HTMLDivElement>(null)

  // Centers the line on mount (opening the calendar, switching views, coming
  // back to today) but not on the per-minute ticks, so it never fights the
  // user's own scrolling.
  useEffect(() => {
    ref.current?.scrollIntoView({ block: "center", inline: "nearest" })
  }, [scrollSignal])

  return (
    <div
      ref={ref}
      aria-hidden
      className="pointer-events-none absolute inset-x-0 z-20 flex -translate-y-1/2 items-center"
      style={{ top: offsetInDay(now) }}
    >
      <span className="-ml-1 size-2 shrink-0 rounded-full bg-calendar-now" />
      <span className="h-px flex-1 bg-calendar-now" />
    </div>
  )
}
