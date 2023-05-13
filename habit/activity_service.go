package habit

import (
	"context"
	"time"
)

type ActivityService struct {
	probe Probe
	repo  ActivityRepository
}

func NewActivityService(
	probe Probe,
	repo ActivityRepository,
) *ActivityService {
	return &ActivityService{probe: probe, repo: repo}
}

func (service *ActivityService) Add(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error) {
	dto.Time = dto.Time.UTC().Truncate(time.Second)
	return service.repo.Add(ctx, habit, dto)
}

func (service *ActivityService) Update(ctx context.Context, dto UpdateActivityDTO) (Activity, error) {
	return service.repo.Update(ctx, dto)
}

func (service *ActivityService) Find(ctx context.Context, find FindActivityDTO) (Activity, error) {
	return service.repo.Find(ctx, find)
}

func (service *ActivityService) Delete(ctx context.Context, activity Activity) error {
	return service.repo.Delete(ctx, activity)
}
