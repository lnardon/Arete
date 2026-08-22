"use client"

import { useEffect, useMemo, useState } from "react"
import { Play, Square } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useActiveTimer, useProjects, useStartTimer, useStopTimer } from "@/hooks/use-pomodoro"
import { formatLocalDate } from "@/lib/date-utils"
import { cn } from "@/lib/utils"

const PRESETS = [25, 50]
const NO_PROJECT = "none"

function formatClock(totalSeconds: number): string {
  const sign = totalSeconds < 0 ? "+" : ""
  const abs = Math.abs(totalSeconds)
  const m = Math.floor(abs / 60)
  const s = abs % 60
  return `${sign}${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`
}

export function PomodoroTimerCard() {
  const { data: active } = useActiveTimer()
  const { data: projects = [] } = useProjects()
  const startTimer = useStartTimer()
  const stopTimer = useStopTimer()

  const [selectedProjectId, setSelectedProjectId] = useState<string>(NO_PROJECT)
  const [plannedMinutes, setPlannedMinutes] = useState(25)
  const [customMinutes, setCustomMinutes] = useState("")
  const [now, setNow] = useState(() => Date.now())

  const isActive = active?.active === true

  useEffect(() => {
    if (!isActive) return
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [isActive])

  const remainingSeconds = useMemo(() => {
    if (isActive) {
      const entry = active.entry
      const startedMs = new Date(entry.startedAt).getTime()
      const totalMs = entry.plannedMinutes * 60_000
      return Math.round((startedMs + totalMs - now) / 1000)
    }
    const minutes = customMinutes.trim() ? Number(customMinutes) : plannedMinutes
    return (Number.isFinite(minutes) ? minutes : 0) * 60
  }, [isActive, active, now, plannedMinutes, customMinutes])

  const activeProject = isActive
    ? projects.find((p) => p.id === active.entry.projectId)
    : projects.find((p) => p.id === selectedProjectId)

  function handleStart() {
    const minutes = customMinutes.trim() ? Number(customMinutes) : plannedMinutes
    if (!Number.isFinite(minutes) || minutes <= 0) return
    startTimer.mutate({
      projectId: selectedProjectId === NO_PROJECT ? null : selectedProjectId,
      plannedMinutes: minutes,
      localDate: formatLocalDate(new Date()),
    })
  }

  function handleStop() {
    stopTimer.mutate()
  }

  return (
    <Card className="border-foreground/20 h-full">
      <CardHeader>
        <CardTitle className="font-display text-lg font-semibold tracking-tight">
          {isActive ? "Timer running" : "Start a timer"}
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6 py-4">
        <div
          className="font-display text-6xl font-semibold tabular-nums tracking-tight"
          style={activeProject ? { color: activeProject.color } : undefined}
        >
          {formatClock(remainingSeconds)}
        </div>

        {activeProject && (
          <span className="text-sm text-muted-foreground">{activeProject.name}</span>
        )}

        {!isActive && (
          <div className="w-full flex flex-col gap-3 max-w-xs">
            <Select value={selectedProjectId} onValueChange={(value) => setSelectedProjectId(value ?? NO_PROJECT)}>
              <SelectTrigger className="w-full border-foreground/20">
                <SelectValue placeholder="No project" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={NO_PROJECT}>No project</SelectItem>
                {projects.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <div className="flex gap-2">
              {PRESETS.map((minutes) => (
                <Button
                  key={minutes}
                  type="button"
                  size="sm"
                  onClick={() => {
                    setPlannedMinutes(minutes)
                    setCustomMinutes("")
                  }}
                  className={cn(
                    "flex-1",
                    (plannedMinutes !== minutes || customMinutes) && "border-foreground/20 bg-transparent hover:bg-muted"
                  )}
                  variant={plannedMinutes === minutes && !customMinutes ? undefined : "outline"}
                >
                  {minutes}m
                </Button>
              ))}
              <Input
                type="number"
                min={1}
                placeholder="Custom"
                value={customMinutes}
                onChange={(e) => setCustomMinutes(e.target.value)}
                className="flex-1 border-foreground/20 placeholder:text-muted-foreground/50"
              />
            </div>
          </div>
        )}

        <Button
          size="lg"
          onClick={isActive ? handleStop : handleStart}
          className={cn("gap-2 px-8", isActive && "bg-destructive text-destructive-foreground hover:bg-destructive/80")}
        >
          {isActive ? <Square className="w-4 h-4" /> : <Play className="w-4 h-4" />}
          {isActive ? "Stop" : "Start"}
        </Button>
      </CardContent>
    </Card>
  )
}
