import { useMemo, useState } from "react"
import { createFileRoute } from "@tanstack/react-router"
import { Flame } from "lucide-react"
import { AppSidebar } from "@/components/app-sidebar"
import { MobileHeader } from "@/components/mobile-header"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { JournalEntryForm } from "@/components/journal-entry-form"
import { JournalEntryItem } from "@/components/journal-entry-item"
import {
  useJournalEntries,
  useJournalEntry,
  useUpsertJournalEntry,
  useDeleteJournalEntry,
} from "@/hooks/use-journal"
import { formatLocalDate } from "@/lib/date-utils"
import type { JournalEntry } from "@/lib/types"

export const Route = createFileRoute('/_app/journal')({
  component: JournalPage,
})

// Consecutive logged days ending today (or yesterday, if today isn't logged
// yet) — computed client-side from the already-fetched entry list rather
// than a dedicated backend endpoint.
function computeStreak(entries: JournalEntry[], today: string): number {
  const dates = new Set(entries.map((e) => e.entryDate))
  const cursor = new Date()
  if (!dates.has(today)) {
    cursor.setDate(cursor.getDate() - 1)
  }

  let streak = 0
  while (dates.has(formatLocalDate(cursor))) {
    streak += 1
    cursor.setDate(cursor.getDate() - 1)
  }
  return streak
}

export default function JournalPage() {
  const today = formatLocalDate(new Date())
  const [editingDate, setEditingDate] = useState<string | null>(null)

  const { data: entries = [] } = useJournalEntries(60)
  const { data: todayEntry } = useJournalEntry(today)
  const upsertEntry = useUpsertJournalEntry()
  const deleteEntry = useDeleteJournalEntry()

  const streak = useMemo(() => computeStreak(entries, today), [entries, today])
  const history = entries.filter((e) => e.entryDate !== today)

  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <AppSidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <MobileHeader />
        <main className="flex-1 overflow-y-auto">
          <div className="max-w-3xl mx-auto px-5 py-8 md:px-8 md:py-12">
            <div className="flex items-start justify-between gap-4 mb-2">
              <div>
                <h2 className="font-display text-2xl font-semibold tracking-wide text-foreground text-balance">
                  Journal
                </h2>
                <p className="text-sm text-muted-foreground mt-1">
                  A couple of minutes to reflect on your day
                </p>
              </div>
              {streak > 0 && (
                <div className="flex items-center gap-1.5 text-sm text-muted-foreground shrink-0 mt-1.5">
                  <Flame className="w-4 h-4 text-orange-500" strokeWidth={1.5} />
                  <span className="tabular-nums">
                    {streak} day{streak === 1 ? "" : "s"}
                  </span>
                </div>
              )}
            </div>

            <div className="my-6 h-px bg-border" />

            <Card className="mb-8">
              <CardHeader>
                <CardTitle>Today</CardTitle>
              </CardHeader>
              <CardContent>
                <JournalEntryForm
                  initialMood={todayEntry?.mood ?? null}
                  initialContent={todayEntry?.content ?? ""}
                  onSave={(mood, content) => upsertEntry.mutate({ date: today, mood, content })}
                  saveLabel={todayEntry ? "Save changes" : "Save entry"}
                  pending={upsertEntry.isPending}
                />
              </CardContent>
            </Card>

            {history.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center border border-dashed border-border rounded-xl">
                <p className="text-sm text-muted-foreground">No past entries yet</p>
                <p className="text-xs text-muted-foreground/60 mt-1">
                  Entries you save will show up here
                </p>
              </div>
            ) : (
              <div className="flex flex-col gap-2">
                {history.map((entry) =>
                  editingDate === entry.entryDate ? (
                    <Card key={entry.id}>
                      <CardContent>
                        <JournalEntryForm
                          initialMood={entry.mood}
                          initialContent={entry.content}
                          onSave={(mood, content) => {
                            upsertEntry.mutate(
                              { date: entry.entryDate, mood, content },
                              { onSuccess: () => setEditingDate(null) }
                            )
                          }}
                          onCancel={() => setEditingDate(null)}
                          saveLabel="Save changes"
                          pending={upsertEntry.isPending}
                        />
                      </CardContent>
                    </Card>
                  ) : (
                    <JournalEntryItem
                      key={entry.id}
                      entry={entry}
                      onEdit={(e) => setEditingDate(e.entryDate)}
                      onDelete={(e) => deleteEntry.mutate(e.entryDate)}
                    />
                  )
                )}
              </div>
            )}
          </div>
        </main>
      </div>
    </div>
  )
}
