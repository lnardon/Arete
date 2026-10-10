import { useMemo } from "react"
import { createFileRoute } from '@tanstack/react-router'
import { AppSidebar } from "@/components/app-sidebar"
import { MobileHeader } from "@/components/mobile-header"
import { DashboardHeader } from "@/components/dashboard-header"
import { DashboardPomodoroChip } from "@/components/dashboard-pomodoro-chip"
import { DashboardScheduleCard } from "@/components/dashboard-schedule-card"
import { DashboardGoalsCard } from "@/components/dashboard-goals-card"
import { DashboardHabitsCard } from "@/components/dashboard-habits-card"
import { useHabits } from "@/hooks/use-habits"
import { useCompletionsForDate } from "@/hooks/use-completions"
import { useGoals } from "@/hooks/use-goals"
import { useCalendarEvents } from "@/hooks/use-calendar-events"
import { GOOGLE_SYNC_REFRESH_MS } from "@/hooks/use-google-calendar"
import { useActiveTimer } from "@/hooks/use-pomodoro"
import { useNow } from "@/hooks/use-now"
import { countEventsLeft } from "@/lib/calendar-layout"
import { addDays, formatLocalDate, getCurrentPeriodKey, getStartOfDay } from "@/lib/date-utils"

export const Route = createFileRoute('/_app/')({
  component: DashboardPage,
})

export default function DashboardPage() {
  // Everything keys off `now`, so a dashboard left open rolls over to the
  // new day (and month) on its own.
  const now = useNow()
  const today = formatLocalDate(now)
  const monthKey = getCurrentPeriodKey("month", now)
  const dayStart = getStartOfDay(now)
  // Built the same way as CalendarView's day range so both share one cache entry.
  const rangeStart = dayStart.toISOString()
  const rangeEnd = addDays(dayStart, 1).toISOString()

  const habitsQuery = useHabits()
  const completionsQuery = useCompletionsForDate(today)
  const goalsQuery = useGoals("month", monthKey)
  const eventsQuery = useCalendarEvents(rangeStart, rangeEnd, { refetchInterval: GOOGLE_SYNC_REFRESH_MS })
  const { data: activeTimer } = useActiveTimer()

  const { data: habits } = habitsQuery
  const completions = completionsQuery.data ?? []
  const goals = goalsQuery.data ?? []
  const events = eventsQuery.data ?? []

  const activeHabits = useMemo(
    () => (habits ?? []).filter((h) => formatLocalDate(new Date(h.createdAt)) <= today),
    [habits, today]
  )
  const habitsDone = activeHabits.filter((h) =>
    completions.some((c) => c.habitId === h.id && c.date === today)
  ).length

  const habitsLoading = habitsQuery.isLoading || completionsQuery.isLoading
  const summaryLoading = habitsLoading || goalsQuery.isLoading || eventsQuery.isLoading
  const summary = summaryLoading
    ? null
    : {
        eventsLeft: countEventsLeft(events, now),
        hasTimedEvents: events.some((e) => !e.allDay),
        habitsDone,
        habitsTotal: activeHabits.length,
        goalsDone: goals.filter((g) => g.completed).length,
        goalsTotal: goals.length,
      }

  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <AppSidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <MobileHeader />
        <main className="flex-1 overflow-y-auto">
          <div className="p-16">
            <DashboardHeader
              now={now}
              summary={summary}
              chip={activeTimer?.active ? <DashboardPomodoroChip entry={activeTimer.entry} /> : null}
            />

            <div className="my-6 h-px bg-border" />

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
              <DashboardScheduleCard
                dayStart={dayStart}
                now={now}
                events={events}
                loading={eventsQuery.isLoading}
                className="h-[560px] lg:h-[calc(100dvh-15rem)] lg:min-h-[520px] lg:sticky lg:top-6"
              />
              <div className="flex flex-col gap-6 min-w-0">
                <DashboardGoalsCard
                  goals={goals}
                  loading={goalsQuery.isLoading}
                  monthName={now.toLocaleDateString("en-US", { month: "long" })}
                />
                <DashboardHabitsCard
                  habits={activeHabits}
                  completions={completions}
                  done={habitsDone}
                  date={today}
                  loading={habitsLoading}
                />
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}
