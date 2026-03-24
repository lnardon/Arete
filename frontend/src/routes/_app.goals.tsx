import { useState } from "react"
import { createFileRoute } from "@tanstack/react-router"
import { ChevronLeft, ChevronRight, Plus } from "lucide-react"
import { AppSidebar } from "@/components/app-sidebar"
import { MobileHeader } from "@/components/mobile-header"
import { GoalItem } from "@/components/goal-item"
import { GoalDialog } from "@/components/goal-dialog"
import { Button } from "@/components/ui/button"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useGoals, useCreateGoal, useUpdateGoal, useToggleGoal, useDeleteGoal } from "@/hooks/use-goals"
import { getCurrentPeriodKey, formatPeriodLabel, navigatePeriodKey } from "@/lib/date-utils"
import type { Goal, GoalPeriodType } from "@/lib/types"

export const Route = createFileRoute('/_app/goals')({
  component: GoalsPage,
})

const PERIOD_TYPES: { value: GoalPeriodType; label: string }[] = [
  { value: 'month', label: 'Month' },
  { value: 'quarter', label: 'Quarter' },
  { value: 'semester', label: 'Semester' },
  { value: 'year', label: 'Year' },
]

function GoalTabPanel({ periodType }: { periodType: GoalPeriodType }) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingGoal, setEditingGoal] = useState<Goal | null>(null)
  const [periodKey, setPeriodKey] = useState(() => getCurrentPeriodKey(periodType))
  const { data: goals = [] } = useGoals(periodType, periodKey)
  const createGoal = useCreateGoal()
  const updateGoal = useUpdateGoal()
  const toggleGoal = useToggleGoal()
  const deleteGoal = useDeleteGoal()

  function handleSave(title: string) {
    if (editingGoal) {
      updateGoal.mutate({ id: editingGoal.id, title })
    } else {
      createGoal.mutate({ title, periodType, periodKey })
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1">
          <button
            onClick={() => setPeriodKey(navigatePeriodKey(periodType, periodKey, -1))}
            className="p-1 rounded hover:bg-muted transition-colors text-muted-foreground hover:text-foreground"
            aria-label="Previous period"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
          <span className="text-sm font-medium w-32 text-center">
            {formatPeriodLabel(periodType, periodKey)}
          </span>
          <button
            onClick={() => setPeriodKey(navigatePeriodKey(periodType, periodKey, 1))}
            className="p-1 rounded hover:bg-muted transition-colors text-muted-foreground hover:text-foreground"
            aria-label="Next period"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
        <Button
          size="sm"
          onClick={() => setDialogOpen(true)}
          className="bg-foreground text-background hover:bg-foreground/80 gap-1.5"
        >
          <Plus className="w-3.5 h-3.5" />
          Add Goal
        </Button>
      </div>

      {goals.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center border border-dashed border-border rounded-xl">
          <p className="text-sm text-muted-foreground">No goals yet</p>
          <p className="text-xs text-muted-foreground/60 mt-1">
            Add a goal to track your progress this period
          </p>
        </div>
      ) : (
        <div className="flex flex-col gap-2">
          {goals.map((goal) => (
            <GoalItem
              key={goal.id}
              goal={goal}
              onToggle={(id) => toggleGoal.mutate({ id })}
              onDelete={(id) => deleteGoal.mutate(id)}
              onEdit={(g) => { setEditingGoal(g); setDialogOpen(true) }}
            />
          ))}
        </div>
      )}

      <GoalDialog
        open={dialogOpen}
        onOpenChange={(open) => { setDialogOpen(open); if (!open) setEditingGoal(null) }}
        goal={editingGoal}
        onSave={handleSave}
      />
    </div>
  )
}

export default function GoalsPage() {
  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <AppSidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <MobileHeader />
        <main className="flex-1 overflow-y-auto">
          <div className="max-w-5xl mx-auto px-5 py-8 md:px-8 md:py-12">
            <div className="mb-2">
              <h2 className="font-display text-2xl font-semibold tracking-wide text-foreground text-balance">
                Goals
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                Set intentions, track progress
              </p>
            </div>

            <div className="my-6 h-px bg-border" />

            <Tabs defaultValue="month">
              <TabsList className="mb-6">
                {PERIOD_TYPES.map(({ value, label }) => (
                  <TabsTrigger key={value} value={value}>
                    {label}
                  </TabsTrigger>
                ))}
              </TabsList>
              {PERIOD_TYPES.map(({ value }) => (
                <TabsContent key={value} value={value}>
                  <GoalTabPanel periodType={value} />
                </TabsContent>
              ))}
            </Tabs>
          </div>
        </main>
      </div>
    </div>
  )
}
