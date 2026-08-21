package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lnardon/arete/internal/repository"
	openai "github.com/openai/openai-go/v3"
)

var validPeriodTypes = map[string]bool{
	"month":    true,
	"quarter":  true,
	"semester": true,
	"year":     true,
}

var validGoalTypes = map[string]bool{
	"binary":  true,
	"numeric": true,
}

var errInvalidPeriodType = errors.New("periodType must be one of: month; quarter; semester; year")
var errInvalidGoalType = errors.New("goalType must be one of: binary; numeric")
var errInvalidMood = errors.New("mood must be an integer between 1 and 5")

type toolHandler func(ctx context.Context, args map[string]any) (string, error)

func mustParameters(schemaJSON string) openai.FunctionParameters {
	var p openai.FunctionParameters
	if err := json.Unmarshal([]byte(schemaJSON), &p); err != nil {
		panic(fmt.Sprintf("ai: invalid tool schema: %v", err))
	}
	return p
}

func jsonResult(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func argString(args map[string]any, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

// argInt reads an integer argument. JSON numbers decode to float64 in a
// map[string]any, so this converts rather than type-asserting to int.
func argInt(args map[string]any, key string) (int, bool) {
	v, ok := args[key].(float64)
	if !ok {
		return 0, false
	}
	return int(v), true
}

func buildTools(habitRepo *repository.HabitRepository, goalRepo *repository.GoalRepository, journalRepo *repository.JournalRepository, userID string, loc *time.Location) ([]openai.ChatCompletionToolUnionParam, map[string]toolHandler) {
	var tools []openai.ChatCompletionToolUnionParam
	handlers := make(map[string]toolHandler)

	register := func(name, description, schemaJSON string, handler toolHandler) {
		tools = append(tools, openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        name,
			Description: openai.String(description),
			Parameters:  mustParameters(schemaJSON),
		}))
		handlers[name] = handler
	}

	register("list_habits",
		"List every habit the user tracks in the order they were created. Takes no input.",
		`{"type": "object", "properties": {}}`,
		func(ctx context.Context, _ map[string]any) (string, error) {
			habits, err := habitRepo.ListHabits(ctx, userID)
			if err != nil {
				return "", err
			}
			return jsonResult(habits)
		},
	)

	register("create_habit",
		"Create a new habit for the user to track daily. Call this when the user asks to start tracking something new.",
		`{
			"type": "object",
			"required": ["name"],
			"properties": {
				"name": {"type": "string", "description": "The name of the new habit to track (e.g. 'Read 20 minutes')."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			name := argString(args, "name")
			if name == "" {
				return "", errors.New("name is required")
			}
			habit, err := habitRepo.CreateHabit(ctx, userID, name)
			if err != nil {
				return "", err
			}
			return jsonResult(habit)
		},
	)

	register("toggle_habit_completion",
		"Mark a habit as done or undo that for a given date. Calling this twice for the same date toggles it back off — check the current state with list_habits or get_completions_range first if you're not sure.",
		`{
			"type": "object",
			"required": ["habitId"],
			"properties": {
				"habitId": {"type": "string", "description": "The ID of the habit to mark done or undone (from list_habits)."},
				"date": {"type": "string", "description": "Date in YYYY-MM-DD format. Omit to use today's date (given in the system prompt)."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			habitID := argString(args, "habitId")
			if habitID == "" {
				return "", errors.New("habitId is required")
			}
			date := argString(args, "date")
			if date == "" {
				date = todayIn(loc)
			}
			if err := habitRepo.ToggleCompletion(ctx, userID, habitID, date); err != nil {
				return "", err
			}
			return jsonResult(map[string]string{"habitId": habitID, "date": date, "status": "toggled"})
		},
	)

	register("delete_habit",
		"Permanently delete a habit and its completion history. This cannot be undone. Only call this after the user has explicitly confirmed the deletion in this conversation (for example they replied 'yes' to your confirmation question) — if you haven't asked yet, ask first instead of calling this tool.",
		`{
			"type": "object",
			"required": ["habitId"],
			"properties": {
				"habitId": {"type": "string", "description": "The ID of the habit to delete (from list_habits)."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			habitID := argString(args, "habitId")
			if habitID == "" {
				return "", errors.New("habitId is required")
			}
			if err := habitRepo.DeleteHabit(ctx, userID, habitID); err != nil {
				return "", err
			}
			return jsonResult(map[string]string{"habitId": habitID, "status": "deleted"})
		},
	)

	register("get_completions_range",
		"Get every habit-completion record within a date range across all of the user's habits. Use this to compute streaks completion rates or answer summary questions like 'how did I do this week'.",
		`{
			"type": "object",
			"required": ["startDate", "endDate"],
			"properties": {
				"startDate": {"type": "string", "description": "Start date in YYYY-MM-DD format (inclusive)."},
				"endDate": {"type": "string", "description": "End date in YYYY-MM-DD format (inclusive)."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			start := argString(args, "startDate")
			end := argString(args, "endDate")
			if start == "" || end == "" {
				return "", errors.New("startDate and endDate are required")
			}
			completions, err := habitRepo.GetCompletionsForRange(ctx, userID, start, end)
			if err != nil {
				return "", err
			}
			return jsonResult(completions)
		},
	)

	register("list_goals",
		"List the user's goals for a specific period (month quarter semester or year).",
		`{
			"type": "object",
			"required": ["periodType", "periodKey"],
			"properties": {
				"periodType": {"type": "string", "description": "One of these values: month; quarter; semester; year."},
				"periodKey": {"type": "string", "description": "Period key matching periodType. Month uses 'YYYY-MM' like 2026-08. Quarter uses 'YYYY-QN' like 2026-Q3. Semester uses 'YYYY-HN' like 2026-H2. Year uses 'YYYY'."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			periodType := argString(args, "periodType")
			periodKey := argString(args, "periodKey")
			if !validPeriodTypes[periodType] {
				return "", errInvalidPeriodType
			}
			goals, err := goalRepo.ListGoals(ctx, userID, periodType, periodKey)
			if err != nil {
				return "", err
			}
			return jsonResult(formatGoals(goals))
		},
	)

	register("create_goal",
		"Create a new goal for a specific period (month quarter semester or year). Goals are either binary (a single done/not-done checkbox) or numeric (progress toward a target count, like practicing tennis 222 times this year).",
		`{
			"type": "object",
			"required": ["title", "periodType", "periodKey"],
			"properties": {
				"title": {"type": "string", "description": "The goal's title."},
				"periodType": {"type": "string", "description": "One of these values: month; quarter; semester; year."},
				"periodKey": {"type": "string", "description": "Period key matching periodType. Month uses 'YYYY-MM' like 2026-08. Quarter uses 'YYYY-QN' like 2026-Q3. Semester uses 'YYYY-HN' like 2026-H2. Year uses 'YYYY'."},
				"goalType": {"type": "string", "description": "One of these values: binary; numeric. Defaults to binary if omitted. Use numeric whenever the user describes a target count (e.g. 'practice tennis 222 times this year')."},
				"targetValue": {"type": "integer", "description": "Required when goalType is numeric: the target count to reach (e.g. 222). Omit for binary goals."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			title := argString(args, "title")
			periodType := argString(args, "periodType")
			periodKey := argString(args, "periodKey")
			if title == "" {
				return "", errors.New("title is required")
			}
			if !validPeriodTypes[periodType] {
				return "", errInvalidPeriodType
			}
			goalType := argString(args, "goalType")
			if goalType == "" {
				goalType = "binary"
			}
			if !validGoalTypes[goalType] {
				return "", errInvalidGoalType
			}
			var targetValue *int
			if v, ok := argInt(args, "targetValue"); ok {
				targetValue = &v
			}
			if goalType == "numeric" && (targetValue == nil || *targetValue <= 0) {
				return "", errors.New("targetValue is required and must be greater than 0 for numeric goals")
			}
			if goalType == "binary" && targetValue != nil {
				return "", errors.New("targetValue is only valid for numeric goals")
			}
			goal, err := goalRepo.CreateGoal(ctx, userID, title, periodType, periodKey, goalType, targetValue)
			if err != nil {
				return "", err
			}
			return jsonResult(formatGoal(goal))
		},
	)

	register("toggle_goal",
		"Mark a binary goal as complete or undo that. Only works for binary goals — for numeric goals (goalType 'numeric'), use log_goal_progress instead.",
		`{
			"type": "object",
			"required": ["goalId"],
			"properties": {
				"goalId": {"type": "string", "description": "The ID of the goal (from list_goals)."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			goalID := argString(args, "goalId")
			if goalID == "" {
				return "", errors.New("goalId is required")
			}
			goal, err := goalRepo.ToggleGoal(ctx, userID, goalID)
			if err != nil {
				return "", err
			}
			return jsonResult(formatGoal(goal))
		},
	)

	register("log_goal_progress",
		"Add to (or subtract from, with a negative delta) a numeric goal's current progress. Use this when the user reports doing something that counts toward a numeric goal, e.g. 'I practiced tennis today' adds 1. Only works for numeric goals — for binary goals, use toggle_goal instead.",
		`{
			"type": "object",
			"required": ["goalId", "delta"],
			"properties": {
				"goalId": {"type": "string", "description": "The ID of the numeric goal (from list_goals)."},
				"delta": {"type": "integer", "description": "Amount to add to the goal's current progress. Use a negative number to correct over-logging."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			goalID := argString(args, "goalId")
			if goalID == "" {
				return "", errors.New("goalId is required")
			}
			delta, ok := argInt(args, "delta")
			if !ok || delta == 0 {
				return "", errors.New("delta is required and must be non-zero")
			}
			goal, err := goalRepo.AddGoalProgress(ctx, userID, goalID, delta)
			if err != nil {
				return "", err
			}
			return jsonResult(formatGoal(goal))
		},
	)

	register("update_goal",
		"Rename an existing goal.",
		`{
			"type": "object",
			"required": ["goalId", "title"],
			"properties": {
				"goalId": {"type": "string", "description": "The ID of the goal to rename (from list_goals)."},
				"title": {"type": "string", "description": "The new title."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			goalID := argString(args, "goalId")
			title := argString(args, "title")
			if goalID == "" || title == "" {
				return "", errors.New("goalId and title are required")
			}
			goal, err := goalRepo.UpdateGoal(ctx, userID, goalID, title, nil, nil)
			if err != nil {
				return "", err
			}
			return jsonResult(formatGoal(goal))
		},
	)

	register("delete_goal",
		"Permanently delete a goal. This cannot be undone. Only call this after the user has explicitly confirmed the deletion in this conversation — if you haven't asked yet, ask first instead of calling this tool.",
		`{
			"type": "object",
			"required": ["goalId"],
			"properties": {
				"goalId": {"type": "string", "description": "The ID of the goal (from list_goals)."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			goalID := argString(args, "goalId")
			if goalID == "" {
				return "", errors.New("goalId is required")
			}
			if err := goalRepo.DeleteGoal(ctx, userID, goalID); err != nil {
				return "", err
			}
			return jsonResult(map[string]string{"goalId": goalID, "status": "deleted"})
		},
	)

	register("list_journal_entries",
		"List the user's most recent journal entries, most recent first. Use this to answer questions about how the user has been feeling or what they've written about recently.",
		`{
			"type": "object",
			"properties": {
				"limit": {"type": "integer", "description": "Max number of entries to return, most recent first. Defaults to 14 if omitted."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			limit, ok := argInt(args, "limit")
			if !ok || limit <= 0 {
				limit = 14
			}
			entries, err := journalRepo.ListEntries(ctx, userID, limit)
			if err != nil {
				return "", err
			}
			return jsonResult(formatJournalEntries(entries))
		},
	)

	register("create_or_update_journal_entry",
		"Create today's (or a given date's) journal entry, or overwrite it if one already exists for that date. Call this whenever the user shares something about their day — journal entries are one per day, so this always safely creates or edits, never duplicates.",
		`{
			"type": "object",
			"required": ["mood", "content"],
			"properties": {
				"entryDate": {"type": "string", "description": "Date in YYYY-MM-DD format. Omit to use today's date (given in the system prompt)."},
				"mood": {"type": "integer", "description": "Mood rating from 1 (rough day) to 5 (great day), inferred from the user's message unless they state one."},
				"content": {"type": "string", "description": "The journal entry text — the user's own words about their day, lightly cleaned up but not rewritten or padded out."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			entryDate := argString(args, "entryDate")
			if entryDate == "" {
				entryDate = todayIn(loc)
			}
			mood, ok := argInt(args, "mood")
			if !ok || mood < 1 || mood > 5 {
				return "", errInvalidMood
			}
			content := argString(args, "content")
			if content == "" {
				return "", errors.New("content is required")
			}
			entry, err := journalRepo.UpsertEntry(ctx, userID, entryDate, mood, content)
			if err != nil {
				return "", err
			}
			return jsonResult(formatJournalEntry(entry))
		},
	)

	register("delete_journal_entry",
		"Permanently delete the journal entry for a given date. This cannot be undone. Only call this after the user has explicitly confirmed the deletion in this conversation — if you haven't asked yet, ask first instead of calling this tool.",
		`{
			"type": "object",
			"required": ["entryDate"],
			"properties": {
				"entryDate": {"type": "string", "description": "Date of the entry to delete, in YYYY-MM-DD format."}
			}
		}`,
		func(ctx context.Context, args map[string]any) (string, error) {
			entryDate := argString(args, "entryDate")
			if entryDate == "" {
				return "", errors.New("entryDate is required")
			}
			if err := journalRepo.DeleteEntry(ctx, userID, entryDate); err != nil {
				return "", err
			}
			return jsonResult(map[string]string{"entryDate": entryDate, "status": "deleted"})
		},
	)

	return tools, handlers
}
