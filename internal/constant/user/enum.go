package user

var (
	UserStatusMap = map[UserStatus]Dto{
		Activate:   {Key: "activate", Value: "啟用"},
		Deactivate: {Key: "deactivate", Value: "停用"},
	}
	UserLevelMap = map[UserLevel]Dto{
		Admin: {Key: "admin", Value: "管理者"},
		User:  {Key: "user", Value: "使用者"},
	}
)

type Dto struct {
	Key   string
	Value string
}

type UserStatus int

const (
	StatusNone UserStatus = 0
	Activate   UserStatus = iota
	Deactivate
)

func (us UserStatus) ToKey() string {
	if v, found := UserStatusMap[us]; found {
		return v.Key
	}
	return ""
}

func (us UserStatus) ToValue() string {
	if v, found := UserStatusMap[us]; found {
		return v.Value
	}
	return ""
}

func UserStatusToEnum(raw string) UserStatus {
	switch raw {
	case "activate", "啟用":
		return Activate
	case "deactivate", "停用":
		return Deactivate
	default:
		return StatusNone
	}
}

type UserLevel int

const (
	LevelNone UserLevel = 0
	Admin     UserLevel = iota
	User
)

func (us UserLevel) ToKey() string {
	if v, found := UserLevelMap[us]; found {
		return v.Key
	}
	return ""
}

func (us UserLevel) ToValue() string {
	if v, found := UserLevelMap[us]; found {
		return v.Value
	}
	return ""
}

func UserLevelToEnum(raw string) UserLevel {
	switch raw {
	case "admin", "管理者":
		return Admin
	case "user", "使用者":
		return User
	default:
		return LevelNone
	}
}
