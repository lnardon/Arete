"use client"
import React from "react"
import { useState, useEffect } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Button } from "@/components/ui/button"
import { Trash2 } from "lucide-react"
import type { CalendarEvent, CalendarEventInput } from "@/lib/types"
import { addDays, formatLocalDate } from "@/lib/date-utils"

const COLOR_SWATCHES = ["#6366f1", "#ec4899", "#22c55e", "#f97316", "#0ea5e9", "#a855f7", "#eab308", "#ef4444"]

function toDatetimeLocalValue(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

interface CalendarEventDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  event: CalendarEvent | null
  defaultStart?: Date
  onSave: (input: CalendarEventInput) => void
  onRequestDelete?: () => void
}

export function CalendarEventDialog({
  open,
  onOpenChange,
  event,
  defaultStart,
  onSave,
  onRequestDelete,
}: CalendarEventDialogProps) {
  const [title, setTitle] = useState("")
  const [description, setDescription] = useState("")
  const [location, setLocation] = useState("")
  const [allDay, setAllDay] = useState(false)
  const [startValue, setStartValue] = useState("")
  const [endValue, setEndValue] = useState("")
  const [color, setColor] = useState(COLOR_SWATCHES[0])

  useEffect(() => {
    if (!open) return

    if (event) {
      setTitle(event.title)
      setDescription(event.description ?? "")
      setLocation(event.location ?? "")
      setAllDay(event.allDay)
      const start = new Date(event.startAt)
      const end = new Date(event.endAt)
      if (event.allDay) {
        setStartValue(formatLocalDate(start))
        // The stored end is exclusive (Google Calendar convention for all-day
        // events); show the user the inclusive last day instead.
        setEndValue(formatLocalDate(addDays(end, -1)))
      } else {
        setStartValue(toDatetimeLocalValue(start))
        setEndValue(toDatetimeLocalValue(end))
      }
      setColor(event.color)
    } else {
      const start = defaultStart ?? new Date()
      const end = new Date(start.getTime() + 60 * 60 * 1000)
      setTitle("")
      setDescription("")
      setLocation("")
      setAllDay(false)
      setStartValue(toDatetimeLocalValue(start))
      setEndValue(toDatetimeLocalValue(end))
      setColor(COLOR_SWATCHES[0])
    }
  }, [open, event, defaultStart])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmedTitle = title.trim()
    if (!trimmedTitle || !startValue || !endValue) return

    let startAt: Date
    let endAt: Date
    if (allDay) {
      startAt = new Date(`${startValue}T00:00:00`)
      endAt = addDays(new Date(`${endValue}T00:00:00`), 1)
    } else {
      startAt = new Date(startValue)
      endAt = new Date(endValue)
    }
    if (endAt <= startAt) return

    onSave({
      title: trimmedTitle,
      description: description.trim() || null,
      location: location.trim() || null,
      startAt: startAt.toISOString(),
      endAt: endAt.toISOString(),
      allDay,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      recurrenceRule: null,
      color,
    })
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md border-foreground/20">
        <DialogHeader>
          <DialogTitle className="font-display text-xl font-semibold tracking-tight">
            {event ? "Edit Event" : "New Event"}
          </DialogTitle>
          <DialogDescription>
            {event ? "Update this appointment" : "Add an appointment to your calendar"}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="py-4 flex flex-col gap-4">
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="e.g. Dentist appointment"
              className="border-foreground/20 focus-visible:ring-foreground/30 placeholder:text-muted-foreground/50"
              autoFocus
            />
            <Textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Notes (optional)"
              className="border-foreground/20 focus-visible:ring-foreground/30 placeholder:text-muted-foreground/50"
              rows={2}
            />
            <Input
              value={location}
              onChange={(e) => setLocation(e.target.value)}
              placeholder="Location (optional)"
              className="border-foreground/20 focus-visible:ring-foreground/30 placeholder:text-muted-foreground/50"
            />

            <div className="flex items-center justify-between">
              <Label className="text-muted-foreground">All day</Label>
              <Switch checked={allDay} onCheckedChange={setAllDay} />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <Label className="text-muted-foreground">Starts</Label>
                <Input
                  type={allDay ? "date" : "datetime-local"}
                  value={startValue}
                  onChange={(e) => setStartValue(e.target.value)}
                  className="border-foreground/20 focus-visible:ring-foreground/30"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label className="text-muted-foreground">Ends</Label>
                <Input
                  type={allDay ? "date" : "datetime-local"}
                  value={endValue}
                  onChange={(e) => setEndValue(e.target.value)}
                  className="border-foreground/20 focus-visible:ring-foreground/30"
                />
              </div>
            </div>

            <div className="flex flex-col gap-1.5">
              <Label className="text-muted-foreground">Color</Label>
              <div className="flex gap-2">
                {COLOR_SWATCHES.map((swatch) => (
                  <button
                    key={swatch}
                    type="button"
                    onClick={() => setColor(swatch)}
                    className="w-6 h-6 rounded-full border-2 transition-transform"
                    style={{
                      backgroundColor: swatch,
                      borderColor: color === swatch ? "var(--foreground)" : "transparent",
                      transform: color === swatch ? "scale(1.1)" : "scale(1)",
                    }}
                    aria-label={swatch}
                  />
                ))}
              </div>
            </div>
          </div>
          <DialogFooter className="sm:justify-between">
            {event && onRequestDelete ? (
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  onOpenChange(false)
                  onRequestDelete()
                }}
                className="border-foreground/20 text-destructive bg-transparent hover:bg-destructive/10"
              >
                <Trash2 className="w-4 h-4 mr-1.5" />
                Delete
              </Button>
            ) : (
              <span />
            )}
            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                className="border-foreground/20 text-foreground bg-transparent hover:text-destructive"
              >
                Cancel
              </Button>
              <Button type="submit" disabled={!title.trim()}>
                {event ? "Save Changes" : "Create Event"}
              </Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
