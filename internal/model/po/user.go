package po

type User struct {
	UserId            int64  `gorm:"column:user_id"`
	Account           string `gorm:"column:account"`
	Password          string `gorm:"column:password"`
	Status            string `gorm:"column:status;type:user_status"`
	Level             string `gorm:"column:level;type:user_level"` // unused
	IsChangedPassword bool   `gorm:"column:is_changed_password"`
	BaseTimeColumns
}

func (User) TableName() string {
	return "users"
}

type UserCond struct {
	UserId            int64
	Accounts          []string
	Password          string
	Status            string
	Level             string
	IsChangedPassword *bool
}

type UserLoginCond struct {
	Account  string
	Password string
}

type UpdateUserData struct {
	Password          string
	Status            string
	IsChangedPassword *bool
}
