"use client"

import React, { useState, useEffect } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

interface GoalDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (title: string) => void
}

export function GoalDialog({ open, onOpenChange, onSave }: GoalDialogProps) {
  const [title, setTitle] = useState("")

  useEffect(() => {
    if (open) setTitle("")
  }, [open])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    const trimmed = title.trim()
    if (!trimmed) return

    onSave(trimmed)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md border-foreground/20">
        <DialogHeader>
          <DialogTitle className="font-display text-xl font-semibold tracking-tight">
            New Goal
          </DialogTitle>
          <DialogDescription className="text-xs tracking-[0.15em] uppercase text-muted-foreground">
            Define a goal for this period
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="py-4">
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="e.g. Read 3 books"
              className="border-foreground/20 focus-visible:ring-foreground/30 placeholder:text-muted-foreground/50"
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              className="border-foreground/20 text-foreground bg-transparent hover:bg-foreground hover:text-background"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={!title.trim()}
              className="bg-foreground text-background hover:bg-foreground/80"
            >
              Add Goal
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
