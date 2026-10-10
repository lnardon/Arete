import { useEffect, useState } from "react"
import { Link } from "@tanstack/react-router"
import { useProjects } from "@/hooks/use-pomodoro"
import { formatClock, getRemainingSeconds } from "@/lib/pomodoro"
import { cn } from "@/lib/utils"
import type { PomodoroEntry } from "@/lib/types"

// Live countdown for a running timer. It ticks every second on its own so the
// rest of the dashboard only re-renders on the minute.
export function DashboardPomodoroChip({ entry }: { entry: PomodoroEntry }) {
  const { data: projects = [] } = useProjects()
  const [nowMs, setNowMs] = useState(() => Date.now())

  useEffect(() => {
    const id = setInterval(() => setNowMs(Date.now()), 1000)
    return () => clearInterval(id)
  }, [])

  const project = projects.find((p) => p.id === entry.projectId)
  const remaining = getRemainingSeconds(entry, nowMs)
  const overtime = remaining < 0

  return (
    <Link
      to="/pomodoro"
      aria-label="Pomodoro timer running — open Pomodoro"
      className="inline-flex items-center gap-2.5 rounded-full border border-border bg-card card-elevated px-3.5 py-1.5 text-sm hover:bg-muted/40 transition-colors"
    >
      <span
        className={cn("size-2 shrink-0 rounded-full motion-safe:animate-pulse", !project && "bg-primary")}
        style={project ? { backgroundColor: project.color } : undefined}
      />
      <span className="text-muted-foreground truncate max-w-40">
        {overtime ? "Overtime" : project?.name ?? "Focus"}
      </span>
      <span className="font-display font-semibold tabular-nums text-foreground">
        {formatClock(remaining)}
      </span>
    </Link>
  )
}
