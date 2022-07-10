package habit

import (
	"context"
	"database/sql"
	"time"
)

type SQLHabitRepository struct {
	DB Querier
}

type Querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func NewHabitRepository(db *sql.DB) *SQLHabitRepository {
	return &SQLHabitRepository{db}
}

func (repo *SQLHabitRepository) Create(ctx context.Context, name string) (Habit, error) {
	row := repo.DB.QueryRowContext(ctx, "INSERT INTO habits(name) VALUES ($1) RETURNING id", name)
	if row.Err() != nil {
		return Habit{}, row.Err()
	}

	var id int
	err := row.Scan(&id)
	if err != nil {
		return Habit{}, err
	}

	return Habit{id, name, []Activity{}}, row.Err()
}

func (repo *SQLHabitRepository) FindByName(ctx context.Context, queryName string) (Habit, error) {
	rows, err := repo.DB.QueryContext(ctx, `
		SELECT
			habits.id,
			habits.name,
			activities.id,
			activities.created_at
		FROM habits
			LEFT JOIN activities ON activities.habit_id = habits.id
		WHERE habits.name = $1`,
		queryName,
	)
	if err != nil {
		return Habit{}, err
	}

	defer rows.Close()

	habits, err := scanRows(rows)
	if err != nil {
		return Habit{}, err
	}

	if len(habits) == 0 {
		return Habit{}, HabitNotFoundErr
	}

	return habits[0], err
}

func (repo *SQLHabitRepository) AddActivity(ctx context.Context, habit Habit, time time.Time) (Activity, error) {
	row := repo.DB.QueryRowContext(ctx, "INSERT INTO activities(habit_id, created_at) VALUES ($1, $2) RETURNING id", habit.ID, time)
	if row.Err() != nil {
		return Activity{}, row.Err()
	}

	var id int
	err := row.Scan(&id)
	if err != nil {
		return Activity{}, err
	}

	return Activity{id, time}, row.Err()
}

func (repo *SQLHabitRepository) List(ctx context.Context) ([]Habit, error) {
	rows, err := repo.DB.QueryContext(ctx, `
		SELECT
			habits.id,
			habits.name,
			activities.id,
			activities.created_at
		FROM habits
			LEFT JOIN activities ON activities.habit_id = habits.id
	`)

	if err != nil {
		return []Habit{}, err
	}

	defer rows.Close()

	return scanRows(rows)
}

func (repo *SQLHabitRepository) DeleteByName(ctx context.Context, name string) error {
	_, err := repo.DB.ExecContext(ctx, "DELETE FROM habits WHERE name = $1", name)
	return err
}

func (repo *SQLHabitRepository) DeleteAll(ctx context.Context) error {
	_, err := repo.DB.ExecContext(ctx, "DELETE FROM habits")
	return err
}

func scanRows(rows *sql.Rows) ([]Habit, error) {
	m := map[int]*Habit{}

	for rows.Next() {
		var habitID int
		var habitName string
		var activityID sql.NullInt32
		var activityCreatedAt sql.NullTime

		err := rows.Scan(&habitID, &habitName, &activityID, &activityCreatedAt)
		if err != nil {
			return []Habit{}, err
		}

		habit, ok := m[habitID]
		if !ok {
			habit = &Habit{habitID, habitName, []Activity{}}
			m[habitID] = habit
		}

		if activityID.Valid {
			activity := Activity{int(activityID.Int32), activityCreatedAt.Time}
			habit.Activities = append(habit.Activities, activity)
		}
	}

	habits := make([]Habit, len(m))
	var i int
	for _, h := range m {
		habits[i] = *h
		i++
	}
	return habits, nil
}
