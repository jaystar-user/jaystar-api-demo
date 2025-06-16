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

func ProvideDepositRecordCommonService(kintoneDepositRecordRepo interfaces.IKintoneDepositRecordRepo) *DepositRecordCommonService {
	return &DepositRecordCommonService{
		kintoneDepositRecordRepo: kintoneDepositRecordRepo,
	}
}

type DepositRecordCommonService struct {
	kintoneDepositRecordRepo interfaces.IKintoneDepositRecordRepo
}

func (srv *DepositRecordCommonService) GetKintoneDepositRecords(ctx context.Context, cond *dto.DepositRecordReq) ([]*bo.KintoneDepositRecord, int, error) {
	poDepositRecordRes, err := srv.kintoneDepositRecordRepo.GetKintoneDepositRecords(ctx, cond)
	if err != nil {
		return nil, 0, xerrors.Errorf("kintoneDepositRecordRepo.GetKintoneDepositRecords: %w", err)
	}

	total, err := strconv.ParseInt(poDepositRecordRes.TotalCount, 10, 64)
	if err != nil {
		return nil, 0, xerrors.Errorf("strconv.ParseInt: %w", err)
	}

	boDepositRecords := make([]*bo.KintoneDepositRecord, 0, len(poDepositRecordRes.Records))

	for _, record := range poDepositRecordRes.Records {
		boDepositRecord, err := record.ToKintoneDepositRecordBo()
		if err != nil {
			return nil, 0, xerrors.Errorf("record.ToKintoneDepositRecordBo: %w", err)
		}
		boDepositRecords = append(boDepositRecords, boDepositRecord)
	}

	return boDepositRecords, int(total), nil
}

func (srv *DepositRecordCommonService) UpdateKintoneDepositRecordsStudentNames(ctx context.Context, oldPointCardName string, newPointCardName string) error {
	var (
		limit  = 500
		offset = 0
		total  int
		err    error
	)

	getReq := &dto.DepositRecordReq{
		StudentName: oldPointCardName,
		Limit:       limit,
		Offset:      offset * limit,
	}

	depositRecords, total, err := srv.GetKintoneDepositRecords(ctx, getReq)
	if err != nil {
		return xerrors.Errorf("GetKintoneDepositRecords: %w", err)
	}

	allRecords := make([]*bo.KintoneDepositRecord, 0, total)
	allRecords = append(allRecords, depositRecords...)

	for ((offset * limit) + limit) < total {
		offset += 1
		getReq.Offset = offset * limit
		depositRecords, total, err = srv.GetKintoneDepositRecords(ctx, getReq)
		if err != nil {
			return xerrors.Errorf("GetKintoneDepositRecords: %w", err)
		}
		allRecords = append(allRecords, depositRecords...)
	}

	if len(allRecords) == 0 {
		return errs.KintoneErr.ResponseEmptyError
	}

	for i := 0; i < len(allRecords); i += 100 {
		end := i + 100
		if end > len(allRecords) {
			end = len(allRecords)
		}

		updateReq := &dto.UpdateDepositRecordsReq{}
		batchRecords := allRecords[i:end]
		updateReq.Records = make([]dto.UpdateDepositRecord, 0, len(batchRecords))
		for _, record := range batchRecords {
			updateDepositRecord := dto.UpdateDepositRecord{}
			updateDepositRecord.Id = record.Id
			updateDepositRecord.Record.StudentName.Value = newPointCardName
			updateReq.Records = append(updateReq.Records, updateDepositRecord)
		}
		_, err = srv.kintoneDepositRecordRepo.UpdateKintoneDepositRecords(ctx, updateReq)
		if err != nil {
			return xerrors.Errorf("kintoneDepositRecordRepo.UpdateKintoneDepositRecords: %w", err)
		}
	}

	return nil
}
