import { cn } from "@/lib/utils"
import { MOOD_EMOJI, MOOD_LABEL } from "@/lib/types"

const MOOD_VALUES = [1, 2, 3, 4, 5]

interface MoodPickerProps {
  value: number | null
  onChange: (mood: number) => void
}

export function MoodPicker({ value, onChange }: MoodPickerProps) {
  return (
    <div className="flex gap-2" role="radiogroup" aria-label="Mood">
      {MOOD_VALUES.map((mood) => (
        <button
          key={mood}
          type="button"
          role="radio"
          aria-checked={value === mood}
          aria-label={MOOD_LABEL[mood]}
          onClick={() => onChange(mood)}
          className={cn(
            "flex h-11 w-11 items-center justify-center rounded-lg border text-xl transition-colors",
            value === mood
              ? "border-primary bg-primary"
              : "border-border hover:bg-muted"
          )}
        >
          {MOOD_EMOJI[mood]}
        </button>
      ))}
    </div>
  )
}
