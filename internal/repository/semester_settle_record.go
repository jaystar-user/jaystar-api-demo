package repository

import (
	"context"
	"gorm.io/gorm"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/po"
	"time"
)

func ProvideSemesterSettleRecordRepository() *SemesterSettleRecordRepository {
	return &SemesterSettleRecordRepository{ICommonRepo: ProvideCommonRepository()}
}

type SemesterSettleRecordRepository struct {
	interfaces.ICommonRepo `wire:"-"`
}

func (repo *SemesterSettleRecordRepository) GetRecords(ctx context.Context, db *gorm.DB, cond *po.SemesterSettleRecordCond, pager *po.Pager) ([]*po.SemesterSettleRecordView, error) {
	records := make([]*po.SemesterSettleRecordView, 0)
	semesterSettleRecord := new(po.SemesterSettleRecord)
	tableName := semesterSettleRecord.TableName()
	if err := db.
		Model(semesterSettleRecord).
		Select(tableName + ".*, students.student_name, students.parent_phone").
		Joins("LEFT JOIN students on " + tableName + ".student_id = students.student_id").
		Scopes(repo.makeGetSemesterSettleRecordsCond(ctx, cond, pager)).
		Find(&records).Error; err != nil {
		return nil, handleDBError(err)
	}

	return records, nil
}

func (repo *SemesterSettleRecordRepository) GetRecordsPager(ctx context.Context, db *gorm.DB, cond *po.SemesterSettleRecordCond, pager *po.Pager) (*po.PagerResult, error) {
	var total int64
	semesterSettleRecord := &po.SemesterSettleRecord{}
	tableName := semesterSettleRecord.TableName()
	if err := db.
		Model(semesterSettleRecord).
		Joins("LEFT JOIN students ON " + tableName + ".student_id = students.student_id").
		Scopes(repo.makeGetSemesterSettleRecordsCond(ctx, cond, nil)).
		Count(&total).Error; err != nil {
		return nil, handleDBError(err)
	}

	return po.NewPagerResult(pager, total), nil
}

func (repo *SemesterSettleRecordRepository) GetRecord(ctx context.Context, db *gorm.DB, cond *po.SemesterSettleRecordCond) (*po.SemesterSettleRecord, error) {
	semesterSettleRecord := &po.SemesterSettleRecord{}

	if err := db.
		Model(semesterSettleRecord).
		Scopes(repo.makeGetSemesterSettleRecordsCond(ctx, cond, nil)).
		First(&semesterSettleRecord).Error; err != nil {
		return nil, handleDBError(err)
	}

	return semesterSettleRecord, nil
}

func (repo *SemesterSettleRecordRepository) AddRecord(ctx context.Context, db *gorm.DB, data *po.SemesterSettleRecord) error {
	if err := db.WithContext(ctx).Create(data).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *SemesterSettleRecordRepository) UpdateRecord(ctx context.Context, db *gorm.DB, cond *po.UpdateSemesterSettleRecordCond, data *po.UpdateSemesterSettleRecordData) error {
	updated := make(map[string]interface{})

	if data.StartTime != nil {
		updated["start_time"] = data.StartTime
	}
	if data.EndTime != nil {
		updated["end_time"] = data.EndTime
	}
	if data.ClearPoints != nil {
		updated["clear_points"] = data.ClearPoints
	}
	if data.IsDeleted != nil {
		updated["is_deleted"] = *data.IsDeleted
	}
	if data.DeletedAt != nil {
		updated["deleted_at"] = *data.DeletedAt
	}

	if err := db.
		Model(&po.SemesterSettleRecord{}).
		Where("record_ref_id = ?", cond.RecordRefId).
		Updates(updated).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *SemesterSettleRecordRepository) GetSemesterSettleRecordRefIds(ctx context.Context, db *gorm.DB, cond *po.SemesterSettleRecordCond) ([]int, error) {
	var semesterSettleRecordRefIds []int

	if err := db.
		Model(&po.SemesterSettleRecord{}).
		Scopes(repo.makeGetSemesterSettleRecordsCond(ctx, cond, nil)).
		Pluck("record_ref_id", &semesterSettleRecordRefIds).
		Error; err != nil {
		return nil, handleDBError(err)
	}

	return semesterSettleRecordRefIds, nil
}

func (repo *SemesterSettleRecordRepository) makeGetSemesterSettleRecordsCond(ctx context.Context, cond *po.SemesterSettleRecordCond, pager *po.Pager) func(db *gorm.DB) *gorm.DB {
	tableName := new(po.SemesterSettleRecord).TableName()
	return func(db *gorm.DB) *gorm.DB {
		if cond != nil {
			if cond.StudentId != 0 {
				db = db.Where(tableName+".student_id = ?", cond.StudentId)
			}
			if cond.RecordRefId != 0 {
				db = db.Where(tableName+".record_ref_id = ?", cond.RecordRefId)
			}
			if !cond.StartTime.IsZero() {
				db = db.Where(tableName+".start_time >= ?", cond.StartTime.Format(time.DateTime))
			}
			if !cond.EndTime.IsZero() {
				db = db.Where(tableName+".end_time <= ?", cond.EndTime.Format(time.DateTime))
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
