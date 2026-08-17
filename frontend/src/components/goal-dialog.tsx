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
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"

import type { Goal, GoalType } from "@/lib/types"

export interface GoalSaveFields {
  title: string
  goalType?: GoalType
  targetValue?: number
  currentValue?: number
}

interface GoalDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  goal?: Goal | null
  onSave: (fields: GoalSaveFields) => void
}

export function GoalDialog({ open, onOpenChange, goal, onSave }: GoalDialogProps) {
  const [title, setTitle] = useState("")
  const [goalType, setGoalType] = useState<GoalType>("binary")
  const [targetValue, setTargetValue] = useState("")
  const [currentValue, setCurrentValue] = useState("")

  useEffect(() => {
    if (open) {
      setTitle(goal?.title ?? "")
      setGoalType(goal?.goalType ?? "binary")
      setTargetValue(goal?.targetValue != null ? String(goal.targetValue) : "")
      setCurrentValue(goal?.currentValue != null ? String(goal.currentValue) : "")
    }
  }, [open, goal])

  const isNumeric = goalType === "numeric"
  const parsedTarget = Number(targetValue)
  const targetValid = !isNumeric || (targetValue.trim() !== "" && Number.isFinite(parsedTarget) && parsedTarget > 0)
  const canSubmit = title.trim() !== "" && targetValid

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    const trimmed = title.trim()
    if (!trimmed || !targetValid) return

    if (!isNumeric) {
      onSave({ title: trimmed, goalType })
      onOpenChange(false)
      return
    }

    const fields: GoalSaveFields = { title: trimmed, goalType, targetValue: parsedTarget }
    if (goal) {
      const trimmedCurrent = currentValue.trim()
      if (trimmedCurrent !== "") {
        const parsedCurrent = Number(trimmedCurrent)
        if (Number.isFinite(parsedCurrent) && parsedCurrent >= 0) {
          fields.currentValue = parsedCurrent
        }
      }
    }
    onSave(fields)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md border-foreground/20">
        <DialogHeader>
          <DialogTitle className="font-display text-xl font-semibold tracking-tight">
            {goal ? "Edit Goal" : "New Goal"}
          </DialogTitle>
          <DialogDescription>
            {goal ? "Modify this goal" : "Define a goal for this period"}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="py-4 flex flex-col gap-4">
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="e.g. Read 3 books"
              className="border-foreground/20 focus-visible:ring-foreground/30 placeholder:text-muted-foreground/50"
              autoFocus
            />

            {!goal && (
              <RadioGroup
                value={goalType}
                onValueChange={(value) => setGoalType(value as GoalType)}
                className="flex gap-4"
              >
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="binary" id="goal-type-binary" />
                  <Label htmlFor="goal-type-binary" className="font-normal cursor-pointer">
                    Done / not done
                  </Label>
                </div>
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="numeric" id="goal-type-numeric" />
                  <Label htmlFor="goal-type-numeric" className="font-normal cursor-pointer">
                    Progress toward a number
                  </Label>
                </div>
              </RadioGroup>
            )}

            {isNumeric && (
              <div className="flex flex-col gap-3">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="goal-target" className="text-muted-foreground">
                    Target
                  </Label>
                  <Input
                    id="goal-target"
                    type="number"
                    min={1}
                    value={targetValue}
                    onChange={(e) => setTargetValue(e.target.value)}
                    placeholder="e.g. 222"
                    className="border-foreground/20 focus-visible:ring-foreground/30 placeholder:text-muted-foreground/50"
                  />
                </div>
                {goal && (
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="goal-current" className="text-muted-foreground">
                      Current progress
                    </Label>
                    <Input
                      id="goal-current"
                      type="number"
                      min={0}
                      value={currentValue}
                      onChange={(e) => setCurrentValue(e.target.value)}
                      className="border-foreground/20 focus-visible:ring-foreground/30"
                    />
                  </div>
                )}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              className="border-foreground/20 text-foreground bg-transparent hover:text-red-600"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={!canSubmit}
              className="bg-foreground text-background hover:bg-foreground/80"
            >
              {goal ? "Save Changes" : "Add Goal"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
