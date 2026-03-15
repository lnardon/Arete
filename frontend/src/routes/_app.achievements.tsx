import { useMemo } from "react"
import { createFileRoute } from '@tanstack/react-router'
import type { Achievement } from "@/lib/achievements"
import { AppSidebar } from "@/components/app-sidebar"
import { MobileHeader } from "@/components/mobile-header"
import { AchievementCard } from "@/components/achievement-card"
import { useHabits } from "@/hooks/use-habits"
import { useCompletionsForRange } from "@/hooks/use-completions"
import { computeAchievements } from "@/lib/achievements"
import { formatLocalDate } from "@/lib/date-utils"

export const Route = createFileRoute('/_app/achievements')({
  component: AchievementsPage,
})

const categoryLabels: Record<string, string> = {
  foundation: "Foundation",
  streak: "Streaks",
  volume: "Volume",
  mastery: "Mastery",
  collection: "Collection",
}

const categoryDescriptions: Record<string, string> = {
  foundation: "Begin your journey with the first steps of discipline",
  streak: "The mark of consistency and unwavering resolve",
  volume: "Strength forged through repetition",
  mastery: "Perfection achieved through daily devotion",
  collection: "Breadth of ambition and diverse pursuit",
}

const categoryOrder = ["foundation", "streak", "volume", "mastery", "collection"]

export default function AchievementsPage() {
  const today = formatLocalDate(new Date())
  const startDate = new Date()
  startDate.setDate(startDate.getDate() - 364)
  const start = formatLocalDate(startDate)

  const { data: habits = [] } = useHabits()
  const { data: completions = [] } = useCompletionsForRange(start, today)

  const achievements = useMemo(
    () => computeAchievements(habits, completions),
    [habits, completions]
  )

  const unlocked = achievements.filter((a: any) => a.unlocked).length
  const total = achievements.length

  const grouped = useMemo(() => {
    const map: Record<string, Array<Achievement>> = {}
    for (const a of achievements) {
      if (!map[a.category]) map[a.category] = []
      map[a.category].push(a)
    }
    return map
  }, [achievements])

  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <AppSidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <MobileHeader />
        <main className="flex-1 overflow-y-auto">
          <div className="max-w-5xl mx-auto px-5 py-8 md:px-8 md:py-12">
            <div className="mb-2">
              <h2 className="font-display text-2xl font-semibold tracking-wide text-foreground text-balance">
                Achievements
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                Glory earned through discipline
              </p>
            </div>

            <div className="my-6 h-px bg-border" />

            <div className="flex items-center justify-between mb-8 border border-border rounded-2xl bg-card p-5">
              <div className="flex flex-col gap-1">
                <p className="label-section">
                  Total Progress
                </p>
                <p className="font-display text-3xl font-semibold text-foreground">
                  {unlocked}
                  <span className="text-lg text-muted-foreground">
                    {" / "}
                    {total}
                  </span>
                </p>
              </div>
              <div className="flex flex-col items-end gap-2">
                <p className="text-xs text-muted-foreground">
                  {total > 0
                    ? `${Math.round((unlocked / total) * 100)}% Complete`
                    : "Begin your journey"}
                </p>
                <div className="w-32 h-1.5 rounded-full bg-border">
                  <div
                    className="h-full rounded-full bg-foreground transition-all duration-500"
                    style={{
                      width: `${total > 0 ? (unlocked / total) * 100 : 0}%`,
                    }}
                  />
                </div>
              </div>
            </div>

            <div className="flex flex-col gap-10">
              {categoryOrder.map((cat) => {
                const items = grouped[cat]
                if (!items || items.length === 0) return null
                return (
                  <section key={cat}>
                    <div className="mb-4">
                      <h3 className="font-display text-lg font-semibold tracking-wide text-foreground">
                        {categoryLabels[cat]}
                      </h3>
                      <p className="text-sm text-muted-foreground mt-0.5">
                        {categoryDescriptions[cat]}
                      </p>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                      {items.map((a) => (
                        <AchievementCard key={a.id} achievement={a} />
                      ))}
                    </div>
                  </section>
                )
              })}
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}
