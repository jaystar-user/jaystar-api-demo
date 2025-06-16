package repository

import (
	"context"
	"jaystar/internal/config"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/kintoneAPI"

	"golang.org/x/xerrors"
)

type KintoneReduceRecordRepository struct {
	cfg        config.IConfigEnv
	kintoneCli *kintoneAPI.KintoneClient
}

func ProvideKintoneReduceRecordRepository(cfg config.IConfigEnv, kintoneCli *kintoneAPI.KintoneClient) *KintoneReduceRecordRepository {
	return &KintoneReduceRecordRepository{
		cfg:        cfg,
		kintoneCli: kintoneCli,
	}
}

func (repo *KintoneReduceRecordRepository) GetKintoneReduceRecords(ctx context.Context, req *dto.ReduceRecordReq) (*dto.ReduceRecordRes, error) {
	cfg := repo.cfg.GetKintoneConfig()

	m, err := kintoneAPI.ParseReqStructToMap(req)
	if err != nil {
		return nil, xerrors.Errorf("KintoneReduceRecordRepository GetKintoneReduceRecords kintoneAPI.ParseReqStructToMap: %w", err)
	}
	query := map[string]string{
		kintone.QueryApp:   cfg.AppId.ReduceRecord,
		kintone.TotalCount: "true",
		kintone.QueryQuery: kintoneAPI.ConvMapToQueryStringByAnd(m),
	}

	boReduceRecordRes := &dto.ReduceRecordRes{}
	err = repo.kintoneCli.Get(ctx, cfg.CommonUserAuthorization, kintone.RecordsPath, query, &boReduceRecordRes)
	if err != nil {
		return nil, xerrors.Errorf("KintoneReduceRecordRepository GetKintoneReduceRecords kintoneCli.Get: %w", err)
	}

	return boReduceRecordRes, nil
}

func (repo *KintoneReduceRecordRepository) UpdateKintoneReduceRecords(ctx context.Context, req *dto.UpdateReduceRecordsReq) (*dto.UpdateReduceRecordsRes, error) {
	cfg := repo.cfg.GetKintoneConfig()

	body := struct {
		dto.KintoneUpdateAppBase
		*dto.UpdateReduceRecordsReq
	}{
		KintoneUpdateAppBase: dto.KintoneUpdateAppBase{
			App: cfg.AppId.ReduceRecord,
		},
		UpdateReduceRecordsReq: req,
	}

	poReduceRecordResp := &dto.UpdateReduceRecordsRes{}
	err := repo.kintoneCli.Put(ctx, cfg.AdminUserAuthorization, kintone.RecordsPath, body, poReduceRecordResp)
	if err != nil {
		return nil, xerrors.Errorf("KintoneReduceRecordRepository UpdateKintoneReduceRecords kintoneCli.Put: %w", err)
	}

	return poReduceRecordResp, nil
}
