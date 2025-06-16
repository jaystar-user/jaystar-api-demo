package repository

import (
	"context"
	"jaystar/internal/config"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/kintoneAPI"
	"log"
	"os"
	"slices"
	"testing"

	"github.com/SeanZhenggg/go-utils/logger"
)

func newKintoneScheduleRepo() *KintoneScheduleRepository {
	os.Setenv("APP_ENV", "dev")
	envConfig := config.ProviderIConfigEnv()
	logger := logger.ProviderILogger(envConfig)
	return ProvideKintoneScheduleRepository(envConfig, kintoneAPI.ProvideKintoneClient(envConfig, logger))
}

func TestGetKintoneSchedules(t *testing.T) {
	repo := newKintoneScheduleRepo()

	schedules, err := repo.GetKintoneSchedules(context.TODO(), &dto.ScheduleReq{StudentName: "沈品言/0975296250"})
	if err != nil {
		panic(err)
	}

	log.Println(schedules)
}

func TestUpdateKintoneSchedules(t *testing.T) {
	type args struct {
		ctx  context.Context
		cond *dto.UpdateSchedulesReq
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
			cond: &dto.UpdateSchedulesReq{
				Records: []dto.UpdateSchedule{
					{
						KintoneUpdateIdBase: dto.KintoneUpdateIdBase{
							Id: 7460,
						},
						Record: dto.UpdateScheduleValue{
							Attendance: dto.Attendance{
								Type: "SUBTABLE",
								Value: []*dto.AttendanceValue{
									{
										Id: "281616",
										Value: dto.AttendanceRecordValue{
											StudentName: dto.StringField{
												NormalField: dto.NormalField{
													Value: "楊鈞傑/0918236165",
												},
											},
										},
									},
									{
										Id: "281618",
										Value: dto.AttendanceRecordValue{
											StudentName: dto.StringField{
												NormalField: dto.NormalField{
													Value: "沈品言/0975296250",
												},
											},
										},
									},
								},
							},
						},
					},
					{
						KintoneUpdateIdBase: dto.KintoneUpdateIdBase{
							Id: 7452,
						},
						Record: dto.UpdateScheduleValue{
							Attendance: dto.Attendance{
								Type: "SUBTABLE",
								Value: []*dto.AttendanceValue{
									{
										Id: "272663",
										Value: dto.AttendanceRecordValue{
											StudentName: dto.StringField{
												NormalField: dto.NormalField{
													Value: "劉安寶/0930518181",
												},
											},
										},
									},
									{
										Id: "272665",
										Value: dto.AttendanceRecordValue{
											StudentName: dto.StringField{
												NormalField: dto.NormalField{
													Value: "黃貝安/0987148407",
												},
											},
										},
									},
									{
										Id: "272667",
										Value: dto.AttendanceRecordValue{
											StudentName: dto.StringField{
												NormalField: dto.NormalField{
													Value: "温宥翔.宥瑄/0910301963",
												},
											},
										},
									},
									{
										Id: "272669",
										Value: dto.AttendanceRecordValue{
											StudentName: dto.StringField{
												NormalField: dto.NormalField{
													Value: "沈品言/0975296250",
												},
											},
										},
									},
								},
							},
						},
					},
					{
						KintoneUpdateIdBase: dto.KintoneUpdateIdBase{
							Id: 7369,
						},
						Record: dto.UpdateScheduleValue{
							Attendance: dto.Attendance{
								Type: "SUBTABLE",
								Value: []*dto.AttendanceValue{
									{
										Id: "271986",
										Value: dto.AttendanceRecordValue{
											StudentName: dto.StringField{
												NormalField: dto.NormalField{
													Value: "楊鈞傑/0918236165",
												},
											},
										},
									},
									{
										Id: "271988",
										Value: dto.AttendanceRecordValue{
											StudentName: dto.StringField{
												NormalField: dto.NormalField{
													Value: "沈品言/0975296250",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		updateIds: []string{"7460", "7452", "7369"},
	}}

	repo := newKintoneScheduleRepo()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := repo.UpdateKintoneSchedules(tt.args.ctx, tt.args.cond)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateKintoneSchedules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, record := range resp.Records {
				if !slices.Contains(tt.updateIds, record.Id) {
					t.Errorf("UpdateKintoneSchedules() updateIds = %v, record id = %v", tt.updateIds, record.Id)
					return
				}
			}
		})
	}
}
