package web

import (
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/controller/web/util"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/ctxUtil"
	"jaystar/internal/utils/errs"
	"jaystar/internal/utils/strUtil"
	"net/http"
	"strconv"
	"time"
)

func ProvideDepositRecordController(depositRecordSrv interfaces.IDepositRecordSrv, logger logger.ILogger, reqParse util.IRequestParse) *DepositRecordCtrl {
	return &DepositRecordCtrl{
		recordSrv: depositRecordSrv,
		logger:    logger,
		reqParse:  reqParse,
	}
}

type DepositRecordCtrl struct {
	recordSrv interfaces.IDepositRecordSrv
	logger    logger.ILogger
	reqParse  util.IRequestParse
}

func (ctrl *DepositRecordCtrl) GetDepositRecords(ctx *gin.Context) {
	req := dto.DepositRecordGetIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	userStore := ctxUtil.GetUserSessionFromCtx(ctx)

	boStudentCond := &bo.StudentCond{}
	boStudentCond.UserId = userStore.UserId
	if req.StudentName != nil {
		boStudentCond.StudentName = *req.StudentName
	}
	if req.ParentPhone != nil {
		boStudentCond.ParentPhone = *req.ParentPhone
	}

	boDepositRecordCond := &bo.DepositRecordCond{
		Pager: po.Pager{
			Index: req.Index,
			Size:  req.Size,
			Order: "charging_date desc",
		},
	}
	if req.DepositedDateStart != nil {
		boDepositRecordCond.ChargingDateStart = *req.DepositedDateStart
	}
	if req.DepositedDateEnd != nil {
		boDepositRecordCond.ChargingDateEnd = *req.DepositedDateEnd
	}

	boDepositRecords, pagerResult, err := ctrl.recordSrv.GetDepositRecords(ctx, boDepositRecordCond, boStudentCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	depositRecordList := make([]dto.DepositRecordGetVO, 0, len(boDepositRecords))
	for _, v := range boDepositRecords {
		depositRecordList = append(depositRecordList, dto.DepositRecordGetVO{
			RecordId:             strconv.FormatInt(v.RecordId, 10),
			StudentName:          strUtil.GetStudentNameByStudentName(v.StudentName),
			ChargingDate:         v.ChargingDate.Format(time.DateOnly),
			TaxId:                v.TaxId,
			AccountLastFiveYards: v.AccountLastFiveYards,
			ChargingAmount:       v.ChargingAmount,
			TeacherName:          v.TeacherName,
			DepositedPoints:      v.DepositedPoints,
			ChargingMethod:       kintone.ChargingMethodToValue(v.ChargingMethod),
			ChargingStatus:       v.ChargingStatus.ToName(),
			ActualChargingAmount: v.ActualChargingAmount,
		})
	}

	listVO := dto.ListVO{
		List: depositRecordList,
		Pager: dto.PagerVO{
			Index: pagerResult.Index,
			Size:  pagerResult.Size,
			Pages: pagerResult.Pages,
			Total: pagerResult.Total,
		},
	}

	SetStandardResponse(ctx, http.StatusOK, listVO)
}

func (ctrl *DepositRecordCtrl) AdminGetDepositRecords(ctx *gin.Context) {
	req := dto.AdminGetDepositRecordsIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boDepositRecordCond := &bo.DepositRecordCond{
		Pager: po.Pager{
			Index: req.Index,
			Size:  req.Size,
			Order: "charging_date desc, record_id desc",
		},
	}
	if req.DepositedDateStart != nil {
		boDepositRecordCond.ChargingDateStart = *req.DepositedDateStart
	}
	if req.DepositedDateEnd != nil {
		boDepositRecordCond.ChargingDateEnd = *req.DepositedDateEnd
	}
	boDepositRecordCond.IsDeleted = req.IsDeleted

	studentCond := &bo.StudentCond{}
	if req.StudentName != nil {
		studentCond.StudentName = *req.StudentName
	}
	if req.ParentPhone != nil {
		studentCond.ParentPhone = *req.ParentPhone
	}

	boDepositRecords, pagerResult, err := ctrl.recordSrv.GetDepositRecords(ctx, boDepositRecordCond, studentCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	depositRecordList := make([]dto.AdminDepositRecordGetVO, 0, len(boDepositRecords))
	for _, v := range boDepositRecords {
		adminDepositRecordGetVO := dto.AdminDepositRecordGetVO{}
		adminDepositRecordGetVO.RecordId = strconv.FormatInt(v.RecordId, 10)
		adminDepositRecordGetVO.StudentName = v.StudentName
		adminDepositRecordGetVO.ParentPhone = v.ParentPhone
		adminDepositRecordGetVO.ChargingDate = v.ChargingDate.Format(time.DateOnly)
		adminDepositRecordGetVO.TaxId = v.TaxId
		adminDepositRecordGetVO.AccountLastFiveYards = v.AccountLastFiveYards
		adminDepositRecordGetVO.ChargingAmount = v.ChargingAmount
		adminDepositRecordGetVO.TeacherName = v.TeacherName
		adminDepositRecordGetVO.DepositedPoints = v.DepositedPoints
		adminDepositRecordGetVO.ChargingMethod = kintone.ChargingMethodToValue(v.ChargingMethod)
		adminDepositRecordGetVO.ChargingStatus = v.ChargingStatus.ToName()
		adminDepositRecordGetVO.ActualChargingAmount = v.ActualChargingAmount
		adminDepositRecordGetVO.IsDeleted = v.IsDeleted
		adminDepositRecordGetVO.CreatedAt = v.CreatedAt.Format(time.RFC3339)
		adminDepositRecordGetVO.UpdatedAt = v.UpdatedAt.Format(time.RFC3339)
		if v.DeletedAt != nil {
			adminDepositRecordGetVO.DeletedAt = v.DeletedAt.Format(time.RFC3339)
		}

		depositRecordList = append(depositRecordList, adminDepositRecordGetVO)
	}

	listVO := dto.ListVO{
		List: depositRecordList,
		Pager: dto.PagerVO{
			Index: pagerResult.Index,
			Size:  pagerResult.Size,
			Pages: pagerResult.Pages,
			Total: pagerResult.Total,
		},
	}

	SetStandardResponse(ctx, http.StatusOK, listVO)
}

func (ctrl *DepositRecordCtrl) KintoneDepositRecordWebhook(ctx *gin.Context) {
	req := dto.KintoneWebhookDepositRecordIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		ctrl.logger.Error(ctx, "depositRecordCtrl KintoneDepositRecordWebhook Bind", err)
		return
	}

	switch req.Type {
	case kintone.AddRecordType:
		ctrl.addDepositRecord(ctx, req)
	case kintone.UpdateRecordType, kintone.UpdateStatusType:
		ctrl.updateDepositRecord(ctx, req)
	case kintone.DeleteRecordType:
		ctrl.deleteDepositRecord(ctx, req)
	}
}

func (ctrl *DepositRecordCtrl) addDepositRecord(ctx *gin.Context, req dto.KintoneWebhookDepositRecordIO) {
	boKintoneDepositRecord, err := ctrl.checkBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "depositRecordCtrl addDepositRecord checkBasicRequestData", err, zap.String("record_id", req.Record.Id.Value))
		return
	}

	boStudentCond := &bo.StudentCond{
		StudentName: boKintoneDepositRecord.StudentName,
		ParentPhone: boKintoneDepositRecord.ParentPhone,
	}

	if err := ctrl.recordSrv.AddDepositRecord(ctx, boKintoneDepositRecord, boStudentCond); err != nil {
		ctrl.logger.Error(ctx, "depositRecordCtrl addDepositRecord AddDepositRecord", err, zap.Int("record_ref_id", boKintoneDepositRecord.Id))
	}
}

