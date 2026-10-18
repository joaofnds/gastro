package transfer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TableSummary struct {
	Rows   int    `json:"rows"`
	SHA256 string `json:"sha256"`
}

type Summary struct {
	Habits      TableSummary `json:"habits"`
	Activities  TableSummary `json:"activities"`
	Groups      TableSummary `json:"groups"`
	Memberships TableSummary `json:"groups_habits"`
}

type habitRow struct {
	ID, UserID, Name string
}

type activityRow struct {
	ID, HabitID, Description string
	CreatedAt                time.Time
}

type groupRow struct {
	ID, UserID, Name string
}

type membershipRow struct {
	GroupID, HabitID, UserID string
}

type snapshot struct {
	habits      []habitRow
	activities  []activityRow
	groups      []groupRow
	memberships []membershipRow
}

func (data snapshot) summary() (Summary, error) {
	habits, err := summarizeRows(data.habits)
	if err != nil {
		return Summary{}, err
	}

	activities, err := summarizeRows(data.activities)
	if err != nil {
		return Summary{}, err
	}

	groups, err := summarizeRows(data.groups)
	if err != nil {
		return Summary{}, err
	}

	memberships, err := summarizeRows(data.memberships)
	if err != nil {
		return Summary{}, err
	}

	return Summary{Habits: habits, Activities: activities, Groups: groups, Memberships: memberships}, nil
}

func summarizeRows[T habitRow | activityRow | groupRow | membershipRow](rows []T) (TableSummary, error) {
	encoded, err := json.Marshal(rows)
	if err != nil {
		return TableSummary{}, ErrInvalidData
	}

	return TableSummary{Rows: len(rows), SHA256: fmt.Sprintf("%x", sha256.Sum256(encoded))}, nil
}

func readSnapshot(ctx context.Context, tx *sql.Tx, schema string) (snapshot, error) {
	if err := validateColumns(ctx, tx, schema); err != nil {
		return snapshot{}, err
	}

	data := snapshot{
		habits:      []habitRow{},
		activities:  []activityRow{},
		groups:      []groupRow{},
		memberships: []membershipRow{},
	}
	queries := []struct {
		query string
		read  func(*sql.Rows) error
	}{
		{"SELECT id, user_id, name FROM " + tableName(schema, "habits") + " ORDER BY id", func(rows *sql.Rows) error {
			var row habitRow
			if err := rows.Scan(&row.ID, &row.UserID, &row.Name); err != nil {
				return ErrRead
			}

			if !canonicalUUIDs(row.ID, row.UserID) {
				return ErrInvalidData
			}

			data.habits = append(data.habits, row)
			return nil
		}},
		{"SELECT id, habit_id, description, created_at FROM " + tableName(schema, "activities") + " ORDER BY id", func(rows *sql.Rows) error {
			var row activityRow
			if err := rows.Scan(&row.ID, &row.HabitID, &row.Description, &row.CreatedAt); err != nil {
				return ErrRead
			}

			if !canonicalUUIDs(row.ID, row.HabitID) || row.CreatedAt.Nanosecond()%1000 != 0 || row.CreatedAt.Year() < 1 || row.CreatedAt.Year() > 9999 {
				return ErrInvalidData
			}

			row.CreatedAt = row.CreatedAt.UTC()
			data.activities = append(data.activities, row)
			return nil
		}},
		{"SELECT id, user_id, name FROM " + tableName(schema, "groups") + " ORDER BY id, user_id", func(rows *sql.Rows) error {
			var row groupRow
			if err := rows.Scan(&row.ID, &row.UserID, &row.Name); err != nil {
				return ErrRead
			}

			if !canonicalUUIDs(row.ID, row.UserID) {
				return ErrInvalidData
			}

			data.groups = append(data.groups, row)
			return nil
		}},
		{"SELECT group_id, habit_id, user_id FROM " + tableName(schema, "groups_habits") + " ORDER BY group_id, habit_id", func(rows *sql.Rows) error {
			var row membershipRow
			if err := rows.Scan(&row.GroupID, &row.HabitID, &row.UserID); err != nil {
				return ErrRead
			}

			if !canonicalUUIDs(row.GroupID, row.HabitID, row.UserID) {
				return ErrInvalidData
			}

			data.memberships = append(data.memberships, row)
			return nil
		}},
	}
	for _, query := range queries {
		if err := readRows(ctx, tx, query.query, query.read); err != nil {
			return snapshot{}, err
		}
	}

	return data, nil
}

func validateColumns(ctx context.Context, tx *sql.Tx, schema string) error {
	tables := []struct {
		name    string
		columns []string
	}{
		{"habits", []string{"id", "name", "user_id"}},
		{"groups", []string{"id", "name", "user_id"}},
		{"activities", []string{"created_at", "description", "habit_id", "id"}},
		{"groups_habits", []string{"group_id", "habit_id", "user_id"}},
	}
	for _, table := range tables {
		rows, err := tx.QueryContext(ctx, "SELECT * FROM "+tableName(schema, table.name)+" LIMIT 0")
		if err != nil {
			return ErrValidation
		}

		columns, err := rows.Columns()
		_ = rows.Close()
		if err != nil {
			return ErrValidation
		}

		slices.Sort(columns)
		if !slices.Equal(columns, table.columns) {
			return ErrValidation
		}
	}

	return nil
}

func readRows(ctx context.Context, tx *sql.Tx, query string, read func(*sql.Rows) error) error {
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return ErrRead
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		if err := read(rows); err != nil {
			return err
		}
	}

	if rows.Err() != nil {
		return ErrRead
	}

	return nil
}

func canonicalUUIDs(values ...string) bool {
	for _, value := range values {
		parsed, err := uuid.Parse(value)
		if err != nil || parsed.String() != value {
			return false
		}
	}

	return true
}

func tableName(schema, table string) string {
	if schema == "" {
		return pq.QuoteIdentifier(table)
	}

	return pq.QuoteIdentifier(schema) + "." + pq.QuoteIdentifier(table)
}

func writeSnapshot(ctx context.Context, tx *sql.Tx, data snapshot, schema string) error {
	for _, row := range data.habits {
		if _, err := tx.ExecContext(ctx, "INSERT INTO "+tableName(schema, "habits")+" (id, user_id, name) VALUES ($1, $2, $3)", row.ID, row.UserID, row.Name); err != nil {
			return ErrWrite
		}
	}

	for _, row := range data.groups {
		if _, err := tx.ExecContext(ctx, "INSERT INTO "+tableName(schema, "groups")+" (id, user_id, name) VALUES ($1, $2, $3)", row.ID, row.UserID, row.Name); err != nil {
			return ErrWrite
		}
	}

	for _, row := range data.activities {
		if _, err := tx.ExecContext(ctx, "INSERT INTO "+tableName(schema, "activities")+" (id, habit_id, description, created_at) VALUES ($1, $2, $3, $4)", row.ID, row.HabitID, row.Description, row.CreatedAt); err != nil {
			return ErrWrite
		}
	}

	for _, row := range data.memberships {
		if _, err := tx.ExecContext(ctx, "INSERT INTO "+tableName(schema, "groups_habits")+" (group_id, habit_id, user_id) VALUES ($1, $2, $3)", row.GroupID, row.HabitID, row.UserID); err != nil {
			return ErrWrite
		}
	}

	return nil
}
