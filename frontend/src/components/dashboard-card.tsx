import type { ReactNode } from "react"
import { Link } from "@tanstack/react-router"
import { ArrowRight } from "lucide-react"
import { cn } from "@/lib/utils"

interface DashboardCardProps {
  title: string
  href: string
  linkLabel?: string
  // Next to the title, e.g. a "2/6" count.
  meta?: ReactNode
  // Before the link, e.g. the Google sync status.
  actions?: ReactNode
  className?: string
  children: ReactNode
}

export function DashboardCard({
  title,
  href,
  linkLabel = "View all",
  meta,
  actions,
  className,
  children,
}: DashboardCardProps) {
  return (
    <section className={cn("border border-border rounded-2xl bg-card p-5 card-elevated flex flex-col gap-4", className)}>
      <div className="flex flex-wrap items-center justify-between gap-x-3 gap-y-1">
        <div className="flex items-baseline gap-2 min-w-0">
          <h3 className="font-display text-base font-semibold tracking-wide text-foreground truncate">
            {title}
          </h3>
          {meta}
        </div>
        <div className="ml-auto flex flex-wrap items-center justify-end gap-x-3 gap-y-1">
          {actions}
          <Link
            to={href}
            className="text-xs text-muted-foreground hover:text-foreground inline-flex items-center gap-1 transition-colors"
          >
            {linkLabel}
            <ArrowRight className="w-3 h-3" />
          </Link>
        </div>
      </div>
      {children}
    </section>
  )
}

export function DashboardCount({ done, total }: { done: number; total: number }) {
  return (
    <span className="font-display text-sm font-semibold text-foreground tabular-nums shrink-0">
      {done}
      <span className="text-muted-foreground">{"/"}{total}</span>
    </span>
  )
}

export function DashboardSkeletonRows() {
  return (
    <div className="flex flex-col gap-2">
      {[1, 2, 3].map((i) => (
        <div key={i} className="h-10 rounded-lg bg-muted animate-pulse" />
      ))}
    </div>
  )
}

export function DashboardEmpty({ message, href, linkLabel }: { message: string; href: string; linkLabel: string }) {
  return (
    <div className="flex flex-col items-center gap-1 py-4 text-center">
      <p className="text-sm text-muted-foreground">{message}</p>
      <Link
        to={href}
        className="text-xs text-muted-foreground hover:text-foreground inline-flex items-center gap-1 transition-colors"
      >
        {linkLabel}
        <ArrowRight className="w-3 h-3" />
      </Link>
    </div>
  )
}
