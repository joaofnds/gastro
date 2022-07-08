package habit

import (
	"context"
	"database/sql"
	"fmt"
)

type SQLHabitRepository struct {
	db *sql.DB
}

func NewHabitRepository(db *sql.DB) *SQLHabitRepository {
	return &SQLHabitRepository{db}
}

func (repo *SQLHabitRepository) Create(ctx context.Context, name string) (Habit, error) {
	row := repo.db.QueryRowContext(ctx, "INSERT INTO habits(name) VALUES ($1) RETURNING id", name)
	if row.Err() != nil {
		return Habit{}, row.Err()
	}

	var id int
	err := row.Scan(&id)
	if err != nil {
		return Habit{}, err
	}

	return Habit{id, name}, row.Err()
}

func (repo *SQLHabitRepository) FindByName(ctx context.Context, queryName string) (Habit, error) {
	row := repo.db.QueryRowContext(ctx, "SELECT id, name FROM habits WHERE name = $1", queryName)
	if row.Err() != nil {
		return Habit{}, row.Err()
	}

	var id int
	var name string
	err := row.Scan(&id, &name)
	if err != nil {
		return Habit{}, fmt.Errorf("failed to parse habit: %w", err)
	}

	return Habit{id, name}, nil
}

func (repo *SQLHabitRepository) List(ctx context.Context) ([]Habit, error) {
	rows, err := repo.db.QueryContext(ctx, "SELECT id, name FROM habits")
	if err != nil {
		return []Habit{}, err
	}

	defer rows.Close()

	habits := []Habit{}

	for rows.Next() {
		var id int
		var name string

		err = rows.Scan(&id, &name)
		if err != nil {
			return []Habit{}, err
		}

		habits = append(habits, Habit{id, name})
	}

	return habits, nil
}

func (repo *SQLHabitRepository) DeleteByName(ctx context.Context, name string) error {
	_, err := repo.db.ExecContext(ctx, "DELETE FROM habits WHERE name = $1", name)
	return err
}

func (repo *SQLHabitRepository) DeleteAll(ctx context.Context) error {
	_, err := repo.db.ExecContext(ctx, "DELETE FROM habits")
	return err
}
