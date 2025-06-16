package dto

import "time"

type KintoneWebhookSemesterSettleRecordIO struct {
	KintoneWebhookIO
	Record SemesterSettleRecord `json:"record"`
}

type AdminGetSemesterSettleRecordsIO struct {
	StudentName *string    `form:"student_name"`
	ParentPhone *string    `form:"parent_phone"`
	StartTime   *time.Time `form:"start_time" time_format:"2006-01-02"`
	EndTime     *time.Time `form:"end_time" time_format:"2006-01-02"`
	IsDeleted   *bool      `form:"is_deleted"`
	*PagerIO
}

type GetSemesterSettleRecordsVO struct {
	RecordId    string  `json:"record_id"`
	StudentName string  `json:"student_name"`
	ParentPhone string  `json:"parent_phone"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	ClearPoints float64 `json:"clear_points"`
}

type AdminGetSemesterSettleRecordsVO struct {
	GetSemesterSettleRecordsVO
	AdminVO
}

type SyncSemesterSettleRecordIO struct {
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

type SettleSemesterPointsIO struct {
	Date string `json:"date"`
}
