package habit

import "context"

type Group struct {
	ID     string  `json:"id" gorm:"default:uuid_generate_v4()"`
	Name   string  `json:"name"`
	UserID string  `json:"user_id"`
	Habits []Habit `json:"habits" gorm:"many2many:groups_habits;ForeignKey:id,user_id;joinForeignKey:group_id,user_id"`
}

type CreateGroupDTO struct {
	Name   string
	UserID string
}

type FindGroupDTO struct {
	GroupID string
	UserID  string
}

type GroupRepository interface {
	Create(ctx context.Context, dto CreateGroupDTO) (Group, error)
	Find(ctx context.Context, dto FindGroupDTO) (Group, error)
	Delete(ctx context.Context, group Group) error
	Join(ctx context.Context, habit Habit, group Group) error
	Leave(ctx context.Context, habit Habit, group Group) error
	GroupsAndHabits(ctx context.Context, userID string) ([]Group, []Habit, error)
}
