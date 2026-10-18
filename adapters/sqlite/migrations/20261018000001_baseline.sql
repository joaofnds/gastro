-- +goose Up
CREATE TABLE habits (
    id TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL
);
CREATE UNIQUE INDEX idx_habit_id_and_user_id ON habits (id, user_id);
CREATE TABLE activities (
    id TEXT NOT NULL PRIMARY KEY,
    habit_id TEXT NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    description TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_activity_habit ON activities(habit_id);
CREATE TABLE groups (
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY(id, user_id)
);
CREATE TABLE groups_habits (
    group_id TEXT NOT NULL,
    habit_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    FOREIGN KEY(habit_id, user_id) REFERENCES habits(id, user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY(group_id, user_id) REFERENCES groups(id, user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY(group_id, habit_id)
);

-- +goose Down
DROP TABLE groups_habits;
DROP TABLE groups;
DROP TABLE activities;
DROP TABLE habits;
