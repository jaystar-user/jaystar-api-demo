package po

import (
	"time"
)

type Student struct {
	StudentId        int64  `gorm:"column:student_id"`
	StudentRefId     int    `gorm:"column:student_ref_id"`
	UserId           int64  `gorm:"column:user_id"`
	StudentName      string `gorm:"column:student_name"`
	ParentName       string `gorm:"column:parent_name"`
	ParentPhone      string `gorm:"column:parent_phone"`
	Mode             string `gorm:"column:mode;type:mode"`
	IsSettleNormally bool   `gorm:"column:is_settle_normally"`
	DeleteRelatedColumns
	BaseTimeColumns
}

type StudentWithBalance struct {
	Student
	Balance float64 `gorm:"column:rest_points"`
}

func (Student) TableName() string {
	return "students"
}

type StudentCond struct {
	UserId           int64
	StudentName      string
	ParentPhone      string
	StudentId        int64
	StudentRefId     int
	Mode             string
	IsDeleted        *bool
	IsSettleNormally *bool
}

type UpdateStudentCond struct {
	StudentId    int64
	StudentRefId int
}

type UpdateStudentData struct {
	UserId           int64
	StudentName      string
	ParentName       string
	ParentPhone      string
	Mode             string
	IsSettleNormally *bool
	IsDeleted        *bool
	DeletedAt        *time.Time
}
