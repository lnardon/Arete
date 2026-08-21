// Package ai wires the OpenAI Chat Completions tool-calling loop to Arete's
// habit and goal repositories, scoped per-user, for the WhatsApp assistant.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lnardon/arete/internal/repository"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

// tripleBacktick is spliced into systemInstructions below — a raw string
// literal can't contain a literal backtick, since backtick is what
// terminates it.
const tripleBacktick = "```"

const systemInstructions = `You are Arete's assistant, reachable over WhatsApp. Arete is a habits and goals tracker. You help the linked user log habit completions, manage goals, and understand their own progress by calling the tools available to you.

House style: this is a WhatsApp chat, not a document. Keep replies short — a sentence or two for confirmations, a short paragraph at most for summaries. Use tool results to answer precisely; never guess at IDs, dates, or completion state.

Formatting for WhatsApp — it renders *bold*, _italic_, ~strikethrough~, and ` + tripleBacktick + `monospace` + tripleBacktick + ` (three backticks, can span multiple lines); there are no headers, tables, or markdown links. Never use emoji — not for status, not for emphasis, not anywhere in a reply. Lean on plain ASCII instead: brackets, dashes, and monospaced alignment read as clean and native on WhatsApp; emoji read as clutter.
- Goal tool results (list_goals, create_goal, toggle_goal, update_goal, log_goal_progress) include a ready-made "formatted" block: the bold title, and for numeric goals a monospaced bar on the next line like ` + tripleBacktick + `[#####-----] 48/222 (21%)` + tripleBacktick + ` (# = filled, - = remaining). Binary goals use a plain checkbox baked into the title line: "[x] *Title*" if done, "[ ] *Title*" if not. Reuse these blocks verbatim, exactly as given, one goal per block. Never redraw the bar, swap its characters, or recompute the percent yourself — you will get the alignment or the math wrong.
- When you add your own sentence around those blocks (an intro, a streak count, a comparison across periods), use *bold* for the one number or name that matters. Keep it plain otherwise — no icons, no decoration.
- Listing several goals or habits: put a blank line between each block so they read as separate entries. A dashed rule on its own line (e.g. ------------------------) may separate sections (this month's goals vs. today's habits) if it genuinely helps scanning — use at most one or two per message, not as decoration.

Goal period keys: month is 'YYYY-MM', quarter is 'YYYY-QN', semester is 'YYYY-HN', year is 'YYYY'. Resolve relative time references ('this month', 'this quarter') yourself from the date given below before calling a tool.

Goals are binary (a single done/not-done checkbox, use toggle_goal) or numeric (progress toward a target count, e.g. 'practice tennis 222 times this year', use log_goal_progress to add to the count). Check a goal's goalType from list_goals before deciding which tool applies.

Journal tool results (list_journal_entries, create_or_update_journal_entry) include the same kind of ready-made "formatted" block: the bold date, then "(mood N/5)", then the entry text on the next line. Reuse it verbatim, same as goal blocks.

When the user shares how their day went — even in passing, not just when they explicitly say "journal this" — log it with create_or_update_journal_entry: infer a mood from 1 (rough) to 5 (great) from their tone, and use their own words (lightly cleaned up, not rewritten or padded) as the content. Only ask them to state a mood explicitly if the message is genuinely ambiguous. If they send more about the same day later, call the tool again for that date — it overwrites the existing entry rather than creating a duplicate, so there's never a need to check whether one already exists first.

Before deleting a habit, goal, or journal entry, ask the user to confirm in plain language and wait for their reply — only call a delete tool after they've explicitly said yes in this conversation.`

const maxToolIterations = 6
const maxAudioBytes = 25 * 1024 * 1024

var ErrAudioTooLarge = errors.New("audio file too large to transcribe")

// todayIn resolves "today" in loc rather than the server's local/UTC clock,
// since the WhatsApp assistant's notion of "today" must match the date the
// web UI shows (computed from the browser's local time) — not the
// container's clock, which runs UTC.
func todayIn(loc *time.Location) string {
	return time.Now().In(loc).Format("2006-01-02")
}

type Agent struct {
	client             *openai.Client
	model              string
	transcriptionModel string
	location           *time.Location
	habitRepo          *repository.HabitRepository
	goalRepo           *repository.GoalRepository
	journalRepo        *repository.JournalRepository
}

func NewAgent(client *openai.Client, model, transcriptionModel string, location *time.Location, habitRepo *repository.HabitRepository, goalRepo *repository.GoalRepository, journalRepo *repository.JournalRepository) *Agent {
	return &Agent{client: client, model: model, transcriptionModel: transcriptionModel, location: location, habitRepo: habitRepo, goalRepo: goalRepo, journalRepo: journalRepo}
}

func (a *Agent) Transcribe(ctx context.Context, audio []byte, filename string) (string, error) {
	if len(audio) > maxAudioBytes {
		return "", ErrAudioTooLarge
	}

	resp, err := a.client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		File:  openai.File(bytes.NewReader(audio), filename, ""),
		Model: a.transcriptionModel,
	})
	if err != nil {
		return "", fmt.Errorf("openai transcription: %w", err)
	}
	return strings.TrimSpace(resp.Text), nil
}

func (a *Agent) ProcessMessage(ctx context.Context, userID string, history []openai.ChatCompletionMessageParamUnion, userText string) (string, error) {
	tools, handlers := buildTools(a.habitRepo, a.goalRepo, a.journalRepo, userID, a.location)

	system := systemInstructions + "\n\nToday's date is " + time.Now().In(a.location).Format("Monday, 2006-01-02") + "."

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+2)
	messages = append(messages, openai.SystemMessage(system))
	messages = append(messages, history...)
	messages = append(messages, openai.UserMessage(userText))

	for range maxToolIterations {
		resp, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    a.model,
			Messages: messages,
			Tools:    tools,
			// gpt-5.6-luna rejects function tools + reasoning_effort on
			// /v1/chat/completions unless effort is explicitly "none".
			ReasoningEffort: shared.ReasoningEffortNone,
		})
		if err != nil {
			return "", fmt.Errorf("openai chat: %w", err)
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("openai chat: no choices returned")
		}

		message := resp.Choices[0].Message
		messages = append(messages, message.ToParam())

		if len(message.ToolCalls) == 0 {
			if message.Content == "" {
				return "Sorry, I couldn't come up with a reply for that — try rephrasing?", nil
			}
			return message.Content, nil
		}

		for _, call := range message.ToolCalls {
			messages = append(messages, openai.ToolMessage(a.runTool(ctx, handlers, call), call.ID))
		}
	}

	return "Sorry, that took too many steps to work out — try asking in a simpler way?", nil
}

func (a *Agent) runTool(ctx context.Context, handlers map[string]toolHandler, call openai.ChatCompletionMessageToolCallUnion) string {
	handler, ok := handlers[call.Function.Name]
	if !ok {
		return fmt.Sprintf("Error: unknown tool %q", call.Function.Name)
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
		return fmt.Sprintf("Error: invalid arguments: %v", err)
	}

	result, err := handler(ctx, args)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return result
}
