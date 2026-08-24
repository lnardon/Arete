import { createFileRoute } from '@tanstack/react-router'
import { AppSidebar } from "@/components/app-sidebar"
import { MobileHeader } from "@/components/mobile-header"
import { CalendarView } from "@/components/calendar-view"

export const Route = createFileRoute('/_app/calendar')({
  component: CalendarPage,
})

export default function CalendarPage() {
  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <AppSidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <MobileHeader />
        <CalendarView />
      </div>
    </div>
  )
}
