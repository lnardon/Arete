export function formatLocalDate(date: Date): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, "0")
  const d = String(date.getDate()).padStart(2, "0")

  return `${y}-${m}-${d}`
}

export function getCurrentPeriodKey(type: 'month' | 'quarter' | 'semester' | 'year'): string {
  const now = new Date()
  const y = now.getFullYear()
  const mo = now.getMonth() + 1 // 1-12

  if (type === 'month') {
    return `${y}-${String(mo).padStart(2, '0')}`
  }

  if (type === 'quarter') {
    const q = Math.ceil(mo / 3)
    return `${y}-Q${q}`
  }

  if (type === 'semester') {
    return mo <= 6 ? `${y}-H1` : `${y}-H2`
  }

  return `${y}`
}

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

export function navigatePeriodKey(type: 'month' | 'quarter' | 'semester' | 'year', key: string, delta: number): string {
  if (type === 'month') {
    const [y, m] = key.split('-').map(Number)
    const d = new Date(y, m - 1 + delta, 1)

    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
  }

  if (type === 'quarter') {
    const y = parseInt(key)
    const q = parseInt(key.split('-Q')[1])
    const total = y * 4 + (q - 1) + delta
    const newY = Math.floor(total / 4)
    const newQ = (((total % 4) + 4) % 4) + 1

    return `${newY}-Q${newQ}`
  }

  if (type === 'semester') {
    const y = parseInt(key)
    const h = parseInt(key.split('-H')[1])
    const total = y * 2 + (h - 1) + delta
    const newY = Math.floor(total / 2)
    const newH = (((total % 2) + 2) % 2) + 1

    return `${newY}-H${newH}`
  }

  return `${parseInt(key) + delta}`
}

export function addDays(date: Date, amount: number): Date {
  const d = new Date(date)
  d.setDate(d.getDate() + amount)
  return d
}

// Monday-start week, to match the ISO week convention used elsewhere in the app.
export function getStartOfWeek(date: Date): Date {
  const d = new Date(date)
  const day = d.getDay() // 0 = Sunday
  const diff = day === 0 ? -6 : 1 - day
  d.setDate(d.getDate() + diff)
  d.setHours(0, 0, 0, 0)
  return d
}

export function getStartOfDay(date: Date): Date {
  const d = new Date(date)
  d.setHours(0, 0, 0, 0)
  return d
}

export function formatHourLabel(hour: number): string {
  if (hour === 0) return '12 AM'
  if (hour === 12) return '12 PM'
  return hour < 12 ? `${hour} AM` : `${hour - 12} PM`
}

export function formatWeekdayLabel(date: Date): string {
  return date.toLocaleDateString('en-US', { weekday: 'short' })
}

export function isSameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

export function formatPeriodLabel(type: 'month' | 'quarter' | 'semester' | 'year', key: string): string {
  if (type === 'month') {
    const [y, m] = key.split('-')
    return `${MONTH_NAMES[parseInt(m, 10) - 1]} ${y}` // 2026-03
  }

  if (type === 'quarter') {
    const [y, q] = key.split('-')
    return `${q} ${y}` // 2026-Q1
  }

  if (type === 'semester') {
    const [y, h] = key.split('-')
    return `${h} ${y}` // 2026-H1
  }

  return key // year
}
