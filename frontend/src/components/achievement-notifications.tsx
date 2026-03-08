import { useMemo, useRef, useEffect } from "react"
import { toast } from "sonner"
import { useHabits } from "@/hooks/use-habits"
import { useCompletionsForRange } from "@/hooks/use-completions"
import { computeAchievements } from "@/lib/achievements"

export function AchievementNotifications() {
  const today = new Date().toISOString().split("T")[0]
  const startDate = new Date()
  startDate.setDate(startDate.getDate() - 364)
  const start = startDate.toISOString().split("T")[0]

  const { data: habits = [] } = useHabits()
  const { data: completions = [] } = useCompletionsForRange(start, today)

  const achievements = useMemo(
    () => computeAchievements(habits, completions),
    [habits, completions]
  )

  const prevUnlockedIds = useRef<Set<string> | null>(null)

  useEffect(() => {
    const currentUnlocked = new Set(
      achievements.filter((a) => a.unlocked).map((a) => a.id)
    )

    if (prevUnlockedIds.current !== null) {
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

    prevUnlockedIds.current = currentUnlocked
  }, [achievements])

  return null
}
