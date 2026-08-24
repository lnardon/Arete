"use client"

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import type { CalendarEvent } from "@/lib/types"

interface CalendarEventDeleteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  event: CalendarEvent | null
  onConfirm: () => void
}

export function CalendarEventDeleteDialog({
  open,
  onOpenChange,
  event,
  onConfirm,
}: CalendarEventDeleteDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md border-foreground/20">
        <DialogHeader>
          <DialogTitle className="font-display text-xl font-semibold tracking-tight">
            Delete Event
          </DialogTitle>
          <DialogDescription className="text-sm text-muted-foreground mt-2">
            {"Are you certain you wish to delete "}
            <span className="font-medium text-foreground">{event?.title}</span>
            {"? This cannot be undone."}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="mt-2">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="border-foreground/20 text-foreground bg-transparent"
          >
            Keep Event
          </Button>
          <Button
            onClick={() => {
              onConfirm()
              onOpenChange(false)
            }}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/80"
          >
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
