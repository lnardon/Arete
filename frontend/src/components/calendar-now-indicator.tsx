"use client"

import { useEffect, useRef, type RefObject } from "react"
import { offsetInDay } from "@/lib/calendar-layout"

interface CalendarNowIndicatorProps {
  now: Date
  // Bumped by the "Today" button so the view re-centers on the line even when
  // it's already showing today (and the indicator doesn't remount).
  scrollSignal: number
  // When the grid scrolls inside its own box (the dashboard card), only that
  // box is scrolled. Without it the line is scrolled into view page-wide.
  scrollContainerRef?: RefObject<HTMLElement | null>
}

export function CalendarNowIndicator({ now, scrollSignal, scrollContainerRef }: CalendarNowIndicatorProps) {
  const ref = useRef<HTMLDivElement>(null)

  // Centers the line on mount (opening the calendar, switching views, coming
  // back to today) but not on the per-minute ticks, so it never fights the
  // user's own scrolling.
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const container = scrollContainerRef?.current
    if (!container) {
      el.scrollIntoView({ block: "center", inline: "nearest" })
      return
    }
    // scrollIntoView would also scroll every scrollable ancestor (the page
    // itself), so move just the container.
    const delta = el.getBoundingClientRect().top - container.getBoundingClientRect().top
    container.scrollBy({ top: delta - container.clientHeight / 2 })
  }, [scrollSignal, scrollContainerRef])

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
