package repository

import (
	"context"
	"golang.org/x/xerrors"
	"jaystar/internal/config"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/kintoneAPI"
)

type KintoneSemesterSettleRecordRepository struct {
	cfg        config.IConfigEnv
	kintoneCli *kintoneAPI.KintoneClient
}

func ProvideKintoneSemesterSettleRecordRepository(cfg config.IConfigEnv, kintoneCli *kintoneAPI.KintoneClient) *KintoneSemesterSettleRecordRepository {
	return &KintoneSemesterSettleRecordRepository{
		cfg:        cfg,
		kintoneCli: kintoneCli,
	}
}

func (repo *KintoneSemesterSettleRecordRepository) GetKintoneSemesterSettleRecords(ctx context.Context, req *dto.SemesterSettleRecordReq) (*dto.SemesterSettleRecordRes, error) {
	m, err := kintoneAPI.ParseReqStructToMap(req)
	if err != nil {
		return nil, xerrors.Errorf("KintoneScheduleRepository GetKintoneSchedules kintoneAPI.ParseReqStructToMap: %w", err)
	}

	cfg := repo.cfg.GetKintoneConfig()
	query := map[string]string{
		kintone.QueryApp:   cfg.AppId.SemesterSettleRecord,
		kintone.TotalCount: "true",
		kintone.QueryQuery: kintoneAPI.ConvMapToQueryStringByAnd(m),
	}

	boScheduleRes := &dto.SemesterSettleRecordRes{}
	err = repo.kintoneCli.Get(ctx, cfg.CommonUserAuthorization, kintone.RecordsPath, query, &boScheduleRes)
	if err != nil {
		return nil, xerrors.Errorf("KintoneScheduleRepository GetKintoneSchedules kintoneCli.Get: %w", err)
	}

	return boScheduleRes, nil
}

func (repo *KintoneSemesterSettleRecordRepository) InsertKintoneSemesterSettleRecords(ctx context.Context, req *dto.InsertSemesterSettleRecordsReq) (*dto.InsertSemesterSettleRecordsRes, error) {
	cfg := repo.cfg.GetKintoneConfig()

	body := struct {
		dto.KintoneUpdateAppBase
		*dto.InsertSemesterSettleRecordsReq
	}{
		KintoneUpdateAppBase: dto.KintoneUpdateAppBase{
			App: cfg.AppId.SemesterSettleRecord,
		},
		InsertSemesterSettleRecordsReq: req,
	}

	insertSemesterSettleRecordsResResp := &dto.InsertSemesterSettleRecordsRes{}
	err := repo.kintoneCli.Post(ctx, cfg.AdminUserAuthorization, kintone.RecordsPath, body, insertSemesterSettleRecordsResResp)
	if err != nil {
		return nil, xerrors.Errorf("kintoneCli.Post: %w", err)
	}

	return insertSemesterSettleRecordsResResp, nil
}
