package repository

import (
	"context"
	"golang.org/x/xerrors"
	"jaystar/internal/config"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/kintoneAPI"
)

type KintonePointCardRepository struct {
	cfg        config.IConfigEnv
	kintoneCli *kintoneAPI.KintoneClient
}

func ProvideKintonePointCardRepository(cfg config.IConfigEnv, kintoneCli *kintoneAPI.KintoneClient) *KintonePointCardRepository {
	return &KintonePointCardRepository{
		cfg:        cfg,
		kintoneCli: kintoneCli,
	}
}

func (repo *KintonePointCardRepository) GetPointCards(ctx context.Context, req *dto.GetPointCardReq) (*dto.GetPointCardRes, error) {
	cfg := repo.cfg.GetKintoneConfig()

	m, err := kintoneAPI.ParseReqStructToMap(req)
	if err != nil {
		return nil, xerrors.Errorf("kintoneAPI.ParseReqStructToMap: %w", err)
	}
	query := map[string]string{
		kintone.QueryApp:   cfg.AppId.PointCard,
		kintone.TotalCount: "true",
		kintone.QueryQuery: kintoneAPI.ConvMapToQueryStringByAnd(m),
	}

	poPointCardResp := &dto.GetPointCardRes{}
	err = repo.kintoneCli.Get(ctx, cfg.CommonUserAuthorization, kintone.RecordsPath, query, poPointCardResp)
	if err != nil {
		return nil, xerrors.Errorf("kintoneCli.Get: %w", err)
	}

	return poPointCardResp, nil
}

func (repo *KintonePointCardRepository) UpdatePointCard(ctx context.Context, req *dto.UpdatePointCardReq) (*dto.UpdatePointCardRes, error) {
	cfg := repo.cfg.GetKintoneConfig()

	body := struct {
		dto.KintoneUpdateAppBase
		*dto.UpdatePointCardReq
	}{
		KintoneUpdateAppBase: dto.KintoneUpdateAppBase{
			App: cfg.AppId.PointCard,
		},
		UpdatePointCardReq: req,
	}

	poUpdatePointCardRes := &dto.UpdatePointCardRes{}
	err := repo.kintoneCli.Put(ctx, cfg.AdminUserAuthorization, kintone.RecordPath, body, poUpdatePointCardRes)
	if err != nil {
		return nil, xerrors.Errorf("KintonePointCardRepository UpdatePointCard kintoneCli.Put: %w", err)
	}

	return poUpdatePointCardRes, nil
}

func (repo *KintonePointCardRepository) UpdatePointCards(ctx context.Context, req *dto.UpdatePointCardsReq) (*dto.UpdatePointCardRes, error) {
	cfg := repo.cfg.GetKintoneConfig()

	body := struct {
		dto.KintoneUpdateAppBase
		*dto.UpdatePointCardsReq
	}{
		KintoneUpdateAppBase: dto.KintoneUpdateAppBase{
			App: cfg.AppId.PointCard,
		},
		UpdatePointCardsReq: req,
	}

	poUpdatePointCardRes := &dto.UpdatePointCardRes{}
	err := repo.kintoneCli.Put(ctx, cfg.AdminUserAuthorization, kintone.RecordsPath, body, poUpdatePointCardRes)
	if err != nil {
		return nil, xerrors.Errorf("kintoneCli.Put: %w", err)
	}

	return poUpdatePointCardRes, nil
}
