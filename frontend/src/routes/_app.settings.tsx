import { useEffect, useState } from "react"
import { createFileRoute } from '@tanstack/react-router'
import { MessageCircle } from "lucide-react"
import { AppSidebar } from "@/components/app-sidebar"
import { MobileHeader } from "@/components/mobile-header"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import { useWhatsAppStatus, useGenerateLinkCode, useUnlinkWhatsApp } from "@/hooks/use-whatsapp"
import type { WhatsAppLinkCode } from "@/lib/types"

export const Route = createFileRoute('/_app/settings')({
  component: SettingsPage,
})

const ASSISTANT_NUMBER = import.meta.env.VITE_WHATSAPP_NUMBER as string | undefined

export default function SettingsPage() {
  const [pendingCode, setPendingCode] = useState<WhatsAppLinkCode | null>(null)
  const [unlinkOpen, setUnlinkOpen] = useState(false)
  const [secondsLeft, setSecondsLeft] = useState(0)

  const { data: status, isLoading } = useWhatsAppStatus({
    // Stop polling once the pending code either links the account or expires
    // — read from the query's own latest data, not component state, so this
    // doesn't need a state-syncing effect.
    refetchInterval: (query) => (pendingCode && !query.state.data?.linked ? 3000 : false),
  })
  const generateCode = useGenerateLinkCode()
  const unlink = useUnlinkWhatsApp()

  useEffect(() => {
    if (!pendingCode) return
    const tick = () => {
      const remaining = Math.max(
        0,
        Math.floor((new Date(pendingCode.expiresAt).getTime() - Date.now()) / 1000)
      )
      setSecondsLeft(remaining)
      if (remaining <= 0) setPendingCode(null)
    }
    tick()
    const interval = setInterval(tick, 1000)
    return () => clearInterval(interval)
  }, [pendingCode])

  function handleConnect() {
    generateCode.mutate(undefined, {
      onSuccess: (data) => setPendingCode(data),
    })
  }

  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <AppSidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <MobileHeader />
        <main className="flex-1 overflow-y-auto">
          <div className="p-16">
            <div className="mb-2">
              <h2 className="font-display text-2xl font-semibold tracking-wide text-foreground text-balance">
                Settings
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                Manage integrations and connected accounts
              </p>
            </div>

            <div className="my-6 h-px bg-border" />

            <Card className="max-w-xl">
              <CardHeader>
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-lg bg-foreground/10 flex items-center justify-center shrink-0">
                    <MessageCircle className="w-5 h-5" strokeWidth={1.5} />
                  </div>
                  <div>
                    <CardTitle>WhatsApp Assistant</CardTitle>
                    <CardDescription>
                      Log habits, manage goals, and ask for summaries right from WhatsApp
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                {isLoading ? (
                  <div className="h-14 rounded-lg bg-muted animate-pulse" />
                ) : status?.linked ? (
                  <div className="flex items-center justify-between rounded-lg border border-border bg-muted/40 px-4 py-3">
                    <div>
                      <p className="text-sm font-medium text-foreground">{status.phoneNumberMasked}</p>
                      <p className="text-xs text-muted-foreground">
                        Connected {new Date(status.linkedAt).toLocaleDateString()}
                      </p>
                    </div>
                    <Button variant="destructive" size="sm" onClick={() => setUnlinkOpen(true)}>
                      Disconnect
                    </Button>
                  </div>
                ) : pendingCode ? (
                  <div className="flex flex-col gap-3">
                    <p className="text-sm text-muted-foreground">
                      {ASSISTANT_NUMBER ? (
                        <>
                          Message <span className="font-medium text-foreground">{ASSISTANT_NUMBER}</span>:
                        </>
                      ) : (
                        "Message the Arete assistant number:"
                      )}
                    </p>
                    <div className="flex items-center justify-between rounded-lg border border-border bg-muted/40 px-4 py-3">
                      <code className="text-lg font-semibold tracking-widest text-foreground">
                        LINK {pendingCode.code}
                      </code>
                      <span className="text-xs text-muted-foreground">
                        {secondsLeft > 0
                          ? `Expires in ${Math.floor(secondsLeft / 60)}:${String(secondsLeft % 60).padStart(2, "0")}`
                          : "Expired"}
                      </span>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleConnect}
                      disabled={generateCode.isPending}
                    >
                      Generate a new code
                    </Button>
                  </div>
                ) : (
                  <Button onClick={handleConnect} disabled={generateCode.isPending}>
                    Connect WhatsApp
                  </Button>
                )}
              </CardContent>
            </Card>
          </div>
        </main>
      </div>

      <Dialog open={unlinkOpen} onOpenChange={setUnlinkOpen}>
        <DialogContent className="sm:max-w-md border-foreground/20">
          <DialogHeader>
            <DialogTitle className="font-display text-xl font-semibold tracking-tight">
              Disconnect WhatsApp
            </DialogTitle>
            <DialogDescription className="text-sm text-muted-foreground mt-2">
              The assistant will stop responding to this number until you reconnect it.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="mt-2">
            <Button variant="outline" onClick={() => setUnlinkOpen(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => {
                unlink.mutate()
                setUnlinkOpen(false)
              }}
            >
              Disconnect
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
