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

func newKintoneStudentRepo() *KintoneStudentRepository {
	os.Setenv("APP_ENV", "dev")
	envConfig := config.ProviderIConfigEnv()
	logger := logger.ProviderILogger(envConfig)
	return ProvideKintoneStudentRepository(envConfig, kintoneAPI.ProvideKintoneClient(envConfig, logger))
}

func TestGetKintoneStudents(t *testing.T) {
	type args struct {
		ctx  context.Context
		cond *dto.StudentReq
	}
	tests := []struct {
		name       string
		args       args
		wantLength int
		wantErr    bool
	}{{
		name:       "normal get",
		args:       args{ctx: context.TODO(), cond: &dto.StudentReq{ParentPhone: "0975296255", Limit: 1, Offset: 0}},
		wantLength: 1,
		wantErr:    false,
	}}

	repo := newKintoneStudentRepo()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := repo.GetKintoneStudents(tt.args.ctx, tt.args.cond)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKintoneStudents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(resp.Records) != tt.wantLength {
				t.Errorf("GetKintoneStudents() error = %v, wantLength %v", err, tt.wantErr)
			}

			t.Log("resp: ", resp)
		})
	}
}
