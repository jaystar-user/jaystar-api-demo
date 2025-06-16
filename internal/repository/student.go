package repository

import (
	"context"
	"gorm.io/gorm"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/errs"
)

func ProvideStudentRepository() *StudentRepo {
	return &StudentRepo{ICommonRepo: ProvideCommonRepository()}
}

type StudentRepo struct {
	interfaces.ICommonRepo `wire:"-"`
}

func (repo *StudentRepo) GetStudents(ctx context.Context, db *gorm.DB, cond *po.StudentCond, pager *po.Pager) ([]*po.StudentWithBalance, error) {
	student := new(po.Student)
	students := make([]*po.StudentWithBalance, 0)
	tableName := student.TableName()

	if err := db.
		Model(student).
		Select(tableName + ".*, pc.rest_points").
		Joins("INNER JOIN point_card pc on pc.student_id = " + tableName + ".student_id").
		Scopes(repo.makeStudentCond(ctx, cond, pager)).
		Find(&students).Error; err != nil {
		return nil, errs.ParseDBError(err)
	}
	return students, nil
}

func (repo *StudentRepo) GetStudentsPager(ctx context.Context, db *gorm.DB, cond *po.StudentCond, pager *po.Pager) (*po.PagerResult, error) {
	var count int64

	if err := db.
		Model(&po.Student{}).
		Scopes(repo.makeStudentCond(ctx, cond, nil)).
		Count(&count).Error; err != nil {
		return nil, errs.ParseDBError(err)
	}
	return po.NewPagerResult(pager, count), nil
}

func (repo *StudentRepo) GetStudent(ctx context.Context, db *gorm.DB, cond *po.StudentCond) (*po.StudentWithBalance, error) {
	student := &po.StudentWithBalance{}
	tableName := student.TableName()
	if err := db.
		Model(&po.Student{}).
		Select(tableName + ".*, pc.rest_points").
		Joins("INNER JOIN point_card pc on pc.student_id = " + tableName + ".student_id").
		Scopes(repo.makeStudentCond(ctx, cond, nil)).
		First(&student).Error; err != nil {
		return nil, errs.ParseDBError(err)
	}
	return student, nil
}

func (repo *StudentRepo) GetStudentRefIds(ctx context.Context, db *gorm.DB, cond *po.StudentCond) ([]int, error) {
	var studentRefIds []int
	if err := db.
		Model(&po.Student{}).
		Scopes(repo.makeStudentCond(ctx, cond, nil)).
		Pluck("student_ref_id", &studentRefIds).
		Error; err != nil {
		return nil, errs.ParseDBError(err)
	}
	return studentRefIds, nil
}

func (repo *StudentRepo) CreateStudent(ctx context.Context, db *gorm.DB, data *po.Student) error {
	if err := db.Create(data).Error; err != nil {
		return errs.ParseDBError(err)
	}

	return nil
}

func (repo *StudentRepo) UpdateStudent(ctx context.Context, db *gorm.DB, cond *po.UpdateStudentCond, data *po.UpdateStudentData) error {
	updated := make(map[string]interface{})

	if data.UserId != 0 {
		updated["user_id"] = data.UserId
	}
	if data.StudentName != "" {
		updated["student_name"] = data.StudentName
	}
	if data.ParentName != "" {
		updated["parent_name"] = data.ParentName
	}
	if data.ParentPhone != "" {
		updated["parent_phone"] = data.ParentPhone
	}
	if data.IsDeleted != nil {
		updated["is_deleted"] = *data.IsDeleted
	}
	if data.Mode != "" {
		updated["mode"] = data.Mode
	}
	if data.IsSettleNormally != nil {
		updated["is_settle_normally"] = *data.IsSettleNormally
	}
	if data.DeletedAt != nil {
		updated["deleted_at"] = *data.DeletedAt
	}

	if err := db.
		Model(&po.Student{}).
		Scopes(repo.makeUpdateStudentCond(ctx, cond)).
		Updates(updated).Error; err != nil {
		return errs.ParseDBError(err)
	}

	return nil
}

func (repo *StudentRepo) makeStudentCond(ctx context.Context, cond *po.StudentCond, pager *po.Pager) func(*gorm.DB) *gorm.DB {
	tableName := new(po.Student).TableName()
	return func(db *gorm.DB) *gorm.DB {
		if cond != nil {
			if cond.StudentId != 0 {
				db = db.Where(tableName+".student_id = ?", cond.StudentId)
			}

			if cond.StudentRefId != 0 {
				db = db.Where(tableName+".student_ref_id = ?", cond.StudentRefId)
			}

			if cond.UserId != 0 {
				db = db.Where(tableName+".user_id = ?", cond.UserId)
			}

			if cond.StudentName != "" {
				db = db.Where(tableName+".student_name = ?", cond.StudentName)
			}

			if cond.ParentPhone != "" {
				db = db.Where(tableName+".parent_phone = ?", cond.ParentPhone)
			}

			if cond.IsDeleted != nil {
				db = db.Where(tableName+".is_deleted = ?", *cond.IsDeleted)
			}

			if cond.Mode != "" {
				db = db.Where(tableName+".mode = ?", cond.Mode)
			}

			if cond.IsSettleNormally != nil {
				db = db.Where(tableName+".is_settle_normally = ?", *cond.IsSettleNormally)
			}
		}

		if pager != nil {
			db.Scopes(parsePaging(pager))
		}

		return db
	}
}

func (repo *StudentRepo) makeUpdateStudentCond(ctx context.Context, cond *po.UpdateStudentCond) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if cond.StudentId != 0 {
			db = db.Where("student_id", cond.StudentId)
		}
		if cond.StudentRefId != 0 {
			db = db.Where("student_ref_id = ?", cond.StudentRefId)
		}
		return db
	}
}
