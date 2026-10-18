package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
)

type column struct {
	name, kind    string
	required, key int
}

type foreignKey struct {
	parent             string
	sequence           int
	from, to           string
	onUpdate, onDelete string
}

func checkStructure(ctx context.Context, db *sql.DB) error {
	tables := []struct {
		name    string
		columns []column
		keys    []foreignKey
	}{
		{"habits", []column{{"id", "TEXT", 1, 1}, {"user_id", "TEXT", 1, 0}, {"name", "TEXT", 1, 0}}, nil},
		{"activities", []column{{"id", "TEXT", 1, 1}, {"habit_id", "TEXT", 1, 0}, {"description", "TEXT", 1, 0}, {"created_at", "DATETIME", 1, 0}}, []foreignKey{{"habits", 0, "habit_id", "id", "NO ACTION", "CASCADE"}}},
		{"groups", []column{{"id", "TEXT", 1, 1}, {"name", "TEXT", 1, 0}, {"user_id", "TEXT", 1, 2}}, nil},
		{"groups_habits", []column{{"group_id", "TEXT", 1, 1}, {"habit_id", "TEXT", 1, 2}, {"user_id", "TEXT", 1, 0}}, []foreignKey{
			{"groups", 0, "group_id", "id", "CASCADE", "CASCADE"},
			{"groups", 1, "user_id", "user_id", "CASCADE", "CASCADE"},
			{"habits", 0, "habit_id", "id", "CASCADE", "CASCADE"},
			{"habits", 1, "user_id", "user_id", "CASCADE", "CASCADE"},
		}},
	}
	for _, table := range tables {
		columns, err := readColumns(ctx, db, table.name)
		if err != nil || !slices.Equal(columns, table.columns) {
			return fmt.Errorf("%w: invalid columns or primary key in %s", ErrUncommissioned, table.name)
		}

		keys, err := readForeignKeys(ctx, db, table.name)
		if err != nil || !slices.Equal(keys, table.keys) {
			return fmt.Errorf("%w: invalid foreign keys in %s", ErrUncommissioned, table.name)
		}
	}

	indexes := []struct {
		table, name string
		unique      int
		columns     []string
	}{
		{"habits", "idx_habit_id_and_user_id", 1, []string{"id", "user_id"}},
		{"activities", "idx_activity_habit", 0, []string{"habit_id"}},
	}
	for _, index := range indexes {
		var unique, partial int
		err := db.QueryRowContext(ctx, `SELECT "unique", partial FROM pragma_index_list(?) WHERE name=?`, index.table, index.name).Scan(&unique, &partial)
		if err != nil || unique != index.unique || partial != 0 {
			return fmt.Errorf("%w: invalid index %s", ErrUncommissioned, index.name)
		}

		columns, err := readIndexColumns(ctx, db, index.name)
		if err != nil || !slices.Equal(columns, index.columns) {
			return fmt.Errorf("%w: invalid index columns in %s", ErrUncommissioned, index.name)
		}
	}

	return nil
}

func readColumns(ctx context.Context, db *sql.DB, table string) ([]column, error) {
	rows, err := db.QueryContext(ctx, `SELECT name, upper(type), "notnull", pk FROM pragma_table_info(?) ORDER BY cid`, table)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var columns []column
	for rows.Next() {
		var value column
		if err := rows.Scan(&value.name, &value.kind, &value.required, &value.key); err != nil {
			return nil, err
		}

		columns = append(columns, value)
	}

	return columns, rows.Err()
}

func readForeignKeys(ctx context.Context, db *sql.DB, table string) ([]foreignKey, error) {
	rows, err := db.QueryContext(ctx, `SELECT "table", seq, "from", "to", on_update, on_delete FROM pragma_foreign_key_list(?) ORDER BY "table", id, seq`, table)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var keys []foreignKey
	for rows.Next() {
		var value foreignKey
		if err := rows.Scan(&value.parent, &value.sequence, &value.from, &value.to, &value.onUpdate, &value.onDelete); err != nil {
			return nil, err
		}

		keys = append(keys, value)
	}

	return keys, rows.Err()
}

func readIndexColumns(ctx context.Context, db *sql.DB, index string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT name FROM pragma_index_info(?) ORDER BY seqno`, index)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var columns []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		columns = append(columns, name)
	}

	return columns, rows.Err()
}
