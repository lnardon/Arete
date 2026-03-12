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
