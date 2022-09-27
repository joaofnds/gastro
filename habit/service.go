package habit

import (
	"context"
	"errors"
	"time"
)

type HabitService struct {
	repo            HabitRepository
	instrumentation HabitInstrumentation
}

func NewHabitService(sqlRepo *SQLHabitRepository, instrumentation HabitInstrumentation) *HabitService {
	return &HabitService{sqlRepo, instrumentation}
}

func (service *HabitService) Create(ctx context.Context, userID, name string) (Habit, error) {
	habit, err := service.repo.Create(ctx, userID, name)

	if err != nil {
		service.instrumentation.LogFailedToCreateHabit(err)
	} else {
		service.instrumentation.LogHabitCreated()
	}

	return habit, err
}

func (service *HabitService) AddActivity(ctx context.Context, habit Habit, date time.Time) (Activity, error) {
	return service.repo.AddActivity(ctx, habit, date.Truncate(time.Second))
}

func (service *HabitService) FindByName(ctx context.Context, userID, name string) (Habit, error) {
	habit, err := service.repo.FindByName(ctx, userID, name)
	if err != nil {
		if errors.Is(err, HabitNotFoundErr) {
			return habit, err
		} else {
			return habit, RepositoryErr
		}
	}

	return habit, nil
}

func (service *HabitService) List(ctx context.Context, userID string) ([]Habit, error) {
	return service.repo.List(ctx, userID)
}

func (service *HabitService) DeleteByName(ctx context.Context, userID, name string) error {
	return service.repo.DeleteByName(ctx, userID, name)
}

func (service *HabitService) DeleteAll(ctx context.Context) error {
	return service.repo.DeleteAll(ctx)
}
