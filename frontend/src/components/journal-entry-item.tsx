import { Pencil, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { MOOD_EMOJI } from "@/lib/types"
import type { JournalEntry } from "@/lib/types"

function formatEntryDateLabel(entryDate: string): string {
  const [y, m, d] = entryDate.split("-").map(Number)
  return new Date(y, m - 1, d).toLocaleDateString(undefined, {
    weekday: "short",
    month: "short",
    day: "numeric",
  })
}

interface JournalEntryItemProps {
  entry: JournalEntry
  onEdit: (entry: JournalEntry) => void
  onDelete: (entry: JournalEntry) => void
}

export function JournalEntryItem({ entry, onEdit, onDelete }: JournalEntryItemProps) {
  return (
    <div className="group flex w-full items-start gap-3 px-4 py-3.5 bg-card border border-border rounded-xl transition-all">
      <span className="text-xl leading-none mt-0.5 shrink-0" aria-hidden="true">
        {MOOD_EMOJI[entry.mood]}
      </span>
      <div className="flex-1 min-w-0 flex flex-col gap-1">
        <span className="text-xs text-muted-foreground tracking-wide">
          {formatEntryDateLabel(entry.entryDate)}
        </span>
        <p className="text-sm text-foreground whitespace-pre-wrap break-words">
          {entry.content}
        </p>
      </div>
      <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => onEdit(entry)}
          className="h-7 w-7 text-muted-foreground hover:text-foreground hover:bg-transparent"
        >
          <Pencil className="w-3.5 h-3.5" />
          <span className="sr-only">Edit</span>
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => onDelete(entry)}
          className="h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-transparent"
        >
          <Trash2 className="w-3.5 h-3.5" />
          <span className="sr-only">Delete</span>
        </Button>
      </div>
    </div>
  )
}
