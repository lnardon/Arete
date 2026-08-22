import { useMemo, type ReactNode } from "react"
import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowRight, BookOpen, CheckCircle2, Circle, Sparkles } from "lucide-react"
import { AppSidebar } from "@/components/app-sidebar"
import { MobileHeader } from "@/components/mobile-header"
import { StatsSummary } from "@/components/stats-summary"
import { Checkbox } from "@/components/ui/checkbox"
import { Progress } from "@/components/ui/progress"
import { useHabits } from "@/hooks/use-habits"
import { useCompletionsForDate, useCompletionsForRange, useToggleCompletion } from "@/hooks/use-completions"
import { useGoals } from "@/hooks/use-goals"
import { useJournalEntry } from "@/hooks/use-journal"
import { formatLocalDate, getCurrentPeriodKey } from "@/lib/date-utils"
import { cn } from "@/lib/utils"
import type { Goal, Habit, HabitCompletion } from "@/lib/types"

export const Route = createFileRoute('/_app/')({
  component: DashboardPage,
})

const DASHBOARD_PREVIEW_LIMIT = 5

function DashboardCard({
  title,
  href,
  linkLabel = "View all",
  children,
}: {
  title: string
  href: string
  linkLabel?: string
  children: ReactNode
}) {
  return (
    <div className="border border-border rounded-2xl bg-card p-5 card-elevated flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h3 className="font-display text-base font-semibold tracking-wide text-foreground">
          {title}
        </h3>
        <Link
          to={href}
          className="text-xs text-muted-foreground hover:text-foreground inline-flex items-center gap-1 transition-colors shrink-0"
        >
          {linkLabel}
          <ArrowRight className="w-3 h-3" />
        </Link>
      </div>
      {children}
    </div>
  )
}

function TodayHabitRow({ habit, date, completions }: { habit: Habit; date: string; completions: HabitCompletion[] }) {
  const completed = completions.some((c) => c.habitId === habit.id && c.date === date)
  const toggle = useToggleCompletion()

  return (
    <label
      htmlFor={`dashboard-habit-${habit.id}`}
      className={cn(
        "flex items-center gap-3 px-3.5 py-2.5 rounded-lg border border-border bg-background cursor-pointer transition-colors",
        completed && "bg-muted/60"
      )}
    >
      <Checkbox
        id={`dashboard-habit-${habit.id}`}
        checked={completed}
        onCheckedChange={() => toggle.mutate({ habitId: habit.id, date })}
        className="h-5 w-5 shrink-0 border-foreground/30 data-checked:border-secondary data-checked:bg-secondary data-checked:text-secondary-foreground"
      />
      <span
        className={cn(
          "flex-1 text-sm tracking-wide select-none",
          completed && "line-through text-muted-foreground"
        )}
      >
        {habit.name}
      </span>
    </label>
  )
}

function GoalSummaryRow({ goal }: { goal: Goal }) {
  if (goal.goalType === "numeric") {
    const target = goal.targetValue ?? 0
    const percent = target > 0 ? Math.min(100, (goal.currentValue / target) * 100) : 0
    return (
      <div className="flex flex-col gap-2 px-3.5 py-2.5 rounded-lg border border-border bg-background">
        <div className="flex items-center justify-between gap-2">
          <span className="text-sm tracking-wide truncate">{goal.title}</span>
          <span className="text-xs text-muted-foreground tabular-nums shrink-0">
            {goal.currentValue} / {target}
          </span>
        </div>
        <Progress value={percent} />
      </div>
    )
  }

  return (
    <div className="flex items-center gap-3 px-3.5 py-2.5 rounded-lg border border-border bg-background">
      {goal.completed ? (
        <CheckCircle2 className="w-4 h-4 shrink-0 text-foreground" strokeWidth={1.5} />
      ) : (
        <Circle className="w-4 h-4 shrink-0 text-muted-foreground" strokeWidth={1.5} />
      )}
      <span
        className={cn(
          "flex-1 text-sm tracking-wide truncate",
          goal.completed && "text-muted-foreground line-through"
        )}
      >
        {goal.title}
      </span>
    </div>
  )
}

