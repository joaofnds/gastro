-- +goose Up
CREATE TABLE habits(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL,
  name varchar NOT NULL
);

CREATE UNIQUE INDEX idx_habit_id_and_user_id ON habits(id, user_id);

-- +goose Down
DROP TABLE habits;
