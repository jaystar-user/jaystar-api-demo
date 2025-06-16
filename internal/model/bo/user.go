package bo

import (
	"jaystar/internal/constant/user"
	"jaystar/internal/model/po"
	"time"
)

type User struct {
	UserId            int64
	Account           string
	Password          string
	Status            user.UserStatus
	Level             user.UserLevel // unused
	IsChangedPassword bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type UserCond struct {
	UserId            int64
	Accounts          []string
	Status            user.UserStatus
	IsChangedPassword *bool
	po.Pager
}

type UserLoginCond struct {
	Account  string
	Password string
}

type CreateUserData struct {
	Account  string
	Password string
}

type UpdateUserData struct {
	Password          *string
	Status            user.UserStatus
	IsChangedPassword *bool
}
