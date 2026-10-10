import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Progress } from "@/components/ui/progress"
import { DashboardCard, DashboardCount, DashboardEmpty, DashboardSkeletonRows } from "@/components/dashboard-card"
import { useAddGoalProgress, useToggleGoal } from "@/hooks/use-goals"
import { cn } from "@/lib/utils"
import type { Goal } from "@/lib/types"

function BinaryGoalRow({ goal, onToggle }: { goal: Goal; onToggle: () => void }) {
  return (
    <label
      htmlFor={`dashboard-goal-${goal.id}`}
      className={cn(
        "flex items-center gap-3 px-3.5 py-2.5 rounded-lg border border-border bg-background cursor-pointer transition-colors",
        goal.completed && "bg-muted/60"
      )}
    >
      <Checkbox
        id={`dashboard-goal-${goal.id}`}
        checked={goal.completed}
        onCheckedChange={onToggle}
        aria-label={goal.title}
        className="h-5 w-5 shrink-0 border-foreground/30 data-checked:border-secondary data-checked:bg-secondary data-checked:text-secondary-foreground"
      />
      <span
        className={cn(
          "flex-1 text-sm tracking-wide truncate select-none",
          goal.completed && "line-through text-muted-foreground"
        )}
      >
        {goal.title}
      </span>
    </label>
  )
}

function NumericGoalRow({ goal, pending, onAddProgress }: { goal: Goal; pending: boolean; onAddProgress: () => void }) {
  const target = goal.targetValue ?? 0
  const percent = target > 0 ? Math.min(100, (goal.currentValue / target) * 100) : 0

  return (
    <div
      className={cn(
        "flex items-center gap-3 px-3.5 py-2.5 rounded-lg border border-border bg-background",
        goal.completed && "bg-muted/60"
      )}
    >
      <Button
        variant="outline"
        onClick={onAddProgress}
        disabled={pending}
        aria-label={`Add 1 to ${goal.title}`}
        className="h-7 min-w-7 px-1.5 rounded-md border-foreground/20 text-xs font-semibold tabular-nums shrink-0"
      >
        +1
      </Button>
      <div className="flex-1 min-w-0 flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <span className={cn("text-sm tracking-wide truncate", goal.completed && "text-muted-foreground")}>
            {goal.title}
          </span>
          <span className="text-xs text-muted-foreground tabular-nums shrink-0">
            {goal.currentValue} / {target}
          </span>
        </div>
        <Progress value={percent} />
      </div>
    </div>
  )
}

interface DashboardGoalsCardProps {
  goals: Goal[]
  loading: boolean
  monthName: string
}

export function DashboardGoalsCard({ goals, loading, monthName }: DashboardGoalsCardProps) {
  const toggleGoal = useToggleGoal()
  const addGoalProgress = useAddGoalProgress()
  const completed = goals.filter((g) => g.completed).length

  return (
    <DashboardCard
      title={`${monthName} Goals`}
      href="/goals"
      meta={!loading && goals.length > 0 && <DashboardCount done={completed} total={goals.length} />}
    >
      {loading ? (
        <DashboardSkeletonRows />
      ) : goals.length === 0 ? (
        <DashboardEmpty message={`No goals set for ${monthName} yet`} href="/goals" linkLabel="Set a goal" />
      ) : (
        <div className="flex flex-col gap-2">
          {goals.map((goal) =>
            goal.goalType === "numeric" ? (
              <NumericGoalRow
                key={goal.id}
                goal={goal}
                pending={addGoalProgress.isPending && addGoalProgress.variables?.id === goal.id}
                onAddProgress={() => addGoalProgress.mutate({ id: goal.id, delta: 1 })}
              />
            ) : (
              <BinaryGoalRow
                key={goal.id}
                goal={goal}
                onToggle={() =>
                  toggleGoal.mutate({ id: goal.id, periodType: goal.periodType, periodKey: goal.periodKey })
                }
              />
            )
          )}
        </div>
      )}
    </DashboardCard>
  )
}
