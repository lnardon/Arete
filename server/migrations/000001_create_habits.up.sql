CREATE TABLE habits (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE habit_completions (
    habit_id   UUID  NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    date       DATE  NOT NULL,
    PRIMARY KEY (habit_id, date)
);
