import { LogOut } from "lucide-react"
import { useNavigate } from "@tanstack/react-router"
import { useQueryClient } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import { useAuth } from "@/lib/auth"
import { cn } from "@/lib/utils"

interface SignOutButtonProps {
  variant?: "sidebar" | "header"
}

export function SignOutButton({ variant = "sidebar" }: SignOutButtonProps) {
  const { logout } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  async function handleSignOut() {
    await api.auth.logout()
    logout()
    queryClient.clear()
    navigate({ to: '/login' })
  }

  if (variant === "header") {
    return (
      <button
        onClick={handleSignOut}
        className="p-2 rounded-md text-muted-foreground hover:text-foreground transition-colors"
        aria-label="Sign out"
      >
        <LogOut className="w-4 h-4" />
      </button>
    )
  }

  return (
    <button
      onClick={handleSignOut}
      className={cn(
        "flex items-center gap-3 w-full px-3 py-2.5 text-sm tracking-wide transition-colors rounded-md",
        "text-sidebar-foreground/70 hover:text-sidebar-foreground hover:bg-sidebar-accent/60"
      )}
    >
      <LogOut className="w-4 h-4 shrink-0" strokeWidth={1.5} />
      <span>Sign out</span>
    </button>
  )
}
