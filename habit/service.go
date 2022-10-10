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

func NewHabitService(sqlRepo HabitRepository, instrumentation HabitInstrumentation) *HabitService {
	return &HabitService{sqlRepo, instrumentation}
}

func (service *HabitService) Create(ctx context.Context, create CreateDTO) (Habit, error) {
	habit, err := service.repo.Create(ctx, create)

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

func (service *HabitService) Find(ctx context.Context, find FindDTO) (Habit, error) {
	habit, err := service.repo.Find(ctx, find)
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

func (service *HabitService) Delete(ctx context.Context, find FindDTO) error {
	return service.repo.Delete(ctx, find)
}

func (service *HabitService) DeleteAll(ctx context.Context) error {
	return service.repo.DeleteAll(ctx)
}
