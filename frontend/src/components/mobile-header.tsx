"use client"

import { Link } from "@tanstack/react-router"
import { Home, BarChart2, Trophy } from "lucide-react"
import { ThemeSwitcher } from "@/components/theme-switcher"
import { SignOutButton } from "@/components/sign-out-button"

export function MobileHeader() {
  return (
    <header className="flex md:hidden items-center justify-between px-4 py-4 border-b border-border bg-background/80 backdrop-blur-md text-foreground">
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
      <nav className="flex items-center gap-1">
        <Link
          to="/"
          activeProps={{ className: "text-foreground bg-accent" }}
          activeOptions={{ exact: true }}
          className="p-2 rounded-md text-muted-foreground hover:text-foreground transition-colors"
        >
          <Home className="w-4 h-4" />
        </Link>
        <Link
          to="/statistics"
          activeProps={{ className: "text-foreground bg-accent" }}
          className="p-2 rounded-md text-muted-foreground hover:text-foreground transition-colors"
        >
          <BarChart2 className="w-4 h-4" />
        </Link>
        <Link
          to="/achievements"
          activeProps={{ className: "text-foreground bg-accent" }}
          className="p-2 rounded-md text-muted-foreground hover:text-foreground transition-colors"
        >
          <Trophy className="w-4 h-4" />
        </Link>
      </nav>
      <div className="flex items-center gap-1">
        <ThemeSwitcher variant="header" />
        <SignOutButton variant="header" />
      </div>
    </header>
  )
}
