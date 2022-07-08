package habit

import (
	"context"

	"go.uber.org/zap"
)

type HabitService struct {
	repo   HabitRepository
	logger *zap.Logger
}

func NewHabitService(sqlRepo *SQLHabitRepository, logger *zap.Logger) *HabitService {
	return &HabitService{sqlRepo, logger}
}

func (service *HabitService) Create(ctx context.Context, name string) (Habit, error) {
	return service.repo.Create(ctx, name)
}

func (service *HabitService) List(ctx context.Context) ([]Habit, error) {
	return service.repo.List(ctx)
}

func (service *HabitService) FindByName(ctx context.Context, name string) (Habit, error) {
	return service.repo.FindByName(ctx, name)
}

func (service *HabitService) DeleteByName(ctx context.Context, name string) error {
	return service.repo.DeleteByName(ctx, name)
}

func (service *HabitService) DeleteAll(ctx context.Context) error {
	return service.repo.DeleteAll(ctx)
}
