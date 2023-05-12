package habit

import (
	"context"
	"errors"
	"time"
)

type Service struct {
	probe        Probe
	habitRepo    HabitRepository
	activityRepo ActivityRepository
	groupRepo    GroupRepository
}

func NewService(
	probe Probe,
	habitRepo HabitRepository,
	activityRepo ActivityRepository,
	groupRepo GroupRepository,
) *Service {
	return &Service{
		probe:        probe,
		habitRepo:    habitRepo,
		activityRepo: activityRepo,
		groupRepo:    groupRepo,
	}
}

func (service *Service) Create(ctx context.Context, create CreateHabitDTO) (Habit, error) {
	habit, err := service.habitRepo.Create(ctx, create)

	if err != nil {
		service.probe.LogFailedToCreateHabit(err)
	} else {
		service.probe.LogHabitCreated()
	}

	return habit, err
}

func (service *Service) Update(ctx context.Context, dto UpdateHabitDTO) error {
	err := service.habitRepo.Update(ctx, dto)
	return service.switchErr(err)
}

func (service *Service) AddActivity(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error) {
	dto.Time = dto.Time.UTC().Truncate(time.Second)
	return service.activityRepo.Add(ctx, habit, dto)
}

func (service *Service) UpdateActivity(ctx context.Context, dto UpdateActivityDTO) (Activity, error) {
	return service.activityRepo.Update(ctx, dto)
}

func (service *Service) FindActivity(ctx context.Context, find FindActivityDTO) (Activity, error) {
	return service.activityRepo.Find(ctx, find)
}

func (service *Service) DeleteActivity(ctx context.Context, activity Activity) error {
	return service.activityRepo.Delete(ctx, activity)
}

func (service *Service) Find(ctx context.Context, find FindHabitDTO) (Habit, error) {
	habit, err := service.habitRepo.Find(ctx, find)
	return habit, service.switchErr(err)
}

func (service *Service) List(ctx context.Context, userID string) ([]Habit, error) {
	return service.habitRepo.List(ctx, userID)
}

func (service *Service) Delete(ctx context.Context, find FindHabitDTO) error {
	return service.habitRepo.Delete(ctx, find)
}

func (service *Service) DeleteAll(ctx context.Context) error {
	return service.habitRepo.DeleteAll(ctx)
}

func (service *Service) CreateGroup(ctx context.Context, dto CreateGroupDTO) (Group, error) {
	return service.groupRepo.Create(ctx, dto)
}

func (service *Service) AddToGroup(ctx context.Context, habit Habit, group Group) error {
	return service.groupRepo.Join(ctx, habit, group)
}

func (service *Service) RemoveFromGroup(ctx context.Context, habit Habit, group Group) error {
	return service.groupRepo.Leave(ctx, habit, group)
}

func (service *Service) FindGroup(ctx context.Context, dto FindGroupDTO) (Group, error) {
	return service.groupRepo.Find(ctx, dto)
}

func (service *Service) DeleteGroup(ctx context.Context, group Group) error {
	return service.groupRepo.Delete(ctx, group)
}

func (service *Service) GroupsAndHabits(ctx context.Context, userID string) ([]Group, []Habit, error) {
	return service.groupRepo.GroupsAndHabits(ctx, userID)
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
