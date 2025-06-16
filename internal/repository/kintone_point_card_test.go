package repository

import (
	"context"
	"github.com/SeanZhenggg/go-utils/logger"
	"jaystar/internal/config"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/kintoneAPI"
	"os"
	"testing"
)

func newKintonePointCardRepo() *KintonePointCardRepository {
	os.Setenv("APP_ENV", "dev")
	envConfig := config.ProviderIConfigEnv()
	logger := logger.ProviderILogger(envConfig)
	return ProvideKintonePointCardRepository(envConfig, kintoneAPI.ProvideKintoneClient(envConfig, logger))
}

func TestUpdatePointCard(t *testing.T) {
	type args struct {
		ctx  context.Context
		cond *dto.UpdatePointCardReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: context.TODO(),
				cond: &dto.UpdatePointCardReq{
					KintoneUpdateIdBase: dto.KintoneUpdateIdBase{
						Id: 1439,
					},
					Record: dto.UpdatePointCardRecord{
						StudentName: dto.NormalField{Value: "沈品言/0975296250"},
					},
				},
			},
			wantErr: false,
		},
	}
	repo := newKintonePointCardRepo()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := repo.UpdatePointCard(tt.args.ctx, tt.args.cond)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePointCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(resp.Revision) == 0 {
				t.Errorf("UpdatePointCard() error revision response = %v", resp.Revision)
			}
		})
	}
}

func TestUpdatePointCards(t *testing.T) {
	type args struct {
		ctx  context.Context
		cond *dto.UpdatePointCardsReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: context.TODO(),
				cond: &dto.UpdatePointCardsReq{
					Records: []dto.UpdatePointCardsRecord{{
						UpdateKey: struct {
							Field string `json:"field"`
							Value string `json:"value"`
						}{Field: "studentName", Value: "小鄭/0937054388"},
						Record: dto.UpdatePointCardsRecordValue{
							ClearPoints: dto.NormalField{Value: "2"},
						},
					},
					},
				},
			},
			wantErr: false,
		},
	}
	repo := newKintonePointCardRepo()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := repo.UpdatePointCards(tt.args.ctx, tt.args.cond)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePointCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(resp.Revision) == 0 {
				t.Errorf("UpdatePointCard() error revision response = %v", resp.Revision)
			}
		})
	}
}
