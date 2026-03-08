import type { Habit, HabitCompletion } from "@/lib/types"

export interface Achievement {
  id: string
  name: string
  description: string
  category: "foundation" | "streak" | "volume" | "mastery" | "collection"
  icon: string
  unlocked: boolean
  progress: number
  goal: number
}

function formatDateKey(d: Date): string {
  return d.toISOString().split("T")[0]
}

function computeCurrentStreak(
  habits: Array<Habit>,
  completions: Array<HabitCompletion>
): number {
  if (habits.length === 0) return 0
  const total = habits.length
  const today = new Date()
  today.setHours(12, 0, 0, 0)
  let streak = 0
  const cursor = new Date(today)
  // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
  while (true) {
    const key = formatDateKey(cursor)
    const done = completions.filter((c) => c.date === key).length
    if (done === total) {
      streak++
      cursor.setDate(cursor.getDate() - 1)
    } else {
      break
    }
  }
  return streak
}

function computeBestStreak(
  habits: Array<Habit>,
  completions: Array<HabitCompletion>
): number {
  if (habits.length === 0) return 0
  const total = habits.length
  const uniqueDates = Array.from(new Set(completions.map((c) => c.date))).sort()
  let best = 0
  let current = 0
  let prevDate: Date | null = null
  for (const dateStr of uniqueDates) {
    const done = completions.filter((c) => c.date === dateStr).length
    const thisDate = new Date(dateStr + "T12:00:00")
    if (done === total) {
      if (
        prevDate &&
        thisDate.getTime() - prevDate.getTime() === 86400000
      ) {
        current++
      } else {
        current = 1
      }
      best = Math.max(best, current)
      prevDate = thisDate
    } else {
      current = 0
      prevDate = null
    }
  }
  return best
}

function countPerfectDays(
  habits: Array<Habit>,
  completions: Array<HabitCompletion>
): number {
  if (habits.length === 0) return 0
  const total = habits.length
  const uniqueDates = new Set(completions.map((c) => c.date))
  let count = 0
  for (const date of uniqueDates) {
    const done = completions.filter((c) => c.date === date).length
    if (done >= total) count++
  }
  return count
}

export function computeAchievements(
  habits: Array<Habit>,
  completions: Array<HabitCompletion>
): Array<Achievement> {
  const totalHabits = habits.length
  const totalCompletions = completions.length
  const currentStreak = computeCurrentStreak(habits, completions)
  const bestStreak = computeBestStreak(habits, completions)
  const perfectDays = countPerfectDays(habits, completions)
  const uniqueActiveDays = new Set(completions.map((c) => c.date)).size

  return [
    {
      id: "first-step",
      name: "Primum Passum",
      description: "Create your first habit",
      category: "foundation",
      icon: "pillar",
      unlocked: totalHabits >= 1,
      progress: Math.min(totalHabits, 1),
      goal: 1,
    },
    {
      id: "first-mark",
      name: "Prima Nota",
      description: "Complete a habit for the first time",
      category: "foundation",
      icon: "check",
      unlocked: totalCompletions >= 1,
      progress: Math.min(totalCompletions, 1),
      goal: 1,
    },
    {
      id: "the-foundation",
      name: "Fundamentum",
      description: "Create three habits to track",
      category: "foundation",
      icon: "temple",
      unlocked: totalHabits >= 3,
      progress: Math.min(totalHabits, 3),
      goal: 3,
    },
    {
      id: "perfect-day",
      name: "Dies Perfecta",
      description: "Complete all habits in a single day",
      category: "mastery",
      icon: "sun",
      unlocked: perfectDays >= 1,
      progress: Math.min(perfectDays, 1),
      goal: 1,
    },
    {
      id: "steadfast",
      name: "Constantia",
      description: "Maintain a 3-day streak of all habits",
      category: "streak",
      icon: "flame",
      unlocked: bestStreak >= 3,
      progress: Math.min(currentStreak > 0 ? currentStreak : bestStreak, 3),
      goal: 3,
    },
    {
      id: "unwavering",
      name: "Perseverantia",
      description: "Maintain a 7-day streak of all habits",
      category: "streak",
      icon: "shield",
      unlocked: bestStreak >= 7,
      progress: Math.min(currentStreak > 0 ? currentStreak : bestStreak, 7),
      goal: 7,
    },
    {
      id: "stoic-resolve",
      name: "Virtus Stoica",
      description: "Maintain a 14-day streak of all habits",
      category: "streak",
      icon: "sword",
      unlocked: bestStreak >= 14,
      progress: Math.min(currentStreak > 0 ? currentStreak : bestStreak, 14),
      goal: 14,
    },
    {
      id: "philosopher",
      name: "Via Philosophi",
      description: "Maintain a 30-day streak of all habits",
      category: "streak",
      icon: "scroll",
      unlocked: bestStreak >= 30,
      progress: Math.min(currentStreak > 0 ? currentStreak : bestStreak, 30),
      goal: 30,
    },
    {
      id: "centurion",
      name: "Centurio",
      description: "Reach 100 total habit completions",
      category: "volume",
      icon: "helmet",
      unlocked: totalCompletions >= 100,
      progress: Math.min(totalCompletions, 100),
      goal: 100,
    },
    {
      id: "legionnaire",
      name: "Legionarius",
      description: "Reach 500 total habit completions",
      category: "volume",
      icon: "eagle",
      unlocked: totalCompletions >= 500,
      progress: Math.min(totalCompletions, 500),
      goal: 500,
    },
    {
      id: "tribune",
      name: "Tribunus",
      description: "Reach 1,000 total habit completions",
      category: "volume",
      icon: "laurel",
      unlocked: totalCompletions >= 1000,
      progress: Math.min(totalCompletions, 1000),
      goal: 1000,
    },
    {
      id: "decathlon",
      name: "Decathlon",
      description: "Achieve 10 perfect days",
      category: "mastery",
      icon: "crown",
      unlocked: perfectDays >= 10,
      progress: Math.min(perfectDays, 10),
      goal: 10,
    },
    {
      id: "month-mastery",
      name: "Imperator Mensis",
      description: "Achieve 30 perfect days",
      category: "mastery",
      icon: "wreath",
      unlocked: perfectDays >= 30,
      progress: Math.min(perfectDays, 30),
      goal: 30,
    },
    {
      id: "pantheon",
      name: "Pantheon",
      description: "Create five different habits",
      category: "collection",
      icon: "columns",
      unlocked: totalHabits >= 5,
      progress: Math.min(totalHabits, 5),
      goal: 5,
    },
    {
      id: "devoted",
      name: "Devotio",
      description: "Track habits for 30 different days",
      category: "collection",
      icon: "calendar",
      unlocked: uniqueActiveDays >= 30,
      progress: Math.min(uniqueActiveDays, 30),
      goal: 30,
    },
    {
      id: "olympiad",
      name: "Olympias",
      description: "Create ten different habits",
      category: "collection",
      icon: "rings",
      unlocked: totalHabits >= 10,
      progress: Math.min(totalHabits, 10),
      goal: 10,
    },
  ]
}
