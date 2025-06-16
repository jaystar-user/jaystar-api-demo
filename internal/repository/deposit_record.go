package repository

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/po"
	"time"
)

func ProvideDepositRecordRepository() *DepositRecordRepo {
	return &DepositRecordRepo{ICommonRepo: ProvideCommonRepository()}
}

type DepositRecordRepo struct {
	interfaces.ICommonRepo `wire:"-"`
}

func (repo *DepositRecordRepo) GetRecords(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond, pager *po.Pager) ([]*po.DepositRecordView, error) {
	records := make([]*po.DepositRecordView, 0)
	depositRecord := &po.DepositRecord{}
	tableName := depositRecord.TableName()
	if err := db.
		Model(depositRecord).
		Select(tableName + ".*, students.student_name, students.parent_phone").
		Joins("LEFT JOIN students on " + tableName + ".student_id = students.student_id").
		Scopes(repo.makeDepositRecordCond(ctx, cond, pager)).
		Find(&records).Error; err != nil {
		return nil, handleDBError(err)
	}
	return records, nil
}

func (repo *DepositRecordRepo) GetRecordsPager(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond, pager *po.Pager) (*po.PagerResult, error) {
	var total int64
	depositRecord := &po.DepositRecord{}
	if err := db.
		Model(depositRecord).
		Joins("LEFT JOIN students on " + depositRecord.TableName() + ".student_id = students.student_id").
		Scopes(repo.makeDepositRecordCond(ctx, cond, nil)).
		Count(&total).Error; err != nil {
		return nil, handleDBError(err)
	}
	return po.NewPagerResult(pager, total), nil
}

func (repo *DepositRecordRepo) GetDepositRecordRefIds(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond) ([]int, error) {
	var recordIds []int

	if err := db.
		Model(&po.DepositRecord{}).
		Scopes(repo.makeDepositRecordCond(ctx, cond, nil)).
		Pluck("record_ref_id", &recordIds).
		Error; err != nil {
		return nil, handleDBError(err)
	}
	return recordIds, nil
}

func (repo *DepositRecordRepo) GetRecord(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond) (*po.DepositRecord, error) {
	record := &po.DepositRecord{}
	if err := db.
		Model(&po.DepositRecord{}).
		Scopes(repo.makeDepositRecordCond(ctx, cond, nil)).
		First(&record).Error; err != nil {
		return nil, handleDBError(err)
	}
	return record, nil
}

func (repo *DepositRecordRepo) AddRecord(ctx context.Context, db *gorm.DB, data *po.DepositRecord) error {
	if err := db.Create(data).Error; err != nil {
		return handleDBError(err)
	}
	return nil
}

func (repo *DepositRecordRepo) UpdateRecord(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond, data *po.UpdateDepositRecordData) error {
	updated := make(map[string]interface{})

	if data.StudentId != 0 {
		updated["student_id"] = data.StudentId
	}
	// 這裡是用自定義的 struct 去寫入 []string 至 postgres array
	// 有包裝過 gorm scan/value/type interface
	if len(data.ChargingMethod.Values) > 0 {
		updated["charging_method"] = data.ChargingMethod
	}
	if data.TaxId != "" {
		updated["tax_id"] = data.TaxId
	}
	if data.ChargingDate != nil {
		updated["charging_date"] = *data.ChargingDate
	}
	if data.AccountLastFiveYards != "" {
		updated["account_last_five_yards"] = data.AccountLastFiveYards
	}
	if data.TeacherName != "" {
		updated["teacher_name"] = data.TeacherName
	}
	if data.ChargingAmount != nil {
		updated["charging_amount"] = *data.ChargingAmount
	}
	if data.DepositedPoints != nil {
		updated["deposited_points"] = *data.DepositedPoints
	}
	if data.ActualChargingAmount != nil {
		updated["actual_charging_amount"] = *data.ActualChargingAmount
	}
	if data.HitStatus != nil {
		updated["hit_status"] = *data.HitStatus
	}
	if data.IsDeleted != nil {
		updated["is_deleted"] = *data.IsDeleted
	}
	if data.DeletedAt != nil {
		updated["deleted_at"] = *data.DeletedAt
	}

	if err := db.
		Model(&po.DepositRecord{}).
		Where("record_ref_id = ?", cond.RecordRefId).
		Updates(updated).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *DepositRecordRepo) GetStudentTotalDepositPoints(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond) ([]*po.StudentTotalDepositPoints, error) {
	records := make([]*po.StudentTotalDepositPoints, 0)
	depositRecord := &po.DepositRecord{}
	tableName := depositRecord.TableName()
	if err := db.
		Model(depositRecord).
		Select(fmt.Sprintf("%s.student_id, sum(%s.deposited_points) as total_deposited_points", tableName, tableName)).
		Scopes(repo.makeDepositRecordCond(ctx, cond, nil)).
		Group(tableName + ".student_id").Find(&records).Error; err != nil {
		return nil, handleDBError(err)
	}

	return records, nil
}

func (repo *DepositRecordRepo) makeDepositRecordCond(ctx context.Context, cond *po.DepositRecordCond, pager *po.Pager) func(db *gorm.DB) *gorm.DB {
	tableName := new(po.DepositRecord).TableName()
	return func(db *gorm.DB) *gorm.DB {
		if cond != nil {
			if cond.StudentId != 0 {
				db = db.Where(tableName+".student_id = ?", cond.StudentId)
			}

			if len(cond.StudentIds) > 0 {
				db = db.Where(tableName+".student_id IN ?", cond.StudentIds)
			}

			if cond.RecordRefId != 0 {
				db = db.Where(tableName+".record_ref_id = ?", cond.RecordRefId)
			}

			if !cond.ChargingDateStart.IsZero() {
				db = db.Where(tableName+".charging_date >= ?", cond.ChargingDateStart.Format(time.DateTime))
			}

			if !cond.ChargingDateEnd.IsZero() {
				db = db.Where(tableName+".charging_date <= ?", cond.ChargingDateEnd.Format(time.DateTime))
			}

			if cond.IsDeleted != nil {
				db = db.Where(tableName+".is_deleted = ?", *cond.IsDeleted)
			}
		}
		if pager != nil {
			db.Scopes(parsePaging(pager))
		}
		return db
	}
}
