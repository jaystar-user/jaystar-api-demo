package interfaces

import (
	"context"
	"gorm.io/gorm"
	"jaystar/internal/model/po"
)

type IStudentRepo interface {
	ICommonRepo
	GetStudents(ctx context.Context, db *gorm.DB, cond *po.StudentCond, pager *po.Pager) ([]*po.StudentWithBalance, error)
	GetStudentsPager(ctx context.Context, db *gorm.DB, cond *po.StudentCond, pager *po.Pager) (*po.PagerResult, error)
	GetStudent(ctx context.Context, db *gorm.DB, cond *po.StudentCond) (*po.StudentWithBalance, error)
	GetStudentRefIds(ctx context.Context, db *gorm.DB, cond *po.StudentCond) ([]int, error)
	CreateStudent(ctx context.Context, db *gorm.DB, data *po.Student) error
	UpdateStudent(ctx context.Context, db *gorm.DB, cond *po.UpdateStudentCond, data *po.UpdateStudentData) error
}

type IUserRepo interface {
	GetUsers(ctx context.Context, db *gorm.DB, cond *po.UserCond, pager *po.Pager) ([]*po.User, error)
	GetUsersPager(ctx context.Context, db *gorm.DB, cond *po.UserCond, pager *po.Pager) (*po.PagerResult, error)
	GetUser(ctx context.Context, db *gorm.DB, cond *po.UserCond) (*po.User, error)
	CreateUser(ctx context.Context, db *gorm.DB, data *po.User) error
	UpdateUser(ctx context.Context, db *gorm.DB, cond *po.UserCond, data *po.UpdateUserData) error
	DeactivateUserWithNoActiveStudent(ctx context.Context, db *gorm.DB) ([]*po.User, error)
}

type IDepositRecordRepo interface {
	ICommonRepo
	GetRecords(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond, pager *po.Pager) ([]*po.DepositRecordView, error)
	GetRecordsPager(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond, pager *po.Pager) (*po.PagerResult, error)
	GetDepositRecordRefIds(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond) ([]int, error)
	GetRecord(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond) (*po.DepositRecord, error)
	AddRecord(ctx context.Context, db *gorm.DB, data *po.DepositRecord) error
	UpdateRecord(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond, data *po.UpdateDepositRecordData) error
	GetStudentTotalDepositPoints(ctx context.Context, db *gorm.DB, cond *po.DepositRecordCond) ([]*po.StudentTotalDepositPoints, error)
}

type IReduceRecordRepo interface {
	ICommonRepo
	GetRecordsWithSettleRecords(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond, pager *po.Pager) ([]*po.ReduceRecordSettleRecordView, error)
	GetRecordsWithSettleRecordsPager(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond, pager *po.Pager) (*po.PagerResult, error)
	GetReduceRecordRefIds(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond) ([]int, error)
	GetRecord(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond) (*po.ReduceRecord, error)
	AddRecord(ctx context.Context, db *gorm.DB, data *po.ReduceRecord) error
	UpdateRecord(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond, data *po.UpdateReduceRecordData) error
	GetStudentTotalReducePoints(ctx context.Context, db *gorm.DB, cond *po.ReduceRecordCond) ([]*po.StudentTotalReducePoints, error)
}

type IScheduleRepo interface {
	ICommonRepo
	GetSchedules(ctx context.Context, db *gorm.DB, cond *po.GetScheduleCond, pager *po.Pager) ([]*po.ScheduleView, error)
	GetSchedulesPager(ctx context.Context, db *gorm.DB, cond *po.GetScheduleCond, pager *po.Pager) (*po.PagerResult, error)
	GetSchedulesByRefId(ctx context.Context, db *gorm.DB, scheduleRefId int) ([]*po.ScheduleView, error)
	AddSchedule(ctx context.Context, db *gorm.DB, data *po.Schedule) error
	UpdateSchedule(ctx context.Context, db *gorm.DB, cond *po.UpdateScheduleCond, data *po.UpdateScheduleData) error
	GetAllScheduleRefIds(ctx context.Context, db *gorm.DB, cond *po.GetScheduleCond) ([]int, error)
}

type ISemesterSettleRecordRepo interface {
	ICommonRepo
	GetRecords(ctx context.Context, db *gorm.DB, cond *po.SemesterSettleRecordCond, pager *po.Pager) ([]*po.SemesterSettleRecordView, error)
	GetRecordsPager(ctx context.Context, db *gorm.DB, cond *po.SemesterSettleRecordCond, pager *po.Pager) (*po.PagerResult, error)
	GetRecord(ctx context.Context, db *gorm.DB, cond *po.SemesterSettleRecordCond) (*po.SemesterSettleRecord, error)
	AddRecord(ctx context.Context, db *gorm.DB, data *po.SemesterSettleRecord) error
	UpdateRecord(ctx context.Context, db *gorm.DB, cond *po.UpdateSemesterSettleRecordCond, data *po.UpdateSemesterSettleRecordData) error
	GetSemesterSettleRecordRefIds(ctx context.Context, db *gorm.DB, cond *po.SemesterSettleRecordCond) ([]int, error)
}

type IPointCardRepo interface {
	ICommonRepo
	GetPointCard(ctx context.Context, db *gorm.DB, cond *po.PointCardCond) (*po.PointCard, error)
	GetPointCards(ctx context.Context, db *gorm.DB, cond *po.PointCardCond) ([]*po.PointCard, error)
	AddPointCard(ctx context.Context, db *gorm.DB, data *po.PointCard) error
	UpdatePointCard(ctx context.Context, db *gorm.DB, cond *po.UpdatePointCardCond, data *po.UpdatePointCardData) error
	GetPointCardRefIds(ctx context.Context, db *gorm.DB, cond *po.PointCardCond) ([]int, error)
}

type ICommonRepo interface {
	ResetFromDeleted(ctx context.Context, db *gorm.DB, tableName string, whereScopes func(db *gorm.DB) *gorm.DB) error
}
