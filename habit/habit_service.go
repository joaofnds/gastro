package habit

import (
	"context"
)

type HabitService struct {
	probe Probe
	repo  HabitRepository
}

func NewHabitService(
	probe Probe,
	habitRepo HabitRepository,
) *HabitService {
	return &HabitService{probe: probe, repo: habitRepo}
}

func (service *HabitService) Create(ctx context.Context, create CreateHabitDTO) (Habit, error) {
	habit, err := service.repo.Create(ctx, create)

	if err != nil {
		service.probe.LogFailedToCreateHabit(err)
	} else {
		service.probe.LogHabitCreated()
	}

	return habit, err
}

func (service *HabitService) Update(ctx context.Context, dto UpdateHabitDTO) error {
	return service.repo.Update(ctx, dto)
}

func (service *HabitService) Find(ctx context.Context, find FindHabitDTO) (Habit, error) {
	return service.repo.Find(ctx, find)
}

func (service *HabitService) List(ctx context.Context, userID string) ([]Habit, error) {
	return service.repo.List(ctx, userID)
}

func (service *HabitService) Delete(ctx context.Context, find FindHabitDTO) error {
	return service.repo.Delete(ctx, find)
}

func (service *HabitService) DeleteAll(ctx context.Context) error {
	return service.repo.DeleteAll(ctx)
}
