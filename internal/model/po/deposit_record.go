package po

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type DepositRecord struct {
	RecordId             int64               `gorm:"column:record_id"`
	RecordRefId          int                 `gorm:"column:record_ref_id"`
	StudentId            int64               `gorm:"column:student_id"`
	ChargingDate         *time.Time          `gorm:"column:charging_date"`
	TaxId                string              `gorm:"column:tax_id"`
	AccountLastFiveYards string              `gorm:"column:account_last_five_yards"`
	ChargingAmount       int                 `gorm:"column:charging_amount"`
	TeacherName          string              `gorm:"column:teacher_name"`
	DepositedPoints      int                 `gorm:"column:deposited_points"`
	ChargingMethod       ChargingMethodArray `gorm:"column:charging_method;type:charging_method[]"`
	HitStatus            bool                `gorm:"column:hit_status"`
	ActualChargingAmount int                 `gorm:"column:actual_charging_amount"`
	DeleteRelatedColumns
	BaseTimeColumns
}

type DepositRecordView struct {
	DepositRecord
	StudentName string `gorm:"column:student_name"` // 欄位是透過 join students 取得
	ParentPhone string `gorm:"column:parent_phone"` // 欄位是透過 join students 取得
}

type DepositRecordCond struct {
	StudentId         int64
	StudentIds        []int64
	RecordRefId       int
	ChargingDateStart time.Time
	ChargingDateEnd   time.Time
	IsDeleted         *bool
}

type UpdateDepositRecordData struct {
	StudentId            int64
	ChargingDate         *time.Time
	TaxId                string
	AccountLastFiveYards string
	ChargingAmount       *int
	TeacherName          string
	DepositedPoints      *int
	ChargingMethod       ChargingMethodArray
	HitStatus            *bool
	ActualChargingAmount *int
	IsDeleted            *bool
	DeletedAt            *time.Time
}

type ChargingMethodArray struct {
	Values []string
}

// Scan 必須要用 pointer receiver method 才能將資料寫進自定義的struct field 裡
func (c *ChargingMethodArray) Scan(value interface{}) error {
	if v, ok := value.(string); ok {
		v = strings.ReplaceAll(v, "{", "")
		v = strings.ReplaceAll(v, "}", "")
		c.Values = strings.Split(v, ",")
	}
	return nil
}

func (c ChargingMethodArray) Value() (driver.Value, error) {
	return fmt.Sprintf("{%s}", strings.Join(c.Values, ",")), nil
}

func (DepositRecord) TableName() string {
	return "deposit_point_records"
}

type StudentTotalDepositPoints struct {
	StudentId            int64 `gorm:"column:student_id"`
	TotalDepositedPoints int   `gorm:"column:total_deposited_points"`
}
