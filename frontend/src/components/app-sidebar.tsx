import { BarChart3, BookOpen, LayoutDashboard, ListChecks, Settings, Target, Timer, Trophy } from "lucide-react"
import { Link, useRouterState } from "@tanstack/react-router"
import { cn } from "@/lib/utils"
import { ThemeSwitcher } from "@/components/theme-switcher"
import { SignOutButton } from "@/components/sign-out-button"

interface NavItem {
  icon: typeof LayoutDashboard
  label: string
  href: string
}

const trunkItem: NavItem = { icon: LayoutDashboard, label: "Dashboard", href: "/" }

const branchItems: NavItem[] = [
  { icon: ListChecks, label: "Habits", href: "/habits" },
  { icon: Target, label: "Goals", href: "/goals" },
  { icon: BookOpen, label: "Journal", href: "/journal" },
  { icon: Timer, label: "Pomodoro", href: "/pomodoro" },
]

const extraItems: NavItem[] = [
  { icon: BarChart3, label: "Statistics", href: "/statistics" },
  { icon: Trophy, label: "Achievements", href: "/achievements" },
  { icon: Settings, label: "Settings", href: "/settings" },
]

function NavLink({ item, isActive }: { item: NavItem; isActive: boolean }) {
  return (
    <Link
      to={item.href}
      className={cn(
        "flex items-center gap-3 w-full px-3 py-2.5 text-sm tracking-wide transition-colors rounded-md",
        isActive
          ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium"
          : "text-sidebar-foreground/70 hover:text-sidebar-foreground hover:bg-sidebar-accent/60"
      )}
    >
      <item.icon className="w-4 h-4 shrink-0" strokeWidth={1.5} />
      <span>{item.label}</span>
    </Link>
  )
}

export function AppSidebar() {
  const pathname = useRouterState({ select: (s) => s.location.pathname }) ?? "/"

  return (
    <aside className="hidden md:flex md:w-64 flex-col bg-sidebar text-sidebar-foreground border-r border-sidebar-border">
      <div className="flex items-center gap-3 px-6 py-8">
        <div className="w-10 h-10 rounded-sm bg-foreground/10 flex items-center justify-center">
          <img src="/logo.png" alt="Arete Logo" className="w-full h-full" />
        </div>
        <div>
          <h1 className="text-lg font-display font-semibold">
            Arete
          </h1>
          <p className="text-[10px] text-sidebar-foreground/50">
            Daily Discipline
          </p>
        </div>
      </div>

      <nav className="flex-1 px-3 py-4 overflow-y-auto">
        <p className="px-3 mb-3 label-section text-sidebar-foreground/50">
          Overview
        </p>

        <NavLink item={trunkItem} isActive={pathname === trunkItem.href} />

        <div className="relative ml-5 pl-4 mt-0.5 border-l border-sidebar-border">
          <span className="absolute -left-[3px] top-0 w-1.5 h-1.5 -translate-x-1/2 rounded-full bg-sidebar-border" />
          <ul className="flex flex-col gap-0.5">
            {branchItems.map((item) => {
              const isActive = pathname === item.href
              return (
                <li key={item.href} className="relative">
                  <span className="absolute -left-4 top-1/2 -translate-y-1/2 w-3 h-px bg-sidebar-border" />
                  <NavLink item={item} isActive={isActive} />
                </li>
              )
            })}
          </ul>
        </div>

        <div className="my-4 h-px bg-sidebar-border" />

        <p className="px-3 mb-3 label-section text-sidebar-foreground/50">
          Extras
        </p>
        <ul className="flex flex-col gap-0.5">
          {extraItems.map((item) => (
            <li key={item.href}>
              <NavLink item={item} isActive={pathname === item.href} />
            </li>
          ))}
        </ul>
      </nav>

      <div className="px-4 py-4 border-t border-sidebar-border space-y-2">
        <ThemeSwitcher variant="sidebar" />
        <SignOutButton variant="sidebar" />
        <div className="px-2">
          <p className="text-[11px] text-sidebar-foreground/45 leading-relaxed">
            {"\"We are what we repeatedly do.\""}
          </p>
          <p className="mt-0.5 text-[11px] text-sidebar-foreground/35">
            — Aristotle
          </p>
        </div>
      </div>
    </aside>
  )
}
