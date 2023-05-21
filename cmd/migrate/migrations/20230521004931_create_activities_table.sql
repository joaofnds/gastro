-- +goose Up
CREATE TABLE activities(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  habit_id uuid NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
  description varchar NOT NULL DEFAULT '',
  created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_habit ON activities(habit_id);

-- +goose Down
DROP TABLE activities;
