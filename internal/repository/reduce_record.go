package repository

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"jaystar/internal/config"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/po"
	"time"
)

func ProvideReduceRecordRepository(cfg config.IConfigEnv) *ReduceRecordRepo {
	return &ReduceRecordRepo{ICommonRepo: ProvideCommonRepository()}
}

type ReduceRecordRepo struct {
	interfaces.ICommonRepo `wire:"-"`
}

func (repo *ReduceRecordRepo) GetRecordsWithSettleRecords(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond, pager *po.Pager) ([]*po.ReduceRecordSettleRecordView, error) {
	records := make([]*po.ReduceRecordSettleRecordView, 0)

	if err := db.
		Raw(
			fmt.Sprintf("? union all ? order by %s offset ? limit ?", pager.Order),
			db.Table("reduce_point_records as rpr").
				Select(`'reduce_record' AS "type", rpr.record_id, rpr.record_ref_id, rpr.student_id, s.student_name, s.parent_phone, rpr."class_level", rpr."class_type", rpr.class_time, rpr.teacher_name, rpr.reduce_points, rpr.is_attended, rpr.is_deleted, rpr.created_at, rpr.updated_at, rpr.deleted_at`).
				Joins("inner join students s on rpr.student_id = s.student_id").
				Scopes(repo.makeSReduceRecordCond(ctx, cond, "rpr")),
			db.Table("semester_settle_records ssr").
				Select(`'settle_record' AS "type", ssr.record_id, ssr.record_ref_id, ssr.student_id, s.student_name, s.parent_phone, null, null, ssr.end_time as class_time, '', ssr.clear_points as reduce_points, false, ssr.is_deleted, ssr.created_at, ssr.updated_at, ssr.deleted_at`).
				Joins("inner join students s on ssr.student_id = s.student_id").
				Scopes(repo.makeSSettleRecordCond(ctx, cond, "ssr")),
			pager.GetOffset(),
			pager.GetSize(),
		).
		Scan(&records).Error; err != nil {
		return nil, handleDBError(err)
	}

	return records, nil
}

func (repo *ReduceRecordRepo) GetRecordsWithSettleRecordsPager(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond, pager *po.Pager) (*po.PagerResult, error) {
	var total int64

	if err := db.
		Table(
			"(?) as p",
			db.Raw(
				fmt.Sprintf("? union all ?"),
				db.Table("reduce_point_records as rpr").
					Select(`COUNT(rpr.record_id) as count`).
					Joins("inner join students s on rpr.student_id = s.student_id").
					Scopes(repo.makeSReduceRecordCond(ctx, cond, "rpr")),
				db.Table("semester_settle_records ssr").
					Select(`COUNT(ssr.record_id) as count`).
					Joins("inner join students s on ssr.student_id = s.student_id").
					Scopes(repo.makeSSettleRecordCond(ctx, cond, "ssr")),
			),
		).Select("sum(p.count)").Find(&total).Error; err != nil {
		return nil, handleDBError(err)
	}

	return po.NewPagerResult(pager, total), nil
}

func (repo *ReduceRecordRepo) GetReduceRecordRefIds(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond) ([]int, error) {
	var reduceRecordRefIds []int

	if err := db.
		Model(&po.ReduceRecord{}).
		Scopes(repo.makeReduceRecordCond(ctx, cond, nil)).
		Pluck("record_ref_id", &reduceRecordRefIds).
		Error; err != nil {
		return nil, handleDBError(err)
	}

	return reduceRecordRefIds, nil
}

func (repo *ReduceRecordRepo) GetRecord(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond) (*po.ReduceRecord, error) {
	record := &po.ReduceRecord{}

	if err := db.
		Model(&po.ReduceRecord{}).
		Scopes(repo.makeReduceRecordCond(ctx, cond, nil)).
		First(&record).Error; err != nil {
		return nil, handleDBError(err)
	}

	return record, nil
}

