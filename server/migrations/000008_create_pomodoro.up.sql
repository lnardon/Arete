CREATE TABLE pomodoro_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_pomodoro_projects_user_id ON pomodoro_projects(user_id);

CREATE TABLE pomodoro_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES pomodoro_projects(id) ON DELETE SET NULL,
    planned_minutes INTEGER NOT NULL CHECK (planned_minutes > 0),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    local_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_pomodoro_entries_user_id ON pomodoro_entries(user_id);
CREATE INDEX idx_pomodoro_entries_user_date ON pomodoro_entries(user_id, local_date);
-- Enforces "one active timer per user": at most one row per user can have
-- ended_at IS NULL at a time.
CREATE UNIQUE INDEX idx_pomodoro_entries_one_active ON pomodoro_entries(user_id) WHERE ended_at IS NULL;
