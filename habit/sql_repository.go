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

func NewSQLHabitRepository(db *sql.DB) *SQLHabitRepository {
	return &SQLHabitRepository{db}
}

func (repo *SQLHabitRepository) Create(ctx context.Context, create CreateDTO) (Habit, error) {
	row := repo.DB.QueryRowContext(
		ctx,
		"INSERT INTO habits(user_id, name) VALUES ($1, $2) RETURNING id",
		create.UserID, create.Name,
	)
	if row.Err() != nil {
		return Habit{}, row.Err()
	}

	var id string
	err := row.Scan(&id)
	if err != nil {
		return Habit{}, err
	}

	h := Habit{ID: id, UserID: create.UserID, Name: create.Name, Activities: []Activity{}}
	return h, row.Err()
}

func (repo *SQLHabitRepository) Find(ctx context.Context, find FindDTO) (Habit, error) {
	return repo.findOne(ctx, `
		SELECT
			habits.id,
			habits.user_id,
			habits.name,
			activities.id,
			activities.created_at
		FROM habits
			LEFT JOIN activities ON activities.habit_id = habits.id
		WHERE habits.id = $1 AND habits.user_id = $2`,
		find.HabitID,
		find.UserID,
	)
}

func (repo *SQLHabitRepository) AddActivity(ctx context.Context, habit Habit, time time.Time) (Activity, error) {
	row := repo.DB.QueryRowContext(
		ctx,
		"INSERT INTO activities(habit_id, created_at) VALUES ($1, $2) RETURNING id",
		habit.ID, time,
	)
	if row.Err() != nil {
		return Activity{}, row.Err()
	}

	var id string
	err := row.Scan(&id)
	if err != nil {
		return Activity{}, err
	}

	return Activity{id, time}, row.Err()
}

func (repo *SQLHabitRepository) List(ctx context.Context, userID string) ([]Habit, error) {
	rows, err := repo.DB.QueryContext(ctx, `
		SELECT
			habits.id,
			habits.user_id,
			habits.name,
			activities.id,
			activities.created_at
		FROM habits
			LEFT JOIN activities ON activities.habit_id = habits.id
		WHERE habits.user_id = $1`,
		userID,
	)
	if err != nil {
		return []Habit{}, err
	}

	defer rows.Close()

	return scanRows(rows)
}

func (repo *SQLHabitRepository) Delete(ctx context.Context, find FindDTO) error {
	r, err := repo.DB.ExecContext(
		ctx,
		"DELETE FROM habits WHERE id = $1 AND user_id = $2",
		find.HabitID,
		find.UserID,
	)
	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return HabitNotFoundErr
	}
	return err
}

func (repo *SQLHabitRepository) DeleteAll(ctx context.Context) error {
	_, err := repo.DB.ExecContext(ctx, "DELETE FROM habits")
	return err
}

func scanRows(rows *sql.Rows) ([]Habit, error) {
	m := map[string]*Habit{}

	for rows.Next() {
		var (
			id                string
			userID            string
			name              string
			activityID        sql.NullString
			activityCreatedAt sql.NullTime
		)

		err := rows.Scan(&id, &userID, &name, &activityID, &activityCreatedAt)
		if err != nil {
			return []Habit{}, err
		}

		habit, ok := m[id]
		if !ok {
			habit = &Habit{ID: id, UserID: userID, Name: name, Activities: []Activity{}}
			m[id] = habit
		}

		if activityID.Valid {
			activity := Activity{activityID.String, activityCreatedAt.Time}
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

func (repo *SQLHabitRepository) findOne(ctx context.Context, query string, args ...any) (Habit, error) {
	rows, err := repo.DB.QueryContext(ctx, query, args...)
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