func (repo *ReduceRecordRepo) AddRecord(ctx context.Context, db *gorm.DB, data *po.ReduceRecord) error {
	if err := db.Create(data).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *ReduceRecordRepo) UpdateRecord(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond, data *po.UpdateReduceRecordData) error {
	updated := make(map[string]interface{})

	if data.StudentId != 0 {
		updated["student_id"] = data.StudentId
	}
	if data.ClassType != "" {
		updated["class_type"] = data.ClassType
	}
	if data.ClassLevel != "" {
		updated["class_level"] = data.ClassLevel
	}
	if data.ClassTime != nil {
		updated["class_time"] = *data.ClassTime
	}
	if data.TeacherName != "" {
		updated["teacher_name"] = data.TeacherName
	}
	if data.ReducePoints != nil {
		updated["reduce_points"] = *data.ReducePoints
	}
	if data.IsAttended != nil {
		updated["is_attended"] = *data.IsAttended
	}
	if data.IsDeleted != nil {
		updated["is_deleted"] = *data.IsDeleted
	}
	if data.DeletedAt != nil {
		updated["deleted_at"] = *data.DeletedAt
	}

	if err := db.
		Model(&po.ReduceRecord{}).
		Where("record_ref_id = ?", cond.RecordRefId).
		Updates(updated).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *ReduceRecordRepo) GetStudentTotalReducePoints(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond) ([]*po.StudentTotalReducePoints, error) {
	records := make([]*po.StudentTotalReducePoints, 0)
	reduceRecord := &po.ReduceRecord{}
	tableName := reduceRecord.TableName()
	if err := db.
		Model(reduceRecord).
		Select(fmt.Sprintf("%s.student_id, sum(%s.reduce_points) as total_reduce_points", tableName, tableName)).
		Scopes(repo.makeReduceRecordCond(ctx, cond, nil)).
		Group(tableName + ".student_id").
		Find(&records).Error; err != nil {
		return nil, handleDBError(err)
	}
	return records, nil
}

func (repo *ReduceRecordRepo) makeReduceRecordCond(ctx context.Context, cond *po.ReduceRecordCond, pager *po.Pager) func(db *gorm.DB) *gorm.DB {
	tableName := new(po.ReduceRecord).TableName()
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

			if !cond.ClassTimeStart.IsZero() {
				db = db.Where(tableName+".class_time >= ?", cond.ClassTimeStart.Format(time.DateTime))
			}

			if !cond.ClassTimeEnd.IsZero() {
				db = db.Where(tableName+".class_time <= ?", cond.ClassTimeEnd.Format(time.DateTime))
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

func (repo *ReduceRecordRepo) makeSReduceRecordCond(ctx context.Context, cond *po.ReduceRecordCond, tableName string) func(db *gorm.DB) *gorm.DB {
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

			if !cond.ClassTimeStart.IsZero() {
				db = db.Where(tableName+".class_time >= ?", cond.ClassTimeStart.Format(time.DateTime))
			}

			if !cond.ClassTimeEnd.IsZero() {
				db = db.Where(tableName+".class_time <= ?", cond.ClassTimeEnd.Format(time.DateTime))
			}

			if cond.IsDeleted != nil {
				db = db.Where(tableName+".is_deleted = ?", *cond.IsDeleted)
			}
		}

		return db
	}
}

func (repo *ReduceRecordRepo) makeSSettleRecordCond(ctx context.Context, cond *po.ReduceRecordCond, tableName string) func(db *gorm.DB) *gorm.DB {
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

			if !cond.ClassTimeStart.IsZero() {
				db = db.Where(tableName+".end_time >= ?", cond.ClassTimeStart.Format(time.DateTime))
			}

			if !cond.ClassTimeEnd.IsZero() {
				db = db.Where(tableName+".end_time <= ?", cond.ClassTimeEnd.Format(time.DateTime))
			}

			if cond.IsDeleted != nil {
				db = db.Where(tableName+".is_deleted = ?", *cond.IsDeleted)
			}
		}

		return db
	}
}
