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
	"jaystar/internal/utils"
	"jaystar/internal/utils/errs"
	"jaystar/internal/utils/strUtil"
	"net/http"
	"strconv"
	"time"
)

func ProvideSemesterSettleRecordController(recordSrv interfaces.ISemesterSettleRecordSrv, logger logger.ILogger, reqParse util.IRequestParse) *SemesterSettleRecordCtrl {
	return &SemesterSettleRecordCtrl{
		recordSrv: recordSrv,
		logger:    logger,
		reqParse:  reqParse,
	}
}

type SemesterSettleRecordCtrl struct {
	recordSrv interfaces.ISemesterSettleRecordSrv
	logger    logger.ILogger
	reqParse  util.IRequestParse
}

func (ctrl *SemesterSettleRecordCtrl) AdminGetSemesterSettleRecords(ctx *gin.Context) {
	req := dto.AdminGetSemesterSettleRecordsIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boSemesterSettleRecordCond := &bo.SemesterSettleRecordCond{
		Pager: po.Pager{
			Index: req.Index,
			Size:  req.Size,
			Order: "end_time desc, record_id desc",
		},
	}
	if req.StartTime != nil {
		boSemesterSettleRecordCond.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		boSemesterSettleRecordCond.EndTime = *req.EndTime
	}
	boSemesterSettleRecordCond.IsDeleted = req.IsDeleted

	boStudentCond := &bo.StudentCond{}
	if req.StudentName != nil {
		boStudentCond.StudentName = *req.StudentName
	}
	if req.ParentPhone != nil {
		boStudentCond.ParentPhone = *req.ParentPhone
	}

	boSemesterSettleRecords, pagerResult, err := ctrl.recordSrv.GetSemesterSettleRecords(ctx, boSemesterSettleRecordCond, boStudentCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	semesterSettleRecordList := make([]dto.AdminGetSemesterSettleRecordsVO, 0, len(boSemesterSettleRecords))
	for _, v := range boSemesterSettleRecords {
		adminGetSemesterSettleRecordsVO := dto.AdminGetSemesterSettleRecordsVO{}
		adminGetSemesterSettleRecordsVO.RecordId = strconv.FormatInt(v.RecordId, 10)
		adminGetSemesterSettleRecordsVO.StudentName = v.StudentName
		adminGetSemesterSettleRecordsVO.ParentPhone = v.ParentPhone
		adminGetSemesterSettleRecordsVO.StartTime = v.StartTime.Format(time.DateTime)
		adminGetSemesterSettleRecordsVO.EndTime = v.EndTime.Format(time.DateTime)
		adminGetSemesterSettleRecordsVO.ClearPoints = v.ClearPoints
		adminGetSemesterSettleRecordsVO.IsDeleted = v.IsDeleted
		adminGetSemesterSettleRecordsVO.CreatedAt = v.CreatedAt.Format(time.RFC3339)
		adminGetSemesterSettleRecordsVO.UpdatedAt = v.UpdatedAt.Format(time.RFC3339)
		if v.DeletedAt != nil {
			adminGetSemesterSettleRecordsVO.DeletedAt = v.DeletedAt.Format(time.RFC3339)
		}

		semesterSettleRecordList = append(semesterSettleRecordList, adminGetSemesterSettleRecordsVO)
	}

	listVO := dto.ListVO{
		List: semesterSettleRecordList,
		Pager: dto.PagerVO{
			Index: pagerResult.Index,
			Size:  pagerResult.Size,
			Pages: pagerResult.Pages,
			Total: pagerResult.Total,
		},
	}

	SetStandardResponse(ctx, http.StatusOK, listVO)
}

func (ctrl *SemesterSettleRecordCtrl) KintoneSemesterSettleRecordWebhook(ctx *gin.Context) {
	req := dto.KintoneWebhookSemesterSettleRecordIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		ctrl.logger.Error(ctx, "SemesterSettleRecordCtrl KintoneSemesterSettleRecordWebhook Bind", err)
		return
	}

	switch req.Type {
	case kintone.AddRecordType:
		ctrl.addSemesterSettleRecord(ctx, req)
	case kintone.UpdateRecordType, kintone.UpdateStatusType:
		ctrl.updateSemesterSettleRecord(ctx, req)
	case kintone.DeleteRecordType:
		ctrl.deleteSemesterSettleRecord(ctx, req)
	}
}

