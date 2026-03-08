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
import type { Habit } from "@/lib/types"

interface DeleteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  habit: Habit | null
  onConfirm: () => void
}

export function DeleteDialog({
  open,
  onOpenChange,
  habit,
  onConfirm,
}: DeleteDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md border-foreground/20">
        <DialogHeader>
          <DialogTitle className="font-display text-xl font-semibold tracking-tight">
            Remove Habit
          </DialogTitle>
          <DialogDescription className="text-sm text-muted-foreground mt-2">
            {"Are you certain you wish to remove "}
            <span className="font-medium text-foreground">
              {habit?.name}
            </span>
            {"? All associated progress will be permanently lost."}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="mt-2">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="border-foreground/20 text-foreground bg-transparent hover:bg-foreground hover:text-background"
          >
            Keep Habit
          </Button>
          <Button
            onClick={() => {
              onConfirm()
              onOpenChange(false)
            }}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/80"
          >
            Remove
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
