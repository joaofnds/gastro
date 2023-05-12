package habit

import "time"

type Group struct {
	ID     string  `json:"id" gorm:"default:uuid_generate_v4()"`
	Name   string  `json:"name"`
	UserID string  `json:"user_id"`
	Habits []Habit `json:"habits" gorm:"many2many:groups_habits;ForeignKey:id,user_id;joinForeignKey:group_id,user_id"`
}

type Habit struct {
	ID         string     `json:"id" gorm:"default:uuid_generate_v4()"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	Activities []Activity `json:"activities" gorm:"foreignKey:HabitID"`
}

type Activity struct {
	ID        string    `json:"id" gorm:"default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"created_at"`
	Desc      string    `json:"description" gorm:"column:description"`
	HabitID   string
}
