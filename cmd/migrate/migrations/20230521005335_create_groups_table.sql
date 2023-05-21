-- +goose Up
CREATE TABLE "groups"(
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name varchar NOT NULL,
  user_id uuid NOT NULL,
  PRIMARY KEY (id, user_id)
);

-- +goose Down
DROP TABLE "groups";
