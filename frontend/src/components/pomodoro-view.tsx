"use client"

import { PomodoroTimerCard } from "@/components/pomodoro-timer-card"
import { PomodoroProjectsPanel } from "@/components/pomodoro-projects-panel"
import { PomodoroChart } from "@/components/pomodoro-chart"
import { PomodoroHistoryList } from "@/components/pomodoro-history-list"

export function PomodoroView() {
  return (
    <main className="flex-1 overflow-y-auto bg-background">
      <div className="max-w-5xl mx-auto px-5 py-8 md:px-8 md:py-12">
        <div className="mb-6">
          <h1 className="font-display text-2xl font-semibold tracking-tight">Pomodoro</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Track focused time against your projects
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          <div className="lg:col-span-2">
            <PomodoroTimerCard />
          </div>
          <PomodoroProjectsPanel />
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <PomodoroChart />
          <PomodoroHistoryList />
        </div>
      </div>
    </main>
  )
}
