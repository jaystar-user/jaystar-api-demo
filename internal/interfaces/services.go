package interfaces

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/model/po"
	"sync"
)

type IStudentSrv interface {
	GetStudents(ctx context.Context, cond *bo.StudentCond) ([]*bo.Student, *po.PagerResult, error)
	UserRegisterAndCreateStudent(ctx context.Context, data *bo.Student) error
	UpdateStudent(ctx context.Context, data *bo.Student) error
	DeleteStudent(ctx context.Context, studentRefId int) error
	BatchSyncStudentsAndUsers(ctx context.Context, cond *bo.SyncStudentCond, wg ...*sync.WaitGroup) error
	GetStudentsSettled(ctx context.Context, db *gorm.DB) ([]*bo.Student, error)
}

type IUserSrv interface {
	GetUsers(ctx context.Context, cond *bo.UserCond) ([]*bo.User, *po.PagerResult, error)
	GetUser(ctx context.Context, cond *bo.UserCond) (*bo.User, error)
	UserLogin(ctx *gin.Context, cond *bo.UserLoginCond) (*bo.User, error)
	UserRegister(ctx context.Context, cond *bo.StudentCond, data *bo.CreateUserData) error
	UpdateUser(ctx context.Context, cond *bo.UserCond, data *bo.UpdateUserData) error
	GetEncryptedPassword(ctx context.Context, data dto.PasswordIO) (string, error)
}

type IDepositRecordSrv interface {
	GetDepositRecords(ctx context.Context, cond *bo.DepositRecordCond, studentCond *bo.StudentCond) ([]*bo.DepositRecord, *po.PagerResult, error)
	AddDepositRecord(ctx context.Context, data *bo.KintoneDepositRecord, studentCond *bo.StudentCond) error
	UpdateDepositRecord(ctx context.Context, cond *bo.DepositRecordCond, studentCond *bo.StudentCond, data *bo.UpdateDepositRecordData) error
	DeleteDepositRecord(ctx context.Context, cond *bo.DepositRecordCond) error
	BatchSyncDepositRecord(ctx context.Context, cond *bo.SyncDepositRecordCond, wg ...*sync.WaitGroup) error
	GetStudentTotalDepositPoints(ctx context.Context, db *gorm.DB, cond *bo.StudentTotalDepositPointsCond) (map[int64]*bo.StudentTotalDepositPoints, error)
}

type IReduceRecordSrv interface {
	GetReduceRecords(ctx context.Context, cond *bo.ReduceRecordCond, studentCond *bo.StudentCond) ([]*bo.ReduceRecord, *po.PagerResult, error)
	AddReduceRecord(ctx context.Context, data *bo.KintoneReduceRecord, studentCond *bo.StudentCond) error
	UpdateReduceRecord(ctx context.Context, cond *bo.ReduceRecordCond, studentCond *bo.StudentCond, data *bo.UpdateReduceRecordData) error
	DeleteReduceRecord(ctx context.Context, cond *bo.ReduceRecordCond) error
	BatchSyncReduceRecord(ctx context.Context, cond *bo.SyncReduceRecordCond, wg ...*sync.WaitGroup) error
	GetStudentTotalReducePoints(ctx context.Context, db *gorm.DB, cond *bo.StudentTotalReducePointsCond) (map[int64]*bo.StudentTotalReducePoints, error)
}

type IScheduleSrv interface {
	GetSchedules(ctx context.Context, cond *bo.GetScheduleCond, studentCond *bo.StudentCond) ([]*bo.Schedule, *po.PagerResult, error)
	GetSchedulesByRefId(ctx context.Context, scheduleRefId int) ([]*bo.Schedule, error)
	AddSchedule(ctx context.Context, data *bo.KintoneSchedule, studentCond *bo.StudentCond) error
	UpdateScheduleById(ctx context.Context, scheduleId int64, studentCond *bo.StudentCond, data *bo.UpdateScheduleData) error
	DeleteSchedule(ctx context.Context, cond *bo.UpdateScheduleCond) error
	BatchSyncSchedule(ctx context.Context, cond *bo.SyncScheduleCond, wg ...*sync.WaitGroup) error
}

type ISemesterSettleRecordSrv interface {
	GetSemesterSettleRecords(ctx context.Context, cond *bo.SemesterSettleRecordCond, studentCond *bo.StudentCond) ([]*bo.SemesterSettleRecord, *po.PagerResult, error)
	AddSemesterSettleRecord(ctx context.Context, data *bo.SemesterSettleRecord, studentCond *bo.StudentCond) error
	UpdateSemesterSettleRecord(ctx context.Context, cond *bo.UpdateSemesterSettleRecordCond, studentCond *bo.StudentCond, data *bo.UpdateSemesterSettleRecordData) error
	DeleteSemesterSettleRecord(ctx context.Context, cond *bo.UpdateSemesterSettleRecordCond) error
	BatchSyncSemesterSettleRecord(ctx context.Context, cond *bo.SyncSemesterSettleRecordCond, wg ...*sync.WaitGroup) error
	SettleSemesterPoints(ctx context.Context, cond *bo.SettleSemesterPointsCond) (err error)
}

type IPointCardSrv interface {
	GetPointCards(ctx context.Context, db *gorm.DB, cond *bo.GetPointCardCond) (map[int64]*bo.PointCard, error)
	UpdateKintonePointCardStudentName(ctx context.Context, oldPointCardName string, newPointCardName string) error
	AddPointCard(ctx context.Context, data *bo.PointCard, studentCond *bo.StudentCond) error
	UpdatePointCard(ctx context.Context, cond *bo.UpdatePointCardCond, studentCond *bo.StudentCond, data *bo.UpdatePointCardRecordData) error
	DeletePointCard(ctx context.Context, cond *bo.UpdatePointCardCond) error
	BatchSyncPointCard(ctx context.Context, cond *bo.SyncPointCardCond, wg ...*sync.WaitGroup) error
	SyncSettledStudentPointCards(ctx context.Context, db *gorm.DB, data []*bo.SyncSettledStudentPointCardData) error
}
