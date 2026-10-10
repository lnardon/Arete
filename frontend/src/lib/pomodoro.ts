import type { PomodoroEntry } from '@/lib/types'

// mm:ss, with a leading "+" once the timer has run past its planned length.
export function formatClock(totalSeconds: number): string {
  const sign = totalSeconds < 0 ? '+' : ''
  const abs = Math.abs(totalSeconds)
  const m = Math.floor(abs / 60)
  const s = abs % 60
  return `${sign}${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

// Seconds left in a running entry; negative once it's in overtime.
export function getRemainingSeconds(entry: PomodoroEntry, nowMs: number): number {
  const startedMs = new Date(entry.startedAt).getTime()
  const totalMs = entry.plannedMinutes * 60_000
  return Math.round((startedMs + totalMs - nowMs) / 1000)
}
