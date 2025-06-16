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

func ProvideReduceRecordController(recordSrv interfaces.IReduceRecordSrv, logger logger.ILogger, reqParse util.IRequestParse) *ReduceRecordCtrl {
	return &ReduceRecordCtrl{
		recordSrv: recordSrv,
		logger:    logger,
		reqParse:  reqParse,
	}
}

type ReduceRecordCtrl struct {
	recordSrv interfaces.IReduceRecordSrv
	logger    logger.ILogger
	reqParse  util.IRequestParse
}

func (ctrl *ReduceRecordCtrl) GetReduceRecords(ctx *gin.Context) {
	req := dto.ReduceRecordGetIO{}
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

	boReduceRecordCond := &bo.ReduceRecordCond{
		Pager: po.Pager{
			Index: req.Index,
			Size:  req.Size,
			Order: "class_time desc",
		},
	}
	if req.ClassTimeStart != nil {
		boReduceRecordCond.ClassTimeStart = *req.ClassTimeStart
	}
	if req.ClassTimeEnd != nil {
		boReduceRecordCond.ClassTimeEnd = *req.ClassTimeEnd
	}

	boReduceRecords, pagerResult, err := ctrl.recordSrv.GetReduceRecords(ctx, boReduceRecordCond, boStudentCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	reduceRecordList := make([]dto.ReduceRecordGetVO, 0, len(boReduceRecords))
	for _, v := range boReduceRecords {
		voRecord := dto.ReduceRecordGetVO{
			RecordId:     strconv.FormatInt(v.RecordId, 10),
			StudentName:  strUtil.GetStudentNameByStudentName(v.StudentName),
			ReducePoints: v.ReducePoints,
			RecordType:   v.RecordType,
		}
		if v.RecordType == "reduce_record" {
			voRecord.ClassLevel = v.ClassLevel.ToValue()
			voRecord.ClassType = v.ClassType.ToValue()
			voRecord.ClassTime = v.ClassTime.Format(time.DateTime)
			voRecord.TeacherName = v.TeacherName
			voRecord.IsAttended = v.IsAttended
		} else {
			voRecord.ClassTime = v.ClassTime.Format(time.DateTime)
		}

		reduceRecordList = append(reduceRecordList, voRecord)
	}

	listVO := dto.ListVO{
		List: reduceRecordList,
		Pager: dto.PagerVO{
			Index: pagerResult.Index,
			Size:  pagerResult.Size,
			Pages: pagerResult.Pages,
			Total: pagerResult.Total,
		},
	}

	SetStandardResponse(ctx, http.StatusOK, listVO)
}

func (ctrl *ReduceRecordCtrl) AdminGetReduceRecords(ctx *gin.Context) {
	req := dto.AdminGetReduceRecordsIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boReduceRecordCond := &bo.ReduceRecordCond{
		Pager: po.Pager{
			Index: req.Index,
			Size:  req.Size,
			Order: "class_time desc, record_id desc",
		},
	}
	if req.ClassTimeStart != nil {
		boReduceRecordCond.ClassTimeStart = *req.ClassTimeStart
	}
	if req.ClassTimeEnd != nil {
		boReduceRecordCond.ClassTimeEnd = *req.ClassTimeEnd
	}
	boReduceRecordCond.IsDeleted = req.IsDeleted

	studentCond := &bo.StudentCond{}
	if req.StudentName != nil {
		studentCond.StudentName = *req.StudentName
	}
	if req.ParentPhone != nil {
		studentCond.ParentPhone = *req.ParentPhone
	}

	boReduceRecords, pagerResult, err := ctrl.recordSrv.GetReduceRecords(ctx, boReduceRecordCond, studentCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	reduceRecordList := make([]dto.AdminGetReduceRecordsVO, 0, len(boReduceRecords))
	for _, v := range boReduceRecords {
		adminGetReduceRecordsVO := dto.AdminGetReduceRecordsVO{}
		adminGetReduceRecordsVO.RecordId = strconv.FormatInt(v.RecordId, 10)
		adminGetReduceRecordsVO.StudentName = v.StudentName
		adminGetReduceRecordsVO.ParentPhone = v.ParentPhone
		adminGetReduceRecordsVO.ReducePoints = v.ReducePoints
		adminGetReduceRecordsVO.RecordType = v.RecordType
		if v.RecordType == "reduce_record" {
			adminGetReduceRecordsVO.ClassLevel = v.ClassLevel.ToValue()
			adminGetReduceRecordsVO.ClassType = v.ClassType.ToValue()
			adminGetReduceRecordsVO.ClassTime = v.ClassTime.Format(time.DateTime)
			adminGetReduceRecordsVO.TeacherName = v.TeacherName
			adminGetReduceRecordsVO.IsAttended = v.IsAttended
		} else {
			adminGetReduceRecordsVO.ClassTime = v.ClassTime.Format(time.DateTime)
		}
		adminGetReduceRecordsVO.IsDeleted = v.IsDeleted
		adminGetReduceRecordsVO.CreatedAt = v.CreatedAt.Format(time.RFC3339)
		adminGetReduceRecordsVO.UpdatedAt = v.UpdatedAt.Format(time.RFC3339)
		if v.DeletedAt != nil {
			adminGetReduceRecordsVO.DeletedAt = v.DeletedAt.Format(time.RFC3339)
		}

		reduceRecordList = append(reduceRecordList, adminGetReduceRecordsVO)
	}

	listVO := dto.ListVO{
		List: reduceRecordList,
		Pager: dto.PagerVO{
			Index: pagerResult.Index,
			Size:  pagerResult.Size,
			Pages: pagerResult.Pages,
			Total: pagerResult.Total,
		},
	}

	SetStandardResponse(ctx, http.StatusOK, listVO)
}

func (ctrl *ReduceRecordCtrl) KintoneReduceRecordWebhook(ctx *gin.Context) {
	req := dto.KintoneWebhookReduceRecordIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		ctrl.logger.Error(ctx, "reduceRecordCtrl KintoneReduceRecordWebhook Bind", err)
		return
	}

	switch req.Type {
	case kintone.AddRecordType:
		ctrl.addReduceRecord(ctx, req)
	case kintone.UpdateRecordType, kintone.UpdateStatusType:
		ctrl.updateReduceRecord(ctx, req)
	case kintone.DeleteRecordType:
		ctrl.deleteReduceRecord(ctx, req)
	}
}

func (ctrl *ReduceRecordCtrl) addReduceRecord(ctx *gin.Context, req dto.KintoneWebhookReduceRecordIO) {
	boKintoneReduceRecord, err := ctrl.checkBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "reduceRecordCtrl addReduceRecord checkBasicRequestData", err, zap.String("record_id", req.Record.Id.Value))
		return
	}

	boStudentCond := &bo.StudentCond{
		StudentName: boKintoneReduceRecord.StudentName,
		ParentPhone: boKintoneReduceRecord.ParentPhone,
	}

	if err = ctrl.recordSrv.AddReduceRecord(ctx, boKintoneReduceRecord, boStudentCond); err != nil {
		ctrl.logger.Error(ctx, "reduceRecordCtrl addReduceRecord AddReduceRecord", err, zap.Int("record_ref_id", boKintoneReduceRecord.Id))
	}
}

