"use client"

import type { Achievement } from "@/lib/achievements"
import { cn } from "@/lib/utils"
import {
  Columns3,
  Crown,
  Flame,
  Scroll,
  Shield,
  Sun,
  Sword,
  Check,
  CalendarDays,
  CircleDot,
} from "lucide-react"
import type { LucideIcon } from "lucide-react"

const iconMap: Record<string, LucideIcon> = {
  pillar: Columns3,
  check: Check,
  temple: Columns3,
  sun: Sun,
  flame: Flame,
  shield: Shield,
  sword: Sword,
  scroll: Scroll,
  helmet: Shield,
  eagle: Crown,
  laurel: Crown,
  crown: Crown,
  wreath: Crown,
  columns: Columns3,
  calendar: CalendarDays,
  rings: CircleDot,
}

export function AchievementCard({
  achievement,
}: {
  achievement: Achievement
}) {
  const Icon = iconMap[achievement.icon] || Shield
  const percentage = Math.round(
    (achievement.progress / achievement.goal) * 100
  )

  return (
    <div
      className={cn(
        "border rounded-lg p-5 flex flex-col gap-4 transition-colors",
        achievement.unlocked
          ? "bg-card border-secondary"
          : "bg-card border-border text-foreground opacity-60"
      )}
    >
      <div className="flex items-start justify-between">
        <div
          className={cn(
            "flex items-center justify-center w-10 h-10 rounded-lg border",
            achievement.unlocked
              ? "bg-secondary border-secondary text-secondary-foreground"
              : "border-border"
          )}
        >
          <Icon
            className={cn("w-5 h-5", !achievement.unlocked && "text-muted-foreground")}
            strokeWidth={1.5}
          />
        </div>
        {achievement.unlocked && (
          <span className="inline-flex items-center gap-1 rounded-md border border-primary bg-primary px-2 py-0.5 text-xs font-medium text-primary-foreground">
            ✓ Achieved
          </span>
        )}
      </div>

      <div className="flex flex-col gap-1">
        <h3 className="font-display text-base font-semibold text-foreground">
          {achievement.name}
        </h3>
        <p className="text-[11px] text-muted-foreground">
          {achievement.description}
        </p>
      </div>

      <div className="flex flex-col gap-2 mt-auto">
        <div className="flex items-center justify-between">
          <span className="text-[10px] text-muted-foreground">
            {achievement.progress} / {achievement.goal}
          </span>
          <span className="text-[10px] text-muted-foreground">
            {percentage}%
          </span>
        </div>
        <div className="w-full h-1.5 rounded-full bg-muted">
          <div
            className={cn(
              "h-full rounded-full transition-all duration-500",
              achievement.unlocked ? "bg-secondary" : "bg-foreground/70"
            )}
            style={{ width: `${percentage}%` }}
          />
        </div>
      </div>
    </div>
  )
}
