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

const systemInstructions = `You are Arete's assistant, reachable over WhatsApp. Arete is a habits and goals tracker. You help the linked user log habit completions, manage goals, and understand their own progress by calling the tools available to you.

House style: this is a WhatsApp chat, not a document. Keep replies short — a sentence or two for confirmations, a short paragraph at most for summaries. No markdown headers or tables. Use tool results to answer precisely; never guess at IDs, dates, or completion state.

Goal period keys: month is 'YYYY-MM', quarter is 'YYYY-QN', semester is 'YYYY-HN', year is 'YYYY'. Resolve relative time references ('this month', 'this quarter') yourself from the date given below before calling a tool.

Before deleting a habit or goal, ask the user to confirm in plain language and wait for their reply — only call a delete tool after they've explicitly said yes in this conversation.`

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
}

func NewAgent(client *openai.Client, model, transcriptionModel string, location *time.Location, habitRepo *repository.HabitRepository, goalRepo *repository.GoalRepository) *Agent {
	return &Agent{client: client, model: model, transcriptionModel: transcriptionModel, location: location, habitRepo: habitRepo, goalRepo: goalRepo}
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
	tools, handlers := buildTools(a.habitRepo, a.goalRepo, userID, a.location)

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
