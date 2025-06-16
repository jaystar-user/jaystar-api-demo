package repository

import (
	"context"
	"gorm.io/gorm"
)

type CommonRepository struct{}

func ProvideCommonRepository() *CommonRepository {
	return &CommonRepository{}
}

func (repo *CommonRepository) ResetFromDeleted(ctx context.Context, db *gorm.DB, tableName string, whereScopes func(db *gorm.DB) *gorm.DB) error {
	value := map[string]interface{}{
		"is_deleted": false,
		"deleted_at": nil,
	}

	if err := db.
		WithContext(ctx).
		Table(tableName).
		Scopes(whereScopes).
		Updates(value).Error; err != nil {
		return handleDBError(err)
	}
	return nil
}
