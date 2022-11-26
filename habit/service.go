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

func (service *Service) CreateGroup(ctx context.Context, dto CreateGroupDTO) (Group, error) {
	return service.repo.CreateGroup(ctx, dto)
}

func (service *Service) AddToGroup(ctx context.Context, habit Habit, group Group) error {
	return service.repo.AddToGroup(ctx, habit, group)
}

func (service *Service) RemoveFromGroup(ctx context.Context, habit Habit, group Group) error {
	return service.repo.RemoveFromGroup(ctx, habit, group)
}

func (service *Service) FindGroup(ctx context.Context, dto FindGroupDTO) (Group, error) {
	return service.repo.FindGroup(ctx, dto)
}

func (service *Service) DeleteGroup(ctx context.Context, group Group) error {
	return service.repo.DeleteGroup(ctx, group)
}

func (service *Service) GroupsAndHabits(ctx context.Context, userID string) ([]Group, []Habit, error) {
	return service.repo.GroupsAndHabits(ctx, userID)
}

func (service *Service) switchErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrNotFound):
		return err
	default:
		return ErrRepository
	}
}
