-- +goose Up
CREATE TABLE IF NOT EXISTS habits (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL,
    name varchar NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_habit_id_and_user_id ON habits (id, user_id);

-- +goose Down
DROP TABLE IF EXISTS habits;
