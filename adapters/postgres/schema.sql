CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS habits(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL,
  name varchar NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_habit_id_and_user_id ON habits(id, user_id);

CREATE TABLE IF NOT EXISTS activities(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  habit_id uuid NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
  description varchar NOT NULL DEFAULT '',
  created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activity_habit ON activities(habit_id);

CREATE TABLE IF NOT EXISTS "public"."groups"(
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name varchar NOT NULL,
  user_id uuid NOT NULL,
  PRIMARY KEY (id, user_id)
);

CREATE TABLE IF NOT EXISTS groups_habits(
  group_id uuid NOT NULL,
  habit_id uuid NOT NULL,
  user_id uuid NOT NULL,
  CONSTRAINT "groups_habits_habit_id_fkey" FOREIGN KEY (habit_id, user_id) REFERENCES habits(id, user_id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT "groups_habits_group_id_fkey" FOREIGN KEY (group_id, user_id) REFERENCES GROUPS (id, user_id) ON DELETE CASCADE ON UPDATE CASCADE,
  PRIMARY KEY (group_id, habit_id)
);