func (ctrl *DepositRecordCtrl) updateDepositRecord(ctx *gin.Context, req dto.KintoneWebhookDepositRecordIO) {
	boKintoneDepositRecord, err := ctrl.checkBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "depositRecordCtrl updateDepositRecord checkBasicRequestData", err, zap.String("record_id", req.Record.Id.Value))
		return
	}

	boStudentCond := &bo.StudentCond{
		StudentName: boKintoneDepositRecord.StudentName,
		ParentPhone: boKintoneDepositRecord.ParentPhone,
	}

	boUpdateDepositRecordData := &bo.UpdateDepositRecordData{
		ChargingDate:         boKintoneDepositRecord.ChargingDate,
		TaxId:                boKintoneDepositRecord.TaxId,
		AccountLastFiveYards: boKintoneDepositRecord.AccountLastFiveYards,
		ChargingAmount:       &boKintoneDepositRecord.ChargingAmount,
		TeacherName:          boKintoneDepositRecord.TeacherName,
		DepositedPoints:      &boKintoneDepositRecord.DepositedPoints,
		ChargingMethod:       boKintoneDepositRecord.ChargingMethod,
		ChargingStatus:       boKintoneDepositRecord.ChargingStatus,
		ActualChargingAmount: &boKintoneDepositRecord.ActualChargingAmount,
	}
	if err = ctrl.recordSrv.UpdateDepositRecord(ctx, &bo.DepositRecordCond{RecordRefId: boKintoneDepositRecord.Id}, boStudentCond, boUpdateDepositRecordData); err != nil {
		ctrl.logger.Error(ctx, "depositRecordCtrl updateDepositRecord UpdateDepositRecord", err, zap.Int("record_ref_id", boKintoneDepositRecord.Id))
	}
}

func (ctrl *DepositRecordCtrl) deleteDepositRecord(ctx *gin.Context, req dto.KintoneWebhookDepositRecordIO) {
	recordId, err := strconv.Atoi(req.RecordId)
	if err != nil {
		ctrl.logger.Error(ctx, "depositRecordCtrl deleteDepositRecord Atoi", err, zap.String("record_id", req.RecordId))
		return
	}

	if err := ctrl.recordSrv.DeleteDepositRecord(ctx, &bo.DepositRecordCond{RecordRefId: recordId}); err != nil {
		ctrl.logger.Error(ctx, "depositRecordCtrl deleteDepositRecord DeleteDepositRecord", err, zap.Int("record_ref_id", recordId))
	}
}

func (ctrl *DepositRecordCtrl) checkBasicRequestData(req dto.KintoneWebhookDepositRecordIO) (*bo.KintoneDepositRecord, error) {
	if !strUtil.IsValidStudentName(req.Record.StudentName.ToString()) {
		return nil, xerrors.Errorf("depositRecordCtrl checkBasicRequestData failed: %w", errs.StudentErr.StudentNameInvalidErr)
	}

	if req.Record.Id.Value == "" {
		return nil, xerrors.Errorf("depositRecordCtrl checkBasicRequestData failed: %w", errs.StudentErr.StudentIdInvalidErr)
	}

	boKintoneDepositRecord, err := req.Record.ToKintoneDepositRecordBo()
	if err != nil {
		return nil, xerrors.Errorf("depositRecordCtrl checkBasicRequestData req.Record.ToKintoneDepositRecordBo: %w", err)
	}

	return boKintoneDepositRecord, nil
}
