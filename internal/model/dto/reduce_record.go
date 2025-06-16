package dto

import (
	"time"
)

type KintoneWebhookReduceRecordIO struct {
	KintoneWebhookIO
	Record ReduceRecordRecord `json:"record"`
}

type ReduceRecordGetIO struct {
	StudentName    *string    `form:"student_name"`
	ParentPhone    *string    `form:"parent_phone"`
	ClassTimeStart *time.Time `form:"class_time_start" time_format:"2006-01-02"`
	ClassTimeEnd   *time.Time `form:"class_time_end" time_format:"2006-01-02"`
	*PagerIO
}

type ReduceRecordGetVO struct {
	RecordType   string  `json:"record_type"`
	RecordId     string  `json:"record_id"`
	StudentName  string  `json:"student_name"`
	ParentPhone  string  `json:"parent_phone,omitempty"`
	ClassLevel   string  `json:"class_level"`
	ClassType    string  `json:"class_type"`
	ClassTime    string  `json:"class_time"`
	TeacherName  string  `json:"teacher_name"`
	ReducePoints float64 `json:"reduce_points"`
	IsAttended   bool    `json:"is_attended"`
}

type ReduceRecordSyncIO struct {
	ClassTimeStart *time.Time `json:"class_time_start"`
	ClassTimeEnd   *time.Time `json:"class_time_end"`
}

type AdminGetReduceRecordsIO struct {
	StudentName    *string    `form:"student_name"`
	ParentPhone    *string    `form:"parent_phone"`
	ClassTimeStart *time.Time `form:"class_time_start" time_format:"2006-01-02"`
	ClassTimeEnd   *time.Time `form:"class_time_end" time_format:"2006-01-02"`
	IsDeleted      *bool      `form:"is_deleted"`
	*PagerIO
}

type AdminGetReduceRecordsVO struct {
	ReduceRecordGetVO
	AdminVO
}
