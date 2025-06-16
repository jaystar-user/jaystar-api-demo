package interfaces

import (
	"context"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/model/po"
)

type IUserCommonSrv interface {
	GetUser(ctx context.Context, cond *bo.UserCond) (*bo.User, error)
	CreateUser(ctx context.Context, data *bo.CreateUserData) (int64, error)
	CreateUserAndStudent(ctx context.Context, data *bo.CreateUserData, studentData *bo.Student) error
	GetHashedPasswordFromEncrypted(password string) (string, error)
	GenHashedPassword(password []byte) (string, error)
	UserLogin(ctx context.Context, cond *bo.UserLoginCond) (*po.User, error)
	GetEncryptedPwd(password string) (string, error)
	DeactivateUserWithNoActiveStudent(ctx context.Context) ([]*po.User, error)
}

type IStudentCommonSrv interface {
	GetKintoneStudents(ctx context.Context, cond *dto.StudentReq) ([]*bo.Student, int, error)
	GetStudent(ctx context.Context, cond *bo.StudentCond) (*bo.Student, error)
}

type IDepositRecordCommonSrv interface {
	GetKintoneDepositRecords(ctx context.Context, cond *dto.DepositRecordReq) ([]*bo.KintoneDepositRecord, int, error)
	UpdateKintoneDepositRecordsStudentNames(ctx context.Context, oldPointCardName string, newPointCardName string) error
}

type IReduceRecordCommonSrv interface {
	GetKintoneReduceRecords(ctx context.Context, cond *dto.ReduceRecordReq) ([]*bo.KintoneReduceRecord, int, error)
	UpdateKintoneReduceRecordsStudentNames(ctx context.Context, oldPointCardName string, newPointCardName string) error
}

type IScheduleCommonSrv interface {
	GetKintoneSchedules(ctx context.Context, cond *dto.ScheduleReq) ([]dto.ScheduleRecord, int, error)
	UpdateKintoneSchedulesStudentNames(ctx context.Context, oldPointCardName string, newPointCardName string) error
}
