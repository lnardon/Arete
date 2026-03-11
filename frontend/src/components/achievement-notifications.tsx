import { useMemo, useRef, useEffect } from "react"
import { toast } from "sonner"
import { useHabits } from "@/hooks/use-habits"
import { useCompletionsForRange } from "@/hooks/use-completions"
import { computeAchievements } from "@/lib/achievements"
import { formatLocalDate } from "@/lib/date-utils"

export function AchievementNotifications() {
  const today = formatLocalDate(new Date())
  const startDate = new Date()
  startDate.setDate(startDate.getDate() - 364)
  const start = formatLocalDate(startDate)

  const { data: habits = [], isSuccess: habitsReady } = useHabits()
  const { data: completions = [], isSuccess: completionsReady } = useCompletionsForRange(start, today)

  const achievements = useMemo(
    () => computeAchievements(habits, completions),
    [habits, completions]
  )

  const initialized = useRef(false)
  const prevUnlockedIds = useRef<Set<string>>(new Set())

  useEffect(() => {
    if (!habitsReady || !completionsReady) return

    const currentUnlocked = new Set(
      achievements.filter((a) => a.unlocked).map((a) => a.id)
    )

    if (initialized.current) {
      for (const id of currentUnlocked) {
        if (!prevUnlockedIds.current.has(id)) {
          const achievement = achievements.find((a) => a.id === id)
          if (achievement) {
            toast.success("Achievement Unlocked!", {
              description: achievement.name,
            })
          }
        }
      }
    }

    initialized.current = true
    prevUnlockedIds.current = currentUnlocked
  }, [achievements, habitsReady, completionsReady])

  return null
}
