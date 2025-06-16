package po

import "time"

type SemesterSettleRecord struct {
	RecordId    int64     `gorm:"column:record_id"`
	RecordRefId int       `gorm:"column:record_ref_id"`
	StudentId   int64     `gorm:"column:student_id"`
	StartTime   time.Time `gorm:"column:start_time"`
	EndTime     time.Time `gorm:"column:end_time"`
	ClearPoints float64   `gorm:"column:clear_points"`
	DeleteRelatedColumns
	BaseTimeColumns
}

func (SemesterSettleRecord) TableName() string {
	return "semester_settle_records"
}

type SemesterSettleRecordView struct {
	SemesterSettleRecord
	StudentName string `gorm:"column:student_name"` // 欄位是透過 join students 取得
	ParentPhone string `gorm:"column:parent_phone"` // 欄位是透過 join students 取得
}

type SemesterSettleRecordCond struct {
	StudentId   int64
	RecordRefId int
	StartTime   time.Time
	EndTime     time.Time
	IsDeleted   *bool
}

type UpdateSemesterSettleRecordCond struct {
	RecordRefId int
}

type UpdateSemesterSettleRecordData struct {
	StudentId   *int64
	StartTime   *time.Time
	EndTime     *time.Time
	ClearPoints *float64
	IsDeleted   *bool
	DeletedAt   *time.Time
}
