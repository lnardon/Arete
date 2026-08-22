"use client"

interface DailyProgressProps {
  completed: number
  total: number
}

export function DailyProgress({ completed, total }: DailyProgressProps) {
  const percentage = total > 0 ? Math.round((completed / total) * 100) : 0

  return (
    <div className="flex flex-col gap-3 px-5 py-4 border border-border rounded-lg bg-card">
      <div className="flex items-baseline justify-between">
        <p className="label-section">
          Daily Progress
        </p>
        <p className="font-display text-lg font-semibold text-foreground">
          {completed}
          <span className="text-muted-foreground">
            {" / "}
            {total}
          </span>
        </p>
      </div>
      <div className="h-1.5 rounded-full bg-border overflow-hidden">
        <div
          className="h-full rounded-full bg-secondary transition-all duration-500 ease-out"
          style={{ width: `${percentage}%` }}
        />
      </div>
      {total > 0 && (
        <p className="text-xs text-muted-foreground">
          {percentage === 100
            ? "All habits completed. Virtue achieved."
            : `${percentage}% of daily habits fulfilled`}
        </p>
      )}
    </div>
  )
}
