ALTER TABLE goals
    ADD COLUMN goal_type VARCHAR(10) NOT NULL DEFAULT 'binary' CHECK (goal_type IN ('binary', 'numeric')),
    ADD COLUMN target_value INTEGER,
    ADD COLUMN current_value INTEGER NOT NULL DEFAULT 0;

ALTER TABLE goals ADD CONSTRAINT goals_numeric_target_check
    CHECK (goal_type = 'binary' OR (target_value IS NOT NULL AND target_value > 0));

ALTER TABLE goals ADD CONSTRAINT goals_current_value_check
    CHECK (current_value >= 0);
