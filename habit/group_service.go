package habit

import (
	"context"
)

type GroupService struct {
	probe Probe
	repo  GroupRepository
}

func NewGroupService(probe Probe, groupRepo GroupRepository) *GroupService {
	return &GroupService{probe: probe, repo: groupRepo}
}

func (service *GroupService) Create(ctx context.Context, dto CreateGroupDTO) (Group, error) {
	return service.repo.Create(ctx, dto)
}

func (service *GroupService) Find(ctx context.Context, dto FindGroupDTO) (Group, error) {
	return service.repo.Find(ctx, dto)
}

func (service *GroupService) Delete(ctx context.Context, group Group) error {
	return service.repo.Delete(ctx, group)
}

func (service *GroupService) Join(ctx context.Context, habit Habit, group Group) error {
	return service.repo.Join(ctx, habit, group)
}

func (service *GroupService) Leave(ctx context.Context, habit Habit, group Group) error {
	return service.repo.Leave(ctx, habit, group)
}

func (service *GroupService) GroupsAndHabits(ctx context.Context, userID string) ([]Group, []Habit, error) {
	return service.repo.GroupsAndHabits(ctx, userID)
}
