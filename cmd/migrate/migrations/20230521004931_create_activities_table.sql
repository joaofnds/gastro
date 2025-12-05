-- +goose Up
CREATE TABLE IF NOT EXISTS activities (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    habit_id uuid NOT NULL REFERENCES habits (id) ON DELETE CASCADE,
    description varchar NOT NULL DEFAULT '',
    created_at timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_activity_habit ON activities (habit_id);

-- +goose Down
DROP TABLE IF EXISTS activities;
