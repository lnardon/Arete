import { createFileRoute } from '@tanstack/react-router'
import { AppSidebar } from "@/components/app-sidebar"
import { MobileHeader } from "@/components/mobile-header"
import { HabitList } from "@/components/habit-list"

export const Route = createFileRoute('/_app/')({
  component: AppPage,
})

export default function AppPage() {
  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <AppSidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <MobileHeader />
        <HabitList />
      </div>
    </div>
  )
}