func (ctrl *ReduceRecordCtrl) updateReduceRecord(ctx *gin.Context, req dto.KintoneWebhookReduceRecordIO) {
	boKintoneReduceRecord, err := ctrl.checkBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "reduceRecordCtrl updateReduceRecord checkBasicRequestData", err, zap.String("record_id", req.Record.Id.Value))
		return
	}

	boStudentCond := &bo.StudentCond{
		StudentName: boKintoneReduceRecord.StudentName,
		ParentPhone: boKintoneReduceRecord.ParentPhone,
	}

	boUpdateReduceRecordData := &bo.UpdateReduceRecordData{
		ClassLevel:   boKintoneReduceRecord.ClassLevel,
		ClassType:    boKintoneReduceRecord.ClassType,
		ClassTime:    boKintoneReduceRecord.ClassTime,
		TeacherName:  boKintoneReduceRecord.TeacherName,
		ReducePoints: &boKintoneReduceRecord.ReducePoints,
		IsAttended:   &boKintoneReduceRecord.AttendStatus,
	}
	if err = ctrl.recordSrv.UpdateReduceRecord(ctx, &bo.ReduceRecordCond{RecordRefId: boKintoneReduceRecord.Id}, boStudentCond, boUpdateReduceRecordData); err != nil {
		ctrl.logger.Error(ctx, "reduceRecordCtrl updateReduceRecord UpdateReduceRecord", err, zap.Int("record_ref_id", boKintoneReduceRecord.Id))
	}
}

func (ctrl *ReduceRecordCtrl) deleteReduceRecord(ctx *gin.Context, req dto.KintoneWebhookReduceRecordIO) {
	recordId, err := strconv.Atoi(req.RecordId)
	if err != nil {
		ctrl.logger.Error(ctx, "reduceRecordCtrl deleteReduceRecord Atoi", err, zap.String("record_id", req.RecordId))
		return
	}

	if err = ctrl.recordSrv.DeleteReduceRecord(ctx, &bo.ReduceRecordCond{RecordRefId: recordId}); err != nil {
		ctrl.logger.Error(ctx, "reduceRecordCtrl deleteReduceRecord DeleteReduceRecord", err, zap.Int("record_ref_id", recordId))
	}
}

func (ctrl *ReduceRecordCtrl) checkBasicRequestData(req dto.KintoneWebhookReduceRecordIO) (*bo.KintoneReduceRecord, error) {
	if !strUtil.IsValidStudentName(req.Record.StudentName.ToString()) {
		return nil, xerrors.Errorf("reduceRecordCtrl checkBasicRequestData failed: %w", errs.StudentErr.StudentNameInvalidErr)
	}

	if req.Record.Id.Value == "" {
		return nil, xerrors.Errorf("reduceRecordCtrl checkBasicRequestData failed: %w", errs.StudentErr.StudentIdInvalidErr)
	}

	boKintoneReduceRecord, err := req.Record.ToKintoneReduceRecord()
	if err != nil {
		return nil, xerrors.Errorf("reduceRecordCtrl checkBasicRequestData req.Record.ToKintoneReduceRecord: %w", err)
	}

	return boKintoneReduceRecord, nil
}
