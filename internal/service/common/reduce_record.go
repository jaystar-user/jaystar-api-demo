package common

import (
	"context"
	"golang.org/x/xerrors"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/errs"
	"strconv"
)

func ProvideReduceRecordCommonService(kintoneReduceRecordRepo interfaces.IKintoneReduceRecordRepo) *ReduceRecordCommonService {
	return &ReduceRecordCommonService{
		kintoneReduceRecordRepo: kintoneReduceRecordRepo,
	}
}

type ReduceRecordCommonService struct {
	kintoneReduceRecordRepo interfaces.IKintoneReduceRecordRepo
}

func (srv *ReduceRecordCommonService) GetKintoneReduceRecords(ctx context.Context, cond *dto.ReduceRecordReq) ([]*bo.KintoneReduceRecord, int, error) {
	poReduceRecordRes, err := srv.kintoneReduceRecordRepo.GetKintoneReduceRecords(ctx, cond)
	if err != nil {
		return nil, 0, xerrors.Errorf("kintoneReduceRecordRepo.GetKintoneReduceRecords: %w", err)
	}

	total, err := strconv.ParseInt(poReduceRecordRes.TotalCount, 10, 64)
	if err != nil {
		return nil, 0, xerrors.Errorf("strconv.ParseInt: %w", err)
	}

	boReduceRecords := make([]*bo.KintoneReduceRecord, 0, len(poReduceRecordRes.Records))

	for _, record := range poReduceRecordRes.Records {
		boReduceRecord, err := record.ToKintoneReduceRecord()
		if err != nil {
			return nil, 0, xerrors.Errorf("record.ToKintoneReduceRecord: %w", err)
		}
		boReduceRecords = append(boReduceRecords, boReduceRecord)
	}

	return boReduceRecords, int(total), nil
}

func (srv *ReduceRecordCommonService) UpdateKintoneReduceRecordsStudentNames(ctx context.Context, oldPointCardName string, newPointCardName string) error {
	var (
		limit  = 500
		offset = 0
		total  int
		err    error
	)

	getReq := &dto.ReduceRecordReq{
		StudentName: oldPointCardName,
		Limit:       limit,
		Offset:      offset,
	}
	reduceRecordRes, total, err := srv.GetKintoneReduceRecords(ctx, getReq)
	if err != nil {
		return xerrors.Errorf("GetKintoneReduceRecords: %w", err)
	}

	allRecords := make([]*bo.KintoneReduceRecord, 0, total)
	allRecords = append(allRecords, reduceRecordRes...)

	for ((offset * limit) + limit) < total {
		offset += 1
		getReq.Offset = offset * limit
		reduceRecordRes, total, err = srv.GetKintoneReduceRecords(ctx, getReq)
		if err != nil {
			return xerrors.Errorf("GetKintoneReduceRecords: %w", err)
		}
		allRecords = append(allRecords, reduceRecordRes...)
	}

	if len(allRecords) == 0 {
		return errs.KintoneErr.ResponseEmptyError
	}

	for i := 0; i < len(allRecords); i += 100 {
		end := i + 100
		if end > len(allRecords) {
			end = len(allRecords)
		}

		updateReq := &dto.UpdateReduceRecordsReq{}
		batchRecords := allRecords[i:end]
		updateReq.Records = make([]dto.UpdateReduceRecord, 0, len(batchRecords))
		for _, record := range batchRecords {
			updateReduceRecord := dto.UpdateReduceRecord{}
			updateReduceRecord.Id = record.Id
			updateReduceRecord.Record.StudentName.Value = newPointCardName
			updateReq.Records = append(updateReq.Records, updateReduceRecord)
		}
		_, err = srv.kintoneReduceRecordRepo.UpdateKintoneReduceRecords(ctx, updateReq)
		if err != nil {
			return xerrors.Errorf("kintoneReduceRecordRepo.UpdateKintoneReduceRecords: %w", err)
		}
	}

	return nil
}
