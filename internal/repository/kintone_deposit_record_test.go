package repository

import (
	"context"
	"github.com/SeanZhenggg/go-utils/logger"
	"jaystar/internal/config"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/kintoneAPI"
	"os"
	"slices"
	"testing"
)

func newKintoneDepositRecordRepo() *KintoneDepositRecordRepository {
	os.Setenv("APP_ENV", "dev")
	envConfig := config.ProviderIConfigEnv()
	logger := logger.ProviderILogger(envConfig)
	return ProvideKintoneDepositRecordRepository(envConfig, kintoneAPI.ProvideKintoneClient(envConfig, logger))
}

func TestUpdateKintoneDepositRecords(t *testing.T) {
	type args struct {
		ctx  context.Context
		cond *dto.UpdateDepositRecordsReq
	}
	tests := []struct {
		name      string
		args      args
		updateIds []string
		wantErr   bool
	}{{
		name: "normal",
		args: args{
			ctx: context.TODO(),
			cond: &dto.UpdateDepositRecordsReq{
				Records: []dto.UpdateDepositRecord{
					{
						KintoneUpdateIdBase: dto.KintoneUpdateIdBase{
							Id: 1730,
						},
						Record: dto.UpdateDepositRecordValue{
							StudentName: dto.NormalField{
								Value: "沈品言/0975296250",
							},
						},
					},
					{
						KintoneUpdateIdBase: dto.KintoneUpdateIdBase{
							Id: 1929,
						},
						Record: dto.UpdateDepositRecordValue{
							StudentName: dto.NormalField{
								Value: "沈品言/0975296250",
							},
						},
					},
				},
			},
		},
		updateIds: []string{"1730", "1929"},
	}}

	repo := newKintoneDepositRecordRepo()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := repo.UpdateKintoneDepositRecords(tt.args.ctx, tt.args.cond)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateKintoneDepositRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, record := range resp.Records {
				if !slices.Contains(tt.updateIds, record.Id) {
					t.Errorf("UpdateKintoneDepositRecords() updateIds = %v, record id = %v", tt.updateIds, record.Id)
					return
				}
			}
		})
	}
}
