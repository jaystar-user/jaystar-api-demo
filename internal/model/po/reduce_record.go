package po

import (
	"time"
)

type ReduceRecord struct {
	RecordId     int64      `gorm:"column:record_id"`
	RecordRefId  int        `gorm:"column:record_ref_id"`
	StudentId    int64      `gorm:"column:student_id"`
	ClassType    string     `gorm:"column:class_type;type:class_type"`
	ClassLevel   string     `gorm:"column:class_level;type:class_level"`
	ClassTime    *time.Time `gorm:"column:class_time"`
	ReducePoints float64    `gorm:"column:reduce_points"`
	TeacherName  string     `gorm:"column:teacher_name"`
	IsAttended   bool       `gorm:"column:is_attended"`
	DeleteRelatedColumns
	BaseTimeColumns
}

type ReduceRecordView struct {
	ReduceRecord
	StudentName string `gorm:"column:student_name"` // 欄位是透過 join students 取得
}

type ReduceRecordSettleRecordView struct {
	ReduceRecord
	StudentName string `gorm:"column:student_name"` // 欄位是透過 join students 取得
	ParentPhone string `gorm:"column:parent_phone"` // 欄位是透過 join students 取得
	RecordType  string `gorm:"column:type"`
}

func (ReduceRecord) TableName() string {
	return "reduce_point_records"
}

type ReduceRecordCond struct {
	RecordRefId    int
	StudentId      int64
	StudentIds     []int64
	ClassTimeStart time.Time
	ClassTimeEnd   time.Time
	IsDeleted      *bool
}

type UpdateReduceRecordData struct {
	StudentId    int64
	ClassType    string
	ClassLevel   string
	ClassTime    *time.Time
	TeacherName  string
	ReducePoints *float64
	IsAttended   *bool
	IsDeleted    *bool
	DeletedAt    *time.Time
}

type StudentTotalReducePoints struct {
	StudentId         int64   `gorm:"column:student_id"`
	TotalReducePoints float64 `gorm:"column:total_reduce_points"`
}
