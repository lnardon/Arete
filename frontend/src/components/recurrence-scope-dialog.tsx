"use client"

import { useState } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import type { RecurrenceScope } from "@/lib/types"

const SCOPE_LABELS: Record<RecurrenceScope, string> = {
  this: "This event",
  following: "This and following events",
  all: "All events",
}

interface RecurrenceScopeDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  mode: "save" | "delete"
  eventTitle: string
  // False when the repeat rule itself changed: a single occurrence can't have
  // its own rule, so only "following" and "all" make sense (as in Google).
  allowThis?: boolean
  onConfirm: (scope: RecurrenceScope) => void
}

export function RecurrenceScopeDialog({
  open,
  onOpenChange,
  mode,
  eventTitle,
  allowThis = true,
  onConfirm,
}: RecurrenceScopeDialogProps) {
  const scopes: RecurrenceScope[] = allowThis ? ["this", "following", "all"] : ["following", "all"]
  // Each prompt starts on the first option; a pick only lasts until it closes.
  const [picked, setPicked] = useState<RecurrenceScope | null>(null)
  const scope = picked && scopes.includes(picked) ? picked : scopes[0]

  function handleOpenChange(next: boolean) {
    if (!next) setPicked(null)
    onOpenChange(next)
  }

  const isDelete = mode === "delete"

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-sm border-foreground/20">
        <DialogHeader>
          <DialogTitle className="font-display text-xl font-semibold tracking-tight">
            {isDelete ? "Delete recurring event" : "Edit recurring event"}
          </DialogTitle>
          <DialogDescription>
            <span className="font-medium text-foreground">{eventTitle}</span>
            {isDelete ? " repeats. Which events do you want to delete?" : " repeats. Which events should change?"}
          </DialogDescription>
        </DialogHeader>
        <RadioGroup value={scope} onValueChange={(value) => setPicked(value as RecurrenceScope)} className="gap-3">
          {scopes.map((s) => (
            <div key={s} className="flex items-center gap-2">
              <RadioGroupItem value={s} id={`recurrence-scope-${s}`} />
              <Label htmlFor={`recurrence-scope-${s}`} className="font-normal cursor-pointer">
                {SCOPE_LABELS[s]}
              </Label>
            </div>
          ))}
        </RadioGroup>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            className="border-foreground/20 text-foreground bg-transparent"
          >
            Cancel
          </Button>
          <Button
            onClick={() => {
              onConfirm(scope)
              handleOpenChange(false)
            }}
            className={isDelete ? "bg-destructive text-destructive-foreground hover:bg-destructive/80" : undefined}
          >
            {isDelete ? "Delete" : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
