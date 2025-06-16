package common

import (
	"context"
	"golang.org/x/xerrors"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/errs"
	"slices"
	"strconv"
)

func ProvideScheduleCommonService(kintoneScheduleRepo interfaces.IKintoneScheduleRepo) *ScheduleCommonService {
	return &ScheduleCommonService{
		kintoneScheduleRepo: kintoneScheduleRepo,
	}
}

type ScheduleCommonService struct {
	kintoneScheduleRepo interfaces.IKintoneScheduleRepo
}

func (srv *ScheduleCommonService) GetKintoneSchedules(ctx context.Context, cond *dto.ScheduleReq) ([]dto.ScheduleRecord, int, error) {
	dtoScheduleRes, err := srv.kintoneScheduleRepo.GetKintoneSchedules(ctx, cond)
	if err != nil {
		return nil, 0, xerrors.Errorf("kintoneScheduleRepo.GetKintoneSchedules: %w", err)
	}

	total, err := strconv.ParseInt(dtoScheduleRes.TotalCount, 10, 64)
	if err != nil {
		return nil, 0, xerrors.Errorf("strconv.ParseInt: %w", err)
	}

	return dtoScheduleRes.Records, int(total), nil
}

func (srv *ScheduleCommonService) UpdateKintoneSchedulesStudentNames(ctx context.Context, oldPointCardName string, newPointCardName string) error {
	var (
		limit  = 500
		offset = 0
		total  int
		err    error
	)

	getReq := &dto.ScheduleReq{
		StudentName: oldPointCardName,
		Limit:       limit,
		Offset:      offset,
	}
	schedules, total, err := srv.GetKintoneSchedules(ctx, getReq)
	if err != nil {
		return xerrors.Errorf("GetKintoneSchedules: %w", err)
	}

	allRecords := make([]dto.ScheduleRecord, 0, total)
	allRecords = append(allRecords, schedules...)

	for ((offset * limit) + limit) < total {
		offset += 1
		getReq.Offset = offset * limit
		schedules, total, err = srv.GetKintoneSchedules(ctx, getReq)
		if err != nil {
			return xerrors.Errorf("GetKintoneSchedules: %w", err)
		}
		allRecords = append(allRecords, schedules...)
	}

	if len(allRecords) == 0 {
		return errs.KintoneErr.ResponseEmptyError
	}

	for i := 0; i < len(allRecords); i += 100 {
		end := i + 100
		if end > len(allRecords) {
			end = len(allRecords)
		}

		updateReq := &dto.UpdateSchedulesReq{}
		batchRecords := allRecords[i:end]
		updateReq.Records = make([]dto.UpdateSchedule, 0, len(batchRecords))
		for _, record := range batchRecords {
			updateSchedule := dto.UpdateSchedule{}
			id, err := record.Id.ToId()
			if err != nil {
				return xerrors.Errorf("record.Id.ToId student: %s, id: %s, err: %w", oldPointCardName, record.Id.Value, err)
			}
			updateSchedule.Id = id
			updateSchedule.Record = dto.UpdateScheduleValue{}
			updateSchedule.Record.Attendance = record.Attendance
			index := slices.IndexFunc(updateSchedule.Record.Attendance.Value, func(attendanceValue *dto.AttendanceValue) bool {
				return attendanceValue.Value.StudentName.ToString() == oldPointCardName
			})

			updateSchedule.Record.Attendance.Value[index].Value.StudentName.Value = newPointCardName

			updateReq.Records = append(updateReq.Records, updateSchedule)
		}
		_, err = srv.kintoneScheduleRepo.UpdateKintoneSchedules(ctx, updateReq)
		if err != nil {
			return xerrors.Errorf("kintoneScheduleRepo.UpdateKintoneSchedules: %w", err)
		}
	}

	return nil
}
