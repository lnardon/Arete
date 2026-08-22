import { createFileRoute } from "@tanstack/react-router"
import { AppSidebar } from "@/components/app-sidebar"
import { MobileHeader } from "@/components/mobile-header"
import { HabitHeatmap } from "@/components/habit-heatmap"
import { StatsSummary } from "@/components/stats-summary"
import { WeeklyChart } from "@/components/weekly-chart"
import { HabitBreakdown } from "@/components/habit-breakdown"
import { useHabits } from "@/hooks/use-habits"
import { useCompletionsForRange } from "@/hooks/use-completions"
import { formatLocalDate } from "@/lib/date-utils"

export const Route = createFileRoute('/_app/statistics')({
  component: StatisticsPage,
})

export default function StatisticsPage() {
  const today = formatLocalDate(new Date())
  const startDate = new Date()
  startDate.setDate(startDate.getDate() - 364)
  const start = formatLocalDate(startDate)

  const { data: habits = [] } = useHabits()
  const { data: completions = [] } = useCompletionsForRange(start, today)

  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <AppSidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <MobileHeader />
        <main className="flex-1 overflow-y-auto">
          <div className="p-16">
            <div className="mb-2">
              <h2 className="font-display text-2xl font-semibold tracking-wide text-foreground text-balance">
                Statistics
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                Measure what matters
              </p>
            </div>

            <div className="my-6 h-px bg-border" />

            <div className="flex flex-col gap-6">
              <HabitHeatmap habits={habits} completions={completions} />
              <StatsSummary habits={habits} completions={completions} />
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <WeeklyChart habits={habits} completions={completions} />
                <HabitBreakdown habits={habits} completions={completions} />
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}
