package bo

import (
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/po"
	"time"
)

type KintoneReduceRecord struct {
	Id                 int
	KintoneStudentName string
	StudentName        string // 學生姓名，不含家長電話的
	ParentPhone        string // 家長電話
	ClassLevel         kintone.ClassLevel
	ClassType          kintone.ClassType
	ClassTime          time.Time
	TeacherName        string
	ReducePoints       float64
	AttendStatus       bool
	Description        string
	CreatedBy          string
	CreatedAt          time.Time
	UpdatedBy          string
	UpdatedAt          time.Time
}

type ReduceRecord struct {
	RecordType   string
	RecordId     int64
	RecordRefId  int
	StudentName  string
	ParentPhone  string
	ClassLevel   kintone.ClassLevel
	ClassType    kintone.ClassType
	ClassTime    time.Time
	TeacherName  string
	ReducePoints float64
	IsAttended   bool
	IsDeleted    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type ReduceRecordCond struct {
	RecordRefId    int
	ClassTimeStart time.Time
	ClassTimeEnd   time.Time
	IsDeleted      *bool
	po.Pager
}

type UpdateReduceRecordData struct {
	ClassLevel   kintone.ClassLevel
	ClassType    kintone.ClassType
	ClassTime    time.Time
	TeacherName  string
	ReducePoints *float64
	IsAttended   *bool
}

type SyncReduceRecordCond struct {
	StudentName    *string
	ParentPhone    *string
	ClassTimeStart *time.Time
	ClassTimeEnd   *time.Time
	Offset         *int
	Limit          *int
}

type StudentTotalReducePointsCond struct {
	StudentIds     []int64
	ClassTimeStart time.Time
	ClassTimeEnd   time.Time
}

type StudentTotalReducePoints struct {
	StudentId         int64
	TotalReducePoints float64
}
