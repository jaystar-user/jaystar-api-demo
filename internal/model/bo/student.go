package bo

import (
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/po"
	"time"
)

type StudentCond struct {
	UserId           int64
	StudentName      string
	ParentPhone      string
	StudentRefId     int
	StudentId        int64
	Mode             *kintone.Mode
	IsSettleNormally *bool
	IsDeleted        *bool
	*po.Pager
}

type Student struct {
	StudentId        int64
	StudentRefId     int
	UserId           int64
	StudentName      string
	ParentName       string
	ParentPhone      string
	Balance          float64
	IsDeleted        bool
	Mode             kintone.Mode
	IsSettleNormally bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type UpdateStudentCond struct {
	RecordRefId int
}

type UpdateStudentData struct {
	UserId           int64
	StudentName      string
	ParentName       string
	ParentPhone      string
	Balance          *float64
	Mode             *kintone.Mode
	IsSettleNormally *bool
}

type DeleteStudentData struct {
	IsDeleted bool
	DeletedAt time.Time
}

type AddStudentAndUserIfNeededResp struct {
	IsUpdated bool
	Account   string
}

type SyncStudentCond struct {
	StudentName *string
	ParentPhone *string
	Limit       *int
	Offset      *int
}
