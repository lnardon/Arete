"use client"

import { useState } from "react"
import { CalendarToolbar, type CalendarViewMode } from "@/components/calendar-toolbar"
import { CalendarDayView } from "@/components/calendar-day-view"
import { CalendarWeekView } from "@/components/calendar-week-view"
import { CalendarEventDialog } from "@/components/calendar-event-dialog"
import { CalendarEventDeleteDialog } from "@/components/calendar-event-delete-dialog"
import {
  useCalendarEvents,
  useCreateCalendarEvent,
  useUpdateCalendarEvent,
  useDeleteCalendarEvent,
} from "@/hooks/use-calendar-events"
import { addDays, getStartOfDay, getStartOfWeek } from "@/lib/date-utils"
import type { CalendarEvent, CalendarEventInput } from "@/lib/types"

export function CalendarView() {
  const [selectedDate, setSelectedDate] = useState(new Date())
  const [viewMode, setViewMode] = useState<CalendarViewMode>("day")
  const [eventDialogOpen, setEventDialogOpen] = useState(false)
  const [editingEvent, setEditingEvent] = useState<CalendarEvent | null>(null)
  const [defaultStart, setDefaultStart] = useState<Date | undefined>(undefined)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deletingEvent, setDeletingEvent] = useState<CalendarEvent | null>(null)

  const rangeStart = viewMode === "day" ? getStartOfDay(selectedDate) : getStartOfWeek(selectedDate)
  const rangeEnd = viewMode === "day" ? addDays(rangeStart, 1) : addDays(rangeStart, 7)

  const { data: events = [], isLoading } = useCalendarEvents(rangeStart.toISOString(), rangeEnd.toISOString())
  const createEvent = useCreateCalendarEvent()
  const updateEvent = useUpdateCalendarEvent()
  const deleteEvent = useDeleteCalendarEvent()

  function handlePrevious() {
    setSelectedDate((d) => addDays(d, viewMode === "day" ? -1 : -7))
  }

  function handleNext() {
    setSelectedDate((d) => addDays(d, viewMode === "day" ? 1 : 7))
  }

  function handleToday() {
    setSelectedDate(new Date())
  }

  function handleCreate() {
    setEditingEvent(null)
    setDefaultStart(undefined)
    setEventDialogOpen(true)
  }

  function handleSlotClick(start: Date) {
    setEditingEvent(null)
    setDefaultStart(start)
    setEventDialogOpen(true)
  }

  function handleEventClick(event: CalendarEvent) {
    setEditingEvent(event)
    setEventDialogOpen(true)
  }

  function handleSave(input: CalendarEventInput) {
    if (editingEvent) {
      updateEvent.mutate({ id: editingEvent.id, event: input })
    } else {
      createEvent.mutate(input)
    }
  }

  function handleRequestDelete() {
    setDeletingEvent(editingEvent)
    setDeleteDialogOpen(true)
  }

  function handleConfirmDelete() {
    if (deletingEvent) {
      deleteEvent.mutate(deletingEvent.id)
    }
  }

  return (
    <main className="flex-1 overflow-y-auto bg-background">
      <div className="p-16">
        <CalendarToolbar
          selectedDate={selectedDate}
          viewMode={viewMode}
          onViewModeChange={setViewMode}
          onPrevious={handlePrevious}
          onNext={handleNext}
          onToday={handleToday}
          onCreate={handleCreate}
        />

        <div className="my-6 h-px bg-border" />

        {isLoading ? (
          <div className="h-96 rounded-xl bg-muted animate-pulse" />
        ) : (
          <div className="overflow-x-auto">
            <div className="min-w-[560px]">
              {viewMode === "day" ? (
                <CalendarDayView
                  date={selectedDate}
                  events={events}
                  onSlotClick={handleSlotClick}
                  onEventClick={handleEventClick}
                />
              ) : (
                <CalendarWeekView
                  weekStart={rangeStart}
                  events={events}
                  onSlotClick={handleSlotClick}
                  onEventClick={handleEventClick}
                />
              )}
            </div>
          </div>
        )}

        <CalendarEventDialog
          open={eventDialogOpen}
          onOpenChange={setEventDialogOpen}
          event={editingEvent}
          defaultStart={defaultStart}
          onSave={handleSave}
          onRequestDelete={editingEvent ? handleRequestDelete : undefined}
        />

        <CalendarEventDeleteDialog
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          event={deletingEvent}
          onConfirm={handleConfirmDelete}
        />
      </div>
    </main>
  )
}
