package bo

import (
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/po"
	"time"
)

type KintoneSchedule struct {
	ScheduleRefId      int
	RecordRefId        *int
	TeacherName        string
	ClassLevel         kintone.ClassLevel
	ClassType          kintone.ClassType
	ClassTime          time.Time
	Description        string
	KintoneStudentName string
	StudentName        string // 學生姓名，不含家長電話的
	ParentPhone        string // 家長電話
	CreatedBy          string
	CreatedAt          time.Time
	UpdatedBy          string
	UpdatedAt          time.Time
}

type Schedule struct {
	ScheduleId    int64
	ScheduleRefId int
	RecordRefId   *int
	StudentId     int64
	StudentName   string
	ParentPhone   string
	ClassLevel    kintone.ClassLevel
	ClassType     kintone.ClassType
	ClassTime     time.Time
	TeacherName   string
	IsDeleted     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

type GetScheduleCond struct {
	ClassTimeStart time.Time
	ClassTimeEnd   time.Time
	IsDeleted      *bool
	po.Pager
}

type UpdateScheduleCond struct {
	ScheduleId    int64
	ScheduleRefId int
	RecordRefId   int
}

type UpdateScheduleData struct {
	RecordRefId *int
	ClassLevel  kintone.ClassLevel
	ClassType   kintone.ClassType
	ClassTime   time.Time
	TeacherName string
}

type SyncScheduleCond struct {
	StudentName    *string
	ParentPhone    *string
	ClassTimeStart *time.Time
	ClassTimeEnd   *time.Time
	Limit          *int
	Offset         *int
}
