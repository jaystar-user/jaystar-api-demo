package web

import (
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/gin-gonic/gin"
	"jaystar/internal/controller/web/util"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/errs"
	"net/http"
	"sync"
)

func ProvideSyncController(
	studentSrv interfaces.IStudentSrv,
	pointCardSrv interfaces.IPointCardSrv,
	depositRecordSrv interfaces.IDepositRecordSrv,
	reduceRecordSrv interfaces.IReduceRecordSrv,
	scheduleSrv interfaces.IScheduleSrv,
	semesterSettleRecordSrv interfaces.ISemesterSettleRecordSrv,
	reqParse util.IRequestParse,
	logger logger.ILogger,
) *SyncCtrl {
	return &SyncCtrl{
		studentSrv:              studentSrv,
		pointCardSrv:            pointCardSrv,
		depositRecordSrv:        depositRecordSrv,
		reduceRecordSrv:         reduceRecordSrv,
		scheduleSrv:             scheduleSrv,
		semesterSettleRecordSrv: semesterSettleRecordSrv,
		reqParse:                reqParse,
		logger:                  logger,
	}
}

type SyncCtrl struct {
	studentSrv              interfaces.IStudentSrv
	pointCardSrv            interfaces.IPointCardSrv
	depositRecordSrv        interfaces.IDepositRecordSrv
	reduceRecordSrv         interfaces.IReduceRecordSrv
	scheduleSrv             interfaces.IScheduleSrv
	semesterSettleRecordSrv interfaces.ISemesterSettleRecordSrv
	reqParse                util.IRequestParse
	logger                  logger.ILogger
}

func (ctrl *SyncCtrl) AdminSyncAllByStudent(ctx *gin.Context) {
	req := dto.SyncAllByStudentIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	studentWg := sync.WaitGroup{}
	// 會員資料
	// 必須先同步學生資料，後續才同步其他跟學生資料相關的資料
	boSyncStudentCond := &bo.SyncStudentCond{StudentName: &req.StudentName, ParentPhone: &req.ParentPhone}
	err := ctrl.studentSrv.BatchSyncStudentsAndUsers(ctx, boSyncStudentCond, &studentWg)
	if err != nil {
		SetStandardResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	studentWg.Wait()

	wg := sync.WaitGroup{}
	// 點數管理
	boSyncPointCardCond := &bo.SyncPointCardCond{StudentName: &req.StudentName, ParentPhone: &req.ParentPhone}
	err = ctrl.pointCardSrv.BatchSyncPointCard(ctx, boSyncPointCardCond, &wg)
	if err != nil {
		SetStandardResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	// 購課記錄
	boSyncDepositRecordCond := &bo.SyncDepositRecordCond{StudentName: &req.StudentName, ParentPhone: &req.ParentPhone}
	err = ctrl.depositRecordSrv.BatchSyncDepositRecord(ctx, boSyncDepositRecordCond, &wg)
	if err != nil {
		SetStandardResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	// 點名管理
	boReduceRecordCond := &bo.SyncReduceRecordCond{StudentName: &req.StudentName, ParentPhone: &req.ParentPhone}
	err = ctrl.reduceRecordSrv.BatchSyncReduceRecord(ctx, boReduceRecordCond, &wg)
	if err != nil {
		SetStandardResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	// 課表管理合併點名
	boScheduleCond := &bo.SyncScheduleCond{StudentName: &req.StudentName, ParentPhone: &req.ParentPhone}
	err = ctrl.scheduleSrv.BatchSyncSchedule(ctx, boScheduleCond, &wg)
	if err != nil {
		SetStandardResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	// 學期制結算記錄
	boSemesterSettleRecordCond := &bo.SyncSemesterSettleRecordCond{StudentName: &req.StudentName, ParentPhone: &req.ParentPhone}
	err = ctrl.semesterSettleRecordSrv.BatchSyncSemesterSettleRecord(ctx, boSemesterSettleRecordCond, &wg)
	if err != nil {
		SetStandardResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	wg.Wait()

	SetStandardResponse(ctx, http.StatusOK, nil)
}

func (ctrl *SyncCtrl) AdminBatchSyncStudentsAndUsers(ctx *gin.Context) {
	if err := ctrl.studentSrv.BatchSyncStudentsAndUsers(ctx, nil); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}

func (ctrl *SyncCtrl) AdminBatchSyncDepositRecord(ctx *gin.Context) {
	req := dto.SyncDepositRecordIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boSyncDepositRecordCond := &bo.SyncDepositRecordCond{
		ChargingDateStart: req.DepositedDateStart,
		ChargingDateEnd:   req.DepositedDateEnd,
	}

	if err := ctrl.depositRecordSrv.BatchSyncDepositRecord(ctx, boSyncDepositRecordCond); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}

func (ctrl *SyncCtrl) AdminBatchSyncReduceRecord(ctx *gin.Context) {
	req := &dto.ReduceRecordSyncIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	reduceRecordReq := &bo.SyncReduceRecordCond{
		ClassTimeStart: req.ClassTimeStart,
		ClassTimeEnd:   req.ClassTimeEnd,
	}

	if err := ctrl.reduceRecordSrv.BatchSyncReduceRecord(ctx, reduceRecordReq); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}

func (ctrl *SyncCtrl) AdminBatchSyncSchedule(ctx *gin.Context) {
	req := &dto.ScheduleSyncIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boSyncScheduleCond := &bo.SyncScheduleCond{
		ClassTimeStart: req.ClassTimeStart,
		ClassTimeEnd:   req.ClassTimeEnd,
	}

	if err := ctrl.scheduleSrv.BatchSyncSchedule(ctx, boSyncScheduleCond); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}

func (ctrl *SyncCtrl) AdminBatchSyncPointCard(ctx *gin.Context) {
	if err := ctrl.pointCardSrv.BatchSyncPointCard(ctx, nil); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}

func (ctrl *SyncCtrl) AdminBatchSyncSemesterSettleRecord(ctx *gin.Context) {
	req := dto.SyncSemesterSettleRecordIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boSyncSemesterSettleRecordCond := &bo.SyncSemesterSettleRecordCond{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	if err := ctrl.semesterSettleRecordSrv.BatchSyncSemesterSettleRecord(ctx, boSyncSemesterSettleRecordCond); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}