export default function DashboardPage() {
  const today = formatLocalDate(new Date())
  const statsStart = new Date()
  statsStart.setDate(statsStart.getDate() - 364)
  const start = formatLocalDate(statsStart)

  const { data: habits = [] } = useHabits()
  const { data: todayCompletions = [] } = useCompletionsForDate(today)
  const { data: rangeCompletions = [] } = useCompletionsForRange(start, today)
  const { data: monthGoals = [] } = useGoals('month', getCurrentPeriodKey('month'))
  const { data: todayEntry } = useJournalEntry(today)

  const activeHabits = useMemo(
    () => habits.filter((h) => formatLocalDate(new Date(h.createdAt)) <= today),
    [habits, today]
  )
  const completedToday = todayCompletions.filter((c) =>
    activeHabits.some((h) => h.id === c.habitId) && c.date === today
  ).length
  const habitPreview = activeHabits.slice(0, DASHBOARD_PREVIEW_LIMIT)

  const goalPreview = monthGoals.slice(0, DASHBOARD_PREVIEW_LIMIT)

  const greeting = new Date().toLocaleDateString("en-US", {
    weekday: "long",
    month: "long",
    day: "numeric",
  })

  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <AppSidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <MobileHeader />
        <main className="flex-1 overflow-y-auto">
          <div className="p-16">
            <div className="mb-2">
              <h2 className="font-display text-2xl font-semibold tracking-wide text-foreground text-balance">
                Dashboard
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                {greeting} — here's where things stand
              </p>
            </div>

            <div className="my-6 h-px bg-border" />

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
              <DashboardCard title="Today's Habits" href="/habits">
                {activeHabits.length === 0 ? (
                  <p className="text-sm text-muted-foreground py-4 text-center">
                    No habits yet — create your first one
                  </p>
                ) : (
                  <>
                    <div className="flex items-baseline justify-between">
                      <p className="label-section">Progress</p>
                      <p className="font-display text-sm font-semibold text-foreground tabular-nums">
                        {completedToday}
                        <span className="text-muted-foreground">{" / "}{activeHabits.length}</span>
                      </p>
                    </div>
                    <div className="flex flex-col gap-2">
                      {habitPreview.map((habit) => (
                        <TodayHabitRow key={habit.id} habit={habit} date={today} completions={todayCompletions} />
                      ))}
                    </div>
                    {activeHabits.length > DASHBOARD_PREVIEW_LIMIT && (
                      <p className="text-xs text-muted-foreground/70 text-center">
                        +{activeHabits.length - DASHBOARD_PREVIEW_LIMIT} more on the Habits page
                      </p>
                    )}
                  </>
                )}
              </DashboardCard>

              <DashboardCard title="This Month's Goals" href="/goals">
                {goalPreview.length === 0 ? (
                  <p className="text-sm text-muted-foreground py-4 text-center">
                    No goals set for this month yet
                  </p>
                ) : (
                  <>
                    <div className="flex flex-col gap-2">
                      {goalPreview.map((goal) => (
                        <GoalSummaryRow key={goal.id} goal={goal} />
                      ))}
                    </div>
                    {monthGoals.length > DASHBOARD_PREVIEW_LIMIT && (
                      <p className="text-xs text-muted-foreground/70 text-center">
                        +{monthGoals.length - DASHBOARD_PREVIEW_LIMIT} more on the Goals page
                      </p>
                    )}
                  </>
                )}
              </DashboardCard>
            </div>

            <div className="mb-6">
              <div className="flex items-center justify-between mb-4">
                <p className="label-section">Statistics</p>
                <Link
                  to="/statistics"
                  className="text-xs text-muted-foreground hover:text-foreground inline-flex items-center gap-1 transition-colors"
                >
                  View full statistics
                  <ArrowRight className="w-3 h-3" />
                </Link>
              </div>
              <StatsSummary habits={habits} completions={rangeCompletions} />
            </div>

            <Link
              to="/journal"
              className="flex items-center gap-4 px-5 py-4 border border-border rounded-2xl bg-card card-elevated hover:bg-muted/40 transition-colors"
            >
              <div
                className={cn(
                  "w-10 h-10 rounded-lg flex items-center justify-center shrink-0",
                  todayEntry ? "bg-foreground/10" : "bg-primary text-primary-foreground"
                )}
              >
                {todayEntry ? (
                  <BookOpen className="w-5 h-5" strokeWidth={1.5} />
                ) : (
                  <Sparkles className="w-5 h-5" strokeWidth={1.5} />
                )}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-foreground">
                  {todayEntry ? "Today's entry is logged" : "You haven't reflected today"}
                </p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {todayEntry ? "Edit your journal entry" : "Take a couple of minutes to write in your journal"}
                </p>
              </div>
              <ArrowRight className="w-4 h-4 text-muted-foreground shrink-0" />
            </Link>
          </div>
        </main>
      </div>
    </div>
  )
}
