"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import {
  REPEAT_PRESETS,
  WEEKDAYS,
  defaultCustom,
  presetLabel,
  type CustomRepeat,
  type RepeatEnd,
  type RepeatFreq,
  type RepeatPreset,
  type RepeatValue,
  type Weekday,
} from "@/lib/recurrence"
import { addDays, formatLocalDate } from "@/lib/date-utils"

const UNIT_LABELS: Record<RepeatFreq, string> = { DAILY: "days", WEEKLY: "weeks", MONTHLY: "months", YEARLY: "years" }
const WEEKDAY_INITIALS: Record<Weekday, string> = { MO: "M", TU: "T", WE: "W", TH: "T", FR: "F", SA: "S", SU: "S" }

interface RecurrencePickerProps {
  value: RepeatValue
  onChange: (value: RepeatValue) => void
  start: Date
  // Shown for a rule the picker can't express, e.g. one made in Google Calendar.
  keepLabel: string
}

export function RecurrencePicker({ value, onChange, start, keepLabel }: RecurrencePickerProps) {
  const selected = value.kind === "preset" ? value.preset : value.kind

  const items: Record<string, string> = { none: "Does not repeat" }
  for (const preset of REPEAT_PRESETS) items[preset] = presetLabel(preset, start)
  if (value.kind === "keep") items.keep = keepLabel
  items.custom = "Custom…"

  function handleSelect(next: string | null) {
    if (!next || next === selected) return
    if (next === "none") onChange({ kind: "none" })
    else if (next === "keep") onChange({ kind: "keep" })
    else if (next === "custom") onChange({ kind: "custom", custom: defaultCustom(start) })
    else onChange({ kind: "preset", preset: next as RepeatPreset })
  }

  return (
    <div className="flex flex-col gap-1.5">
      <Label className="text-muted-foreground">Repeat</Label>
      <Select value={selected} onValueChange={handleSelect} items={items}>
        <SelectTrigger className="w-full border-foreground/20">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {Object.entries(items).map(([key, label]) => (
            <SelectItem key={key} value={key}>
              {label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {value.kind === "custom" && (
        <CustomRepeatForm
          custom={value.custom}
          start={start}
          onChange={(custom) => onChange({ kind: "custom", custom })}
        />
      )}
    </div>
  )
}

interface CustomRepeatFormProps {
  custom: CustomRepeat
  start: Date
  onChange: (custom: CustomRepeat) => void
}

function CustomRepeatForm({ custom, start, onChange }: CustomRepeatFormProps) {
  const update = (patch: Partial<CustomRepeat>) => onChange({ ...custom, ...patch })
  const untilDate = custom.end.kind === "until" ? custom.end.date : formatLocalDate(addDays(start, 30))
  const count = custom.end.kind === "count" ? custom.end.count : 10

  function setEnd(kind: RepeatEnd["kind"]) {
    if (kind === "never") update({ end: { kind: "never" } })
    if (kind === "until") update({ end: { kind: "until", date: untilDate } })
    if (kind === "count") update({ end: { kind: "count", count } })
  }

  return (
    <div className="mt-1 flex flex-col gap-3 rounded-xl border border-foreground/10 p-3">
      <div className="flex items-center gap-2 text-sm">
        <span className="text-muted-foreground">Every</span>
        <Input
          type="number"
          min={1}
          max={99}
          value={custom.interval}
          onChange={(e) => update({ interval: Math.min(99, Math.max(1, Number(e.target.value) || 1)) })}
          className="w-16 border-foreground/20"
          aria-label="Repeat interval"
        />
        <Select
          value={custom.freq}
          onValueChange={(freq) => freq && update({ freq: freq as RepeatFreq })}
          items={UNIT_LABELS}
        >
          <SelectTrigger className="border-foreground/20">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {(Object.keys(UNIT_LABELS) as RepeatFreq[]).map((freq) => (
              <SelectItem key={freq} value={freq}>
                {UNIT_LABELS[freq]}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {custom.freq === "WEEKLY" && (
        <ToggleGroup
          multiple
          value={custom.weekdays}
          onValueChange={(days: Weekday[]) => update({ weekdays: days.length > 0 ? days : custom.weekdays })}
          spacing={1}
          aria-label="Repeat on"
        >
          {WEEKDAYS.map((day) => (
            <ToggleGroupItem
              key={day}
              value={day}
              aria-label={day}
              className="size-8 rounded-full p-0 text-xs aria-pressed:bg-primary aria-pressed:text-primary-foreground"
            >
              {WEEKDAY_INITIALS[day]}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
      )}

      <RadioGroup value={custom.end.kind} onValueChange={(kind) => setEnd(kind as RepeatEnd["kind"])} className="gap-2">
        <Label className="text-muted-foreground">Ends</Label>
        <div className="flex items-center gap-2">
          <RadioGroupItem value="never" id="repeat-end-never" />
          <Label htmlFor="repeat-end-never" className="font-normal cursor-pointer">
            Never
          </Label>
        </div>
        <div className="flex items-center gap-2">
          <RadioGroupItem value="until" id="repeat-end-until" />
          <Label htmlFor="repeat-end-until" className="font-normal cursor-pointer">
            On
          </Label>
          <Input
            type="date"
            value={untilDate}
            min={formatLocalDate(start)}
            disabled={custom.end.kind !== "until"}
            onChange={(e) => e.target.value && update({ end: { kind: "until", date: e.target.value } })}
            className="w-auto border-foreground/20"
            aria-label="Last day"
          />
        </div>
        <div className="flex items-center gap-2">
          <RadioGroupItem value="count" id="repeat-end-count" />
          <Label htmlFor="repeat-end-count" className="font-normal cursor-pointer">
            After
          </Label>
          <Input
            type="number"
            min={1}
            max={1000}
            value={count}
            disabled={custom.end.kind !== "count"}
            onChange={(e) =>
              update({ end: { kind: "count", count: Math.min(1000, Math.max(1, Number(e.target.value) || 1)) } })
            }
            className="w-20 border-foreground/20"
            aria-label="Number of occurrences"
          />
          <span className="text-sm text-muted-foreground">occurrences</span>
        </div>
      </RadioGroup>
    </div>
  )
}
