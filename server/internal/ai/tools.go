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

var errInvalidPeriodType = errors.New("periodType must be one of: month; quarter; semester; year")

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

func buildTools(habitRepo *repository.HabitRepository, goalRepo *repository.GoalRepository, userID string, loc *time.Location) ([]openai.ChatCompletionToolUnionParam, map[string]toolHandler) {
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
			return jsonResult(goals)
		},
	)

	register("create_goal",
		"Create a new goal for a specific period (month quarter semester or year).",
		`{
			"type": "object",
			"required": ["title", "periodType", "periodKey"],
			"properties": {
				"title": {"type": "string", "description": "The goal's title."},
				"periodType": {"type": "string", "description": "One of these values: month; quarter; semester; year."},
				"periodKey": {"type": "string", "description": "Period key matching periodType. Month uses 'YYYY-MM' like 2026-08. Quarter uses 'YYYY-QN' like 2026-Q3. Semester uses 'YYYY-HN' like 2026-H2. Year uses 'YYYY'."}
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
			goal, err := goalRepo.CreateGoal(ctx, userID, title, periodType, periodKey)
			if err != nil {
				return "", err
			}
			return jsonResult(goal)
		},
	)

	register("toggle_goal",
		"Mark a goal as complete or undo that.",
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
			return jsonResult(goal)
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
			goal, err := goalRepo.UpdateGoal(ctx, userID, goalID, title)
			if err != nil {
				return "", err
			}
			return jsonResult(goal)
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

	return tools, handlers
}
