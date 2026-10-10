import { Checkbox } from "@/components/ui/checkbox"
import { DashboardCard, DashboardCount, DashboardEmpty, DashboardSkeletonRows } from "@/components/dashboard-card"
import { useToggleCompletion } from "@/hooks/use-completions"
import { cn } from "@/lib/utils"
import type { Habit, HabitCompletion } from "@/lib/types"

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
        aria-label={habit.name}
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

interface DashboardHabitsCardProps {
  // Habits active on `date`, i.e. created on or before it.
  habits: Habit[]
  completions: HabitCompletion[]
  done: number
  date: string
  loading: boolean
}

export function DashboardHabitsCard({ habits, completions, done, date, loading }: DashboardHabitsCardProps) {
  const percentage = habits.length > 0 ? Math.round((done / habits.length) * 100) : 0

  return (
    <DashboardCard
      title="Today's Habits"
      href="/habits"
      meta={!loading && habits.length > 0 && <DashboardCount done={done} total={habits.length} />}
    >
      {loading ? (
        <DashboardSkeletonRows />
      ) : habits.length === 0 ? (
        <DashboardEmpty message="No habits yet" href="/habits" linkLabel="Create your first habit" />
      ) : (
        <>
          <div className="flex flex-col gap-1.5">
            <div className="h-1.5 rounded-full bg-border overflow-hidden">
              <div
                className="h-full rounded-full bg-secondary transition-all duration-500 ease-out"
                style={{ width: `${percentage}%` }}
              />
            </div>
            {percentage === 100 && (
              <p className="text-xs text-muted-foreground">All habits completed. Virtue achieved.</p>
            )}
          </div>
          <div className="flex flex-col gap-2">
            {habits.map((habit) => (
              <TodayHabitRow key={habit.id} habit={habit} date={date} completions={completions} />
            ))}
          </div>
        </>
      )}
    </DashboardCard>
  )
}
