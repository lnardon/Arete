import { RRule } from 'rrule'
import { formatLocalDate } from '@/lib/date-utils'

// Recurrence rules use RFC 5545 lines, exactly as Google Calendar stores them
// ("RRULE:FREQ=WEEKLY;BYDAY=FR"). The picker works on a RepeatValue and only
// builds lines at save time, from the event's start date, so a preset like
// "Weekly on Friday" follows the start when the user moves it.

export const WEEKDAYS = ['MO', 'TU', 'WE', 'TH', 'FR', 'SA', 'SU'] as const
export type Weekday = (typeof WEEKDAYS)[number]

const JS_DAY_TO_WEEKDAY: Weekday[] = ['SU', 'MO', 'TU', 'WE', 'TH', 'FR', 'SA']
const WEEKDAY_NAMES: Record<Weekday, string> = {
  MO: 'Monday', TU: 'Tuesday', WE: 'Wednesday', TH: 'Thursday', FR: 'Friday', SA: 'Saturday', SU: 'Sunday',
}

export type RepeatPreset = 'daily' | 'weekly' | 'monthly' | 'yearly' | 'weekdays'
export const REPEAT_PRESETS: RepeatPreset[] = ['daily', 'weekly', 'monthly', 'yearly', 'weekdays']

export type RepeatFreq = 'DAILY' | 'WEEKLY' | 'MONTHLY' | 'YEARLY'

export type RepeatEnd =
  | { kind: 'never' }
  | { kind: 'until'; date: string } // YYYY-MM-DD, inclusive
  | { kind: 'count'; count: number }

export interface CustomRepeat {
  freq: RepeatFreq
  interval: number
  weekdays: Weekday[] // weekly only
  end: RepeatEnd
}

export type RepeatValue =
  | { kind: 'none' }
  | { kind: 'preset'; preset: RepeatPreset }
  | { kind: 'custom'; custom: CustomRepeat }
  // A rule the picker can't express (typically made in Google Calendar),
  // left exactly as it is unless the user picks something else.
  | { kind: 'keep' }

export function weekdayOf(date: Date): Weekday {
  return JS_DAY_TO_WEEKDAY[date.getDay()]
}

// Which occurrence of its weekday the date is in its month: 1-4, or -1 for
// the fifth, which not every month has, so it becomes "last".
function weekdayOrdinal(date: Date): number {
  const n = Math.ceil(date.getDate() / 7)
  return n === 5 ? -1 : n
}

const ORDINALS: Record<number, string> = { 1: 'first', 2: 'second', 3: 'third', 4: 'fourth', [-1]: 'last' }

export function presetRule(preset: RepeatPreset, start: Date): string {
  switch (preset) {
    case 'daily':
      return 'FREQ=DAILY'
    case 'weekly':
      return `FREQ=WEEKLY;BYDAY=${weekdayOf(start)}`
    case 'monthly':
      return `FREQ=MONTHLY;BYDAY=${weekdayOrdinal(start)}${weekdayOf(start)}`
    case 'yearly':
      return 'FREQ=YEARLY'
    case 'weekdays':
      return 'FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR'
  }
}

export function presetLabel(preset: RepeatPreset, start: Date): string {
  const weekday = WEEKDAY_NAMES[weekdayOf(start)]
  switch (preset) {
    case 'daily':
      return 'Daily'
    case 'weekly':
      return `Weekly on ${weekday}`
    case 'monthly':
      return `Monthly on the ${ORDINALS[weekdayOrdinal(start)]} ${weekday}`
    case 'yearly':
      return `Annually on ${start.toLocaleDateString('en-US', { month: 'long', day: 'numeric' })}`
    case 'weekdays':
      return 'Every weekday (Monday to Friday)'
  }
}

export function defaultCustom(start: Date): CustomRepeat {
  return { freq: 'WEEKLY', interval: 1, weekdays: [weekdayOf(start)], end: { kind: 'never' } }
}

// UNTIL is a date on all-day events and a UTC date-time otherwise (RFC 5545):
// the end of the chosen day in the user's zone, so that day is included.
function untilStamp(date: string, allDay: boolean): string {
  if (allDay) return date.replaceAll('-', '')
  return new Date(`${date}T23:59:59`).toISOString().replace(/[-:]/g, '').replace(/\.\d{3}/, '')
}

export function customRule(custom: CustomRepeat, allDay: boolean): string {
  const parts = [`FREQ=${custom.freq}`]
  if (custom.interval > 1) parts.push(`INTERVAL=${custom.interval}`)
  if (custom.freq === 'WEEKLY' && custom.weekdays.length > 0) {
    parts.push(`BYDAY=${WEEKDAYS.filter((d) => custom.weekdays.includes(d)).join(',')}`)
  }
  if (custom.end.kind === 'until') parts.push(`UNTIL=${untilStamp(custom.end.date, allDay)}`)
  if (custom.end.kind === 'count') parts.push(`COUNT=${custom.end.count}`)
  return parts.join(';')
}

