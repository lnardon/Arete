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
import { Button } from "@/components/ui/button"
import type { PomodoroProject } from "@/lib/types"

const COLOR_SWATCHES = ["#6366f1", "#ec4899", "#22c55e", "#f97316", "#0ea5e9", "#a855f7", "#eab308", "#ef4444"]

interface PomodoroProjectDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  project: PomodoroProject | null
  onSave: (name: string, color: string) => void
}

export function PomodoroProjectDialog({
  open,
  onOpenChange,
  project,
  onSave,
}: PomodoroProjectDialogProps) {
  const [name, setName] = useState("")
  const [color, setColor] = useState(COLOR_SWATCHES[0])

  useEffect(() => {
    if (open) {
      setName(project?.name ?? "")
      setColor(project?.color ?? COLOR_SWATCHES[0])
    }
  }, [open, project])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = name.trim()
    if (!trimmed) return
    onSave(trimmed, color)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md border-foreground/20">
        <DialogHeader>
          <DialogTitle className="font-display text-xl font-semibold tracking-tight">
            {project ? "Edit Project" : "New Project"}
          </DialogTitle>
          <DialogDescription>
            {project ? "Modify this project" : "Track time against a new project"}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="py-4 flex flex-col gap-4">
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Client work"
              className="border-foreground/20 focus-visible:ring-foreground/30 placeholder:text-muted-foreground/50"
              autoFocus
            />
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
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              className="border-foreground/20 text-foreground bg-transparent hover:text-destructive"
            >
              Cancel
            </Button>
            <Button type="submit" disabled={!name.trim()}>
              {project ? "Save Changes" : "Create Project"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