func (ctrl *SemesterSettleRecordCtrl) AdminSemesterSettlePoints(ctx *gin.Context) {
	req := dto.SettleSemesterPointsIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		ctrl.logger.Error(ctx, "SemesterSettleRecordCtrl SettleSemesterPointsIO Bind", err)
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	date, err := time.ParseInLocation(time.DateOnly, req.Date, utils.GetLocation())
	if err != nil {
		ctrl.logger.Error(ctx, "SemesterSettleRecordCtrl SettleSemesterPointsIO Parse", err)
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	err = ctrl.recordSrv.SettleSemesterPoints(ctx, &bo.SettleSemesterPointsCond{Date: date})
	if err != nil {
		ctrl.logger.Error(ctx, "SemesterSettleRecordCtrl SettleSemesterPoints", err)
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}

func (ctrl *SemesterSettleRecordCtrl) addSemesterSettleRecord(ctx *gin.Context, req dto.KintoneWebhookSemesterSettleRecordIO) {
	boKintoneSemesterSettleRecord, err := ctrl.checkBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "SemesterSettleRecordCtrl addSemesterSettleRecord checkBasicRequestData", err, zap.String("record_id", req.Record.Id.Value))
		return
	}

	boStudentCond := &bo.StudentCond{
		StudentName: boKintoneSemesterSettleRecord.StudentName,
		ParentPhone: boKintoneSemesterSettleRecord.ParentPhone,
	}

	if err = ctrl.recordSrv.AddSemesterSettleRecord(ctx, boKintoneSemesterSettleRecord, boStudentCond); err != nil {
		ctrl.logger.Error(ctx, "SemesterSettleRecordCtrl addSemesterSettleRecord AddSemesterSettleRecord", err, zap.Int("record_ref_id", boKintoneSemesterSettleRecord.RecordRefId))
	}
}

func (ctrl *SemesterSettleRecordCtrl) updateSemesterSettleRecord(ctx *gin.Context, req dto.KintoneWebhookSemesterSettleRecordIO) {
	boKintoneSemesterSettleRecord, err := ctrl.checkBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "SemesterSettleRecordCtrl updateSemesterSettleRecord checkBasicRequestData", err, zap.String("record_id", req.Record.Id.Value))
		return
	}

	boStudentCond := &bo.StudentCond{
		StudentName: boKintoneSemesterSettleRecord.StudentName,
		ParentPhone: boKintoneSemesterSettleRecord.ParentPhone,
	}

	boUpdateSemesterSettleRecordData := &bo.UpdateSemesterSettleRecordData{
		StartTime:   boKintoneSemesterSettleRecord.StartTime,
		EndTime:     boKintoneSemesterSettleRecord.EndTime,
		ClearPoints: &boKintoneSemesterSettleRecord.ClearPoints,
	}
	if err = ctrl.recordSrv.UpdateSemesterSettleRecord(ctx, &bo.UpdateSemesterSettleRecordCond{RecordRefId: boKintoneSemesterSettleRecord.RecordRefId}, boStudentCond, boUpdateSemesterSettleRecordData); err != nil {
		ctrl.logger.Error(ctx, "SemesterSettleRecordCtrl updateSemesterSettleRecord UpdateSemesterSettleRecord", err, zap.Int("record_ref_id", boKintoneSemesterSettleRecord.RecordRefId))
	}
}

func (ctrl *SemesterSettleRecordCtrl) deleteSemesterSettleRecord(ctx *gin.Context, req dto.KintoneWebhookSemesterSettleRecordIO) {
	recordId, err := strconv.Atoi(req.RecordId)
	if err != nil {
		ctrl.logger.Error(ctx, "SemesterSettleRecordCtrl deleteSemesterSettleRecord Atoi", err, zap.String("record_id", req.RecordId))
		return
	}

	if err = ctrl.recordSrv.DeleteSemesterSettleRecord(ctx, &bo.UpdateSemesterSettleRecordCond{RecordRefId: recordId}); err != nil {
		ctrl.logger.Error(ctx, "SemesterSettleRecordCtrl deleteSemesterSettleRecord DeleteSemesterSettleRecord", err, zap.Int("record_ref_id", recordId))
	}
}

func (ctrl *SemesterSettleRecordCtrl) checkBasicRequestData(req dto.KintoneWebhookSemesterSettleRecordIO) (*bo.SemesterSettleRecord, error) {
	if !strUtil.IsValidStudentName(req.Record.StudentName.ToString()) {
		return nil, xerrors.Errorf("SemesterSettleRecordCtrl) checkBasicRequestData failed: %w", errs.StudentErr.StudentNameInvalidErr)
	}

	if req.Record.Id.Value == "" {
		return nil, xerrors.Errorf("SemesterSettleRecordCtrl) checkBasicRequestData failed: %w", errs.StudentErr.StudentIdInvalidErr)
	}

	boKintoneSemesterSettleRecord, err := req.Record.ToSemesterSettleRecord()
	if err != nil {
		return nil, xerrors.Errorf("SemesterSettleRecordCtrl) checkBasicRequestData req.Record.ToSemesterSettleRecord: %w", err)
	}

	return boKintoneSemesterSettleRecord, nil
}
