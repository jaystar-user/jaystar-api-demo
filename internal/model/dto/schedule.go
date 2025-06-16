package dto

import (
	"time"
)

type KintoneWebhookScheduleIO struct {
	KintoneWebhookIO
	Record ScheduleRecord `json:"record"`
}

type ScheduleGetIO struct {
	StudentName    *string   `form:"student_name"`
	ParentPhone    *string   `form:"parent_phone"`
	ClassTimeStart time.Time `form:"class_time_start" time_format:"2006-01-02"`
	ClassTimeEnd   time.Time `form:"class_time_end" time_format:"2006-01-02"`
	*PagerIO
}

type ScheduleSyncIO struct {
	ClassTimeStart *time.Time `json:"class_time_start"`
	ClassTimeEnd   *time.Time `json:"class_time_end"`
}

type ScheduleVO struct {
	Id          string `json:"id"`
	TeacherName string `json:"teacher_name"`
	StudentName string `json:"student_name"`
	ParentPhone string `json:"parent_phone,omitempty"`
	ClassLevel  string `json:"class_level"`
	ClassType   string `json:"class_type"`
	ClassTime   string `json:"class_time"`
}

type AdminGetSchedulesIO struct {
	StudentName    *string    `form:"student_name"`
	ParentPhone    *string    `form:"parent_phone"`
	ClassTimeStart *time.Time `form:"class_time_start" time_format:"2006-01-02"`
	ClassTimeEnd   *time.Time `form:"class_time_end" time_format:"2006-01-02"`
	IsDeleted      *bool      `form:"is_deleted"`
	*PagerIO
}

type AdminGetScheduleVO struct {
	ScheduleVO
	AdminVO
}
