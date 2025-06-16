package repository

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"jaystar/internal/config"
	"jaystar/internal/constant/user"
	"jaystar/internal/model/po"
)

func ProvideUserRepository(cfg config.IConfigEnv) *UserRepo {
	return &UserRepo{
		cfg: cfg,
	}
}

type UserRepo struct {
	cfg config.IConfigEnv
}

func (repo *UserRepo) GetUsers(ctx context.Context, db *gorm.DB, cond *po.UserCond, pager *po.Pager) ([]*po.User, error) {
	poUsers := make([]*po.User, 0)

	if err := db.
		Model(&po.User{}).
		Scopes(repo.makeUserCond(ctx, cond, pager)).
		Find(&poUsers).Error; err != nil {
		return nil, handleDBError(err)
	}

	return poUsers, nil
}

func (repo *UserRepo) GetUsersPager(ctx context.Context, db *gorm.DB, cond *po.UserCond, pager *po.Pager) (*po.PagerResult, error) {
	var count int64

	if err := db.
		Model(&po.User{}).
		Scopes(repo.makeUserCond(ctx, cond, pager)).
		Count(&count).Error; err != nil {
		return nil, handleDBError(err)
	}

	return po.NewPagerResult(pager, count), nil
}

func (repo *UserRepo) GetUser(ctx context.Context, db *gorm.DB, cond *po.UserCond) (*po.User, error) {
	poUser := &po.User{}

	if err := db.
		Model(&po.User{}).
		Scopes(repo.makeUserCond(ctx, cond, nil)).
		First(&poUser).Error; err != nil {
		return nil, handleDBError(err)
	}
	return poUser, nil
}

func (repo *UserRepo) CreateUser(ctx context.Context, db *gorm.DB, data *po.User) error {
	if err := db.Create(data).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *UserRepo) UpdateUser(ctx context.Context, db *gorm.DB, cond *po.UserCond, data *po.UpdateUserData) error {
	updated := make(map[string]interface{})

	if data.Password != "" {
		updated["password"] = data.Password
	}
	if data.Status != "" {
		updated["status"] = data.Status
	}
	if data.IsChangedPassword != nil {
		updated["is_changed_password"] = *data.IsChangedPassword
	}

	if err := db.
		Model(&po.User{}).
		Scopes(repo.makeUserCond(ctx, cond, nil)).
		Updates(updated).Error; err != nil {
		return handleDBError(err)
	}

	return nil
}

func (repo *UserRepo) DeactivateUserWithNoActiveStudent(ctx context.Context, db *gorm.DB) ([]*po.User, error) {
	var updatedUsers []*po.User
	userWithActiveStudentSubQuery := db.Model(&po.Student{}).Select("user_id").Where("is_deleted = ?", false).Group("user_id")
	if err := db.
		Model(&updatedUsers).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "user_id"}, {Name: "account"}}}).
		Where("status != ?", user.Deactivate.ToKey()).
		Not("user_id IN (?)", userWithActiveStudentSubQuery).
		Update("status", user.Deactivate.ToKey()).
		Error; err != nil {
		return nil, handleDBError(err)
	}

	return updatedUsers, nil
}

func (repo *UserRepo) makeUserCond(ctx context.Context, cond *po.UserCond, pager *po.Pager) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if cond != nil {
			if cond.UserId != 0 {
				db = db.Where("user_id = ?", cond.UserId)
			}
			if len(cond.Accounts) > 0 {
				db = db.Where("account in ?", cond.Accounts)
			}
			if cond.Password != "" {
				db = db.Where("password = ?", cond.Password)
			}
			if cond.Status != "" {
				db = db.Where("status = ?", cond.Status)
			}
			if cond.Level != "" {
				db = db.Where("level = ?", cond.Level)
			}
			if cond.IsChangedPassword != nil {
				db = db.Where("is_changed_password = ?", *cond.IsChangedPassword)
			}
		}

		if pager != nil {
			db.Scopes(parsePaging(pager))
		}

		return db
	}
}
