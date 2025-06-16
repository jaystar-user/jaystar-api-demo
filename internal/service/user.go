package service

import (
	"context"
	"errors"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/xerrors"
	"jaystar/internal/constant/user"
	"jaystar/internal/database"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/errs"
)

func ProvideUserService(
	db database.IPostgresDB,
	userRepo interfaces.IUserRepo,
	userCommonSrv interfaces.IUserCommonSrv,
	studentCommonSrv interfaces.IStudentCommonSrv,
	logger logger.ILogger,
) *UserService {
	return &UserService{
		userRepo:         userRepo,
		DB:               db,
		userCommonSrv:    userCommonSrv,
		studentCommonSrv: studentCommonSrv,
		logger:           logger,
	}
}

type UserService struct {
	userRepo         interfaces.IUserRepo
	DB               database.IPostgresDB
	userCommonSrv    interfaces.IUserCommonSrv
	studentCommonSrv interfaces.IStudentCommonSrv
	logger           logger.ILogger
}

func (srv *UserService) GetUsers(ctx context.Context, cond *bo.UserCond) ([]*bo.User, *po.PagerResult, error) {
	poUserCond := &po.UserCond{
		UserId:            cond.UserId,
		Accounts:          cond.Accounts,
		Status:            cond.Status.ToKey(),
		IsChangedPassword: cond.IsChangedPassword,
	}

	poPagerCond := &po.Pager{
		Index: cond.Index,
		Size:  cond.Size,
	}

	db := srv.DB.Session()
	poUsers, err := srv.userRepo.GetUsers(ctx, db, poUserCond, poPagerCond)
	if err != nil {
		return nil, nil, xerrors.Errorf("userService GetUsers userRepo.GetUsers: %w", err)
	}

	poPager, err := srv.userRepo.GetUsersPager(ctx, db, poUserCond, poPagerCond)
	if err != nil {
		return nil, nil, xerrors.Errorf("userService GetUsers userRepo.GetUsersPager: %w", err)
	}

	boUsers := make([]*bo.User, 0, len(poUsers))
	for _, poUser := range poUsers {
		boUsers = append(boUsers, &bo.User{
			UserId:            poUser.UserId,
			Account:           poUser.Account,
			Password:          poUser.Password,
			Status:            user.UserStatusToEnum(poUser.Status),
			Level:             user.UserLevelToEnum(poUser.Level),
			IsChangedPassword: poUser.IsChangedPassword,
			CreatedAt:         poUser.CreatedAt,
			UpdatedAt:         poUser.UpdatedAt,
		})
	}

	return boUsers, poPager, nil
}

func (srv *UserService) GetUser(ctx context.Context, cond *bo.UserCond) (*bo.User, error) {
	boUser, err := srv.userCommonSrv.GetUser(ctx, cond)
	if err != nil {
		if errors.Is(err, errs.DbErr.NoRow) {
			return nil, errs.UserErr.UserNotFoundErr
		}
		return nil, err
	}
	return boUser, nil
}

func (srv *UserService) UserLogin(ctx *gin.Context, cond *bo.UserLoginCond) (*bo.User, error) {
	poUser, err := srv.userCommonSrv.UserLogin(ctx, cond)
	if err != nil {
		return nil, xerrors.Errorf("userService UserLogin userCommonSrv.UserLogin: %w", err)
	}

	userStatus := user.UserStatusToEnum(poUser.Status)
	if userStatus == user.Deactivate {
		return nil, xerrors.Errorf("userService UserLogin status Deactivate: %w", errs.UserErr.AccountDeactivatedErr)
	}

	boUser := &bo.User{
		UserId:            poUser.UserId,
		Account:           poUser.Account,
		Password:          poUser.Password,
		Status:            userStatus,
		Level:             user.UserLevelToEnum(poUser.Level),
		IsChangedPassword: poUser.IsChangedPassword,
		CreatedAt:         poUser.CreatedAt,
		UpdatedAt:         poUser.UpdatedAt,
	}

	return boUser, nil
}

func (srv *UserService) UserRegister(ctx context.Context, cond *bo.StudentCond, data *bo.CreateUserData) error {
	boStudentReq := &dto.StudentReq{
		StudentName: cond.StudentName,
		ParentPhone: cond.ParentPhone,
	}
	students, _, err := srv.studentCommonSrv.GetKintoneStudents(ctx, boStudentReq)
	if err != nil {
		return xerrors.Errorf("userService UserRegister studentCommonSrv.GetKintoneStudents: %w", err)
	}

	if len(students) == 0 {
		return xerrors.Errorf("userService UserRegister: %w", errs.StudentErr.GetKintoneStudentNotFoundErr)
	}

	return srv.userCommonSrv.CreateUserAndStudent(ctx, data, students[0])
}

func (srv *UserService) UpdateUser(ctx context.Context, cond *bo.UserCond, data *bo.UpdateUserData) error {
	boUser, err := srv.userCommonSrv.GetUser(ctx, cond)
	if err != nil {
		if errors.Is(err, errs.DbErr.NoRow) {
			return xerrors.Errorf("userService UpdateUser userCommonSrv.GetUser not found: %w", errs.UserErr.UserNotFoundErr)
		}
		return xerrors.Errorf("userService UpdateUser userCommonSrv.GetUser: %w", err)
	}

	poUpdateUserData := &po.UpdateUserData{}

	if data.Password != nil {
		if *data.Password == "" {
			return xerrors.Errorf("userService UpdateUser data.Password empty: %w", errs.UserErr.UpdatePasswordEmptyError)
		}

		hashedPwd, err := srv.userCommonSrv.GetHashedPasswordFromEncrypted(*data.Password)
		if err != nil {
			return xerrors.Errorf("userService UpdateUser getHashedPasswordFromEncrypted: %w", err)
		}

		poUpdateUserData.Password = hashedPwd
	}

	if data.IsChangedPassword != nil {
		poUpdateUserData.IsChangedPassword = data.IsChangedPassword
	}

	if data.Status != user.StatusNone {
		poUpdateUserData.Status = data.Status.ToKey()
	}

	db := srv.DB.Session()
	if err := srv.userRepo.UpdateUser(ctx, db, &po.UserCond{UserId: boUser.UserId}, poUpdateUserData); err != nil {
		return xerrors.Errorf("userService UpdateUser userRepo.UpdateUser: %w", err)
	}

	return nil
}

func (srv *UserService) GetEncryptedPassword(ctx context.Context, data dto.PasswordIO) (string, error) {
	pwd, err := srv.userCommonSrv.GetEncryptedPwd(*data.Password)
	if err != nil {
		return "", xerrors.Errorf("userService GetEncryptedPassword userCommonSrv.GetEncryptedPwd: %w", err)
	}

	return pwd, nil
}
