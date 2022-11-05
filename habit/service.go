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

func (service *Service) Create(ctx context.Context, create CreateHabitDTO) (Habit, error) {
	habit, err := service.repo.Create(ctx, create)

	if err != nil {
		service.instrumentation.LogFailedToCreateHabit(err)
	} else {
		service.instrumentation.LogHabitCreated()
	}

	return habit, err
}

func (service *Service) Update(ctx context.Context, dto UpdateHabitDTO) error {
	err := service.repo.Update(ctx, dto)
	return service.switchErr(err)
}

func (service *Service) AddActivity(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error) {
	dto.Time = dto.Time.UTC().Truncate(time.Second)
	return service.repo.AddActivity(ctx, habit, dto)
}

func (service *Service) UpdateActivity(ctx context.Context, dto UpdateActivityDTO) (Activity, error) {
	return service.repo.UpdateActivity(ctx, dto)
}

func (service *Service) FindActivity(ctx context.Context, find FindActivityDTO) (Activity, error) {
	return service.repo.FindActivity(ctx, find)
}

func (service *Service) DeleteActivity(ctx context.Context, activity Activity) error {
	return service.repo.DeleteActivity(ctx, activity)
}

func (service *Service) Find(ctx context.Context, find FindHabitDTO) (Habit, error) {
	habit, err := service.repo.Find(ctx, find)
	return habit, service.switchErr(err)
}

func (service *Service) List(ctx context.Context, userID string) ([]Habit, error) {
	return service.repo.List(ctx, userID)
}

func (service *Service) Delete(ctx context.Context, find FindHabitDTO) error {
	return service.repo.Delete(ctx, find)
}

func (service *Service) DeleteAll(ctx context.Context) error {
	return service.repo.DeleteAll(ctx)
}

func (service *Service) switchErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, ErrNotFound) {
		return err
	}

	return ErrRepository
}
