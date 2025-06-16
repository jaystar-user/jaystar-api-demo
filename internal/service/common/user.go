package common

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"github.com/SeanZhenggg/go-utils/snowflake/autoId"
	"github.com/forgoer/openssl"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/xerrors"
	"io"
	"jaystar/internal/constant/user"
	"jaystar/internal/database"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/errs"
)

func ProvideUserCommonService(db database.IPostgresDB, userRepo interfaces.IUserRepo, studentRepo interfaces.IStudentRepo) *UserCommonService {
	return &UserCommonService{
		DB:          db,
		userRepo:    userRepo,
		studentRepo: studentRepo,
	}
}

type UserCommonService struct {
	DB          database.IPostgresDB
	userRepo    interfaces.IUserRepo
	studentRepo interfaces.IStudentRepo
}

func (srv *UserCommonService) GetUser(ctx context.Context, cond *bo.UserCond) (*bo.User, error) {
	poUserCond := &po.UserCond{}

	// 以 userId 為優先條件
	// 沒帶 userId 則以帳號為條件
	if cond.UserId != 0 {
		poUserCond.UserId = cond.UserId
	} else if len(cond.Accounts) == 1 {
		poUserCond.Accounts = cond.Accounts
	}

	if cond.Status != user.StatusNone {
		poUserCond.Status = cond.Status.ToKey()
	}

	poUserCond.IsChangedPassword = cond.IsChangedPassword

	db := srv.DB.Session()
	poUser, err := srv.userRepo.GetUser(ctx, db, poUserCond)
	if err != nil {
		return nil, xerrors.Errorf("userCommonService GetUser userRepo.GetUser: %w", err)
	}

	boUser := &bo.User{
		UserId:            poUser.UserId,
		Account:           poUser.Account,
		Password:          poUser.Password,
		Status:            user.UserStatusToEnum(poUser.Status),
		Level:             user.UserLevelToEnum(poUser.Level),
		IsChangedPassword: poUser.IsChangedPassword,
		CreatedAt:         poUser.CreatedAt,
		UpdatedAt:         poUser.UpdatedAt,
	}

	return boUser, nil
}

func (srv *UserCommonService) CreateUser(ctx context.Context, data *bo.CreateUserData) (int64, error) {
	poUserData, err := srv.genCreateUserData(data)
	if err != nil {
		return 0, xerrors.Errorf("userCommonService CreateUser genCreateUserData: %w", err)
	}

	db := srv.DB.Session()
	if err := srv.userRepo.CreateUser(ctx, db, poUserData); err != nil {
		return 0, xerrors.Errorf("userCommonService CreateUser userRepo.CreateUser: %w", err)
	}

	return poUserData.UserId, nil
}

