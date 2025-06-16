package repository

import (
	"context"
	"gorm.io/gorm"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/po"
)

func ProvidePointCardRepository() *PointCardRepo {
	return &PointCardRepo{ICommonRepo: ProvideCommonRepository()}
}

type PointCardRepo struct {
	interfaces.ICommonRepo `wire:"-"`
}

func (repo *PointCardRepo) GetPointCard(ctx context.Context, db *gorm.DB, cond *po.PointCardCond) (*po.PointCard, error) {
	record := &po.PointCard{}

	if err := db.
		Model(&po.PointCard{}).
		Scopes(repo.makePointCardCond(ctx, cond)).
		First(&record).Error; err != nil {
		return nil, handleDBError(err)
	}

	return record, nil
}

func (repo *PointCardRepo) GetPointCards(ctx context.Context, db *gorm.DB, cond *po.PointCardCond) ([]*po.PointCard, error) {
	records := make([]*po.PointCard, 0)

	if err := db.
		Model(&po.PointCard{}).
		Scopes(repo.makePointCardCond(ctx, cond)).
		Find(&records).Error; err != nil {
		return nil, handleDBError(err)
	}

	return records, nil
}

func (repo *PointCardRepo) AddPointCard(ctx context.Context, db *gorm.DB, data *po.PointCard) error {
	if err := db.Create(data).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *PointCardRepo) UpdatePointCard(ctx context.Context, db *gorm.DB, cond *po.UpdatePointCardCond, data *po.UpdatePointCardData) error {
	updated := make(map[string]interface{})

	if data.StudentId != 0 {
		updated["student_id"] = data.StudentId
	}
	if data.RestPoints != nil {
		updated["rest_points"] = *data.RestPoints
	}
	if data.IsDeleted != nil {
		updated["is_deleted"] = *data.IsDeleted
	}
	if data.DeletedAt != nil {
		updated["deleted_at"] = *data.DeletedAt
	}

	if err := db.
		Model(&po.PointCard{}).
		Scopes(repo.makeUpdatePointCardCond(ctx, cond)).
		Updates(updated).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *PointCardRepo) GetPointCardRefIds(ctx context.Context, db *gorm.DB, cond *po.PointCardCond) ([]int, error) {
	var recordRefIds []int

	if err := db.
		Model(&po.PointCard{}).
		Scopes(repo.makePointCardCond(ctx, cond)).
		Pluck("record_ref_id", &recordRefIds).
		Error; err != nil {
		return nil, handleDBError(err)
	}

	return recordRefIds, nil
}

func (repo *PointCardRepo) makePointCardCond(ctx context.Context, cond *po.PointCardCond) func(db *gorm.DB) *gorm.DB {
	tableName := new(po.PointCard).TableName()
	return func(db *gorm.DB) *gorm.DB {
		if cond != nil {
			if cond.RecordRefId != 0 {
				db = db.Where(tableName+".record_ref_id = ?", cond.RecordRefId)
			}
			if len(cond.StudentIds) > 0 {
				db = db.Where(tableName+".student_id IN ?", cond.StudentIds)
			}
			if cond.IsDeleted != nil {
				db = db.Where(tableName+".is_deleted = ?", *cond.IsDeleted)
			}
		}
		return db
	}
}

func (repo *PointCardRepo) makeUpdatePointCardCond(ctx context.Context, cond *po.UpdatePointCardCond) func(db *gorm.DB) *gorm.DB {
	tableName := new(po.PointCard).TableName()
	return func(db *gorm.DB) *gorm.DB {
		if cond != nil {
			if cond.RecordRefId != 0 {
				db = db.Where(tableName+".record_ref_id = ?", cond.RecordRefId)
			}
			if cond.StudentId != 0 {
				db = db.Where(tableName+".student_id = ?", cond.StudentId)
			}
		}
		return db
	}
}
