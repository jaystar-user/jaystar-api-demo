package bo

import (
	kintoneConst "jaystar/internal/constant/kintone"
	"jaystar/internal/model/po"
	"time"
)

type KintoneDepositRecord struct {
	Id                   int
	ChargingMethod       []kintoneConst.ChargingMethod
	ActualChargingAmount int
	Gender               string
	DepositedPoints      int
	TeacherName          string
	Description          string
	ParentPhone          string
	ChargingDate         time.Time
	ChargingAmount       int
	ParentName           string // deprecated
	AccountLastFiveYards string
	TaxId                string
	KintoneStudentName   string
	StudentName          string // 學生姓名，不含家長電話的
	ChargingStatus       kintoneConst.ChargingStatus
	CreatedBy            string
	CreatedAt            time.Time
	UpdatedBy            string
	UpdatedAt            time.Time
}

type DepositRecord struct {
	RecordId             int64
	RecordRefId          int
	StudentName          string
	ParentPhone          string
	ChargingDate         time.Time
	TaxId                string
	AccountLastFiveYards string
	ChargingAmount       int
	TeacherName          string
	DepositedPoints      int
	ChargingMethod       []kintoneConst.ChargingMethod
	ChargingStatus       kintoneConst.ChargingStatus
	ActualChargingAmount int
	IsDeleted            bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
}

type DepositRecordCond struct {
	RecordRefId       int
	ChargingDateStart time.Time
	ChargingDateEnd   time.Time
	IsDeleted         *bool
	po.Pager
}

type UpdateDepositRecordData struct {
	ChargingDate         time.Time
	TaxId                string
	AccountLastFiveYards string
	ChargingAmount       *int
	TeacherName          string
	DepositedPoints      *int
	ChargingMethod       []kintoneConst.ChargingMethod
	ChargingStatus       kintoneConst.ChargingStatus
	ActualChargingAmount *int
}

type SyncDepositRecordCond struct {
	StudentName       *string
	ParentPhone       *string
	ChargingDateStart *time.Time
	ChargingDateEnd   *time.Time
	Offset            *int
	Limit             *int
}

type StudentTotalDepositPointsCond struct {
	StudentIds        []int64
	ChargingDateStart time.Time
	ChargingDateEnd   time.Time
}

type StudentTotalDepositPoints struct {
	StudentId          int64
	TotalDepositPoints int
}
