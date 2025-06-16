package repository

import (
	"gorm.io/gorm"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/errs"
)

func parsePaging(pager *po.Pager) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pager != nil {
			db = db.Order(pager.Order).Limit(pager.GetSize()).Offset(pager.GetOffset())
		}

		return db
	}
}

func handleDBError(err error) error {
	return errs.ParseDBError(err)
}
