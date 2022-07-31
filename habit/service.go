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

func NewHabitService(sqlRepo *SQLHabitRepository, instumentation HabitInstrumentation) *HabitService {
	return &HabitService{sqlRepo, instumentation}
}

func (service *HabitService) Create(ctx context.Context, name string) (Habit, error) {
	habit, err := service.repo.Create(ctx, name)

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

func (service *HabitService) FindByName(ctx context.Context, name string) (Habit, error) {
	habit, err := service.repo.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, HabitNotFoundErr) {
			return habit, err
		} else {
			return habit, RepositoryErr
		}
	}

	return habit, nil
}

func (service *HabitService) List(ctx context.Context) ([]Habit, error) {
	return service.repo.List(ctx)
}

func (service *HabitService) DeleteByName(ctx context.Context, name string) error {
	return service.repo.DeleteByName(ctx, name)
}

func (service *HabitService) DeleteAll(ctx context.Context) error {
	return service.repo.DeleteAll(ctx)
}
