package repository

import (
	"context"
	"golang.org/x/xerrors"
	"jaystar/internal/config"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/kintoneAPI"
)

type KintoneDepositRecordRepository struct {
	cfg        config.IConfigEnv
	kintoneCli *kintoneAPI.KintoneClient
}

func ProvideKintoneDepositRecordRepository(cfg config.IConfigEnv, kintoneCli *kintoneAPI.KintoneClient) *KintoneDepositRecordRepository {
	return &KintoneDepositRecordRepository{
		cfg:        cfg,
		kintoneCli: kintoneCli,
	}
}

func (repo *KintoneDepositRecordRepository) GetKintoneDepositRecords(ctx context.Context, req *dto.DepositRecordReq) (*dto.DepositRecordRes, error) {
	cfg := repo.cfg.GetKintoneConfig()

	m, err := kintoneAPI.ParseReqStructToMap(req)
	if err != nil {
		return nil, xerrors.Errorf("KintoneDepositRecordRepository GetKintoneDepositRecords kintoneAPI.ParseReqStructToMap: %w", err)
	}
	query := map[string]string{
		kintone.QueryApp:   cfg.AppId.DepositRecord,
		kintone.TotalCount: "true",
		kintone.QueryQuery: kintoneAPI.ConvMapToQueryStringByAnd(m),
	}

	poDepositRecordResp := &dto.DepositRecordRes{}
	err = repo.kintoneCli.Get(ctx, cfg.CommonUserAuthorization, kintone.RecordsPath, query, poDepositRecordResp)
	if err != nil {
		return nil, xerrors.Errorf("KintoneDepositRecordRepository GetKintoneDepositRecords kintoneCli.Get: %w", err)
	}

	return poDepositRecordResp, nil
}

func (repo *KintoneDepositRecordRepository) UpdateKintoneDepositRecords(ctx context.Context, req *dto.UpdateDepositRecordsReq) (*dto.UpdateDepositRecordsRes, error) {
	cfg := repo.cfg.GetKintoneConfig()

	body := struct {
		dto.KintoneUpdateAppBase
		*dto.UpdateDepositRecordsReq
	}{
		KintoneUpdateAppBase: dto.KintoneUpdateAppBase{
			App: cfg.AppId.DepositRecord,
		},
		UpdateDepositRecordsReq: req,
	}

	poDepositRecordResp := &dto.UpdateDepositRecordsRes{}
	err := repo.kintoneCli.Put(ctx, cfg.AdminUserAuthorization, kintone.RecordsPath, body, poDepositRecordResp)
	if err != nil {
		return nil, xerrors.Errorf("KintoneDepositRecordRepository UpdateKintoneDepositRecords kintoneCli.Put: %w", err)
	}

	return poDepositRecordResp, nil
}
