package repository

import (
	"context"
	"jaystar/internal/config"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/kintoneAPI"

	"golang.org/x/xerrors"
)

type KintoneScheduleRepository struct {
	cfg        config.IConfigEnv
	kintoneCli *kintoneAPI.KintoneClient
}

func ProvideKintoneScheduleRepository(cfg config.IConfigEnv, kintoneCli *kintoneAPI.KintoneClient) *KintoneScheduleRepository {
	return &KintoneScheduleRepository{
		cfg:        cfg,
		kintoneCli: kintoneCli,
	}
}

func (repo *KintoneScheduleRepository) GetKintoneSchedules(ctx context.Context, req *dto.ScheduleReq) (*dto.ScheduleRes, error) {
	m, err := kintoneAPI.ParseReqStructToMap(req)
	if err != nil {
		return nil, xerrors.Errorf("KintoneScheduleRepository GetKintoneSchedules kintoneAPI.ParseReqStructToMap: %w", err)
	}

	cfg := repo.cfg.GetKintoneConfig()
	query := map[string]string{
		kintone.QueryApp:   cfg.AppId.ScheduleRecord,
		kintone.TotalCount: "true",
		kintone.QueryQuery: kintoneAPI.ConvMapToQueryStringByAnd(m),
	}

	boScheduleRes := &dto.ScheduleRes{}
	err = repo.kintoneCli.Get(ctx, cfg.CommonUserAuthorization, kintone.RecordsPath, query, &boScheduleRes)
	if err != nil {
		return nil, xerrors.Errorf("KintoneScheduleRepository GetKintoneSchedules kintoneCli.Get: %w", err)
	}

	return boScheduleRes, nil
}

func (repo *KintoneScheduleRepository) UpdateKintoneSchedules(ctx context.Context, req *dto.UpdateSchedulesReq) (*dto.UpdateSchedulesRes, error) {
	cfg := repo.cfg.GetKintoneConfig()

	body := struct {
		dto.KintoneUpdateAppBase
		*dto.UpdateSchedulesReq
	}{
		KintoneUpdateAppBase: dto.KintoneUpdateAppBase{
			App: cfg.AppId.ScheduleRecord,
		},
		UpdateSchedulesReq: req,
	}

	poScheduleResp := &dto.UpdateSchedulesRes{}
	err := repo.kintoneCli.Put(ctx, cfg.AdminUserAuthorization, kintone.RecordsPath, body, poScheduleResp)
	if err != nil {
		return nil, xerrors.Errorf("KintoneScheduleRepository UpdateKintoneSchedules kintoneCli.Put: %w", err)
	}

	return poScheduleResp, nil
}
