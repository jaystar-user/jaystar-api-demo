package job

import (
	"jaystar/internal/constant/log"
	"jaystar/internal/cronjob"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"time"
)

type JobController struct {
	semesterSettleRecordSrv interfaces.ISemesterSettleRecordSrv
}

func ProvideController(semesterSettleRecordSrv interfaces.ISemesterSettleRecordSrv) *JobController {
	return &JobController{
		semesterSettleRecordSrv: semesterSettleRecordSrv,
	}
}

func (ctrl *JobController) SemesterSettlement(ctx *cronjob.Context) {
	err := ctrl.semesterSettleRecordSrv.SettleSemesterPoints(ctx, &bo.SettleSemesterPointsCond{Date: time.Now()})
	if err != nil {
		cronjob.SetActionLogs(ctx, log.ErrorMessage, err)
	}
}
