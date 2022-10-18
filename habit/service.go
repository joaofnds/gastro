package habit

import (
	"context"
	"errors"
	"time"
)

type Service struct {
	repo            Repository
	instrumentation Instrumentation
}

func NewService(sqlRepo Repository, instrumentation Instrumentation) *Service {
	return &Service{sqlRepo, instrumentation}
}

func (service *Service) Create(ctx context.Context, create CreateDTO) (Habit, error) {
	habit, err := service.repo.Create(ctx, create)

	if err != nil {
		service.instrumentation.LogFailedToCreateHabit(err)
	} else {
		service.instrumentation.LogHabitCreated()
	}

	return habit, err
}

func (service *Service) AddActivity(ctx context.Context, habit Habit, date time.Time) (Activity, error) {
	return service.repo.AddActivity(ctx, habit, date.UTC().Truncate(time.Second))
}

func (service *Service) FindActivity(ctx context.Context, find FindActivityDTO) (Activity, error) {
	return service.repo.FindActivity(ctx, find)
}

func (service *Service) DeleteActivity(ctx context.Context, activity Activity) error {
	return service.repo.DeleteActivity(ctx, activity)
}

func (service *Service) Find(ctx context.Context, find FindDTO) (Habit, error) {
	habit, err := service.repo.Find(ctx, find)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return habit, err
		} else {
			return habit, ErrRepository
		}
	}

	return habit, nil
}

func (service *Service) List(ctx context.Context, userID string) ([]Habit, error) {
	return service.repo.List(ctx, userID)
}

func (service *Service) Delete(ctx context.Context, find FindDTO) error {
	return service.repo.Delete(ctx, find)
}

func (service *Service) DeleteAll(ctx context.Context) error {
	return service.repo.DeleteAll(ctx)
}
