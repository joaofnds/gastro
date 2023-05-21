-- +goose Up
CREATE TABLE groups_habits(
  group_id uuid NOT NULL,
  habit_id uuid NOT NULL,
  user_id uuid NOT NULL,
  CONSTRAINT "groups_habits_habit_id_fkey" FOREIGN KEY (habit_id, user_id) REFERENCES habits(id, user_id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT "groups_habits_group_id_fkey" FOREIGN KEY (group_id, user_id) REFERENCES GROUPS (id, user_id) ON DELETE CASCADE ON UPDATE CASCADE,
  PRIMARY KEY (group_id, habit_id)
);

-- +goose Down
DROP TABLE groups_habits;
