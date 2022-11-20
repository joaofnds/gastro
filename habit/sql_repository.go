package habit

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SQLRepository struct {
	DB Querier
}

type Querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db}
}

func (repo *SQLRepository) Create(ctx context.Context, create CreateHabitDTO) (Habit, error) {
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

func (repo *SQLRepository) Find(ctx context.Context, find FindHabitDTO) (Habit, error) {
	return repo.findOne(ctx, `
		SELECT
			habits.id,
			habits.user_id,
			habits.name,
			activities.id,
			activities.description,
			activities.created_at
		FROM habits
			LEFT JOIN activities ON activities.habit_id = habits.id
		WHERE habits.id = $1 AND habits.user_id = $2`,
		find.HabitID,
		find.UserID,
	)
}

func (repo *SQLRepository) Update(ctx context.Context, dto UpdateHabitDTO) error {
	result, err := repo.DB.ExecContext(
		ctx, `
    UPDATE habits
    SET name = $1
    WHERE id = $2`,
		dto.Name, dto.HabitID,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (repo *SQLRepository) AddActivity(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error) {
	row := repo.DB.QueryRowContext(
		ctx,
		"INSERT INTO activities(habit_id, description, created_at) VALUES ($1, $2, $3) RETURNING id",
		habit.ID, dto.Desc, dto.Time,
	)
	if row.Err() != nil {
		return Activity{}, row.Err()
	}

	var id string
	err := row.Scan(&id)
	if err != nil {
		return Activity{}, err
	}

	return Activity{ID: id, Desc: dto.Desc, CreatedAt: dto.Time}, row.Err()
}

func (repo *SQLRepository) FindActivity(ctx context.Context, find FindActivityDTO) (Activity, error) {
	row := repo.DB.QueryRowContext(
		ctx,
		`
			SELECT
				activities.id, activities.description, activities.created_at
			FROM
				activities
				INNER JOIN habits ON habits.id = activities.habit_id
			WHERE
				habits.user_id = $1
				AND habits.id = $2
				AND activities.id = $3
		`,
		find.UserID, find.HabitID, find.ActivityID,
	)

	if row.Err() != nil {
		return Activity{}, row.Err()
	}

	var (
		id        string
		desc      string
		createdAt time.Time
	)
	if err := row.Scan(&id, &desc, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Activity{}, ErrNotFound
		}
		return Activity{}, err
	}

	return Activity{ID: id, Desc: desc, CreatedAt: createdAt.UTC()}, row.Err()
}

func (repo *SQLRepository) UpdateActivity(ctx context.Context, dto UpdateActivityDTO) (Activity, error) {
	row := repo.DB.QueryRowContext(
		ctx,
		`
			UPDATE activities
			SET description = $1
			WHERE activities.id = $2
			RETURNING id, description, created_at
		`,
		dto.Desc, dto.ActivityID,
	)

	if row.Err() != nil {
		return Activity{}, row.Err()
	}

	var (
		id        string
		desc      string
		createdAt time.Time
	)
	err := row.Scan(&id, &desc, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Activity{}, ErrNotFound
		}
		return Activity{}, err
	}

	return Activity{ID: id, Desc: desc, CreatedAt: createdAt.UTC()}, row.Err()
}

func (repo *SQLRepository) DeleteActivity(ctx context.Context, activity Activity) error {
	_, err := repo.DB.ExecContext(ctx, "DELETE FROM activities WHERE id = $1", activity.ID)
	return err
}

func (repo *SQLRepository) List(ctx context.Context, userID string) ([]Habit, error) {
	habits, err := repo.userHabits(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toList(habits), nil
}

func (repo *SQLRepository) CreateGroup(ctx context.Context, dto CreateGroupDTO) (Group, error) {
	row := repo.DB.QueryRowContext(
		ctx,
		"INSERT INTO groups(name, user_id) VALUES($1, $2) RETURNING id",
		dto.Name, dto.UserID,
	)

	if row.Err() != nil {
		return Group{}, row.Err()
	}

	var id string
	if err := row.Scan(&id); err != nil {
		return Group{}, err
	}

	return Group{ID: id, Name: dto.Name}, nil
}

func (repo *SQLRepository) AddToGroup(ctx context.Context, habit Habit, group Group) error {
	_, err := repo.DB.ExecContext(
		ctx,
		"INSERT INTO groups_habits (group_id, habit_id, user_id) VALUES ($1, $2, $3)",
		group.ID, habit.ID, habit.UserID,
	)
	return err
}

func (repo *SQLRepository) RemoveFromGroup(ctx context.Context, habit Habit, group Group) error {
	_, err := repo.DB.ExecContext(
		ctx,
		"DELETE FROM groups_habits WHERE group_id = $1 AND habit_id = $2 AND user_id = $3",
		group.ID, habit.ID, habit.UserID,
	)
	return err
}

func (repo *SQLRepository) GroupsAndHabits(ctx context.Context, userID string) ([]Group, []Habit, error) {
	habits, err := repo.userHabits(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	rows, err := repo.DB.QueryContext(ctx, `
		SELECT
		    groups.id,
		    groups.name,
			groups_habits.habit_id
		FROM groups
			LEFT JOIN groups_habits ON groups.id = groups_habits.group_id and groups.user_id = groups_habits.user_id
		WHERE
		    groups.user_id = $1
	`, userID)

	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var (
		groupID   string
		groupName string
		habitID   sql.NullString
	)

	groups := map[string]*Group{}

	for rows.Next() {
		if err := rows.Scan(&groupID, &groupName, &habitID); err != nil {
			return nil, nil, err
		}

		group, ok := groups[groupID]
		if !ok {
			group = &Group{ID: groupID, Name: groupName}
			groups[groupID] = group
		}

		if habitID.Valid {
			habit := habits[habitID.String]
			group.Habits = append(group.Habits, *habit)
			delete(habits, habit.ID)
		}
	}

	return toList(groups), toList(habits), nil
}

func (repo *SQLRepository) FindGroup(ctx context.Context, dto FindGroupDTO) (Group, error) {
	rows, err := repo.DB.QueryContext(ctx, `
		SELECT
		    groups.id,
		    groups.name,
		    groups.user_id,
			habits.id,
			habits.name,
			activities.id,
			activities.description,
			activities.created_at
		FROM groups
			LEFT JOIN groups_habits ON groups.id = groups_habits.group_id AND groups.user_id = groups_habits.user_id
			LEFT JOIN habits ON groups_habits.habit_id = habits.id AND groups_habits.user_id = habits.user_id
			LEFT JOIN activities ON habits.id = activities.habit_id
		WHERE groups.id = $1 AND groups.user_id = $2
	`, dto.GroupID, dto.UserID)

	if err != nil {
		return Group{}, err
	}

	defer rows.Close()

	groups := map[string]*Group{}
	habits := map[string]Habit{}

	for rows.Next() {
		var (
			groupID           string
			groupName         string
			userID            string
			habitID           sql.NullString
			name              sql.NullString
			activityID        sql.NullString
			activityDesc      sql.NullString
			activityCreatedAt sql.NullTime
		)

		err := rows.Scan(&groupID, &groupName, &userID, &habitID, &name, &activityID, &activityDesc, &activityCreatedAt)
		if err != nil {
			return Group{}, err
		}
		group, ok := groups[groupID]
		if !ok {
			group = &Group{ID: groupID, Name: groupName}
			groups[groupID] = group
		}

		if habitID.Valid {
			habit, ok := habits[habitID.String]
			if !ok {
				habit = Habit{ID: habitID.String, UserID: userID, Name: name.String, Activities: []Activity{}}
				habits[habitID.String] = habit
			}

			group.Habits = append(group.Habits, habit)

			if activityID.Valid {
				activity := Activity{
					ID:        activityID.String,
					Desc:      activityDesc.String,
					CreatedAt: activityCreatedAt.Time.UTC(),
				}
				habit.Activities = append(habit.Activities, activity)
			}
		}
	}

	if len(groups) != 1 {
		return Group{}, ErrNotFound
	}

	return *groups[dto.GroupID], nil
}

func (repo *SQLRepository) Delete(ctx context.Context, find FindHabitDTO) error {
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
		return ErrNotFound
	}
	return err
}

func (repo *SQLRepository) DeleteAll(ctx context.Context) error {
	_, err := repo.DB.ExecContext(ctx, "DELETE FROM habits")
	return err
}

func (repo *SQLRepository) userHabits(ctx context.Context, userID string) (map[string]*Habit, error) {
	rows, err := repo.DB.QueryContext(ctx, `
		SELECT
			habits.id,
			habits.user_id,
			habits.name,
			activities.id,
			activities.description,
			activities.created_at
		FROM habits
			LEFT JOIN activities ON activities.habit_id = habits.id
		WHERE habits.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	habits, err := scanHabits(rows)
	if err != nil {
		return nil, err
	}

	return habits, nil
}

func scanHabits(rows *sql.Rows) (map[string]*Habit, error) {
	m := map[string]*Habit{}

	for rows.Next() {
		var (
			id                string
			userID            string
			name              string
			activityID        sql.NullString
			activityDesc      sql.NullString
			activityCreatedAt sql.NullTime
		)

		err := rows.Scan(&id, &userID, &name, &activityID, &activityDesc, &activityCreatedAt)
		if err != nil {
			return nil, err
		}

		habit, ok := m[id]
		if !ok {
			habit = &Habit{ID: id, UserID: userID, Name: name, Activities: []Activity{}}
			m[id] = habit
		}

		if activityID.Valid {
			activity := Activity{
				ID:        activityID.String,
				Desc:      activityDesc.String,
				CreatedAt: activityCreatedAt.Time.UTC(),
			}
			habit.Activities = append(habit.Activities, activity)
		}
	}

	return m, nil
}

func toList[T Habit | Group](m map[string]*T) []T {
	result := make([]T, len(m))
	var i int
	for _, h := range m {
		result[i] = *h
		i++
	}
	return result
}

func (repo *SQLRepository) findOne(ctx context.Context, query string, args ...any) (Habit, error) {
	rows, err := repo.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return Habit{}, err
	}

	defer rows.Close()

	habitMap, err := scanHabits(rows)
	if err != nil {
		return Habit{}, err
	}

	habits := toList(habitMap)

	if len(habits) == 0 {
		return Habit{}, ErrNotFound
	}

	return habits[0], nil
}
