package po

import (
	"time"
)

type Schedule struct {
	ScheduleId    int64      `gorm:"column:schedule_id"`
	ScheduleRefId int        `gorm:"column:schedule_ref_id"`
	RecordRefId   *int       `gorm:"column:record_ref_id"`
	StudentId     int64      `gorm:"column:student_id"`
	TeacherName   string     `gorm:"column:teacher_name"`
	ClassType     string     `gorm:"column:class_type;type:class_type"`
	ClassLevel    string     `gorm:"column:class_level;type:class_level"`
	ClassTime     *time.Time `gorm:"column:class_time"`
	DeleteRelatedColumns
	BaseTimeColumns
}

type ScheduleView struct {
	Schedule
	StudentName string `gorm:"column:student_name"` // 欄位是透過 join students 取得
	ParentPhone string `gorm:"column:parent_phone"` // 欄位是透過 join students 取得
}

func (Schedule) TableName() string {
	return "class_schedule"
}

type GetScheduleCond struct {
	StudentId      int64
	ScheduleRefId  int
	IsDeleted      *bool
	ClassTimeStart time.Time
	ClassTimeEnd   time.Time
}

type UpdateScheduleCond struct {
	ScheduleId    int64
	ScheduleRefId int
	RecordRefId   int
}

type UpdateScheduleData struct {
	StudentId   int64
	RecordRefId *int
	ClassType   string
	ClassLevel  string
	ClassTime   *time.Time
	TeacherName string
	IsDeleted   *bool
	DeletedAt   *time.Time
}
