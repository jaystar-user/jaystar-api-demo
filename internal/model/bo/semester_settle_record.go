package bo

import (
	"jaystar/internal/model/po"
	"time"
)

type SemesterSettleRecord struct {
	RecordId           int64
	RecordRefId        int
	KintoneStudentName string
	StudentName        string // 學生姓名，不含家長電話的
	ParentPhone        string // 家長電話
	StartTime          time.Time
	EndTime            time.Time
	ClearPoints        float64
	CreatedBy          string
	CreatedAt          time.Time
	UpdatedBy          string
	IsDeleted          bool
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

type SemesterSettleRecordCond struct {
	StartTime time.Time
	EndTime   time.Time
	IsDeleted *bool
	po.Pager
}

type UpdateSemesterSettleRecordCond struct {
	RecordRefId int
}

type UpdateSemesterSettleRecordData struct {
	StartTime   time.Time
	EndTime     time.Time
	ClearPoints *float64
}

type SyncSemesterSettleRecordCond struct {
	StudentName *string
	ParentPhone *string
	StartTime   *time.Time
	EndTime     *time.Time
	Offset      *int
	Limit       *int
}

type SettleSemesterPointsCond struct {
	Date time.Time
}
