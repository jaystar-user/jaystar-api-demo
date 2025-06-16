package repository

import (
	"context"
	"golang.org/x/xerrors"
	"jaystar/internal/config"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/kintoneAPI"
)

type KintoneStudentRepository struct {
	cfg        config.IConfigEnv
	kintoneCli *kintoneAPI.KintoneClient
}

func ProvideKintoneStudentRepository(cfg config.IConfigEnv, kintoneCli *kintoneAPI.KintoneClient) *KintoneStudentRepository {
	return &KintoneStudentRepository{
		cfg:        cfg,
		kintoneCli: kintoneCli,
	}
}

func (repo *KintoneStudentRepository) GetKintoneStudents(ctx context.Context, req *dto.StudentReq) (*dto.StudentRes, error) {
	cfg := repo.cfg.GetKintoneConfig()

	m, err := kintoneAPI.ParseReqStructToMap(req)
	if err != nil {
		return nil, xerrors.Errorf("KintoneStudentRepository GetKintoneStudent kintoneAPI.ParseReqStructToMap: %w", err)
	}

	query := map[string]string{
		kintone.QueryApp:   cfg.AppId.StudentInfo,
		kintone.TotalCount: "true",
		kintone.QueryQuery: kintoneAPI.ConvMapToQueryStringByAnd(m),
	}

	poStudentRes := &dto.StudentRes{}
	err = repo.kintoneCli.Get(ctx, cfg.CommonUserAuthorization, kintone.RecordsPath, query, poStudentRes)
	if err != nil {
		return nil, xerrors.Errorf("KintoneStudentRepository GetKintoneStudents kintoneCli.Get: %w", err)
	}

	return poStudentRes, nil
}
