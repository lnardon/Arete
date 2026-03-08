"use client"

import { Moon, Sun } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useTheme } from "@/components/theme-provider"
import { cn } from "@/lib/utils"

interface ThemeSwitcherProps {
  className?: string
  variant?: "sidebar" | "header"
}

export function ThemeSwitcher({ className, variant = "sidebar" }: ThemeSwitcherProps) {
  const { theme, toggleTheme } = useTheme()

  const isSidebar = variant === "sidebar"

  return (
    <Button
      type="button"
      variant="ghost"
      size={isSidebar ? "sm" : "icon-sm"}
      onClick={toggleTheme}
      className={cn(
        "shrink-0",
        isSidebar
          ? "w-full justify-start gap-3 px-3 py-2.5 text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
          : "text-foreground hover:bg-muted",
        className
      )}
      aria-label={theme === "dark" ? "Switch to light theme" : "Switch to dark theme"}
    >
      {theme === "dark" ? (
        <Sun className="h-4 w-4" strokeWidth={1.5} />
      ) : (
        <Moon className="h-4 w-4" strokeWidth={1.5} />
      )}
      {isSidebar && (
        <span className="text-sm tracking-wide">
          {theme === "dark" ? "Light" : "Dark"}
        </span>
      )}
    </Button>
  )
}
