"use client"

import { Link, useNavigate, useRouterState } from "@tanstack/react-router"
import { Home, BarChart2, Trophy, Menu, Moon, Sun, LogOut } from "lucide-react"
import { useTheme } from "@/components/theme-provider"
import { useAuth } from "@/lib/auth"
import { useQueryClient } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export function MobileHeader() {
  const { theme, toggleTheme } = useTheme()
  const { logout } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  const isHome = pathname === "/"
  const isStats = pathname === "/statistics"
  const isAchievements = pathname === "/achievements"

  async function handleSignOut() {
    await api.auth.logout()
    logout()
    queryClient.clear()
    navigate({ to: "/login" })
  }

  return (
    <header className="flex md:hidden items-center justify-between px-4 py-4 border-b border-border bg-background text-foreground">
      <Link
        to="/"
        className="flex items-center gap-3 hover:opacity-90 transition-opacity"
      >
        <div className="w-8 h-8 rounded-xl bg-foreground/8 flex items-center justify-center">
          <svg
            viewBox="0 0 24 24"
            fill="none"
            className="w-4 h-4"
            stroke="currentColor"
            strokeWidth="1.5"
          >
            <path d="M12 2L4 6v6c0 5.25 3.4 10.15 8 11.25C16.6 22.15 20 17.25 20 12V6l-8-4z" />
            <path d="M9 12l2 2 4-4" />
          </svg>
        </div>
        <div>
          <h1 className="text-sm font-display font-semibold">
            Arete
          </h1>
          <p className="text-[9px] text-muted-foreground">
            Daily Discipline
          </p>
        </div>
      </Link>
      <DropdownMenu>
        <DropdownMenuTrigger className="cursor-pointer hover:bg-muted dark:hover:bg-muted/50 inline-flex items-center justify-center size-8 rounded-lg transition-all outline-none">
          <Menu className="w-5 h-5" />
          <span className="sr-only">Open menu</span>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-48 text-sm border border-white p-2">
          <DropdownMenuItem className={isHome ? "font-semibold" : ""}>
            <Link to="/" className="flex items-center gap-2 w-full">
              <Home className="w-4 h-4" />
              Today
              {isHome && <span className="ml-auto w-1.5 h-1.5 rounded-full bg-primary" />}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem className={isStats ? "font-semibold" : ""}>
            <Link to="/statistics" className="flex items-center gap-2 w-full">
              <BarChart2 className="w-4 h-4" />
              Statistics
              {isStats && <span className="ml-auto w-1.5 h-1.5 rounded-full bg-primary" />}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem className={isAchievements ? "font-semibold" : ""}>
            <Link to="/achievements" className="flex items-center gap-2 w-full">
              <Trophy className="w-4 h-4" />
              Achievements
              {isAchievements && <span className="ml-auto w-1.5 h-1.5 rounded-full bg-primary" />}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem className="cursor-pointer" onClick={toggleTheme}>
            {theme === "dark" ? (
              <Sun className="w-4 h-4" strokeWidth={1.5} />
            ) : (
              <Moon className="w-4 h-4" strokeWidth={1.5} />
            )}
            {theme === "dark" ? "Light mode" : "Dark mode"}
          </DropdownMenuItem>
          <DropdownMenuItem variant="destructive" onClick={handleSignOut}>
            <LogOut className="w-4 h-4" strokeWidth={1.5} />
            Sign out
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}