func (srv *UserCommonService) CreateUserAndStudent(ctx context.Context, data *bo.CreateUserData, studentData *bo.Student) error {
	if data.Account == "" || data.Password == "" {
		return errs.UserErr.AccountOrPasswordInvalidErr
	}

	poUserData, err := srv.genCreateUserData(data)
	if err != nil {
		return xerrors.Errorf("genCreateUserData: %w", err)
	}

	// transaction
	var isDuplicated bool
	tx := srv.DB.Session().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			return
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	tx.SavePoint("BeforeCreateUser")
	if err = srv.userRepo.CreateUser(ctx, tx, poUserData); err != nil {
		if errors.Is(err, errs.DbErr.UniqueViolation) {
			isDuplicated = true
			tx.RollbackTo("BeforeCreateUser")
		} else {
			return xerrors.Errorf("userRepo.CreateUser: %w", err)
		}
	}

	if isDuplicated {
		poUser, err := srv.userRepo.GetUser(ctx, tx, &po.UserCond{Accounts: []string{data.Account}})
		if err != nil {
			return xerrors.Errorf("userRepo.GetUser: %w", err)
		}
		if poUser != nil {
			poUserData.UserId = poUser.UserId
		}
	}

	studentId, err := autoId.DefaultSnowFlake.GenNextId()
	if err != nil {
		return xerrors.Errorf("autoId.DefaultSnowFlake.GenNextId: %w", err)
	}

	poStudent := &po.Student{
		StudentId:        studentId,
		StudentRefId:     studentData.StudentRefId,
		UserId:           poUserData.UserId,
		StudentName:      studentData.StudentName,
		ParentName:       studentData.ParentName,
		ParentPhone:      studentData.ParentPhone,
		Mode:             studentData.Mode.ToKey(),
		IsSettleNormally: studentData.IsSettleNormally,
	}
	if err := srv.studentRepo.CreateStudent(ctx, tx, poStudent); err != nil {
		return xerrors.Errorf("studentRepo.CreateStudent: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return xerrors.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (srv *UserCommonService) GetHashedPasswordFromEncrypted(password string) (string, error) {
	pwd, err := srv.getDecryptedPwd(password)
	if err != nil {
		return "", xerrors.Errorf("userCommonService getHashedPasswordFromEncrypted srv.getDecryptedPwd: %w", err)
	}

	return srv.GenHashedPassword(pwd)
}

func (srv *UserCommonService) GenHashedPassword(password []byte) (string, error) {
	// 字串加鹽
	passwordBts, err := srv.appendPasswordWithSaltRight(password)
	if err != nil {
		return "", xerrors.Errorf("userCommonService getHashedPasswordFromEncrypted srv.padPasswordWithSaltRight: %w", err)
	}

	// 雜湊
	hashedPwd, err := bcrypt.GenerateFromPassword(passwordBts, bcrypt.MinCost)
	if err != nil {
		return "", xerrors.Errorf("userCommonService getHashedPasswordFromEncrypted bcrypt.GenerateFromPassword: %w", err)
	}

	return string(hashedPwd), nil
}

func (srv *UserCommonService) UserLogin(ctx context.Context, cond *bo.UserLoginCond) (*po.User, error) {
	poUserLoginCond := &po.UserCond{}
	poUserLoginCond.Accounts = []string{cond.Account}

	db := srv.DB.Session()
	poUser, err := srv.userRepo.GetUser(ctx, db, poUserLoginCond)
	if err != nil {
		if errors.Is(err, errs.DbErr.NoRow) {
			return nil, xerrors.Errorf("userCommonService UserLogin userRepo.GetUser: %w", errs.UserErr.AccOrPwdVerificationFailedErr)
		}
		return nil, xerrors.Errorf("userCommonService UserLogin userRepo.GetUser: %w", err)
	}

	pwd, err := srv.getDecryptedPwd(cond.Password)
	if err != nil {
		return nil, xerrors.Errorf("userCommonService UserLogin.getDecryptedPwd: %w", err)
	}

	salted, err := srv.appendPasswordWithSaltRight(pwd)
	if err != nil {
		return nil, xerrors.Errorf("userCommonService UserLogin.appendPasswordWithSaltRight: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(poUser.Password), salted); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, xerrors.Errorf("userCommonService UserLogin bcrypt.CompareHashAndPassword: %w", errs.UserErr.AccOrPwdVerificationFailedErr)
		}
		return nil, xerrors.Errorf("userCommonService UserLogin bcrypt.CompareHashAndPassword: %w", err)
	}

	return poUser, nil
}

func (srv *UserCommonService) DeactivateUserWithNoActiveStudent(ctx context.Context) ([]*po.User, error) {
	db := srv.DB.Session()
	return srv.userRepo.DeactivateUserWithNoActiveStudent(ctx, db)
}

func (srv *UserCommonService) getDecryptedPwd(password string) ([]byte, error) {
	// base64 解碼
	decodeBts, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return nil, xerrors.Errorf("userCommonService getDecryptedPwd base64.StdEncoding.DecodeString: %w", err)
	}
	// AES ECB secret key 解密
	decryptPwd, err := openssl.AesECBDecrypt(decodeBts, []byte(user.PasswordAesECBKey), openssl.PKCS7_PADDING)
	if err != nil {
		return nil, xerrors.Errorf("userCommonService getDecryptedPwd openssl.AesECBDecrypt: %w", err)
	}

	return decryptPwd, nil
}

func (srv *UserCommonService) GetEncryptedPwd(password string) (string, error) {
	// AES ECB secret key 加密
	encryptedPwd, err := openssl.AesECBEncrypt([]byte(password), []byte(user.PasswordAesECBKey), openssl.PKCS7_PADDING)
	if err != nil {
		return "", xerrors.Errorf("userCommonService GetEncryptedPwd openssl.AesECBEncrypt: %w", err)
	}

	// base64 編碼
	encodedPwd := base64.StdEncoding.EncodeToString(encryptedPwd)

	return encodedPwd, nil
}

func (srv *UserCommonService) appendPasswordWithSaltRight(password []byte) ([]byte, error) {
	rd := bytes.NewBuffer(password)
	if _, err := rd.Write([]byte(user.PasswordSalt)); err != nil {
		return nil, xerrors.Errorf("userCommonService padPasswordWithSaltRight rd.Write: %w", err)
	}
	passwordBts, err := io.ReadAll(rd)
	if err != nil {
		return nil, xerrors.Errorf("userCommonService padPasswordWithSaltRight rd.Read: %w", err)
	}

	return passwordBts, nil
}

func (srv *UserCommonService) genCreateUserData(data *bo.CreateUserData) (*po.User, error) {
	userId, err := autoId.DefaultSnowFlake.GenNextId()
	if err != nil {
		return nil, xerrors.Errorf("userCommonService genCreateUserData autoId.DefaultSnowFlake.GenNextId: %w", err)
	}

	hashedPwd, err := srv.GetHashedPasswordFromEncrypted(data.Password)
	if err != nil {
		return nil, xerrors.Errorf("userCommonService genCreateUserData getHashedPasswordFromEncrypted: %w", err)
	}

	poUserData := &po.User{
		UserId:            userId,
		Account:           data.Account,
		Password:          hashedPwd,
		Status:            user.Activate.ToKey(),
		Level:             user.User.ToKey(),
		IsChangedPassword: false,
	}

	return poUserData, nil
}
