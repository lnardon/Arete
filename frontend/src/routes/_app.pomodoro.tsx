import { createFileRoute } from '@tanstack/react-router'
import { AppSidebar } from "@/components/app-sidebar"
import { MobileHeader } from "@/components/mobile-header"
import { PomodoroView } from "@/components/pomodoro-view"

export const Route = createFileRoute('/_app/pomodoro')({
  component: PomodoroPage,
})

export default function PomodoroPage() {
  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <AppSidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <MobileHeader />
        <PomodoroView />
      </div>
    </div>
  )
}
