package dto

type AdminGetUsersIO struct {
	Accounts          *string `form:"account"`
	Status            *bool   `form:"status"`
	IsChangedPassword *bool   `form:"is_changed_password"`
	*PagerIO
}

type UserGetIO struct {
	UserId   *string `form:"user_id"`
	Accounts *string `form:"accounts"`
}

type UserRegIO struct {
	Account     string `json:"account" binding:"required"`
	Password    string `json:"password" binding:"required"`
	ParentPhone string `json:"parent_phone" binding:"required"`
	StudentName string `json:"student_name" binding:"required"`
}

type UserLoginIO struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserIO struct {
	Password          *string `json:"password"`
	Status            *bool   `json:"status"`
	IsChangedPassword *bool   `json:"is_changed_password"`
}

type UpdateUserPasswordIO struct {
	Password string `json:"password" binding:"required"`
}

type UserVO struct {
	UserId            string `json:"user_id"`
	Level             string `json:"user_level"`
	Account           string `json:"account"`
	IsChangedPassword bool   `json:"is_changed_password"`
}

type AdminUserVO struct {
	UserId            string `json:"user_id"`
	Account           string `json:"account"`
	Status            string `json:"status"`
	Level             string `json:"user_level"`
	IsChangedPassword bool   `json:"is_changed_password"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type PasswordIO struct {
	Password *string `json:"password"`
}
