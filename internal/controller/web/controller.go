package web

import (
	"github.com/gin-gonic/gin"
	"jaystar/internal/controller/web/middleware"
)

func ProvideController(
	userCtrl *UserCtrl,
	studentCtrl *StudentCtrl,
	scheduleCtrl *ScheduleCtrl,
	depositRecordCtrl *DepositRecordCtrl,
	reduceRecordCtrl *ReduceRecordCtrl,
	semesterSettleRecordCtrl *SemesterSettleRecordCtrl,
	pointCardCtrl *PointCardCtrl,
	syncCtrl *SyncCtrl,
) *Controller {
	return &Controller{
		UserCtrl:                 userCtrl,
		StudentCtrl:              studentCtrl,
		ScheduleCtrl:             scheduleCtrl,
		DepositRecordCtrl:        depositRecordCtrl,
		ReduceRecordCtrl:         reduceRecordCtrl,
		SemesterSettleRecordCtrl: semesterSettleRecordCtrl,
		PointCardCtrl:            pointCardCtrl,
		SyncCtrl:                 syncCtrl,
	}
}

type Controller struct {
	UserCtrl                 *UserCtrl
	StudentCtrl              *StudentCtrl
	ScheduleCtrl             *ScheduleCtrl
	DepositRecordCtrl        *DepositRecordCtrl
	ReduceRecordCtrl         *ReduceRecordCtrl
	SemesterSettleRecordCtrl *SemesterSettleRecordCtrl
	PointCardCtrl            *PointCardCtrl
	SyncCtrl                 *SyncCtrl
}

func SetStandardResponse(ctx *gin.Context, statusCode int, data interface{}) {
	middleware.SetResp(ctx, statusCode, data)
}
