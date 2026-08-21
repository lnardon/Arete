"use client"

import { Link, useNavigate, useRouterState } from "@tanstack/react-router"
import { LayoutDashboard, ListChecks, BarChart2, Trophy, Target, BookOpen, Settings, Menu, Moon, Sun, LogOut } from "lucide-react"
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

  const isDashboard = pathname === "/"
  const isHabits = pathname === "/habits"
  const isGoals = pathname === "/goals"
  const isJournal = pathname === "/journal"
  const isStats = pathname === "/statistics"
  const isAchievements = pathname === "/achievements"
  const isSettings = pathname === "/settings"

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
        <div className="w-10 h-10 rounded-sm bg-foreground/10 flex items-center justify-center">
          <img src="/logo.png" alt="Arete Logo" className="w-full h-full" />
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
          <DropdownMenuItem className={isDashboard ? "font-semibold" : ""}>
            <Link to="/" className="flex items-center gap-2 w-full">
              <LayoutDashboard className="w-4 h-4" />
              Dashboard
              {isDashboard && <span className="ml-auto w-1.5 h-1.5 rounded-full bg-primary" />}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem className={isHabits ? "font-semibold" : ""}>
            <Link to="/habits" className="flex items-center gap-2 w-full">
              <ListChecks className="w-4 h-4" />
              Habits
              {isHabits && <span className="ml-auto w-1.5 h-1.5 rounded-full bg-primary" />}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem className={isGoals ? "font-semibold" : ""}>
            <Link to="/goals" className="flex items-center gap-2 w-full">
              <Target className="w-4 h-4" />
              Goals
              {isGoals && <span className="ml-auto w-1.5 h-1.5 rounded-full bg-primary" />}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem className={isJournal ? "font-semibold" : ""}>
            <Link to="/journal" className="flex items-center gap-2 w-full">
              <BookOpen className="w-4 h-4" />
              Journal
              {isJournal && <span className="ml-auto w-1.5 h-1.5 rounded-full bg-primary" />}
            </Link>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
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
          <DropdownMenuItem className={isSettings ? "font-semibold" : ""}>
            <Link to="/settings" className="flex items-center gap-2 w-full">
              <Settings className="w-4 h-4" />
              Settings
              {isSettings && <span className="ml-auto w-1.5 h-1.5 rounded-full bg-primary" />}
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
