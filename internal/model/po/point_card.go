package po

import "time"

type PointCard struct {
	RecordId    int64   `gorm:"column:record_id;autoIncrement"`
	RecordRefId int     `gorm:"column:record_ref_id"`
	StudentId   int64   `gorm:"column:student_id"`
	RestPoints  float64 `gorm:"column:rest_points"`
	DeleteRelatedColumns
	BaseTimeColumns
}

func (PointCard) TableName() string {
	return "point_card"
}

type PointCardCond struct {
	StudentIds  []int64
	RecordRefId int
	IsDeleted   *bool
}

type UpdatePointCardCond struct {
	StudentId   int64
	RecordRefId int
}

type UpdatePointCardData struct {
	StudentId  int64
	RestPoints *float64
	IsDeleted  *bool
	DeletedAt  *time.Time
}
