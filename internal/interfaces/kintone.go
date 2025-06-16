package interfaces

import (
	"context"
	"jaystar/internal/model/dto"
)

type IKintoneStudentRepo interface {
	GetKintoneStudents(ctx context.Context, req *dto.StudentReq) (*dto.StudentRes, error)
}

type IKintoneDepositRecordRepo interface {
	GetKintoneDepositRecords(ctx context.Context, req *dto.DepositRecordReq) (*dto.DepositRecordRes, error)
	UpdateKintoneDepositRecords(ctx context.Context, req *dto.UpdateDepositRecordsReq) (*dto.UpdateDepositRecordsRes, error)
}

type IKintoneReduceRecordRepo interface {
	GetKintoneReduceRecords(ctx context.Context, req *dto.ReduceRecordReq) (*dto.ReduceRecordRes, error)
	UpdateKintoneReduceRecords(ctx context.Context, req *dto.UpdateReduceRecordsReq) (*dto.UpdateReduceRecordsRes, error)
}

type IKintoneScheduleRepo interface {
	GetKintoneSchedules(ctx context.Context, req *dto.ScheduleReq) (*dto.ScheduleRes, error)
	UpdateKintoneSchedules(ctx context.Context, req *dto.UpdateSchedulesReq) (*dto.UpdateSchedulesRes, error)
}

type IKintonePointCardRepo interface {
	GetPointCards(ctx context.Context, req *dto.GetPointCardReq) (*dto.GetPointCardRes, error)
	UpdatePointCard(ctx context.Context, req *dto.UpdatePointCardReq) (*dto.UpdatePointCardRes, error)
	UpdatePointCards(ctx context.Context, req *dto.UpdatePointCardsReq) (*dto.UpdatePointCardRes, error)
}

type IKintoneSemesterSettleRecordRepo interface {
	GetKintoneSemesterSettleRecords(ctx context.Context, req *dto.SemesterSettleRecordReq) (*dto.SemesterSettleRecordRes, error)
	InsertKintoneSemesterSettleRecords(ctx context.Context, req *dto.InsertSemesterSettleRecordsReq) (*dto.InsertSemesterSettleRecordsRes, error)
}
