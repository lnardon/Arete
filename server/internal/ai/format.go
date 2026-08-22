package ai

import (
	"fmt"
	"strings"

	"github.com/lnardon/arete/internal/models"
)

const progressBarWidth = 10

// goalView augments a Goal with a ready-made WhatsApp-formatted display line
// so the model never has to hand-draw a progress bar or do the percent math
// itself — it only has to reuse Formatted verbatim (see systemInstructions).
type goalView struct {
	models.Goal
	Formatted string `json:"formatted"`
}

func formatGoal(g models.Goal) goalView {
	return goalView{Goal: g, Formatted: goalLine(g)}
}

func formatGoals(goals []models.Goal) []goalView {
	views := make([]goalView, len(goals))
	for i, g := range goals {
		views[i] = formatGoal(g)
	}
	return views
}

func goalLine(g models.Goal) string {
	if g.GoalType != "numeric" {
		box := "[ ]"
		if g.Completed {
			box = "[x]"
		}
		return fmt.Sprintf("%s *%s*", box, g.Title)
	}

	target := 0
	if g.TargetValue != nil {
		target = *g.TargetValue
	}
	percent := 0
	if target > 0 {
		percent = g.CurrentValue * 100 / target
	}
	status := ""
	if g.Completed {
		status = " (done)"
	}
	// The bar goes in its own ```monospace``` block — WhatsApp renders
	// everything outside of one in a proportional font, where "#" and "-"
	// have different widths and the bar would drift out of alignment.
	return fmt.Sprintf("*%s*%s\n```[%s] %d/%d (%d%%)```", g.Title, status, progressBar(g.CurrentValue, target), g.CurrentValue, target, percent)
}

// progressBar renders current/target as a fixed-width bar of # (filled) and
// - (empty) characters — overshoot (current > target) still shows a full
// bar, it just doesn't grow past it.
// journalView augments a JournalEntry with a ready-made WhatsApp-formatted
// display line, mirroring goalView so the model reuses Formatted verbatim
// instead of inventing its own layout (see systemInstructions).
type journalView struct {
	models.JournalEntry
	Formatted string `json:"formatted"`
}

func formatJournalEntry(e models.JournalEntry) journalView {
	return journalView{JournalEntry: e, Formatted: journalLine(e)}
}

func formatJournalEntries(entries []models.JournalEntry) []journalView {
	views := make([]journalView, len(entries))
	for i, e := range entries {
		views[i] = formatJournalEntry(e)
	}
	return views
}

// journalLine deliberately avoids emoji for the mood — WhatsApp house style
// (see systemInstructions) bans emoji entirely, so mood is spelled out as a
// plain "N/5" instead of an emoji face.
func journalLine(e models.JournalEntry) string {
	return fmt.Sprintf("*%s* (mood %d/5)\n%s", e.EntryDate, e.Mood, e.Content)
}

// pomodoroEntryView augments a PomodoroEntry with a ready-made
// WhatsApp-formatted display line, mirroring goalView/journalView.
type pomodoroEntryView struct {
	models.PomodoroEntry
	Formatted string `json:"formatted"`
}

func formatPomodoroEntry(e models.PomodoroEntry) pomodoroEntryView {
	return pomodoroEntryView{PomodoroEntry: e, Formatted: pomodoroEntryLine(e)}
}

func formatPomodoroEntries(entries []models.PomodoroEntry) []pomodoroEntryView {
	views := make([]pomodoroEntryView, len(entries))
	for i, e := range entries {
		views[i] = formatPomodoroEntry(e)
	}
	return views
}

func pomodoroEntryLine(e models.PomodoroEntry) string {
	start := e.StartedAt.Format("15:04")
	if e.EndedAt == nil {
		return fmt.Sprintf("*Timer running* — started %s, planned %d min", start, e.PlannedMinutes)
	}
	elapsed := int(e.EndedAt.Sub(e.StartedAt).Minutes())
	end := e.EndedAt.Format("15:04")
	return fmt.Sprintf("*%s* %s-%s (%d min, planned %d)", e.LocalDate, start, end, elapsed, e.PlannedMinutes)
}

func progressBar(current, target int) string {
	filled := 0
	if target > 0 {
		filled = current * progressBarWidth / target
	}
	if filled > progressBarWidth {
		filled = progressBarWidth
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("#", filled) + strings.Repeat("-", progressBarWidth-filled)
}
