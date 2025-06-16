package repository

import (
	"context"
	"gorm.io/gorm"
	"jaystar/internal/config"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/po"
	"time"
)

func ProvideScheduleRepository(cfg config.IConfigEnv) *ScheduleRepo {
	return &ScheduleRepo{ICommonRepo: ProvideCommonRepository()}
}

type ScheduleRepo struct {
	interfaces.ICommonRepo `wire:"-"`
}

func (repo *ScheduleRepo) GetSchedules(ctx context.Context, db *gorm.DB, cond *po.GetScheduleCond, pager *po.Pager) ([]*po.ScheduleView, error) {
	schedules := make([]*po.ScheduleView, 0)
	schedule := &po.Schedule{}
	tableName := schedule.TableName()
	if err := db.
		Model(schedule).
		Select(tableName + ".*, students.student_name, students.parent_phone").
		Joins("LEFT JOIN students on " + tableName + ".student_id = students.student_id").
		Scopes(repo.makeGetScheduleCond(ctx, cond, pager)).
		Find(&schedules).Error; err != nil {
		return nil, handleDBError(err)
	}

	return schedules, nil
}

func (repo *ScheduleRepo) GetSchedulesPager(ctx context.Context, db *gorm.DB, cond *po.GetScheduleCond, pager *po.Pager) (*po.PagerResult, error) {
	var total int64
	schedule := &po.Schedule{}
	if err := db.
		Model(schedule).
		Joins("LEFT JOIN students on " + schedule.TableName() + ".student_id = students.student_id").
		Scopes(repo.makeGetScheduleCond(ctx, cond, nil)).
		Count(&total).Error; err != nil {
		return nil, handleDBError(err)
	}

	return po.NewPagerResult(pager, total), nil
}

func (repo *ScheduleRepo) GetSchedulesByRefId(ctx context.Context, db *gorm.DB, scheduleRefId int) ([]*po.ScheduleView, error) {
	schedules := make([]*po.ScheduleView, 0)
	schedule := &po.Schedule{}
	tableName := schedule.TableName()
	if err := db.
		Model(schedule).
		Select(tableName+".*, students.student_name, students.parent_phone").
		Joins("LEFT JOIN students on "+tableName+".student_id = students.student_id").
		Where("schedule_ref_id = ?", scheduleRefId).
		Find(&schedules).Error; err != nil {
		return nil, handleDBError(err)
	}

	return schedules, nil
}

func (repo *ScheduleRepo) AddSchedule(ctx context.Context, db *gorm.DB, data *po.Schedule) error {
	if err := db.Create(data).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *ScheduleRepo) UpdateSchedule(ctx context.Context, db *gorm.DB, cond *po.UpdateScheduleCond, data *po.UpdateScheduleData) error {
	updated := make(map[string]interface{})

	if data.StudentId != 0 {
		updated["student_id"] = data.StudentId
	}
	if data.RecordRefId != nil {
		updated["record_ref_id"] = data.RecordRefId
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
	if data.IsDeleted != nil {
		updated["is_deleted"] = *data.IsDeleted
	}
	if data.DeletedAt != nil {
		updated["deleted_at"] = *data.DeletedAt
	}

	if err := db.
		Model(&po.Schedule{}).
		Scopes(repo.makeUpdateScheduleCond(ctx, cond)).
		Updates(updated).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *ScheduleRepo) GetAllScheduleRefIds(ctx context.Context, db *gorm.DB, cond *po.GetScheduleCond) ([]int, error) {
	var scheduleRefId []int
	if err := db.
		Model(&po.Schedule{}).
		Distinct().
		Scopes(repo.makeGetScheduleCond(ctx, cond, nil)).
		Pluck("schedule_ref_id", &scheduleRefId).
		Error; err != nil {
		return nil, handleDBError(err)
	}

	return scheduleRefId, nil
}

func (repo *ScheduleRepo) makeGetScheduleCond(ctx context.Context, cond *po.GetScheduleCond, pager *po.Pager) func(db *gorm.DB) *gorm.DB {
	schedule := po.Schedule{}
	tableName := schedule.TableName()
	return func(db *gorm.DB) *gorm.DB {
		if cond != nil {
			if cond.StudentId != 0 {
				db = db.Where(tableName+".student_id = ?", cond.StudentId)
			}
			if cond.ScheduleRefId != 0 {
				db = db.Where(tableName+".schedule_ref_id = ?", cond.ScheduleRefId)
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

func (repo *ScheduleRepo) makeUpdateScheduleCond(ctx context.Context, cond *po.UpdateScheduleCond) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if cond != nil {
			if cond.ScheduleId != 0 {
				db = db.Where("schedule_id = ?", cond.ScheduleId)
			}
			if cond.ScheduleRefId != 0 {
				db = db.Where("schedule_ref_id = ?", cond.ScheduleRefId)
			}
			if cond.RecordRefId != 0 {
				db = db.Where("record_ref_id = ?", cond.RecordRefId)
			}
		}
		return db
	}
}
