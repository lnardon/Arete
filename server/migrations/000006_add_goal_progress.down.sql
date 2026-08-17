ALTER TABLE goals DROP CONSTRAINT IF EXISTS goals_current_value_check;
ALTER TABLE goals DROP CONSTRAINT IF EXISTS goals_numeric_target_check;

ALTER TABLE goals
    DROP COLUMN IF EXISTS current_value,
    DROP COLUMN IF EXISTS target_value,
    DROP COLUMN IF EXISTS goal_type;
