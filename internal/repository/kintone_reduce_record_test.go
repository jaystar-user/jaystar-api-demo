package repository

import (
	"context"
	"jaystar/internal/config"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/kintoneAPI"
	"os"
	"slices"
	"testing"

	"github.com/SeanZhenggg/go-utils/logger"
)

func newKintoneReduceRecordRepo() *KintoneReduceRecordRepository {
	os.Setenv("APP_ENV", "dev")
	envConfig := config.ProviderIConfigEnv()
	logger := logger.ProviderILogger(envConfig)
	return ProvideKintoneReduceRecordRepository(envConfig, kintoneAPI.ProvideKintoneClient(envConfig, logger))
}

func TestUpdateKintoneReduceRecords(t *testing.T) {
	type args struct {
		ctx  context.Context
		cond *dto.UpdateReduceRecordsReq
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
			cond: &dto.UpdateReduceRecordsReq{
				Records: []dto.UpdateReduceRecord{
					{
						KintoneUpdateIdBase: dto.KintoneUpdateIdBase{
							Id: 23612,
						},
						Record: dto.UpdateReduceRecordValue{
							StudentName: dto.NormalField{
								Value: "沈品言/0975296255",
							},
						},
					},
					{
						KintoneUpdateIdBase: dto.KintoneUpdateIdBase{
							Id: 23448,
						},
						Record: dto.UpdateReduceRecordValue{
							StudentName: dto.NormalField{
								Value: "沈品言/0975296255",
							},
						},
					},
				},
			},
		},
		updateIds: []string{"23612", "23448"},
	}}

	repo := newKintoneReduceRecordRepo()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := repo.UpdateKintoneReduceRecords(tt.args.ctx, tt.args.cond)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateKintoneReduceRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, record := range resp.Records {
				if !slices.Contains(tt.updateIds, record.Id) {
					t.Errorf("UpdateKintoneReduceRecords() updateIds = %v, record id = %v", tt.updateIds, record.Id)
					return
				}
			}
		})
	}
}
