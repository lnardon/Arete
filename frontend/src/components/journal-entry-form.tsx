import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { MoodPicker } from "@/components/mood-picker"

interface JournalEntryFormProps {
  initialMood: number | null
  initialContent: string
  onSave: (mood: number, content: string) => void
  onCancel?: () => void
  saveLabel: string
  pending?: boolean
}

export function JournalEntryForm({
  initialMood,
  initialContent,
  onSave,
  onCancel,
  saveLabel,
  pending,
}: JournalEntryFormProps) {
  const [mood, setMood] = useState(initialMood)
  const [content, setContent] = useState(initialContent)

  useEffect(() => {
    setMood(initialMood)
    setContent(initialContent)
  }, [initialMood, initialContent])

  const canSubmit = mood !== null && content.trim() !== ""

  return (
    <div className="flex flex-col gap-3">
      <MoodPicker value={mood} onChange={setMood} />
      <Textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="How was your day?"
        className="min-h-28 border-foreground/20 focus-visible:ring-foreground/30 placeholder:text-muted-foreground/50"
      />
      <div className="flex items-center justify-end gap-2">
        {onCancel && (
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            className="border-foreground/20 text-foreground bg-transparent hover:text-red-600"
          >
            Cancel
          </Button>
        )}
        <Button
          type="button"
          disabled={!canSubmit || pending}
          onClick={() => mood !== null && onSave(mood, content.trim())}
          className="bg-foreground text-background hover:bg-foreground/80"
        >
          {saveLabel}
        </Button>
      </div>
    </div>
  )
}