const isRRuleLine = (line: string) => /^RRULE:/i.test(line)

// Compares two rules part by part, ignoring order, case and WKST, so a rule
// Google wrote ("FREQ=WEEKLY;WKST=SU;BYDAY=FR") matches the preset it is.
function normalizeRule(rule: string): string {
  return rule
    .replace(/^RRULE:/i, '')
    .toUpperCase()
    .split(';')
    .filter((part) => part && !part.startsWith('WKST='))
    .sort()
    .join(';')
}

// Builds the lines to save. Only the RRULE line is replaced: EXDATE/RDATE
// lines that came from Google are kept, and a rule equivalent to the existing
// one is sent back verbatim so the server sees no change to the series.
export function buildRecurrence(value: RepeatValue, start: Date, allDay: boolean, existing: string[] | null): string[] | null {
  if (value.kind === 'none') return null
  if (value.kind === 'keep') return existing

  const rule = value.kind === 'preset' ? presetRule(value.preset, start) : customRule(value.custom, allDay)
  const current = existing?.find(isRRuleLine)
  if (current && normalizeRule(current) === normalizeRule(rule)) return existing
  return [`RRULE:${rule}`, ...(existing ?? []).filter((line) => !isRRuleLine(line))]
}

function ruleParts(rule: string): Map<string, string> | null {
  const parts = new Map<string, string>()
  for (const part of rule.replace(/^RRULE:/i, '').split(';')) {
    const [key, value] = part.split('=')
    if (!key || !value) return null
    parts.set(key.toUpperCase(), value.toUpperCase())
  }
  return parts
}

function untilToDate(until: string): string | null {
  const m = until.match(/^(\d{4})(\d{2})(\d{2})(?:T(\d{2})(\d{2})(\d{2})(Z)?)?$/)
  if (!m) return null
  const [, y, mo, d, h, mi, s, z] = m
  if (!h) return `${y}-${mo}-${d}`
  const date = z
    ? new Date(Date.UTC(+y, +mo - 1, +d, +h, +mi, +s))
    : new Date(+y, +mo - 1, +d, +h, +mi, +s)
  return formatLocalDate(date)
}

// Reads the custom form back out of a rule, or null when the rule uses parts
// the form has no controls for.
function toCustom(rule: string): CustomRepeat | null {
  const parts = ruleParts(rule)
  if (!parts) return null
  const allowed = new Set(['FREQ', 'INTERVAL', 'BYDAY', 'COUNT', 'UNTIL', 'WKST'])
  if ([...parts.keys()].some((key) => !allowed.has(key))) return null

  const freq = parts.get('FREQ') as RepeatFreq
  if (!['DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY'].includes(freq)) return null
  const interval = Number(parts.get('INTERVAL') ?? 1)

  let weekdays: Weekday[] = []
  const byday = parts.get('BYDAY')
  if (byday) {
    if (freq !== 'WEEKLY') return null
    const days = byday.split(',')
    if (!days.every((d): d is Weekday => (WEEKDAYS as readonly string[]).includes(d))) return null
    weekdays = days
  }

  let end: RepeatEnd = { kind: 'never' }
  const count = parts.get('COUNT')
  const until = parts.get('UNTIL')
  if (count) end = { kind: 'count', count: Number(count) }
  if (until) {
    const date = untilToDate(until)
    if (!date) return null
    end = { kind: 'until', date }
  }
  return { freq, interval, weekdays, end }
}

// Works out what the picker shows for an existing series.
export function parseRepeat(lines: string[] | null, start: Date): RepeatValue {
  const rrules = (lines ?? []).filter(isRRuleLine)
  if (rrules.length === 0) return { kind: 'none' }
  if (rrules.length > 1) return { kind: 'keep' }

  const rule = normalizeRule(rrules[0])
  const preset = REPEAT_PRESETS.find((p) => normalizeRule(presetRule(p, start)) === rule)
  if (preset) return { kind: 'preset', preset }

  const custom = toCustom(rrules[0])
  return custom ? { kind: 'custom', custom } : { kind: 'keep' }
}

// Human-readable text for a rule, e.g. "Every week on Friday".
export function summarizeRecurrence(lines: string[] | null): string {
  const rrule = lines?.find(isRRuleLine)
  if (!rrule) return 'Does not repeat'
  try {
    const text = RRule.fromString(rrule.replace(/^RRULE:/i, '')).toText()
    return text.charAt(0).toUpperCase() + text.slice(1)
  } catch {
    return 'Custom repeat rule'
  }
}
