import { useEffect, useState } from 'react'

const MINUTE_MS = 60_000

// Current time, refreshed on each minute boundary and when the tab becomes
// visible again (background tabs throttle timers, so a stale value would linger).
export function useNow(): Date {
  const [now, setNow] = useState(() => new Date())

  useEffect(() => {
    let timeout: ReturnType<typeof setTimeout>
    const tick = () => {
      const d = new Date()
      setNow(d)
      timeout = setTimeout(tick, MINUTE_MS - (d.getTime() % MINUTE_MS))
    }
    timeout = setTimeout(tick, MINUTE_MS - (Date.now() % MINUTE_MS))

    const onVisibilityChange = () => {
      if (document.visibilityState !== 'visible') return
      clearTimeout(timeout)
      tick()
    }
    document.addEventListener('visibilitychange', onVisibilityChange)

    return () => {
      clearTimeout(timeout)
      document.removeEventListener('visibilitychange', onVisibilityChange)
    }
  }, [])

  return now
}
