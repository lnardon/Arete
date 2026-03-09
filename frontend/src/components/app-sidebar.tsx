import { BarChart3, LayoutDashboard, Trophy } from "lucide-react"
import { Link, useRouterState } from "@tanstack/react-router"
import { cn } from "@/lib/utils"
import { ThemeSwitcher } from "@/components/theme-switcher"
import { SignOutButton } from "@/components/sign-out-button"

const navItems = [
  { icon: LayoutDashboard, label: "Today", href: "/" },
  { icon: BarChart3, label: "Statistics", href: "/statistics" },
  { icon: Trophy, label: "Achievements", href: "/achievements" },
]

export function AppSidebar() {
  const pathname = useRouterState({ select: (s) => s.location.pathname }) ?? "/"

  return (
    <aside className="hidden md:flex md:w-64 flex-col bg-sidebar text-sidebar-foreground border-r border-sidebar-border">
      <div className="flex items-center gap-3 px-6 py-8">
        <div className="w-9 h-9 rounded-xl bg-foreground/8 flex items-center justify-center">
          <svg
            viewBox="0 0 24 24"
            fill="none"
            className="w-5 h-5"
            stroke="currentColor"
            strokeWidth="1.5"
          >
            <path d="M12 2L4 6v6c0 5.25 3.4 10.15 8 11.25C16.6 22.15 20 17.25 20 12V6l-8-4z" />
            <path d="M9 12l2 2 4-4" />
          </svg>
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

      <nav className="flex-1 px-3 py-4">
        <p className="px-3 mb-3 label-section text-sidebar-foreground/50">
          Navigation
        </p>
        <ul className="flex flex-col gap-0.5">
          {navItems.map((item) => {
            const isActive = pathname === item.href
            return (
              <li key={item.label}>
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
              </li>
            )
          })}
        </ul>
      </nav>

      <div className="px-4 py-4 border-t border-sidebar-border space-y-3">
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
