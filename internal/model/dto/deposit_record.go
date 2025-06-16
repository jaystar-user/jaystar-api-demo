package dto

import (
	"time"
)

type KintoneWebhookDepositRecordIO struct {
	KintoneWebhookIO
	Record KintoneDepositRecordRecordDto `json:"record"`
}

type DepositRecordGetIO struct {
	StudentName        *string    `form:"student_name"`
	ParentPhone        *string    `form:"parent_phone"`
	DepositedDateStart *time.Time `form:"deposited_date_start" time_format:"2006-01-02"`
	DepositedDateEnd   *time.Time `form:"deposited_date_end" time_format:"2006-01-02"`
	*PagerIO
}

type DepositRecordGetVO struct {
	RecordId             string   `json:"record_id"`
	StudentName          string   `json:"student_name"`
	ParentPhone          string   `json:"parent_phone,omitempty"`
	ChargingDate         string   `json:"charging_date"`
	TaxId                string   `json:"tax_id"`
	AccountLastFiveYards string   `json:"account_last_five_yards"`
	ChargingAmount       int      `json:"charging_amount"`
	TeacherName          string   `json:"teacher_name"`
	DepositedPoints      int      `json:"deposited_points"`
	ChargingMethod       []string `json:"charging_method"`
	ChargingStatus       string   `json:"charging_status"`
	ActualChargingAmount int      `json:"actual_charging_amount"`
}

type AdminDepositRecordGetVO struct {
	DepositRecordGetVO
	AdminVO
}

type SyncDepositRecordIO struct {
	DepositedDateStart *time.Time `json:"deposited_date_start"`
	DepositedDateEnd   *time.Time `json:"deposited_date_end"`
}

type AdminGetDepositRecordsIO struct {
	StudentName        *string    `form:"student_name"`
	ParentPhone        *string    `form:"parent_phone"`
	DepositedDateStart *time.Time `form:"deposited_date_start" time_format:"2006-01-02"`
	DepositedDateEnd   *time.Time `form:"deposited_date_end" time_format:"2006-01-02"`
	IsDeleted          *bool      `form:"is_deleted"`
	*PagerIO
}
